package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// ─── FENSTER-ANWESENHEIT ──────────────────────────────────────────────────────
//
// How does the app know its window was closed? It used to rely on the browser
// starter process (unreliable: msedge hands the task to a running Edge and
// exits immediately) and on "no request for X seconds". Both were either too
// eager (server suicide on a slow start) or too sluggish (app kept running
// after closing and the watchdog pulled go2rtc back up).
//
//
// An open connection is reliable instead: the page keeps one open; as long as
// Server-Sent-Events-Kanal (/api/alive) offen. Solange mindestens einer offen
// it exists, the window is there. If the last one drops (window/tab closed) and
// no new one arrives within uiCloseGrace (which catches a reload), the app
// shuts down immediately and cleanly — including go2rtc, without a restart.

var (
	uiConns     atomic.Int64 // offene /api/alive-Verbindungen
	uiEverSeen  atomic.Bool  // was one ever open at all?
	uiGoneSince atomic.Int64 // UnixNano, since when none is open (0 = some open)

	// shuttingDown is set as soon as the program exits. While it is set, go2rtc
	// no longer starts (no watchdog restart during shutdown).
	shuttingDown atomic.Bool
)

// uiCloseGrace: how long to wait after the last connection drops before
// shutting down — enough for a reload (F5) to re-establish the connection.
var uiCloseGrace = 4 * time.Second

// handleAlive keeps an SSE channel open. The abort (window closed) arrives via
// the request context.
func handleAlive(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	uiConns.Add(1)
	uiEverSeen.Store(true)
	uiGoneSince.Store(0)
	defer func() {
		if uiConns.Add(-1) <= 0 {
			uiGoneSince.Store(time.Now().UnixNano())
		}
	}()

	fmt.Fprint(w, "retry: 2000\ndata: hello\n\n")
	fl.Flush()

	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			if _, err := fmt.Fprint(w, "data: ping\n\n"); err != nil {
				return
			}
			fl.Flush()
		}
	}
}

// uiPresenceWatcher calls onGone() once the window is really gone:
// connected at least once, now no connection, and that for
// laenger als uiCloseGrace.
func uiPresenceWatcher(onGone func()) {
	for {
		time.Sleep(500 * time.Millisecond)
		if uiConns.Load() > 0 {
			continue
		}
		if !uiEverSeen.Load() {
			continue // never connected yet -> startup phase, stay patient
		}
		gone := uiGoneSince.Load()
		if gone == 0 {
			continue
		}
		if time.Since(time.Unix(0, gone)) >= uiCloseGrace {
			onGone()
			return
		}
	}
}
