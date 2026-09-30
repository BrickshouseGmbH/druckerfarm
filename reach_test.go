package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// The three cases that must be told apart in practice, each
// against real sockets — not against a mock.
func TestProbePortEchteSockets(t *testing.T) {
	// Open port: a real listener.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	if got := probePort("127.0.0.1", port, 2*time.Second); !got.Offen {
		t.Fatalf("offener Port nicht erkannt: %+v", got)
	}

	// Closed port: the same machine, but nobody is listening.
	ln2, _ := net.Listen("tcp", "127.0.0.1:0")
	zu := ln2.Addr().(*net.TCPAddr).Port
	ln2.Close()
	got := probePort("127.0.0.1", zu, 2*time.Second)
	if got.Offen {
		t.Fatal("geschlossener Port als offen gemeldet")
	}
	if !strings.HasPrefix(got.Grund, "abgelehnt") {
		t.Fatalf("Grund sollte 'abgelehnt' sein, ist %q", got.Grund)
	}

	// No answer: an address that does not route.
	got = probePort("192.0.2.1", 322, 300*time.Millisecond)
	if got.Offen {
		t.Fatal("nicht erreichbare Adresse als offen gemeldet")
	}
	if got.Grund == "" {
		t.Fatal("kein Grund angegeben")
	}
}

// The verdict is the real value: it should tell the user where to
// look. These cases mirror exactly the situation from the go2rtc log.
func TestUrteil(t *testing.T) {
	p := func(port int, offen bool, grund string) portResult {
		return portResult{Port: port, Offen: offen, Grund: grund}
	}
	cases := []struct {
		name  string
		ports []portResult
		will  string
	}{
		{"alles offen",
			[]portResult{p(322, true, ""), p(990, true, ""), p(8883, true, "")},
			"Kamera erreichbar"},
		{"Geraet aus — nichts antwortet",
			[]portResult{p(322, false, "keine Antwort"), p(990, false, "keine Antwort"), p(8883, false, "keine Antwort")},
			"Gerät aus oder nicht im Netz"},
		{"Drucker laeuft, Kamera abgelehnt (wie .51)",
			[]portResult{p(322, false, "abgelehnt — Gerät antwortet, Dienst läuft nicht"), p(990, true, ""), p(8883, true, "")},
			"Kamera ist abgeschaltet"},
		{"Drucker laeuft, Kamera stumm (wie die h-Serie)",
			[]portResult{p(322, false, "keine Antwort"), p(990, true, ""), p(8883, true, "")},
			"Kamera antwortet nicht"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := urteil(c.ports)
			if !strings.Contains(got, c.will) {
				t.Fatalf("Urteil %q enthaelt nicht %q", got, c.will)
			}
		})
	}
}

// Three ports are checked at once — otherwise a run over
// 42 printers with 3 seconds each takes unreasonably long.
func TestCheckPrinterParallel(t *testing.T) {
	start := time.Now()
	r := checkPrinter(Printer{IP: "192.0.2.1", Name: "tot"}, 700*time.Millisecond)
	dauer := time.Since(start)
	if len(r.Ports) != 3 {
		t.Fatalf("erwartet 3 Ports, bekommen %d", len(r.Ports))
	}
	if dauer > 1500*time.Millisecond {
		t.Fatalf("Ports offenbar nacheinander geprueft: %v", dauer)
	}
	for _, p := range r.Ports {
		if p.Was == "" {
			t.Fatal("Portbeschriftung fehlt")
		}
	}
}
