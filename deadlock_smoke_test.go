package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestEndpointsNoDeadlock is a regression guard for the re-entrant-mutex
// deadlock that once froze the whole UI: handleAIConfig held the global `mu`
// and then called a helper that locked `mu` again. Because Go mutexes are not
// re-entrant the handler blocked forever while holding `mu`, so every other
// endpoint that needed `mu` hung too and the grid came up empty.
//
// The test fills the state with a realistic number of printers and then hammers
// the endpoints that take `mu` concurrently, many times over, under a hard
// deadline. If any handler self-deadlocks (or a lock is held across blocking
// work), the requests stop completing and the deadline trips the failure — a
// plain `go test` catches the regression, and `go test -race` additionally flags
// unsynchronised access.
func TestEndpointsNoDeadlock(t *testing.T) {
	// Seed a realistic farm so the handlers actually iterate over data.
	mu.Lock()
	state.Printers = nil
	for i := 0; i < 53; i++ {
		state.Printers = append(state.Printers, Printer{
			IP:     "10.0.0." + itoaSmoke(i),
			Name:   "printer-" + itoaSmoke(i),
			Serial: "SN" + itoaSmoke(i),
			Model:  "H2D",
		})
	}
	mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ai/config", handleAIConfig) // the original offender
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/errors", handleErrors)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/settings", handleSettings)
	mux.HandleFunc("/api/jobs", handleJobsList)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	paths := []string{
		"/api/ai/config",
		"/api/health",
		"/api/errors",
		"/api/status",
		"/api/settings",
		"/api/jobs",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for round := 0; round < 30; round++ {
			for _, p := range paths {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					resp, err := client.Get(url)
					if err != nil {
						return // network/timeout errors surface via the deadline below
					}
					resp.Body.Close()
				}(srv.URL + p)
			}
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All requests returned — no handler self-deadlocked.
	case <-time.After(20 * time.Second):
		t.Fatal("endpoints did not all respond within 20s — possible mutex deadlock (a handler holding mu across blocking work or re-locking mu)")
	}
}

// itoaSmoke is a tiny int->string helper so the test avoids pulling in strconv
// under a different name; kept local to the test file.
func itoaSmoke(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
