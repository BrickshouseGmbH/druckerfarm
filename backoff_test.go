package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackoffGrowsAndIsCapped(t *testing.T) {
	prev := time.Duration(0)
	for i := 1; i <= 12; i++ {
		d := backoffFor(i)
		if d < 15*time.Second {
			t.Fatalf("fails=%d: %v is below the floor", i, d)
		}
		if d > snapMaxBackoff {
			t.Fatalf("fails=%d: %v exceeds the cap", i, d)
		}
		if d < prev {
			t.Fatalf("fails=%d: backoff shrank from %v to %v", i, prev, d)
		}
		prev = d
	}
	if backoffFor(1) >= backoffFor(4) {
		t.Fatal("backoff should grow with repeated failures")
	}
}

// Der Kern des Problems: ein hängender Drucker darf nicht dauerhaft einen der
// vier Transcode-Plätze belegen und die gesunden ausbremsen.
func TestFailingStreamStopsBeingAsked(t *testing.T) {
	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/frame.jpeg" {
			atomic.AddInt64(&calls, 1)
		}
		w.WriteHeader(http.StatusOK) // 200 mit leerem Rumpf = kein Frame
	}))
	defer srv.Close()

	port, _ := strconv.Atoi(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))
	oldPort := go2rtcPort
	go2rtcPort = port
	defer func() { go2rtcPort = oldPort }()

	resetSnapState()
	setPrinters(Printer{Name: "kaputt", IP: "10.0.0.1"})

	// Erster Versuch schlägt fehl und setzt die Pause
	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=1", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("want 1 frame request, got %d", got)
	}
	if rec.Header().Get("X-Snapshot-Retry-In") == "" {
		t.Fatal("the UI needs to know how long the stream is paused")
	}

	// Weitere Anfragen dürfen go2rtc nicht mehr belasten
	for i := 0; i < 20; i++ {
		r2 := httptest.NewRecorder()
		handleSnapshot(r2, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=1", nil))
		if r2.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", r2.Code)
		}
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("paused stream was asked for a frame %d times — it must stay at 1", got)
	}
}

// Ein gesunder Drucker darf durch einen kaputten nicht ausgebremst werden.
func TestHealthyStreamIsNotBlockedByABrokenOne(t *testing.T) {
	var slowCalls, fastCalls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("src"), "kaputt") {
			atomic.AddInt64(&slowCalls, 1)
			time.Sleep(300 * time.Millisecond) // hängt
			w.WriteHeader(http.StatusOK)
			return
		}
		atomic.AddInt64(&fastCalls, 1)
		w.Write(fakeJPEG)
	}))
	defer srv.Close()

	port, _ := strconv.Atoi(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))
	oldPort := go2rtcPort
	go2rtcPort = port
	defer func() { go2rtcPort = oldPort }()

	resetSnapState()
	var ps []Printer
	for i := 1; i <= 8; i++ {
		ps = append(ps, Printer{Name: fmt.Sprintf("kaputt %d", i), IP: fmt.Sprintf("10.0.0.%d", i)})
	}
	ps = append(ps, Printer{Name: "gut", IP: "10.0.1.1"})
	setPrinters(ps...)

	// Erst die kaputten anstoßen, damit sie in die Pause gehen
	for i := 1; i <= 8; i++ {
		rec := httptest.NewRecorder()
		handleSnapshot(rec, httptest.NewRequest("GET", fmt.Sprintf("/api/snapshot/10.0.0.%d?max_age=1", i), nil))
	}

	start := time.Now()
	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest("GET", "/api/snapshot/10.0.1.1?max_age=1", nil))
	took := time.Since(start)

	if rec.Code != 200 {
		t.Fatalf("healthy printer got %d: %s", rec.Code, rec.Body.String())
	}
	if took > 250*time.Millisecond {
		t.Fatalf("healthy printer waited %v behind the broken ones", took)
	}
	t.Logf("gesunder Drucker nach %v bedient, %d Anfragen an kaputte Streams", took, atomic.LoadInt64(&slowCalls))
}

// Nach einem erfolgreichen Abruf muss die Sperre verschwinden.
func TestRecoveryClearsThePause(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Write(fakeJPEG)
	}))
	defer srv.Close()

	port, _ := strconv.Atoi(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))
	oldPort := go2rtcPort
	go2rtcPort = port
	defer func() { go2rtcPort = oldPort }()

	resetSnapState()
	setPrinters(Printer{Name: "wackelig", IP: "10.0.0.1"})

	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=1", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}

	// Pause von Hand aufheben, so als wäre sie abgelaufen, und Drucker heilen
	fail.Store(false)
	e := snapEntryFor(streamName(Printer{Name: "wackelig", IP: "10.0.0.1"}))
	e.mu.Lock()
	e.nextTry = time.Time{}
	e.ts = time.Time{}
	e.mu.Unlock()

	rec2 := httptest.NewRecorder()
	handleSnapshot(rec2, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=1", nil))
	if rec2.Code != 200 {
		t.Fatalf("recovered printer got %d", rec2.Code)
	}
	e.mu.Lock()
	fails, next := e.fails, e.nextTry
	e.mu.Unlock()
	if fails != 0 || !next.IsZero() {
		t.Fatalf("pause was not cleared: fails=%d next=%v", fails, next)
	}
}

// ─── Logo ─────────────────────────────────────────────────────────────────────

func TestLogoIsEmbeddedAndServed(t *testing.T) {
	if len(logoSVG) < 500 {
		t.Fatalf("logo looks empty: %d bytes", len(logoSVG))
	}
	if !strings.Contains(string(logoSVG), "<svg") {
		t.Fatal("embedded file is not an SVG")
	}
	rec := httptest.NewRecorder()
	handleLogo(rec, httptest.NewRequest("GET", "/logo.svg", nil))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Fatalf("content-type %q", ct)
	}
	if rec.Body.Len() != len(logoSVG) {
		t.Fatal("served logo differs from the embedded one")
	}
}

// ─── Komponenten-Aktualisierung ───────────────────────────────────────────────

func TestGo2rtcVersionParsing(t *testing.T) {
	// Der echte Ausgabetext des Programms
	cases := map[string]string{
		"go2rtc version 1.9.14+dev.dc1685e.dirty (dc1685e.dirty) windows/amd64": "v1.9.14",
		"go2rtc version 1.9.4 linux/amd64":                                      "v1.9.4",
		"":                                                                      "",
		"irgendwas anderes":                                                     "",
	}
	for in, want := range cases {
		got := parseGo2rtcVersionLine(in)
		if got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}

func TestComponentHashIsRemembered(t *testing.T) {
	dir := useTempAppDir(t)
	oldData := dataFile
	dataFile = dir + "/config.json"
	defer func() { dataFile = oldData }()

	mu.Lock()
	state.ComponentSHA = nil
	mu.Unlock()

	rememberComponentHash("ffmpeg", "abc123")
	if got := componentHash("ffmpeg"); got != "abc123" {
		t.Fatalf("got %q", got)
	}
	if componentHash("go2rtc") != "" {
		t.Fatal("unknown component should be empty")
	}
	mu.Lock()
	state.ComponentSHA = nil
	mu.Unlock()
}
