package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// ─── MITHOEREN STATT NUR FRAGEN ───────────────────────────────────────────────
//
// Die bisherige Suche hat M-SEARCH verschickt und auf Antworten an ihrem
// eigenen, zufaellig gewaehlten Port gewartet. Das setzt voraus, dass die
// Geraete auf M-SEARCH ueberhaupt antworten — und genau das tun sie
// offenbar nicht zuverlaessig.
//
// Drucker melden sich von sich aus: in regelmaessigen Abstaenden schicken sie
// eine NOTIFY-Nachricht an 239.255.255.250:2021. Wer die hoeren will, muss auf
// Port 2021 lauschen und der Multicast-Gruppe beitreten. Das passiert hier —
// dauerhaft im Hintergrund, damit auch ein IP-Wechsel von selbst auffaellt.

type gehoert struct {
	Drucker   DiscoveredPrinter
	Zeitpunkt time.Time
}

var (
	lauschMu      sync.Mutex
	lauschFunde   = map[string]gehoert{} // Seriennummer -> zuletzt gehoert
	lauschPakete  int                    // alle UDP-Pakete, auch fremde
	lauschSockets []string               // welche Sockets offen sind
	lauschFehler  []string               // und welche nicht
)

// startDiscoveryListener oeffnet so viele Empfangswege wie moeglich und laesst
// sie offen. Scheitert einer, laufen die anderen weiter — auf einem Rechner mit
// mehreren Netzkarten oder strenger Firewall ist das der Normalfall.
func startDiscoveryListener() {
	// 1. Allgemeiner Empfang: faengt Broadcast und Unicast auf Port 2021.
	if c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: ssdpPort}); err == nil {
		merkeSocket("0.0.0.0:2021 (Broadcast)")
		go lauschAuf(c, "broadcast")
	} else {
		merkeFehler("0.0.0.0:2021: " + err.Error())
	}

	// 2. Multicast je Netzkarte. Ohne Gruppenbeitritt kommen die NOTIFY-Pakete
	//    gar nicht erst bei uns an.
	gruppe := &net.UDPAddr{IP: net.ParseIP(ssdpMulticast), Port: ssdpPort}
	ifaces, err := net.Interfaces()
	if err != nil {
		merkeFehler("Netzkarten nicht lesbar: " + err.Error())
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
			merkeFehler(ifi.Name + ": " + err.Error())
			continue
		}
		c.SetReadBuffer(1 << 20)
		merkeSocket(ifi.Name + " → " + ssdpMulticast + ":2021")
		go lauschAuf(c, ifi.Name)
	}
}

func merkeSocket(s string) {
	lauschMu.Lock()
	lauschSockets = append(lauschSockets, s)
	lauschMu.Unlock()
	log.Printf("Suche: hoere mit auf %s", s)
}

func merkeFehler(s string) {
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

// gehoerteDrucker liefert, was in den letzten Minuten zu hoeren war.
func gehoerteDrucker(maxAlter time.Duration) []DiscoveredPrinter {
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

// lauschBericht sagt, ob ueberhaupt etwas ankommt. Ohne diese Auskunft ist
// "es findet nichts" nicht von "es hoert nichts" zu unterscheiden.
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
