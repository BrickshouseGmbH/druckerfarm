package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ─── DRUCK STARTEN ────────────────────────────────────────────────────────────
//
// The command is project_file and expects a path on the device. For
// local jobs project_id, profile_id, task_id and subtask_id
// are each "0"; that is intended and not a placeholder.
//
// On color mapping: ams_mapping maps the file's colors to the trays and
// is filled from the back — for one color [-1,-1,-1,-1,tray]. How many
// colors a file needs is however inside the file itself (in the 3MF), and
// that lives on the printer. So only the single-color mapping is
// offered; anything else would be guessing.

type printStartReq struct {
	IP        string `json:"ip"`
	Datei     string `json:"datei"`
	UseAMS    bool   `json:"use_ams"`
	Fach      int    `json:"fach"`              // AMS-Fach global, -1 = externe Rolle
	Mapping   []int  `json:"mapping,omitempty"` // color->tray, -1 = do not map
	Timelapse bool   `json:"timelapse"`
	Leveling  bool   `json:"leveling"`
	FlowCali  bool   `json:"flow_cali"`
}

// amsMapping builds the mapping list for a single-color job.
// amsMappingListe builds the mapping from several colors. The list has fixed
// length 16 (there are no more trays), padded with -1 on the left.
func amsMappingListe(faecher []int) string {
	m := make([]int, 0, len(faecher))
	for _, f := range faecher {
		if f < 0 {
			m = append(m, -1)
		} else {
			m = append(m, f)
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func amsMapping(fach int) string {
	m := []int{-1, -1, -1, -1, fach}
	b, _ := json.Marshal(m)
	return string(b)
}

func projectFilePayload(r printStartReq, seq string) (string, error) {
	datei := strings.TrimSpace(r.Datei)
	if datei == "" {
		return "", fmt.Errorf("keine Datei angegeben")
	}
	// Path separators and backtracking are as unwanted here as
	// when deleting — the name comes from a list, not from imagination.
	if strings.ContainsAny(datei, `/\`) || strings.Contains(datei, "..") {
		return "", fmt.Errorf("unzulässiger Dateiname %q", datei)
	}
	if !strings.HasSuffix(strings.ToLower(datei), ".3mf") &&
		!strings.HasSuffix(strings.ToLower(datei), ".gcode") {
		return "", fmt.Errorf("nur .3mf und .gcode können gedruckt werden")
	}

	mapping := "[]"
	useAMS := r.UseAMS
	if len(r.Mapping) > 0 {
		// Keep only the actually mapped colors; the
		// protocol fills the list right-aligned and pads left with -1.
		mapping = amsMappingListe(r.Mapping)
		for _, v := range r.Mapping {
			if v >= 0 {
				useAMS = true
			}
		}
	} else if r.UseAMS {
		mapping = amsMapping(r.Fach)
	}

	// Two different commands depending on file type. This caused
	// error 0x07FF8012: a raw .gcode file has no internal path
	// "Metadata/plate_1.gcode" — that only exists in a .3mf archive. For
	// .gcode the correct command is "gcode_file" with the file name.
	istGcode := strings.HasSuffix(strings.ToLower(datei), ".gcode")

	if istGcode {
		payload := map[string]any{
			"print": map[string]any{
				"sequence_id": seq,
				"command":     "gcode_file",
				// Per protocol only the file name, NO leading slash.
				// The slash caused the "unsupported path" error.
				"param": datei,
			},
		}
		b, err := json.Marshal(payload)
		return string(b), err
	}

	payload := map[string]any{
		"print": map[string]any{
			"sequence_id":    seq,
			"command":        "project_file",
			"param":          "Metadata/plate_1.gcode",
			"project_id":     "0",
			"profile_id":     "0",
			"task_id":        "0",
			"subtask_id":     "0",
			"subtask_name":   strings.TrimSuffix(datei, ".3mf"),
			"file":           "",
			"url":            "ftp:///" + datei,
			"md5":            "",
			"timelapse":      r.Timelapse,
			"bed_type":       "auto",
			"bed_levelling":  r.Leveling,
			"flow_cali":      r.FlowCali,
			"vibration_cali": true,
			"layer_inspect":  true,
			"ams_mapping":    json.RawMessage(mapping),
			"use_ams":        useAMS,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func handlePrintStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body printStartReq
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
	var kopie Printer
	if ziel != nil {
		kopie = *ziel
	}
	mu.Unlock()
	if ziel == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	if kopie.Serial == "" {
		http.Error(w, "keine Seriennummer hinterlegt", http.StatusBadRequest)
		return
	}

	// Never overwrite a running job.
	if s := mqttMgr.GetStatus(kopie.IP); s != nil {
		switch strings.ToUpper(s.GcodeState) {
		case "RUNNING", "PAUSE", "PREPARE", "SLICING":
			http.Error(w, "Der Drucker arbeitet gerade — erst abbrechen oder abwarten",
				http.StatusConflict)
			return
		}
	}

	seq := nextSeq()
	payload, err := projectFilePayload(body, seq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := mqttMgr.SendRaw(kopie, "project_file", seq, payload); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "datei": body.Datei, "gestartet": time.Now()})
}
