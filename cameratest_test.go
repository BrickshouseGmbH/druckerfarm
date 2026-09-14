package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Die Liste der geprueften Formen muss beide Schemata und beide Pfade abdecken —
// und den Zugangscode enthalten, sonst weist der Drucker ab.
func TestKameraKandidaten(t *testing.T) {
	k := kameraKandidaten(Printer{IP: "10.0.0.9", Code: "geheim"})
	if len(k) != 4 {
		t.Fatalf("erwartet 4 Formen, bekommen %d", len(k))
	}
	schemata := map[string]bool{}
	pfade := map[string]bool{}
	for _, x := range k {
		schemata[x.Schema] = true
		pfade[x.Pfad] = true
		if !strings.Contains(x.Adresse, "bblp:geheim@10.0.0.9:322") {
			t.Fatalf("Adresse unvollstaendig: %s", x.Adresse)
		}
	}
	for _, s := range []string{"rtsps", "rtspx"} {
		if !schemata[s] {
			t.Fatalf("Schema %s fehlt", s)
		}
	}
	if len(pfade) != 2 {
		t.Fatalf("erwartet 2 Pfade, bekommen %d", len(pfade))
	}
}

// probiereAdresse muss den Inhalt pruefen, nicht den Statuscode: go2rtc
// antwortet auch mit 200, wenn es kein Bild bekommen hat. Genau dieser Fall hat
// uns frueher wochenlang Erfolg vorgegaukelt.
func TestProbiereAdressePrueftInhalt(t *testing.T) {
	jpeg := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 500)...)

	faelle := []struct {
		name    string
		handler http.HandlerFunc
		willOK  bool
		grund   string
	}{
		{"echtes Bild", func(w http.ResponseWriter, r *http.Request) { w.Write(jpeg) }, true, ""},
		{"200 mit leerem Rumpf", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }, false, "kein Bild"},
		{"200 mit Textmuell", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("nope")) }, false, "kein Bild"},
		{"Fehlerstatus", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }, false, "HTTP 500"},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			srv := httptest.NewServer(f.handler)
			defer srv.Close()
			var port int
			fmt.Sscanf(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"), "%d", &port)
			alt := go2rtcPort
			go2rtcPort = port
			defer func() { go2rtcPort = alt }()

			res := probiereAdresse("rtsps://x", 3*time.Second)
			if res.OK != f.willOK {
				t.Fatalf("OK=%v, erwartet %v (Grund %q)", res.OK, f.willOK, res.Grund)
			}
			if f.grund != "" && !strings.Contains(res.Grund, f.grund) {
				t.Fatalf("Grund %q enthaelt nicht %q", res.Grund, f.grund)
			}
			if f.willOK && res.Bytes < 100 {
				t.Fatalf("Bildgroesse nicht gemeldet: %d", res.Bytes)
			}
		})
	}
}

// Wurde eine Form am Geraet gemessen, muss die Konfiguration sie zuerst nehmen —
// unabhaengig davon, was der Modellname vermuten laesst.
func TestGemessenesSchemaSchlaegtModellname(t *testing.T) {
	mu.Lock()
	altSchema := state.KameraSchema
	state.KameraSchema = map[string]string{"10.0.0.7": "rtspx|/streaming/live/0"}
	mu.Unlock()
	defer func() { mu.Lock(); state.KameraSchema = altSchema; mu.Unlock() }()

	// H2D wuerde sonst rtsps zuerst bekommen
	q := kameraQuellen(Printer{IP: "10.0.0.7", Code: "c", Model: "H2D"}, map[string]string{"10.0.0.7": "rtspx|/streaming/live/0"})
	if !strings.HasPrefix(q[0], "rtspx://") || !strings.HasSuffix(q[0], "/streaming/live/0") {
		t.Fatalf("gemessene Form nicht zuerst: %s", q[0])
	}
	if len(q) != 4 {
		t.Fatalf("die uebrigen Formen muessen als Rueckfall bleiben, sind %d", len(q))
	}
	// keine Dopplung
	gesehen := map[string]bool{}
	for _, a := range q {
		if gesehen[a] {
			t.Fatalf("Adresse doppelt: %s", a)
		}
		gesehen[a] = true
	}
}

// Der Deadlock aus der ersten Fassung: writeGo2rtcYaml haelt die Sperre und
// buildYaml griff erneut darauf zu. Der Test haette ewig gewartet — deshalb mit
// eigener Frist, damit ein Rueckfall als Fehlschlag endet und nicht als Haenger.
func TestBuildYamlSperrtNichtDoppelt(t *testing.T) {
	fertig := make(chan string, 1)
	go func() {
		mu.Lock()
		gemessen := map[string]string{"10.0.0.1": "rtsps|/streaming/live/1"}
		y := buildYaml([]Printer{{Name: "a", IP: "10.0.0.1", Code: "c", Model: "H2D"}}, gemessen)
		mu.Unlock()
		fertig <- y
	}()
	select {
	case y := <-fertig:
		if !strings.Contains(y, "rtsps://bblp:c@10.0.0.1:322/streaming/live/1") {
			t.Fatalf("gemessene Form fehlt:\n%s", y)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("buildYaml haengt — greift es wieder selbst auf die Sperre zu?")
	}
}
