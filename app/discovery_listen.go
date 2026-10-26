package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// ─── MITHOEREN STATT NUR FRAGEN ───────────────────────────────────────────────
//
// The previous search sent M-SEARCH and waited for responses on its
// own randomly chosen port. That assumes the
// devices answer M-SEARCH at all — and that is exactly what they
// apparently do not do reliably.
//
// Printers announce themselves: at regular intervals they send
// a NOTIFY message to 239.255.255.250:2021. To hear it you must
// listen on port 2021 and join the multicast group. That happens here —
// permanently in the background, so an IP change is noticed on its own.

type gehoert struct {
	Drucker   DiscoveredPrinter
	Zeitpunkt time.Time
}

var (
	lauschMu      sync.Mutex
	lauschFunde   = map[string]gehoert{} // Seriennummer -> zuletzt gehoert
	lauschPakete  int                    // all UDP packets, including foreign
	lauschSockets []string               // which sockets are open
	lauschFehler  []string               // and which are not
)

// startDiscoveryListener opens as many receive paths as possible and leaves
// them open. If one fails, the others keep running — on a machine with
// multiple NICs or a strict firewall this is the normal case.
func startDiscoveryListener() {
	// 1. General receive: catches broadcast and unicast on port 2021.
	if c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: ssdpPort}); err == nil {
		noteSocket("0.0.0.0:2021 (Broadcast)")
		go lauschAuf(c, "broadcast")
	} else {
		noteError("0.0.0.0:2021: " + err.Error())
	}

	// 2. Multicast per NIC. Without joining the group the NOTIFY packets
	//    do not even reach us.
	gruppe := &net.UDPAddr{IP: net.ParseIP(ssdpMulticast), Port: ssdpPort}
	ifaces, err := net.Interfaces()
	if err != nil {
		noteError("Netzkarten nicht lesbar: " + err.Error())
		return
	}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagMulticast == 0 {
			continue
		}
		if ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		c, err := net.ListenMulticastUDP("udp4", &ifi, gruppe)
		if err != nil {
			noteError(ifi.Name + ": " + err.Error())
			continue
		}
		c.SetReadBuffer(1 << 20)
		noteSocket(ifi.Name + " → " + ssdpMulticast + ":2021")
		go lauschAuf(c, ifi.Name)
	}
}

func noteSocket(s string) {
	lauschMu.Lock()
	lauschSockets = append(lauschSockets, s)
	lauschMu.Unlock()
	log.Printf("Suche: hoere mit auf %s", s)
}

func noteError(s string) {
	lauschMu.Lock()
	lauschFehler = append(lauschFehler, s)
	lauschMu.Unlock()
	log.Printf("Suche: kein Empfang auf %s", s)
}

func lauschAuf(c *net.UDPConn, wo string) {
	defer c.Close()
	buf := make([]byte, 8192)
	for {
		n, src, err := c.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Suche: Empfang auf %s beendet: %v", wo, err)
			return
		}
		lauschMu.Lock()
		lauschPakete++
		lauschMu.Unlock()

		d, ok := parseSSDPResponse(buf[:n], src.IP.String())
		if !ok {
			continue
		}
		schluessel := d.Serial
		if schluessel == "" {
			schluessel = d.IP
		}
		lauschMu.Lock()
		lauschFunde[schluessel] = gehoert{Drucker: d, Zeitpunkt: time.Now()}
		lauschMu.Unlock()
	}
}

// heardPrinters returns what was heard in the last few minutes.
func heardPrinters(maxAlter time.Duration) []DiscoveredPrinter {
	lauschMu.Lock()
	defer lauschMu.Unlock()
	var out []DiscoveredPrinter
	for _, g := range lauschFunde {
		if time.Since(g.Zeitpunkt) <= maxAlter {
			out = append(out, g.Drucker)
		}
	}
	return out
}

// lauschBericht reports whether anything arrives at all. Without it
// "it finds nothing" cannot be told from "it hears nothing".
func lauschBericht() map[string]any {
	lauschMu.Lock()
	defer lauschMu.Unlock()
	sockets := append([]string(nil), lauschSockets...)
	fehler := append([]string(nil), lauschFehler...)
	return map[string]any{
		"empfangswege":    sockets,
		"nicht_moeglich":  fehler,
		"pakete_gesamt":   lauschPakete,
		"geraete_gehoert": len(lauschFunde),
	}
}
