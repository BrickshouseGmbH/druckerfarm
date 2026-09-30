package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ─── SENDESEITE (/broadcast) ──────────────────────────────────────────────────
//
// A decoupled read-only view of the printers, meant for streaming (OBS) or for
// publishing (Cloudflare tunnel). It runs in its own browser tab
// with its own state — the main window's navigation does not touch it.
//
// The SAME page as the control UI is served; an injected
// options block (__BROADCAST_OPTS__) switches the UI into broadcast mode:
// grid only, no header, no control buttons.

func serveBroadcast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	page := replaceBroadcastOpts(pageWithVersion(), broadcastOptsJSON())
	w.Write([]byte(page))
}

// replaceBroadcastOpts inserts the options block into the page.
func replaceBroadcastOpts(page, opts string) string {
	return strings.ReplaceAll(page, "__BROADCAST_OPTS__", opts)
}

// broadcastOptsJSON builds the options block the broadcast page reads on start.
func broadcastOptsJSON() string { return broadcastOptsJSONRemote(false) }

// broadcastOptsJSONRemote also sets the remote flag: over the tunnel
// (remote=true) the broadcast page deliberately runs in snapshot mode, because
// live WebRTC does not traverse the tunnel.
func broadcastOptsJSONRemote(remote bool) string {
	mu.Lock()
	view := state.BroadcastView
	click := state.BroadcastClick
	cols := state.BroadcastCols
	mu.Unlock()
	if view != "online" && view != "offline" {
		view = "all"
	}
	if cols < 1 || cols > 20 {
		cols = 9
	}
	b, _ := json.Marshal(map[string]any{"view": view, "click": click, "cols": cols, "remote": remote})
	return string(b)
}

func broadcastURL() string {
	return fmt.Sprintf("http://localhost:%d/broadcast", appPort)
}

// handleBroadcastConfig reads/writes the broadcast page settings.
func handleBroadcastConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			View  *string `json:"view"`
			Click *bool   `json:"click"`
			Cols  *int    `json:"cols"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		if body.View != nil && (*body.View == "all" || *body.View == "online" || *body.View == "offline") {
			state.BroadcastView = *body.View
		}
		if body.Click != nil {
			state.BroadcastClick = *body.Click
		}
		if body.Cols != nil && *body.Cols >= 1 && *body.Cols <= 20 {
			state.BroadcastCols = *body.Cols
		}
		mu.Unlock()
		saveState()
	}

	mu.Lock()
	view := state.BroadcastView
	click := state.BroadcastClick
	cols := state.BroadcastCols
	mu.Unlock()
	if view == "" {
		view = "all"
	}
	if cols == 0 {
		cols = 9
	}
	writeJSON(w, map[string]any{"view": view, "click": click, "cols": cols, "url": broadcastURL()})
}
