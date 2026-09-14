package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// minimal but valid JPEG header + EOI
var fakeJPEG = append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x41}, 64)...)

type fakeG2 struct {
	srv      *httptest.Server
	hits     int64
	maxPar   int64
	curPar   int64
	emptyFor string
	delay    time.Duration
}

func startFakeG2(t *testing.T, emptyFor string, delay time.Duration) *fakeG2 {
	t.Helper()
	f := &fakeG2{emptyFor: emptyFor, delay: delay}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/frame.jpeg" {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt64(&f.hits, 1)
		cur := atomic.AddInt64(&f.curPar, 1)
		for {
			old := atomic.LoadInt64(&f.maxPar)
			if cur <= old || atomic.CompareAndSwapInt64(&f.maxPar, old, cur) {
				break
			}
		}
		defer atomic.AddInt64(&f.curPar, -1)
		time.Sleep(f.delay)

		src := r.URL.Query().Get("src")
		w.Header().Set("Content-Type", "image/jpeg")
		if src == f.emptyFor {
			// exactly what go2rtc does when it cannot transcode: 200, empty body
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Write(fakeJPEG)
	}))

	// point the app at the fake go2rtc
	u := strings.TrimPrefix(f.srv.URL, "http://127.0.0.1:")
	p, err := strconv.Atoi(u)
	if err != nil {
		t.Fatalf("port parse: %v", err)
	}
	go2rtcPort = p
	t.Cleanup(f.srv.Close)
	return f
}

func resetSnapState() {
	snapMu.Lock()
	snapCache = map[string]*snapEntry{}
	snapMu.Unlock()
}

func setPrinters(ps ...Printer) {
	mu.Lock()
	state.Printers = ps
	mu.Unlock()
}

func get(path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// The old frontend built the stream name itself and got it wrong. The endpoint
// must derive it from the printer via streamName().
func TestSnapshotUsesServerSideStreamName(t *testing.T) {
	f := startFakeG2(t, "", 0)
	resetSnapState()
	setPrinters(Printer{Name: "Woobly 7", IP: "10.0.0.7"})

	var seen string
	f.srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query().Get("src")
		w.Write(fakeJPEG)
	})

	rec := get("/api/snapshot/10.0.0.7")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if want := streamName(Printer{Name: "Woobly 7", IP: "10.0.0.7"}); seen != want {
		t.Fatalf("stream name: got %q, want %q", seen, want)
	}
	if seen != "woobly-7" {
		t.Fatalf("unexpected stream name %q", seen)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("content-type %q", ct)
	}
}

// go2rtc answers 200 with an empty body when it cannot produce a frame.
// That must not reach the UI as a successful (blank) image.
func TestEmptyBodyBecomesError(t *testing.T) {
	startFakeG2(t, "cam-a", 0)
	resetSnapState()
	setPrinters(Printer{Name: "cam a", IP: "10.0.0.1"})

	rec := get("/api/snapshot/10.0.0.1")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	// Die Meldung soll den tatsächlichen Grund nennen, nicht pauschal ffmpeg
	// verdächtigen — ffmpeg ist auf der Farm nachweislich installiert.
	body := rec.Body.String()
	if !strings.Contains(body, "kein Bild") {
		t.Fatalf("reason missing: %q", body)
	}
	if strings.Contains(strings.ToLower(body), "ffmpeg fehlt im path") {
		t.Fatalf("blanket ffmpeg blame is back: %q", body)
	}
}

func TestUnknownPrinter(t *testing.T) {
	startFakeG2(t, "", 0)
	resetSnapState()
	setPrinters(Printer{Name: "cam a", IP: "10.0.0.1"})

	if rec := get("/api/snapshot/10.0.0.99"); rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
	if rec := get("/api/snapshot/"); rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// Repeated polling inside the interval must be served from the cache instead of
// triggering a new transcode for every request.
func TestCacheHonoursMaxAge(t *testing.T) {
	f := startFakeG2(t, "", 0)
	resetSnapState()
	setPrinters(Printer{Name: "cam a", IP: "10.0.0.1"})

	for i := 0; i < 5; i++ {
		if rec := get("/api/snapshot/10.0.0.1?max_age=10"); rec.Code != http.StatusOK {
			t.Fatalf("req %d: %d", i, rec.Code)
		}
	}
	if h := atomic.LoadInt64(&f.hits); h != 1 {
		t.Fatalf("want 1 upstream fetch, got %d", h)
	}
}

// 36 printers polling at once must not turn into 36 parallel transcodes.
func TestConcurrencyIsCapped(t *testing.T) {
	f := startFakeG2(t, "", 40*time.Millisecond)
	resetSnapState()

	var ps []Printer
	for i := 1; i <= 36; i++ {
		ps = append(ps, Printer{Name: fmt.Sprintf("Woobly %d", i), IP: fmt.Sprintf("10.0.0.%d", i)})
	}
	setPrinters(ps...)

	var wg sync.WaitGroup
	for i := 1; i <= 36; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if rec := get(fmt.Sprintf("/api/snapshot/10.0.0.%d?max_age=1", i)); rec.Code != http.StatusOK {
				t.Errorf("printer %d: %d", i, rec.Code)
			}
		}(i)
	}
	wg.Wait()

	if max := atomic.LoadInt64(&f.maxPar); max > snapMaxConcurrent {
		t.Fatalf("max parallel upstream requests %d > limit %d", max, snapMaxConcurrent)
	}
	if h := atomic.LoadInt64(&f.hits); h != 36 {
		t.Fatalf("want 36 fetches (one per stream), got %d", h)
	}
}

// Concurrent requests for the same printer must collapse into one transcode.
func TestSameStreamCollapses(t *testing.T) {
	f := startFakeG2(t, "", 30*time.Millisecond)
	resetSnapState()
	setPrinters(Printer{Name: "cam a", IP: "10.0.0.1"})

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); get("/api/snapshot/10.0.0.1?max_age=5") }()
	}
	wg.Wait()

	if h := atomic.LoadInt64(&f.hits); h != 1 {
		t.Fatalf("want 1 upstream fetch, got %d", h)
	}
}

// max_age below the floor must be clamped, otherwise a 0.1s poll would hammer go2rtc.
func TestMaxAgeFloor(t *testing.T) {
	f := startFakeG2(t, "", 0)
	resetSnapState()
	setPrinters(Printer{Name: "cam a", IP: "10.0.0.1"})

	for i := 0; i < 4; i++ {
		get("/api/snapshot/10.0.0.1?max_age=0.01")
	}
	if h := atomic.LoadInt64(&f.hits); h != 1 {
		t.Fatalf("floor not applied: %d upstream fetches", h)
	}
}
