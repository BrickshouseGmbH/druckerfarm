package main

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// During the pause no image may be served — otherwise the program
// pauses its loops but keeps fetching images as soon as the
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

// The switch reports back what actually happened. Twice
// dasselbe zu setzen darf nichts ausloesen.
func TestNetzPauseSchalter(t *testing.T) {
	netzPause.Store(false)
	defer netzPause.Store(false)

	erst := setNetworkPause(true)
	if erst["pausiert"] != true || erst["geaendert"] != true {
		t.Fatalf("erstes Anhalten: %+v", erst)
	}
	if !netPaused() {
		t.Fatal("Zustand nicht gesetzt")
	}
	nochmal := setNetworkPause(true)
	if nochmal["geaendert"] != false {
		t.Fatalf("zweites Anhalten darf nichts tun: %+v", nochmal)
	}
}

// The state is deliberately NOT persisted: a switch left on overnight
// makes everything look dead the next morning.
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

// The endpoint returns the state and accepts it.
func TestNetzPauseEndpunkt(t *testing.T) {
	netzPause.Store(false)
	defer netzPause.Store(false)

	rec := httptest.NewRecorder()
	handleNetworkPause(rec, httptest.NewRequest(http.MethodGet, "/api/network/pause", nil))
	if !strings.Contains(rec.Body.String(), `"pausiert":false`) {
		t.Fatalf("GET meldet falsch: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	handleNetworkPause(rec, httptest.NewRequest(http.MethodPost, "/api/network/pause",
		strings.NewReader(`{"pausiert":true}`)))
	if rec.Code != 200 || !netPaused() {
		t.Fatalf("POST wirkungslos: HTTP %d, Zustand %v", rec.Code, netPaused())
	}
}

// The supervisor must not pull go2rtc back up during the pause — otherwise
// the pause is over by itself within seconds. The bug was found during a
// -race run, because such a restart reached into the next test
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

	// An operation that ends immediately — like a stopped go2rtc.
	cmd := exec.Command(fastExit())
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

func fastExit() string {
	for _, p := range []string{"/bin/true", "/usr/bin/true"} {
		if fileExists(p) {
			return p
		}
	}
	return "true"
}
