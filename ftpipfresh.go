package main

import (
	"log"
	"time"
)

// ─── IP SELF-HEALING VIA SSDP ─────────────────────────────────────────────────
//
// Printers wander to new IPs over DHCP. MQTT/camera find them again by serial
// number (the discovery listener), but a stored IP that has gone stale makes FTP
// (File Sync, media, the FTP check) run into a dead address and time out on port
// 990 — even though the printer is "Online".
//
// Two distinct failure modes are handled here:
//
//  1. A printer truly moved: the address we stored is no longer announced at all.
//     refreshStoredIPs switches it to a currently-announced address.
//
//  2. A dual-NIC printer (e.g. the X1E with a wired *and* a wifi interface) has
//     two IPs under one serial. SSDP may announce one while the FTP server only
//     listens on the other, so port 990 times out on the stored (announced)
//     address even though it is current. switchToWorkingFTPIP probes the sibling
//     addresses on 990 and switches to whichever actually answers.
//
// No DHCP reservation needed.

// heardIPsBySerial returns every distinct IP the SSDP listener currently
// associates with each serial (passive listener, backed by a short active
// search). A dual-NIC printer shows up under one serial with several IPs.
func heardIPsBySerial() map[string][]string {
	m := map[string][]string{}
	add := func(list []DiscoveredPrinter) {
		for _, d := range list {
			if d.Serial == "" || d.IP == "" {
				continue
			}
			dup := false
			for _, ip := range m[d.Serial] {
				if ip == d.IP {
					dup = true
					break
				}
			}
			if !dup {
				m[d.Serial] = append(m[d.Serial], d.IP)
			}
		}
	}
	add(heardPrinters(5 * time.Minute))
	// A short active search fills gaps for printers that have not announced
	// themselves in the last few minutes. Fails silently where the OS forbids
	// binding the SSDP multicast port; passive data still carries the result.
	if found, err := discoverPrinters(searchTargets(), 2, 1500*time.Millisecond); err == nil {
		add(found)
	}
	return m
}

// applyIPSwitch persists a new IP for the printer with the given serial:
// updates state under the lock, saves, reconnects MQTT and refreshes the
// camera streams. Returns the previous stored IP.
func applyIPSwitch(serial, oldIP, newIP, name string) {
	mu.Lock()
	for i := range state.Printers {
		if state.Printers[i].Serial == serial {
			state.Printers[i].IP = newIP
		}
	}
	mu.Unlock()
	saveState()
	log.Printf("IP aktualisiert: %s %s -> %s", name, oldIP, newIP)
	mqttMgr.Disconnect(oldIP)
	go func() {
		np := Printer{}
		mu.RLock()
		for i := range state.Printers {
			if state.Printers[i].Serial == serial {
				np = state.Printers[i]
				break
			}
		}
		mu.RUnlock()
		if np.IP != "" {
			mqttMgr.Connect(np)
		}
	}()
	restartGo2rtcAsync()
}

// refreshStoredIPs corrects printers that have truly moved: their stored IP is
// no longer announced by SSDP at all. It deliberately does NOT touch a printer
// whose stored IP is still among the announced addresses — for a dual-NIC
// printer that would risk switching away from the FTP-capable interface. Where
// several new addresses are known, one that answers on port 990 is preferred.
// Returns how many printers were switched.
func refreshStoredIPs() int {
	bySerial := heardIPsBySerial()
	if len(bySerial) == 0 {
		return 0
	}

	type move struct{ serial, name, oldIP, newIP string }
	var moves []move

	mu.RLock()
	for _, p := range state.Printers {
		if p.Serial == "" {
			continue
		}
		ips, ok := bySerial[p.Serial]
		if !ok || len(ips) == 0 {
			continue
		}
		// Stored IP still announced → current, leave it alone.
		stillValid := false
		for _, ip := range ips {
			if ip == p.IP {
				stillValid = true
				break
			}
		}
		if stillValid {
			continue
		}
		// Genuinely moved. Prefer a candidate that answers on 990.
		newIP := ips[0]
		for _, ip := range ips {
			if probePort(ip, 990, 2*time.Second).Offen {
				newIP = ip
				break
			}
		}
		moves = append(moves, move{p.Serial, p.Name, p.IP, newIP})
	}
	mu.RUnlock()

	for _, m := range moves {
		applyIPSwitch(m.serial, m.oldIP, m.newIP, m.name)
	}
	return len(moves)
}

// switchToWorkingFTPIP is called when the stored IP does not answer on port 990.
// For a dual-NIC printer the FTP server may live on the *other* address the
// printer announces under the same serial. It probes every sibling IP on 990
// and, if one opens, persists the switch and returns it. altIPs is the list of
// addresses SSDP heard for this printer's serial.
func switchToWorkingFTPIP(p Printer, altIPs []string) (string, bool) {
	for _, ip := range altIPs {
		if ip == p.IP {
			continue
		}
		if probePort(ip, 990, 3*time.Second).Offen {
			applyIPSwitch(p.Serial, p.IP, ip, p.Name)
			return ip, true
		}
	}
	return "", false
}
