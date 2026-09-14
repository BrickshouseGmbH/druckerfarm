package main

import (
	"encoding/json"
	"net/http"
)

// ─── PER-DRUCKER-SCHALTER (Blinken / Kamera) ─────────────────────────────────

// blinkAusFuer sagt, ob das Fehler-Blinken für diesen Drucker abgeschaltet ist.
func blinkAusFuer(ip string) bool {
	mu.Lock()
	defer mu.Unlock()
	return state.BlinkAus[ip]
}

// kameraAusFuer sagt, ob die Kamera dieses Druckers dauerhaft aus ist.
func kameraAusFuer(ip string) bool {
	mu.Lock()
	defer mu.Unlock()
	return state.KameraAus[ip]
}

// handleBlink schaltet das Fehler-Blinken für EINEN Drucker an/aus.
func handleBlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP  string `json:"ip"`
		Aus bool   `json:"aus"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IP == "" {
		http.Error(w, "ip/aus erforderlich", http.StatusBadRequest)
		return
	}
	mu.Lock()
	if state.BlinkAus == nil {
		state.BlinkAus = map[string]bool{}
	}
	if body.Aus {
		state.BlinkAus[body.IP] = true
	} else {
		delete(state.BlinkAus, body.IP)
	}
	// Kopie für evtl. Licht-Reset
	var ziel *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == body.IP {
			p := state.Printers[i]
			ziel = &p
			break
		}
	}
	mu.Unlock()
	saveState()
	// Beim Abschalten sofort normales Licht herstellen (falls es gerade blinkt).
	if body.Aus && ziel != nil {
		go func(p Printer) { _ = mqttMgr.SetChamberLight(p, "on", 500, 500) }(*ziel)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "ip": body.IP, "aus": body.Aus})
}

// handleCameraOff schaltet die Kamera EINES Druckers dauerhaft an/aus (Anzeige
// „Private", kein Stream). Rein lokal — go2rtc wird nicht verändert.
func handleCameraOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP  string `json:"ip"`
		Aus bool   `json:"aus"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IP == "" {
		http.Error(w, "ip/aus erforderlich", http.StatusBadRequest)
		return
	}
	mu.Lock()
	if state.KameraAus == nil {
		state.KameraAus = map[string]bool{}
	}
	if body.Aus {
		state.KameraAus[body.IP] = true
	} else {
		delete(state.KameraAus, body.IP)
	}
	mu.Unlock()
	saveState()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "ip": body.IP, "aus": body.Aus})
}
