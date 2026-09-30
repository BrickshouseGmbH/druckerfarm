package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// printerByIP returns a copy of the stored printer for the IP.
func printerByIP(ip string) *Printer {
	mu.Lock()
	defer mu.Unlock()
	for i := range state.Printers {
		if state.Printers[i].IP == ip {
			p := state.Printers[i]
			return &p
		}
	}
	return nil
}

// sendUpgradeConfirm issues the (undocumented, best-effort) command that
// confirms the firmware update the printer already knows about. Same payload
// as handleFwUpdateStart, but callable internally (e.g. from fwAutoLoop).
func sendUpgradeConfirm(ip string) error {
	z := printerByIP(ip)
	if z == nil {
		return errNichtGefunden
	}
	payload := `{"upgrade":{"sequence_id":"` + nextSeq() + `","command":"upgrade_confirm","src_id":1}}`
	return mqttMgr.PublishRaw(*z, payload)
}

var errNichtGefunden = errNF("Drucker nicht gefunden")

type errNF string

func (e errNF) Error() string { return string(e) }

// handleXcam toggles the built-in AI/camera monitoring (AI monitoring)
// and the abort sensitivity (Halt sensitivity) on the printer. Best-effort;
// not every firmware/model reacts. Not persisted, because the
// printer reports its own state.
func handleXcam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP          string `json:"ip"`
		Monitoring  *bool  `json:"monitoring,omitempty"`
		Sensitivity string `json:"sensitivity,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	z := printerByIP(body.IP)
	if z == nil {
		http.Error(w, "Drucker nicht gefunden", http.StatusNotFound)
		return
	}
	// Fill missing fields from the last known status so a
	// single control does not overwrite the other value.
	mon := true
	sens := strings.TrimSpace(body.Sensitivity)
	if st := mqttMgr.GetStatus(body.IP); st != nil {
		mon = st.AiMonitoring
		if sens == "" && st.Xcam != nil {
			if v, ok := st.Xcam["halt_print_sensitivity"].(string); ok {
				sens = v
			}
		}
	}
	if body.Monitoring != nil {
		mon = *body.Monitoring
	}
	if err := mqttMgr.SetXcam(*z, mon, sens); err != nil {
		log.Printf("XCAM-SET %s fehlgeschlagen: %v", body.IP, err)
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	log.Printf("XCAM-SET an %s: monitoring=%v sensitivity=%q (Best-Effort)", body.IP, mon, sens)
	writeJSON(w, map[string]any{"ok": true, "monitoring": mon, "sensitivity": sens})
}

// handleFwSchedule schedules a firmware update as a "job": if a print is running,
// the update is queued and starts automatically after the print (fwAutoLoop).
// If the printer is idle, the update is triggered immediately.
func handleFwSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP   string `json:"ip"`
		Auto bool   `json:"auto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if printerByIP(body.IP) == nil {
		http.Error(w, "Drucker nicht gefunden", http.StatusNotFound)
		return
	}
	if !body.Auto {
		mu.Lock()
		delete(state.PendingFwUpdate, body.IP)
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true, "scheduled": false})
		return
	}
	// Is a print running? Then queue it, otherwise start immediately.
	drucktGerade := false
	if st := mqttMgr.GetStatus(body.IP); st != nil {
		g := strings.ToUpper(st.GcodeState)
		drucktGerade = g == "RUNNING" || g == "PREPARE" || g == "PAUSE" || g == "PAUSED" || g == "SLICING"
	}
	if drucktGerade {
		mu.Lock()
		if state.PendingFwUpdate == nil {
			state.PendingFwUpdate = map[string]bool{}
		}
		state.PendingFwUpdate[body.IP] = true
		mu.Unlock()
		saveState()
		log.Printf("FW-UPDATE geplant fuer %s — startet nach Druckende", body.IP)
		writeJSON(w, map[string]any{"ok": true, "scheduled": true})
		return
	}
	// Frei -> sofort ausloesen.
	if err := sendUpgradeConfirm(body.IP); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	log.Printf("FW-UPDATE-START an %s (Drucker frei) gesendet", body.IP)
	writeJSON(w, map[string]any{"ok": true, "scheduled": false, "started": true})
}

// fwAutoLoop starts queued firmware updates as soon as the respective printer
// has finished printing (gcode_state FINISH/IDLE/FAILED).
func fwAutoLoop() {
	for {
		time.Sleep(30 * time.Second)
		mu.Lock()
		pending := make([]string, 0, len(state.PendingFwUpdate))
		for ip, an := range state.PendingFwUpdate {
			if an {
				pending = append(pending, ip)
			}
		}
		mu.Unlock()
		for _, ip := range pending {
			st := mqttMgr.GetStatus(ip)
			if st == nil || !st.Online {
				continue // offline: spaeter erneut versuchen
			}
			g := strings.ToUpper(st.GcodeState)
			fertig := g == "FINISH" || g == "IDLE" || g == "FAILED" || g == ""
			if !fertig {
				continue
			}
			if err := sendUpgradeConfirm(ip); err != nil {
				log.Printf("FW-UPDATE (auto) %s fehlgeschlagen: %v — bleibt geplant", ip, err)
				continue
			}
			log.Printf("FW-UPDATE (auto) an %s nach Druckende gesendet", ip)
			mu.Lock()
			delete(state.PendingFwUpdate, ip)
			mu.Unlock()
			saveState()
		}
	}
}
