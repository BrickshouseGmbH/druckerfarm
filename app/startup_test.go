package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func withShortTimeouts(t *testing.T, idle, never, tick time.Duration) {
	t.Helper()
	oi, on, ot := uiIdleTimeout, uiNeverGrace, uiIdlePollTick
	uiIdleTimeout, uiNeverGrace, uiIdlePollTick = idle, never, tick
	t.Cleanup(func() { uiIdleTimeout, uiNeverGrace, uiIdlePollTick = oi, on, ot })
	lastRequest.Store(0)
}

// The browser process exiting early must not shut the app down while the UI is
// still polling — that was the ERR_CONNECTION_REFUSED at first start.
func TestWaitUntilUIIdleStaysWhileUIIsPolling(t *testing.T) {
	withShortTimeouts(t, 300*time.Millisecond, 10*time.Second, 20*time.Millisecond)

	stop := make(chan struct{})
	go func() { // simulate the dashboard polling
		tk := time.NewTicker(50 * time.Millisecond)
		defer tk.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tk.C:
				noteRequest()
			}
		}
	}()

	done := make(chan struct{})
	go func() { waitUntilUIIdle(); close(done) }()

	select {
	case <-done:
		close(stop)
		t.Fatal("shut down while the UI was still active")
	case <-time.After(700 * time.Millisecond):
	}

	close(stop) // window really closed now
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("did not shut down after the UI went quiet")
	}
}

func TestWaitUntilUIIdleGivesUpIfNobodyEverConnects(t *testing.T) {
	withShortTimeouts(t, time.Hour, 250*time.Millisecond, 20*time.Millisecond)
	start := time.Now()
	waitUntilUIIdle()
	if d := time.Since(start); d < 200*time.Millisecond || d > 3*time.Second {
		t.Fatalf("grace period not honoured: %v", d)
	}
}

func TestMiddlewareRecordsRequests(t *testing.T) {
	lastRequest.Store(0)
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/status", nil))
	if lastRequest.Load() == 0 {
		t.Fatal("request was not recorded")
	}
}

// A second instance must not silently half-start: binding has to fail visibly.
func TestSecondInstanceCannotBindThePort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	if _, err := net.Listen("tcp", ln.Addr().String()); err == nil {
		t.Fatal("expected the second bind to fail")
	}
}
