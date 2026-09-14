package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

// ─── UPLOAD-HISTORIE ──────────────────────────────────────────────────────────
//
// Was wann auf welchen Drucker geladen wurde, stand bisher nur in der Anzeige
// des laufenden Vorgangs — nach einem Neustart war es weg. Die Historie liegt
// jetzt in der Konfiguration in %APPDATA% und ueberlebt Neustarts.
//
// Aufgehoben werden hoechstens uploadLogMax Eintraege; die aeltesten fallen
// hinten heraus. Bei 42 Druckern und mehreren Dateien je Auftrag waeren sonst
// schnell zehntausende Zeilen beisammen.

const uploadLogMax = 2000

type UploadEintrag struct {
	Zeit   time.Time `json:"zeit"`
	IP     string    `json:"ip"`
	Name   string    `json:"name"`
	Datei  string    `json:"datei"`
	Bytes  int64     `json:"bytes,omitempty"`
	Erfolg bool      `json:"erfolg"`
	Fehler string    `json:"fehler,omitempty"`
	Dauer  int64     `json:"dauer_ms,omitempty"`
}

// merkeUpload haengt einen Eintrag an und speichert. Fehler beim Speichern
// duerfen den Upload selbst nicht stoeren, deshalb wird hier nichts
// zurueckgemeldet.
func merkeUpload(e UploadEintrag) {
	if e.Zeit.IsZero() {
		e.Zeit = time.Now()
	}
	mu.Lock()
	state.UploadLog = append(state.UploadLog, e)
	if len(state.UploadLog) > uploadLogMax {
		state.UploadLog = state.UploadLog[len(state.UploadLog)-uploadLogMax:]
	}
	mu.Unlock()
	saveState()
}

func handleUploadLog(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		liste := make([]UploadEintrag, len(state.UploadLog))
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
			"grenze":    uploadLogMax,
		})

	case http.MethodDelete:
		mu.Lock()
		anzahl := len(state.UploadLog)
		state.UploadLog = nil
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"geleert": anzahl})

	default:
		http.Error(w, "GET oder DELETE erwartet", http.StatusMethodNotAllowed)
	}
}
