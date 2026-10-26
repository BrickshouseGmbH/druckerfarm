package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// A printer that announces itself — exactly what the old search could not
// hear, because it only waited for replies on its own port.
const notifyBeispiel = "NOTIFY * HTTP/1.1\r\n" +
	"HOST: 239.255.255.250:2021\r\n" +
	"NT: urn:lan-printer:device:3dprinter:1\r\n" +
	"NTS: ssdp:alive\r\n" +
	"USN: 01P00A123456789\r\n" +
	"DevModel.printer.local: H2D\r\n" +
	"DevName.printer.local: woobly-h11\r\n" +
	"DevConnect.printer.local: lan\r\n" +
	"DevBind.printer.local: free\r\n" +
	"DevVersion.printer.local: 01.08.02.00\r\n\r\n"

// The core: an unsolicited announcement is recognised as a printer.
func TestNotifyWirdErkannt(t *testing.T) {
	d, ok := parseSSDPResponse([]byte(notifyBeispiel), "192.168.189.127")
	if !ok {
		t.Fatal("NOTIFY nicht als Drucker erkannt — dann findet die Suche nie etwas")
	}
	if d.Serial != "01P00A123456789" {
		t.Fatalf("Seriennummer falsch: %q", d.Serial)
	}
	if d.Model != "H2D" || d.Name != "woobly-h11" {
		t.Fatalf("Kopfzeilen falsch gelesen: %+v", d)
	}
	if d.IP != "192.168.189.127" {
		t.Fatalf("IP muss vom Absender kommen: %q", d.IP)
	}
}

// And the receive path itself: a real UDP socket, a real packet.
func TestLauscherHoertEchtesPaket(t *testing.T) {
	// Own socket on a free port — the fixed 2021 is not guaranteed free
	// during the test run.
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	port := c.LocalAddr().(*net.UDPAddr).Port

	lauschMu.Lock()
	lauschFunde = map[string]gehoert{}
	vorher := lauschPakete
	lauschMu.Unlock()

	go lauschAuf(c, "test")

	sender, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()
	if _, err := sender.Write([]byte(notifyBeispiel)); err != nil {
		t.Fatal(err)
	}
	// Plus a foreign packet that must not count as a printer
	sender.Write([]byte("NOTIFY * HTTP/1.1\r\nNT: upnp:rootdevice\r\nSERVER: Fritz!Box\r\n\r\n"))

	frist := time.Now().Add(3 * time.Second)
	for time.Now().Before(frist) {
		if len(heardPrinters(time.Minute)) > 0 {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}

	gefunden := heardPrinters(time.Minute)
	if len(gefunden) != 1 {
		t.Fatalf("erwartet genau 1 Drucker, bekommen %d", len(gefunden))
	}
	if gefunden[0].Serial != "01P00A123456789" {
		t.Fatalf("falscher Drucker: %+v", gefunden[0])
	}

	lauschMu.Lock()
	pakete := lauschPakete - vorher
	lauschMu.Unlock()
	if pakete < 2 {
		t.Fatalf("beide Pakete muessen gezaehlt werden, gezaehlt: %d", pakete)
	}
}

// Too-old announcements drop out — a printer silent for hours
// should not count as present.
func TestAlteMeldungenFallenHeraus(t *testing.T) {
	lauschMu.Lock()
	lauschFunde = map[string]gehoert{
		"neu": {Drucker: DiscoveredPrinter{Serial: "neu"}, Zeitpunkt: time.Now()},
		"alt": {Drucker: DiscoveredPrinter{Serial: "alt"}, Zeitpunkt: time.Now().Add(-30 * time.Minute)},
	}
	lauschMu.Unlock()
	got := heardPrinters(10 * time.Minute)
	if len(got) != 1 || got[0].Serial != "neu" {
		t.Fatalf("Altersgrenze greift nicht: %+v", got)
	}
}

// The report must say something even when nothing arrives — otherwise
// "finds nothing" cannot be told from "hears nothing".
func TestLauschBerichtIstAussagekraeftig(t *testing.T) {
	b := lauschBericht()
	for _, feld := range []string{"empfangswege", "nicht_moeglich", "pakete_gesamt", "geraete_gehoert"} {
		if _, ok := b[feld]; !ok {
			t.Fatalf("Feld %q fehlt im Bericht", feld)
		}
	}
}

// A known device at a new address must count as a move, not as a
// new printer — otherwise duplicates are created and the old entry stays dead.
func TestUmzugWirdErkannt(t *testing.T) {
	mu.Lock()
	alt := state.Printers
	state.Printers = []Printer{{Name: "h11", IP: "192.168.189.127", Serial: "01P00A123456789"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = alt; mu.Unlock() }()

	d, _ := parseSSDPResponse([]byte(notifyBeispiel), "192.168.189.200")
	mu.Lock()
	serieZuIP := map[string]string{}
	for _, p := range state.Printers {
		if p.Serial != "" {
			serieZuIP[p.Serial] = p.IP
		}
	}
	mu.Unlock()

	alteIP, kennenWir := serieZuIP[d.Serial]
	if !kennenWir {
		t.Fatal("Seriennummer nicht wiedererkannt")
	}
	if alteIP == d.IP {
		t.Fatal("das waere kein Umzug")
	}
	if !strings.HasPrefix(alteIP, "192.168.189.127") {
		t.Fatalf("alte Adresse falsch: %s", alteIP)
	}
}
