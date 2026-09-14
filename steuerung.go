package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ─── TEMPERATUR UND FILAMENT STEUERN ──────────────────────────────────────────
//
// Beide Wege laufen ueber dieselbe MQTT-Quittung wie Pause/Weiter/Stop. Ohne
// Developer Mode am Geraet lehnt die Firmware sie mit "mqtt message verify
// failed" ab — die Meldung sagt das dann und verweist auf den Developer Mode.

const (
	nozzleMax = 300
	bedMax    = 120
)

// findePrinter sucht einen Drucker anhand der IP. Gibt eine Kopie zurueck,
// damit der Aufrufer nicht unter der Sperre arbeiten muss.
func findePrinter(ip string) (Printer, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, p := range state.Printers {
		if p.IP == ip {
			return p, true
		}
	}
	return Printer{}, false
}

type tempReq struct {
	IP   string `json:"ip"`
	Was  string `json:"was"` // "nozzle" oder "bed"
	Temp int    `json:"temp"`
}

// tempGcode baut die G-Code-Zeile. M104 heizt die Duese, M140 das Bett; beide
// setzen nur den Sollwert und warten nicht (kein M109/M190), damit die
// Oberflaeche nicht blockiert.
func tempGcode(was string, temp int) (string, error) {
	switch was {
	case "nozzle":
		if temp < 0 || temp > nozzleMax {
			return "", fmt.Errorf("Düsentemperatur muss zwischen 0 und %d°C liegen", nozzleMax)
		}
		return fmt.Sprintf("M104 S%d", temp), nil
	case "bed":
		if temp < 0 || temp > bedMax {
			return "", fmt.Errorf("Betttemperatur muss zwischen 0 und %d°C liegen", bedMax)
		}
		return fmt.Sprintf("M140 S%d", temp), nil
	}
	return "", fmt.Errorf("unbekanntes Ziel %q", was)
}

func gcodePayload(line, seq string) string {
	m := map[string]any{
		"print": map[string]any{
			"sequence_id": seq,
			"command":     "gcode_line",
			"param":       line + "\n",
		},
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func handleSetTemp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body tempReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	line, err := tempGcode(body.Was, body.Temp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, ok := findePrinter(body.IP)
	if !ok {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	seq := naechsteSeq()
	if err := mqttMgr.SendRaw(p, "gcode_line", seq, gcodePayload(line, seq)); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "gesendet": line})
}

type filamentReq struct {
	IP        string `json:"ip"`
	AmsID     int    `json:"ams_id"`
	TrayID    int    `json:"tray_id"`
	Type      string `json:"tray_type"`     // PLA, PETG, ABS …
	Color     string `json:"tray_color"`    // RRGGBB oder RRGGBBAA
	InfoIdx   string `json:"tray_info_idx"` // Profil-Kennung, kodiert Marke+Typ (z. B. GFA00)
	NozzleMin int    `json:"nozzle_temp_min"`
	NozzleMax int    `json:"nozzle_temp_max"`
}

// normFarbe macht aus einer Farbe die vom Drucker erwartete Form RRGGBBAA.
func normFarbe(c string) string {
	c = strings.TrimPrefix(strings.TrimSpace(c), "#")
	c = strings.ToUpper(c)
	switch len(c) {
	case 6:
		return c + "FF"
	case 8:
		return c
	default:
		return "00000000"
	}
}

func filamentPayload(f filamentReq, seq string) (string, error) {
	if f.AmsID < 0 || f.AmsID > 3 || f.TrayID < 0 || f.TrayID > 3 {
		return "", fmt.Errorf("AMS oder Fach außerhalb 0–3")
	}
	if f.Type == "" {
		return "", fmt.Errorf("kein Materialtyp angegeben")
	}
	m := map[string]any{
		"print": map[string]any{
			"sequence_id":     seq,
			"command":         "ams_filament_setting",
			"ams_id":          f.AmsID,
			"tray_id":         f.TrayID,
			"tray_info_idx":   f.InfoIdx,
			"tray_color":      normFarbe(f.Color),
			"nozzle_temp_min": f.NozzleMin,
			"nozzle_temp_max": f.NozzleMax,
			"tray_type":       strings.ToUpper(f.Type),
		},
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func handleSetFilament(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body filamentReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	seq := naechsteSeq()
	payload, err := filamentPayload(body, seq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, ok := findePrinter(body.IP)
	if !ok {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	if err := mqttMgr.SendRaw(p, "ams_filament_setting", seq, payload); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
