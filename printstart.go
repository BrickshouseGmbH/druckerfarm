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
// Der Befehl heisst project_file und erwartet einen Pfad auf dem Geraet. Fuer
// oertliche Auftraege sind project_id, profile_id, task_id und subtask_id
// jeweils "0"; das ist so vorgesehen und nicht etwa ein Platzhalter.
//
// Zur Farbzuordnung: ams_mapping ordnet den Farben der Datei die Faecher zu und
// wird von hinten gefuellt — bei einer Farbe also [-1,-1,-1,-1,fach]. Wie viele
// Farben eine Datei braucht, steht allerdings in der Datei selbst (im 3MF), und
// die liegt auf dem Drucker. Deshalb wird hier nur die einfarbige Zuordnung
// angeboten; alles andere waere geraten.

type printStartReq struct {
	IP        string `json:"ip"`
	Datei     string `json:"datei"`
	UseAMS    bool   `json:"use_ams"`
	Fach      int    `json:"fach"`              // AMS-Fach global, -1 = externe Rolle
	Mapping   []int  `json:"mapping,omitempty"` // Farbe->Fach, -1 = nicht zuordnen
	Timelapse bool   `json:"timelapse"`
	Leveling  bool   `json:"leveling"`
	FlowCali  bool   `json:"flow_cali"`
}

// amsMapping baut die Zuordnungsliste fuer einen einfarbigen Auftrag.
// amsMappingListe baut die Zuordnung aus mehreren Farben. Die Liste hat feste
// Laenge 16 (mehr Faecher gibt es nicht), links mit -1 aufgefuellt.
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
	// Pfadtrenner und Rueckwaertsschritte sind hier so wenig erwuenscht wie
	// beim Loeschen — der Name kommt aus einer Liste, nicht aus der Fantasie.
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
		// Von den Farben nur die tatsaechlich zugeordneten behalten; das
		// Protokoll fuellt die Liste rechtsbuendig und links mit -1 auf.
		mapping = amsMappingListe(r.Mapping)
		for _, v := range r.Mapping {
			if v >= 0 {
				useAMS = true
			}
		}
	} else if r.UseAMS {
		mapping = amsMapping(r.Fach)
	}

	// Zwei verschiedene Befehle je nach Dateiart. Das war die Ursache des
	// Fehlers 0x07FF8012: eine rohe .gcode-Datei hat keinen internen Pfad
	// "Metadata/plate_1.gcode" — der steckt nur in einem .3mf-Archiv. Fuer
	// .gcode ist der Befehl "gcode_file" mit dem Dateinamen der richtige.
	istGcode := strings.HasSuffix(strings.ToLower(datei), ".gcode")

	if istGcode {
		payload := map[string]any{
			"print": map[string]any{
				"sequence_id": seq,
				"command":     "gcode_file",
				// Laut Protokoll nur der Dateiname, KEIN fuehrender Schraegstrich.
				// Der Slash war die Ursache des Fehlers "unsupported path".
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

	// Niemals einen laufenden Auftrag ueberschreiben.
	if s := mqttMgr.GetStatus(kopie.IP); s != nil {
		switch strings.ToUpper(s.GcodeState) {
		case "RUNNING", "PAUSE", "PREPARE", "SLICING":
			http.Error(w, "Der Drucker arbeitet gerade — erst abbrechen oder abwarten",
				http.StatusConflict)
			return
		}
	}

	seq := naechsteSeq()
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
