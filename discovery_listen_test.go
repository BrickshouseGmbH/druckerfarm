package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// Ein Drucker, der sich von selbst meldet — genau das, was die alte Suche nicht
// hoeren konnte, weil sie nur auf Antworten an ihren eigenen Port wartete.
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

// Der Kern: eine unaufgeforderte Meldung wird als Drucker erkannt.
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

// Und der Empfangsweg selbst: ein echter UDP-Socket, ein echtes Paket.
func TestLauscherHoertEchtesPaket(t *testing.T) {
	// Eigener Socket auf einem freien Port — der feste 2021 ist im Testlauf
	// nicht garantiert frei.
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
	// Dazu ein fremdes Paket, das nicht als Drucker zaehlen darf
	sender.Write([]byte("NOTIFY * HTTP/1.1\r\nNT: upnp:rootdevice\r\nSERVER: Fritz!Box\r\n\r\n"))

	frist := time.Now().Add(3 * time.Second)
	for time.Now().Before(frist) {
		if len(gehoerteDrucker(time.Minute)) > 0 {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}

	gefunden := gehoerteDrucker(time.Minute)
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

// Zu alte Meldungen fallen heraus — ein Drucker, der seit Stunden schweigt,
// soll nicht als anwesend gelten.
func TestAlteMeldungenFallenHeraus(t *testing.T) {
	lauschMu.Lock()
	lauschFunde = map[string]gehoert{
		"neu": {Drucker: DiscoveredPrinter{Serial: "neu"}, Zeitpunkt: time.Now()},
		"alt": {Drucker: DiscoveredPrinter{Serial: "alt"}, Zeitpunkt: time.Now().Add(-30 * time.Minute)},
	}
	lauschMu.Unlock()
	got := gehoerteDrucker(10 * time.Minute)
	if len(got) != 1 || got[0].Serial != "neu" {
		t.Fatalf("Altersgrenze greift nicht: %+v", got)
	}
}

// Der Bericht muss auch dann etwas sagen, wenn nichts ankommt — sonst ist
// "findet nichts" nicht von "hoert nichts" zu unterscheiden.
func TestLauschBerichtIstAussagekraeftig(t *testing.T) {
	b := lauschBericht()
	for _, feld := range []string{"empfangswege", "nicht_moeglich", "pakete_gesamt", "geraete_gehoert"} {
		if _, ok := b[feld]; !ok {
			t.Fatalf("Feld %q fehlt im Bericht", feld)
		}
	}
}

// Ein bekanntes Geraet unter neuer Adresse muss als Umzug gelten, nicht als
// neuer Drucker — sonst legt man Dubletten an und der alte Eintrag bleibt tot.
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
