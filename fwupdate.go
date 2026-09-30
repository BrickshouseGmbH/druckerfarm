package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── DRUCKER-FIRMWARE-UPDATES ─────────────────────────────────────────────────
//
// How the printer knows an update exists: "LAN only" cuts the
// cloud binding (account, remote control), not the network. The printer keeps
// checking with the vendor over its own channel whether new firmware
// is available, and puts the result into its status report (upgrade_state /
// new_ver_list). This program only reads it — it triggers NO update.
//
// Found updates are remembered locally and stay as a badge after the
// printer name until the printer has actually
// installed the new build. Cleanup happens at the next status poll: if
// the printer no longer reports the module or it already runs the new
// number, the badge disappears.

// UpgradeMessage is a single open update as it comes from the status.
type UpgradeMessage struct {
	Modul   string `json:"modul"`   // "ota", "ams/0" …
	Aktuell string `json:"aktuell"` // running build
	Neu     string `json:"neu"`     // available build
}

// FwModuleUpdate is a remembered update with the time it was found.
type FwModuleUpdate struct {
	Modul    string    `json:"modul"`
	Aktuell  string    `json:"aktuell"`
	Neu      string    `json:"neu"`
	Gefunden time.Time `json:"gefunden"`
}

// parseUpgradeState reads the upgrade_state block from the pushall.
func parseUpgradeState(v interface{}) []UpgradeMessage {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	var out []UpgradeMessage

	// Path 1 — the detailed list (newer firmware, P1/A1/H2 …). It names
	// the current and new number per module.
	if lst, ok := m["new_ver_list"].([]interface{}); ok {
		for _, e := range lst {
			em, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := em["name"].(string)
			cur, _ := em["cur_ver"].(string)
			neu, _ := em["new_ver"].(string)
			name = strings.TrimSpace(name)
			neu = strings.TrimSpace(neu)
			if name == "" || neu == "" {
				continue
			}
			// Only real upward jumps. Some firmware lists the
			// current build even when nothing is pending.
			if cur != "" && !newerVersion(cur, neu) {
				continue
			}
			out = append(out, UpgradeMessage{Modul: name, Aktuell: cur, Neu: neu})
		}
	}
	if len(out) > 0 {
		return out
	}

	// Path 2 — the X1 series (also X1E) has no new_ver_list but adds a
	// single field per module: ota_new_version_number, ams_new_version_number,
	// ahb_new_version_number, ext_new_version_number.
	//
	// Wichtig: new_version_state taugt NICHT als Schalter. Ein echter X1E meldet
	// new_version_state=1 UND trotzdem ota_new_version_number="01.03.00.00" — da
	// an update exists. The only thing that matters: is the number field filled
	// (and not a placeholder). The current build is not included here;
	// it is compared later against get_version (nochOffen) so already
	// installed items disappear again.
	feld := func(key, modul string) {
		val, _ := m[key].(string)
		val = strings.TrimSpace(val)
		if val == "" || istNullVersion(val) {
			return
		}
		out = append(out, UpgradeMessage{Modul: modul, Neu: val})
	}
	feld("ota_new_version_number", "ota")
	feld("ams_new_version_number", "ams")
	feld("ahb_new_version_number", "ahb")
	feld("ext_new_version_number", "ext")
	return out
}

// istNullVersion erkennt Platzhalter wie "00.00.00.00".
func istNullVersion(v string) bool {
	for _, r := range v {
		if r != '0' && r != '.' {
			return false
		}
	}
	return true
}

// aktuelleVersionen returns the current module levels from get_version.
func aktuelleVersionen(st *PrinterStatus) map[string]string {
	m := map[string]string{}
	if st == nil || st.Info == nil {
		return m
	}
	if st.Info.Firmware != "" {
		m["ota"] = st.Info.Firmware
	}
	for _, a := range st.Info.AMS {
		m[strings.ToLower(a.Name)] = a.SW
		// For the X1 message "ams_new_version_number" without index: the
		// lowest AMS level decides. If any AMS is below the
		// new number, the update counts as open.
		if cur, ok := m["ams"]; !ok || (a.SW != "" && newerVersion(a.SW, cur)) {
			m["ams"] = a.SW
		}
	}
	return m
}

// nochOffen removes from an update list everything already
// installed on the printer. If the status is unknown (offline), the list stays.
func nochOffen(list []FwModuleUpdate, st *PrinterStatus) []FwModuleUpdate {
	if st == nil {
		return list
	}
	aktuell := aktuelleVersionen(st)
	out := []FwModuleUpdate{}
	for _, u := range list {
		cur := aktuell[strings.ToLower(u.Modul)]
		if cur == "" {
			cur = u.Aktuell // fall back to the build seen at discovery
		}
		if newerVersion(cur, u.Neu) {
			out = append(out, u)
		}
	}
	return out
}

// fwForDisplay returns the open updates per printer for display and clears
// installed ones silently (persists only when something changed).
// Called at every status poll — this is the "compare on load".
func fwForDisplay() map[string][]FwModuleUpdate {
	mu.Lock()
	defer mu.Unlock()
	res := map[string][]FwModuleUpdate{}
	if state.FwUpdates == nil {
		return res
	}
	dirty := false
	for ip, list := range state.FwUpdates {
		st := mqttMgr.GetStatus(ip)
		offen := nochOffen(list, st)
		if st != nil && st.Online && len(offen) != len(list) {
			// The printer installed (at least) one of the updates.
			if len(offen) == 0 {
				delete(state.FwUpdates, ip)
			} else {
				state.FwUpdates[ip] = offen
			}
			dirty = true
		}
		if len(offen) > 0 {
			res[ip] = offen
		}
	}
	if dirty {
		go saveState()
	}
	return res
}

// handleFwUpdateScan freshly polls every connected printer (pushall) and
// remembers which module has an update pending. Triggers NOTHING.
func handleFwUpdateScan(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	angefragt := 0
	for _, p := range printers {
		// pushall brings upgrade_state, get_version the current firmware —
		// only with both can it cleanly compare what is already installed.
		// get_version needs the serial number.
		ok := mqttMgr.RequestStatus(p)
		mqttMgr.RequestVersion(p)
		if ok {
			angefragt++
		}
	}
	// Wait briefly until the responses have arrived.
	time.Sleep(2500 * time.Millisecond)

	geprueft, mitUpdate := scanMergeFw(printers)
	saveState()

	// Include raw upgrade_state messages — for diagnostics, in case a model
	// seine Update-Info anders aufbaut als erwartet.
	roh := map[string]string{}
	mqttMgr.mu.RLock()
	for _, p := range printers {
		if r, ok := mqttMgr.lastUpgrade[p.IP]; ok && r != "" {
			roh[p.IP] = r
		}
	}
	mqttMgr.mu.RUnlock()

	log.Printf("FW-UPDATE-SCAN: %d angefragt, %d geprueft, %d mit Update", angefragt, geprueft, mitUpdate)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"angefragt":  angefragt,
		"geprueft":   geprueft,
		"mit_update": mitUpdate,
		"roh":        roh,
	})
}

// handleFwUpdateStart triggers the firmware update on the printer. IMPORTANT: This
// is deliberately risky and not officially documented — it works
// only with Developer Mode enabled and is best-effort. The UI warns
// clearly beforehand; here only the command is issued.
func handleFwUpdateStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	var ziel *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == body.IP {
			ziel = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if ziel == nil {
		http.Error(w, "Drucker nicht gefunden", http.StatusNotFound)
		return
	}

	// Best-known form from the community: confirm the pending update.
	// No module/URL — the printer knows the available update itself.
	payload := `{"upgrade":{"sequence_id":"` + nextSeq() + `","command":"upgrade_confirm","src_id":1}}`
	if err := mqttMgr.PublishRaw(*ziel, payload); err != nil {
		log.Printf("FW-UPDATE-START %s fehlgeschlagen: %v", body.IP, err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	log.Printf("FW-UPDATE-START an %s gesendet (Best-Effort, Developer Mode noetig)", body.IP)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// scanMergeFw merges the fresh upgrade_state messages into memory.
func scanMergeFw(printers []Printer) (geprueft, mitUpdate int) {
	mu.Lock()
	defer mu.Unlock()
	if state.FwUpdates == nil {
		state.FwUpdates = map[string][]FwModuleUpdate{}
	}
	for _, p := range printers {
		st := mqttMgr.GetStatus(p.IP)
		if st == nil || !st.Online {
			continue // offline: do not touch the old state
		}
		geprueft++
		frisch := []FwModuleUpdate{}
		for _, u := range st.Upgrade {
			g := time.Now()
			for _, a := range state.FwUpdates[p.IP] {
				if a.Modul == u.Modul && a.Neu == u.Neu {
					g = a.Gefunden // Fundzeitpunkt bewahren
				}
			}
			frisch = append(frisch, FwModuleUpdate{Modul: u.Modul, Aktuell: u.Aktuell, Neu: u.Neu, Gefunden: g})
		}
		frisch = nochOffen(frisch, st)
		if len(frisch) == 0 {
			delete(state.FwUpdates, p.IP)
		} else {
			state.FwUpdates[p.IP] = frisch
			mitUpdate++
		}
	}
	return
}
