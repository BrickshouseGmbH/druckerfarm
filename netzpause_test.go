package main

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Waehrend der Pause darf kein Bild mehr ausgeliefert werden — sonst haelt das
// Programm zwar seine Schleifen an, holt aber weiter Bilder, sobald die
// Oberflaeche fragt.
func TestSnapshotWaehrendPauseAbgelehnt(t *testing.T) {
	netzPause.Store(true)
	defer netzPause.Store(false)

	req := httptest.NewRequest(http.MethodGet, "/api/snapshot/10.0.0.1", nil)
	rec := httptest.NewRecorder()
	handleSnapshot(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("HTTP %d, erwartet 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "angehalten") {
		t.Fatalf("Grund fehlt: %q", rec.Body.String())
	}
}

// Der Schalter meldet zurueck, was tatsaechlich geschehen ist. Zweimal
// dasselbe zu setzen darf nichts ausloesen.
func TestNetzPauseSchalter(t *testing.T) {
	netzPause.Store(false)
	defer netzPause.Store(false)

	erst := setzeNetzPause(true)
	if erst["pausiert"] != true || erst["geaendert"] != true {
		t.Fatalf("erstes Anhalten: %+v", erst)
	}
	if !netzPausiert() {
		t.Fatal("Zustand nicht gesetzt")
	}
	nochmal := setzeNetzPause(true)
	if nochmal["geaendert"] != false {
		t.Fatalf("zweites Anhalten darf nichts tun: %+v", nochmal)
	}
}

// Der Zustand wird bewusst NICHT gespeichert: ein Schalter, der ueber Nacht
// stehen bleibt, laesst am naechsten Morgen alles tot aussehen.
func TestNetzPauseWirdNichtGespeichert(t *testing.T) {
	netzPause.Store(true)
	defer netzPause.Store(false)
	mu.Lock()
	roh := stateAlsText()
	mu.Unlock()
	if strings.Contains(strings.ToLower(roh), "netzpause") || strings.Contains(strings.ToLower(roh), "pausiert") {
		t.Fatalf("Pausezustand darf nicht in der Konfiguration landen:\n%s", roh)
	}
}

// Der Endpunkt liefert den Zustand und nimmt ihn entgegen.
func TestNetzPauseEndpunkt(t *testing.T) {
	netzPause.Store(false)
	defer netzPause.Store(false)

	rec := httptest.NewRecorder()
	handleNetzPause(rec, httptest.NewRequest(http.MethodGet, "/api/network/pause", nil))
	if !strings.Contains(rec.Body.String(), `"pausiert":false`) {
		t.Fatalf("GET meldet falsch: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	handleNetzPause(rec, httptest.NewRequest(http.MethodPost, "/api/network/pause",
		strings.NewReader(`{"pausiert":true}`)))
	if rec.Code != 200 || !netzPausiert() {
		t.Fatalf("POST wirkungslos: HTTP %d, Zustand %v", rec.Code, netzPausiert())
	}
}

// Der Aufpasser darf go2rtc waehrend der Pause nicht wieder hochziehen — sonst
// ist die Pause nach wenigen Sekunden von selbst vorbei. Der Fehler ist beim
// Lauf mit -race aufgefallen, weil ein solcher Neustart in den naechsten Test
// hineinlief.
func TestAufpasserStartetWaehrendPauseNicht(t *testing.T) {
	netzPause.Store(true)
	defer netzPause.Store(false)

	go2rtcMu.Lock()
	altWanted := go2rtcWanted
	go2rtcWanted = true
	gen := go2rtcGen
	go2rtcMu.Unlock()
	defer func() { go2rtcMu.Lock(); go2rtcWanted = altWanted; go2rtcMu.Unlock() }()

	// Ein Vorgang, der sofort endet — wie ein beendetes go2rtc.
	cmd := exec.Command(schnellesEnde())
	if err := cmd.Start(); err != nil {
		t.Skipf("kein Hilfsprogramm zum Testen vorhanden: %v", err)
	}

	fertig := make(chan struct{})
	go func() { superviseGo2rtc(cmd, gen); close(fertig) }()

	select {
	case <-fertig:
	case <-time.After(3 * time.Second):
		t.Fatal("superviseGo2rtc haengt — plant es einen Neustart?")
	}

	go2rtcMu.Lock()
	fails := go2rtcFails
	go2rtcMu.Unlock()
	if fails != 0 {
		t.Fatalf("waehrend der Pause darf kein Fehlversuch gezaehlt werden, sind %d", fails)
	}
}

func schnellesEnde() string {
	for _, p := range []string{"/bin/true", "/usr/bin/true"} {
		if fileExists(p) {
			return p
		}
	}
	return "true"
}
