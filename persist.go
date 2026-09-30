package main

import (
	"encoding/json"
	"os"
)

// ─── BETRIEBSDATEN GETRENNT SPEICHERN ─────────────────────────────────────────
//
// config.json should hold only settings (printers, cameras, theme,
// language, filters …). Operational and history data do not belong there and
// used to bloat the file heavily. They now live separately:
//
//   upload_log.json  — the complete upload history (no limit)
//   runtime.json     — runtime per printer, repair markers, open
//                      firmware updates (small, constantly updated caches)
//
// On start these files are loaded; if missing, the values are migrated once
// from an old config.json and config.json is then rewritten cleanly
// without these fields.

// runtimeData bundles the small runtime caches for runtime.json.
type runtimeData struct {
	Laufzeit  map[string]int64            `json:"laufzeit,omitempty"`
	Reparatur map[string]RepairFlag       `json:"reparatur,omitempty"`
	FwUpdates map[string][]FwModuleUpdate `json:"fw_updates,omitempty"`
}

// legacyConfigOps reads the four operational fields from an old config.json where
// they were still stored (for the one-time migration).
type legacyConfigOps struct {
	UploadLog []UploadEntry               `json:"upload_log"`
	Laufzeit  map[string]int64            `json:"laufzeit"`
	Reparatur map[string]RepairFlag       `json:"reparatur"`
	FwUpdates map[string][]FwModuleUpdate `json:"fw_updates"`
}

// saveUploadLog writes the upload history to its own file.
func saveUploadLog() {
	if uploadLogFile == "" {
		return
	}
	mu.RLock()
	liste := make([]UploadEntry, len(state.UploadLog))
	copy(liste, state.UploadLog)
	mu.RUnlock()
	data, err := json.MarshalIndent(liste, "", "  ")
	if err != nil {
		return
	}
	atomicWrite(uploadLogFile, data, 0644)
}

// saveRuntime writes the small runtime caches to runtime.json.
func saveRuntime() {
	if runtimeFile == "" {
		return
	}
	mu.RLock()
	rt := runtimeData{Laufzeit: state.Laufzeit, Reparatur: state.Reparatur, FwUpdates: state.FwUpdates}
	data, _ := json.MarshalIndent(rt, "", "  ")
	mu.RUnlock()
	atomicWrite(runtimeFile, data, 0644)
}

// loadOrMigrateOps loads upload_log.json and runtime.json into memory. If one
// of the files is missing, the values are migrated from the (possibly old)
// config.json and the new files are created. configRaw is the raw config.json.
func loadOrMigrateOps(configRaw []byte) {
	var alt legacyConfigOps
	haltAlt := len(configRaw) > 0 && json.Unmarshal(configRaw, &alt) == nil
	migriert := false

	// Upload-Historie
	if b, err := os.ReadFile(uploadLogFile); err == nil {
		var ul []UploadEntry
		if json.Unmarshal(b, &ul) == nil {
			mu.Lock()
			state.UploadLog = ul
			mu.Unlock()
		}
	} else if haltAlt && len(alt.UploadLog) > 0 {
		mu.Lock()
		state.UploadLog = alt.UploadLog
		mu.Unlock()
		saveUploadLog()
		migriert = true
	}

	// Laufzeit / Reparatur / FwUpdates
	if b, err := os.ReadFile(runtimeFile); err == nil {
		var rt runtimeData
		if json.Unmarshal(b, &rt) == nil {
			mu.Lock()
			state.Laufzeit = rt.Laufzeit
			state.Reparatur = rt.Reparatur
			state.FwUpdates = rt.FwUpdates
			mu.Unlock()
		}
	} else if haltAlt && (alt.Laufzeit != nil || alt.Reparatur != nil || alt.FwUpdates != nil) {
		mu.Lock()
		state.Laufzeit = alt.Laufzeit
		state.Reparatur = alt.Reparatur
		state.FwUpdates = alt.FwUpdates
		mu.Unlock()
		saveRuntime()
		migriert = true
	}

	// If something was taken from an old config.json, rewrite config.json once
	// cleanly without the operational fields.
	if migriert {
		mu.Lock()
		data, _ := json.MarshalIndent(state, "", "  ")
		mu.Unlock()
		atomicWrite(dataFile, data, 0644)
	}
}
