package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Ein doppelter Schlüssel lässt go2rtc die GESAMTE streams-Sektion verwerfen —
// nachgewiesen gegen das echte go2rtc. Deshalb darf die Datei keine enthalten.
func TestYamlHasNoDuplicateKeys(t *testing.T) {
	printers := []Printer{
		{Name: "woobly 1", IP: "10.0.0.1", Code: "a"},
		{Name: "Woobly-1", IP: "10.0.0.2", Code: "b"}, // ergibt denselben Namen
		{Name: "WOOBLY 1", IP: "10.0.0.3", Code: "c"}, // und nochmal
		{Name: "woobly 2", IP: "10.0.0.4", Code: "d"},
	}
	yaml := buildYaml(printers, nil)

	keys := map[string]int{}
	for _, line := range strings.Split(yaml, "\n") {
		if strings.HasPrefix(line, "  ") && strings.HasSuffix(strings.TrimSpace(line), ":") &&
			!strings.HasPrefix(strings.TrimSpace(line), "-") {
			keys[strings.TrimSpace(line)]++
		}
	}
	for k, n := range keys {
		if n > 1 {
			t.Fatalf("Schlüssel %q kommt %dx vor — go2rtc würde alle Streams verwerfen:\n%s", k, n, yaml)
		}
	}
	// Alle vier Drucker müssen vertreten sein
	for _, p := range printers {
		if !strings.Contains(yaml, p.IP) {
			t.Fatalf("%s fehlt in der Konfiguration:\n%s", p.IP, yaml)
		}
	}
	t.Logf("erzeugte Schlüssel: %d für %d Drucker", len(keys), len(printers))
}

func TestStreamNameForResolvesDuplicates(t *testing.T) {
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{
		{Name: "woobly 1", IP: "10.0.0.1"},
		{Name: "Woobly-1", IP: "10.0.0.2"},
	}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	a, b := streamNameFor("10.0.0.1"), streamNameFor("10.0.0.2")
	if a == b {
		t.Fatalf("beide Drucker bekommen denselben Streamnamen: %q", a)
	}
	if a == "" || b == "" {
		t.Fatalf("Streamname fehlt: %q %q", a, b)
	}
	t.Logf("%q und %q", a, b)
}

// Die Diagnose muss das Nötige liefern, ohne Zugangscodes preiszugeben.
func TestDiagnosticsAreCompleteAndSafe(t *testing.T) {
	dir := t.TempDir()
	oldDir := appDir
	appDir = dir
	defer func() { appDir = oldDir }()

	os.WriteFile(filepath.Join(dir, "go2rtc.log"), []byte("Zeile A\nZeile B\nZeile C\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "druckerfarm.log"), []byte("Start\n"), 0o644)

	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{Name: "woobly 1", IP: "10.0.0.1", Code: "streng-geheim", Serial: "S1"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	rec := httptest.NewRecorder()
	handleDiagnostics(rec, httptest.NewRequest("GET", "/api/diagnostics", nil))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	body := rec.Body.String()

	if strings.Contains(body, "streng-geheim") {
		t.Fatal("der Zugangscode darf nicht in der Diagnose stehen")
	}
	for _, want := range []string{"go2rtc_present", "port_reachable", "streams_loaded", "go2rtc_log", "name_collisions", "should_run"} {
		if !strings.Contains(body, want) {
			t.Fatalf("%s fehlt in der Diagnose", want)
		}
	}
	var d struct {
		Log      []string `json:"go2rtc_log"`
		Printers int      `json:"printers"`
	}
	json.Unmarshal(rec.Body.Bytes(), &d)
	if len(d.Log) != 3 || d.Log[2] != "Zeile C" {
		t.Fatalf("Log nicht korrekt gelesen: %v", d.Log)
	}
	if d.Printers != 1 {
		t.Fatalf("Druckerzahl falsch: %d", d.Printers)
	}
}

func TestTailFileHandlesMissing(t *testing.T) {
	lines := tailFile(filepath.Join(t.TempDir(), "gibtsnicht.log"), 10)
	if len(lines) != 1 || !strings.Contains(lines[0], "no such file") {
		t.Fatalf("fehlende Datei nicht sauber gemeldet: %v", lines)
	}
}

// go2rtc muss seine Ausgabe in eine Datei schreiben — sonst ist jede Fehlersuche Raten.
func TestGo2rtcOutputIsCaptured(t *testing.T) {
	dir := t.TempDir()
	oldDir := appDir
	appDir = dir
	defer func() {
		appDir = oldDir
		stopGo2rtc()
	}()

	os.WriteFile(filepath.Join(dir, "go2rtc"), []byte("#!/bin/sh\necho 'ich sage warum ich sterbe' >&2\nexit 3\n"), 0o755)
	go2rtcMu.Lock()
	go2rtcWanted, go2rtcFails = false, 0
	go2rtcMu.Unlock()

	startGo2rtc()
	waitFor(t, filepath.Join(dir, "go2rtc.log"), "ich sage warum")
	stopGo2rtc()
}

func waitFor(t *testing.T, path, needle string) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if b, err := os.ReadFile(path); err == nil && strings.Contains(string(b), needle) {
			t.Logf("mitgeschrieben: %s", strings.TrimSpace(string(b)))
			return
		}
		sleepShort()
	}
	t.Fatalf("%q wurde nicht in %s mitgeschrieben", needle, path)
}

func sleepShort() { time.Sleep(100 * time.Millisecond) }

// Die H-Reihe antwortet auf rtspx:// mit einer Umleitung und liefert kein Bild.
// Fuer sie muss rtsps:// zuerst stehen; fuer die X1-Reihe bleibt es umgekehrt.
// Angegeben werden immer beide, damit ein Geraet auch dann ein Bild liefert,
// wenn es sich anders verhaelt als sein Modellname vermuten laesst.
func TestKameraQuellenReihenfolge(t *testing.T) {
	faelle := []struct {
		modell, zuerst string
	}{
		{"H2D", "rtsps://"},
		{"H2C", "rtsps://"},
		{"h2s", "rtsps://"},
		{"P2S", "rtsps://"},
		{"X1C", "rtspx://"},
		{"X1E", "rtspx://"},
		{"P1S", "rtspx://"},
		{"", "rtspx://"},
	}
	for _, f := range faelle {
		q := kameraQuellen(Printer{Model: f.modell, IP: "10.0.0.1", Code: "abc"}, nil)
		if len(q) != 2 {
			t.Fatalf("%s: erwartet 2 Adressen, bekommen %d", f.modell, len(q))
		}
		if !strings.HasPrefix(q[0], f.zuerst) {
			t.Fatalf("%s: erste Adresse sollte mit %s beginnen: %s", f.modell, f.zuerst, q[0])
		}
		// Die jeweils andere Form muss als Rueckfallebene dabei sein.
		if strings.HasPrefix(q[1], f.zuerst) {
			t.Fatalf("%s: zweite Adresse ist dieselbe Form: %s", f.modell, q[1])
		}
		for _, a := range q {
			if !strings.Contains(a, "bblp:abc@10.0.0.1:322/streaming/live/1") {
				t.Fatalf("%s: Adresse unvollstaendig: %s", f.modell, a)
			}
		}
	}
}

// In der Konfiguration muessen beide Adressen unter demselben Namen stehen.
func TestYamlEnthaeltBeideAdressen(t *testing.T) {
	y := buildYaml([]Printer{{Name: "h11", Model: "H2D", IP: "10.0.0.5", Code: "xy"}}, nil)
	if !strings.Contains(y, "rtsps://bblp:xy@10.0.0.5:322") {
		t.Fatalf("rtsps fehlt:\n%s", y)
	}
	if !strings.Contains(y, "rtspx://bblp:xy@10.0.0.5:322") {
		t.Fatalf("rtspx fehlt:\n%s", y)
	}
	// rtsps muss bei der H-Reihe zuerst kommen
	if strings.Index(y, "rtsps://") > strings.Index(y, "rtspx://") {
		t.Fatalf("bei H2D muss rtsps zuerst stehen:\n%s", y)
	}
}


