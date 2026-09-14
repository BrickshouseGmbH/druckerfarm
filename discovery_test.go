package main

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const realResponse = "HTTP/1.1 200 OK\r\n" +
	"HOST: 239.255.255.250:2021\r\n" +
	"Server: UPnP/1.0\r\n" +
	"Location: 192.168.189.85\r\n" +
	"NT: urn:schemas-upnp-org:device:printer:1\r\n" +
	"USN: 03W00C123456789\r\n" +
	"Cache-Control: max-age=1800\r\n" +
	"DevModel.suffix: X1E\r\n" +
	"DevName.suffix: woobly 12\r\n" +
	"DevSignal.suffix: -50dBm\r\n" +
	"DevConnect.suffix: lan\r\n" +
	"DevBind.suffix: free\r\n" +
	"DevVersion.suffix: 01.08.00.00\r\n\r\n"

// Header names differ slightly between firmware versions, so the parser keys on
// the prefix rather than an exact match.
const plainResponse = "HTTP/1.1 200 OK\r\n" +
	"USN: 01P00A987654321\r\n" +
	"DevModel: P1S\r\n" +
	"DevName: Halle 2 links\r\n" +
	"DevConnect: lan\r\n" +
	"DevBind: occupied\r\n" +
	"DevVersion: 01.07.00.00\r\n\r\n"

// Firmware-Varianten haengen ein Suffix an die Feldnamen — beide Formen muessen
// gelesen werden.
func TestParseResponseWithSuffixedHeaders(t *testing.T) {
	d, ok := parseSSDPResponse([]byte(realResponse), "192.168.189.85")
	if !ok {
		t.Fatal("suffixed headers were not recognised")
	}
	if d.Model != "X1E" || d.Name != "woobly 12" || d.Serial != "03W00C123456789" {
		t.Fatalf("wrong fields: %+v", d)
	}
	if d.Connect != "lan" || d.Bind != "free" || d.Version != "01.08.00.00" {
		t.Fatalf("wrong fields: %+v", d)
	}
}

func TestParseResponse(t *testing.T) {
	d, ok := parseSSDPResponse([]byte(plainResponse), "10.0.0.5")
	if !ok {
		t.Fatal("response was not recognised")
	}
	if d.IP != "10.0.0.5" || d.Model != "P1S" || d.Name != "Halle 2 links" {
		t.Fatalf("wrong fields: %+v", d)
	}
	if d.Serial != "01P00A987654321" || d.Connect != "lan" || d.Bind != "occupied" || d.Version != "01.07.00.00" {
		t.Fatalf("wrong fields: %+v", d)
	}
}

// Other devices on the network answer ssdp:all too — they must be filtered out.
func TestForeignDevicesAreIgnored(t *testing.T) {
	router := "HTTP/1.1 200 OK\r\nUSN: uuid:abcd-1234\r\nST: upnp:rootdevice\r\nSERVER: Router/1.0\r\n\r\n"
	if _, ok := parseSSDPResponse([]byte(router), "10.0.0.1"); ok {
		t.Fatal("a router without printer headers must not be reported")
	}
	if _, ok := parseSSDPResponse([]byte("garbage"), "10.0.0.1"); ok {
		t.Fatal("garbage must not be reported")
	}
	// missing serial → unusable
	noSerial := "HTTP/1.1 200 OK\r\nDevModel: X1C\r\nDevName: Test\r\n\r\n"
	if _, ok := parseSSDPResponse([]byte(noSerial), "10.0.0.2"); ok {
		t.Fatal("entry without serial must not be reported")
	}
}

func TestNotifyBroadcastIsAccepted(t *testing.T) {
	notify := "NOTIFY * HTTP/1.1\r\nNTS: ssdp:alive\r\nUSN: 03W00C111\r\nDevModel: X1E\r\nDevName: Woobly 1\r\n\r\n"
	d, ok := parseSSDPResponse([]byte(notify), "10.0.0.9")
	if !ok || d.Serial != "03W00C111" {
		t.Fatalf("periodic NOTIFY should be usable: %+v ok=%v", d, ok)
	}
}

func TestMSearchLooksRight(t *testing.T) {
	m := string(buildMSearch("239.255.255.250:2021"))
	for _, want := range []string{"M-SEARCH * HTTP/1.1", "HOST: 239.255.255.250:2021", `MAN: "ssdp:discover"`, "ST: ssdp:all"} {
		if !strings.Contains(m, want) {
			t.Fatalf("%q missing from:\n%s", want, m)
		}
	}
	if !strings.HasSuffix(m, "\r\n\r\n") {
		t.Fatal("request must end with a blank line")
	}
}

func TestBroadcastCalculation(t *testing.T) {
	_, n, _ := net.ParseCIDR("192.168.189.85/24")
	if bc := broadcastAddr(n); bc.String() != "192.168.189.255" {
		t.Fatalf("got %v", bc)
	}
	_, n2, _ := net.ParseCIDR("10.1.2.3/16")
	if bc := broadcastAddr(n2); bc.String() != "10.1.255.255" {
		t.Fatalf("got %v", bc)
	}
}

// End-to-end over real UDP against a stand-in printer, including the repeated
// sends: a printer that drops the first two packets must still be found.
func TestDiscoveryOverUDPSurvivesPacketLoss(t *testing.T) {
	pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	var received int64
	go func() {
		buf := make([]byte, 2048)
		for {
			n, src, err := pc.ReadFromUDP(buf)
			if err != nil {
				return
			}
			if !strings.HasPrefix(string(buf[:n]), "M-SEARCH") {
				continue
			}
			// Erst ab dem dritten Versuch antworten — simuliert Paketverlust
			if atomic.AddInt64(&received, 1) < 3 {
				continue
			}
			pc.WriteToUDP([]byte(plainResponse), src)
		}
	}()

	target := fmt.Sprintf("127.0.0.1:%d", pc.LocalAddr().(*net.UDPAddr).Port)
	list, err := discoverPrinters([]string{target}, discoverSearchRounds, 2500*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 printer, got %d (%d M-SEARCH received)", len(list), atomic.LoadInt64(&received))
	}
	if list[0].Model != "P1S" || list[0].Serial != "01P00A987654321" {
		t.Fatalf("wrong data: %+v", list[0])
	}
	if got := atomic.LoadInt64(&received); got != int64(discoverSearchRounds) {
		t.Fatalf("expected %d sends, printer saw %d", discoverSearchRounds, got)
	}
}

// The same printer answering several times must appear once.
func TestDuplicatesAreCollapsed(t *testing.T) {
	pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			n, src, err := pc.ReadFromUDP(buf)
			if err != nil {
				return
			}
			if strings.HasPrefix(string(buf[:n]), "M-SEARCH") {
				pc.WriteToUDP([]byte(plainResponse), src)
				pc.WriteToUDP([]byte(plainResponse), src)
			}
		}
	}()

	target := fmt.Sprintf("127.0.0.1:%d", pc.LocalAddr().(*net.UDPAddr).Port)
	list, _ := discoverPrinters([]string{target}, 2, 1500*time.Millisecond)
	if len(list) != 1 {
		t.Fatalf("want 1 entry after dedupe, got %d", len(list))
	}
}

func TestSearchTargetsIncludeMulticast(t *testing.T) {
	ts := searchTargets()
	if len(ts) == 0 || !strings.HasPrefix(ts[0], ssdpMulticast+":") {
		t.Fatalf("multicast target missing: %v", ts)
	}
	for _, tgt := range ts {
		if !strings.HasSuffix(tgt, fmt.Sprintf(":%d", ssdpPort)) {
			t.Fatalf("wrong port in %q — must be %d, not 1900", tgt, ssdpPort)
		}
	}
}
