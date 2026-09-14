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

func withPrinters(t *testing.T, n int) {
	t.Helper()
	var ps []Printer
	for i := 1; i <= n; i++ {
		ps = append(ps, Printer{Name: fmt.Sprintf("woobly %d", i), IP: fmt.Sprintf("10.0.0.%d", i)})
	}
	mu.Lock()
	old := state.Printers
	state.Printers = ps
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); state.Printers = old; mu.Unlock() })
}

func resetFails(t *testing.T) {
	t.Helper()
	go2rtcMu.Lock()
	old := go2rtcFails
	go2rtcFails = 0
	go2rtcMu.Unlock()
	t.Cleanup(func() { go2rtcMu.Lock(); go2rtcFails = old; go2rtcMu.Unlock() })
}

// Bei vielen Druckern muss das Intervall angehoben werden — sonst treffen mehr
// Bildanfragen ein, als go2rtc verarbeiten kann.
func TestIntervalScalesWithPrinterCount(t *testing.T) {
	resetFails(t)
	cases := []struct {
		printers    int
		wantAtLeast time.Duration
	}{
		{1, snapMinInterval},
		{6, 2 * time.Second},
		{42, 14 * time.Second}, // 42 / 3 pro Sekunde
	}
	for _, c := range cases {
		withPrinters(t, c.printers)
		got := effectiveSnapInterval(2 * time.Second)
		if got < c.wantAtLeast {
			t.Errorf("%d Drucker: %v, erwartet mindestens %v", c.printers, got, c.wantAtLeast)
		}
		t.Logf("%2d Drucker → alle %v", c.printers, got)
	}
}

// Ein größeres Wunschintervall darf nicht verkleinert werden.
func TestLongerIntervalIsKept(t *testing.T) {
	resetFails(t)
	withPrinters(t, 3)
	if got := effectiveSnapInterval(30 * time.Second); got != 30*time.Second {
		t.Fatalf("30 s wurden auf %v geändert", got)
	}
}

// Stürzt go2rtc wiederholt ab, wird weiter zurückgefahren.
func TestIntervalBacksOffAfterCrashes(t *testing.T) {
	withPrinters(t, 42)
	go2rtcMu.Lock()
	old := go2rtcFails
	go2rtcFails = 0
	go2rtcMu.Unlock()
	defer func() { go2rtcMu.Lock(); go2rtcFails = old; go2rtcMu.Unlock() }()

	base := effectiveSnapInterval(2 * time.Second)
	go2rtcMu.Lock()
	go2rtcFails = 3
	go2rtcMu.Unlock()
	after := effectiveSnapInterval(2 * time.Second)

	if after <= base {
		t.Fatalf("nach Abstürzen nicht zurückgefahren: %v → %v", base, after)
	}
	t.Logf("ohne Abstürze alle %v, nach 3 Abstürzen alle %v", base, after)
}

// Die Oberfläche muss erfahren, welches Intervall wirklich gilt.
func TestResponseAnnouncesEffectiveInterval(t *testing.T) {
	resetFails(t)
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/frame.jpeg" {
			atomic.AddInt64(&hits, 1)
			w.Write(fakeJPEG)
			return
		}
		w.Write([]byte(`{"woobly-1":{}}`))
	}))
	defer srv.Close()
	port, _ := strconv.Atoi(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))
	oldPort := go2rtcPort
	go2rtcPort = port
	defer func() { go2rtcPort = oldPort }()

	withPrinters(t, 42)
	resetSnapState()

	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=2", nil))
	if rec.Code != 200 {
		t.Fatalf("code %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Snapshot-Throttled") != "1" {
		t.Fatal("die Drosselung wurde nicht gemeldet")
	}
	iv, _ := strconv.Atoi(rec.Header().Get("X-Snapshot-Interval"))
	if iv < 14 {
		t.Fatalf("gemeldetes Intervall zu kurz: %d s", iv)
	}

	// Zweite Anfrage kurz danach darf go2rtc nicht erneut belasten
	rec2 := httptest.NewRecorder()
	handleSnapshot(rec2, httptest.NewRequest("GET", "/api/snapshot/10.0.0.1?max_age=2", nil))
	if got := atomic.LoadInt64(&hits); got != 1 {
		t.Fatalf("trotz Drosselung %d Bildanfragen", got)
	}
	t.Logf("gemeldet: alle %d s, gedrosselt=%s", iv, rec.Header().Get("X-Snapshot-Throttled"))
}
