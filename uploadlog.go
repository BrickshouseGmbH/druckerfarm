package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

// ─── UPLOAD-HISTORIE ──────────────────────────────────────────────────────────
//
// What was uploaded to which printer and when used to live only in the display
// of the running job — after a restart it was gone. The history now lives
// in its own file (upload_log.json) in %APPDATA% and survives restarts.
// Everything is kept without a limit.

type UploadEntry struct {
	Zeit   time.Time `json:"zeit"`
	IP     string    `json:"ip"`
	Name   string    `json:"name"`
	Datei  string    `json:"datei"`
	Bytes  int64     `json:"bytes,omitempty"`
	Erfolg bool      `json:"erfolg"`
	Fehler string    `json:"fehler,omitempty"`
	Dauer  int64     `json:"dauer_ms,omitempty"`
}

// noteUpload appends an entry and saves. A save error must not disturb the
// upload itself, so nothing is
// zurueckgemeldet.
func noteUpload(e UploadEntry) {
	if e.Zeit.IsZero() {
		e.Zeit = time.Now()
	}
	mu.Lock()
	state.UploadLog = append(state.UploadLog, e)
	mu.Unlock()
	saveUploadLog()
}

func handleUploadLog(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		liste := make([]UploadEntry, len(state.UploadLog))
		copy(liste, state.UploadLog)
		mu.Unlock()
		// Neueste zuerst — danach sucht man.
		sort.Slice(liste, func(a, b int) bool { return liste[a].Zeit.After(liste[b].Zeit) })

		erfolge, fehler := 0, 0
		for _, e := range liste {
			if e.Erfolg {
				erfolge++
			} else {
				fehler++
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"eintraege": liste,
			"gesamt":    len(liste),
			"erfolge":   erfolge,
			"fehler":    fehler,
			"grenze":    0, // 0 = unbegrenzt
		})

	case http.MethodDelete:
		mu.Lock()
		anzahl := len(state.UploadLog)
		state.UploadLog = nil
		mu.Unlock()
		saveUploadLog()
		writeJSON(w, map[string]any{"geleert": anzahl})

	default:
		http.Error(w, "GET oder DELETE erwartet", http.StatusMethodNotAllowed)
	}
}
