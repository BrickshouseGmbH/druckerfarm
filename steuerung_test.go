package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTempGcode(t *testing.T) {
	f := []struct {
		was    string
		temp   int
		will   string
		fehler bool
	}{
		{"nozzle", 220, "M104 S220", false},
		{"bed", 60, "M140 S60", false},
		{"nozzle", 0, "M104 S0", false},
		{"nozzle", 350, "", true}, // ueber Grenze
		{"bed", 130, "", true},
		{"nozzle", -5, "", true},
		{"kammer", 50, "", true}, // unbekannt
	}
	for _, c := range f {
		g, err := tempGcode(c.was, c.temp)
		if c.fehler && err == nil {
			t.Fatalf("%s %d: Fehler erwartet", c.was, c.temp)
		}
		if !c.fehler {
			if err != nil {
				t.Fatalf("%s %d: %v", c.was, c.temp, err)
			}
			if g != c.will {
				t.Fatalf("%s %d -> %q, erwartet %q", c.was, c.temp, g, c.will)
			}
		}
	}
}

func TestGcodePayloadGueltig(t *testing.T) {
	p := gcodePayload("M104 S210", "9001")
	if !json.Valid([]byte(p)) {
		t.Fatalf("kein gueltiges JSON: %s", p)
	}
	if !strings.Contains(p, `"command":"gcode_line"`) {
		t.Fatalf("falscher Befehl: %s", p)
	}
	if !strings.Contains(p, `M104 S210`) {
		t.Fatalf("G-Code fehlt: %s", p)
	}
}

func TestNormFarbe(t *testing.T) {
	f := map[string]string{
		"FF0000":   "FF0000FF",
		"#00ff00":  "00FF00FF",
		"0000FFFF": "0000FFFF",
		"":         "00000000",
		"xyz":      "00000000",
	}
	for ein, will := range f {
		if got := normFarbe(ein); got != will {
			t.Fatalf("%q -> %q, erwartet %q", ein, got, will)
		}
	}
}

func TestFilamentPayload(t *testing.T) {
	p, err := filamentPayload(filamentReq{
		AmsID: 1, TrayID: 2, Type: "petg", Color: "00FF00",
		InfoIdx: "GFG00", NozzleMin: 220, NozzleMax: 260}, "9002")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid([]byte(p)) {
		t.Fatalf("kein gueltiges JSON: %s", p)
	}
	var w struct {
		Print map[string]any `json:"print"`
	}
	json.Unmarshal([]byte(p), &w)
	for _, feld := range []string{"ams_id", "tray_id", "tray_info_idx", "tray_color", "nozzle_temp_min", "nozzle_temp_max", "tray_type", "command"} {
		if _, ok := w.Print[feld]; !ok {
			t.Fatalf("Feld %q fehlt", feld)
		}
	}
	if w.Print["tray_type"] != "PETG" {
		t.Fatalf("Typ nicht gross geschrieben: %v", w.Print["tray_type"])
	}
	if w.Print["tray_color"] != "00FF00FF" {
		t.Fatalf("Farbe falsch: %v", w.Print["tray_color"])
	}
}

func TestFilamentPayloadWeistUnsinnAb(t *testing.T) {
	for _, f := range []filamentReq{
		{AmsID: 5, TrayID: 0, Type: "PLA"},
		{AmsID: 0, TrayID: 9, Type: "PLA"},
		{AmsID: 0, TrayID: 0, Type: ""},
	} {
		if _, err := filamentPayload(f, "9003"); err == nil {
			t.Fatalf("durchgelassen: %+v", f)
		}
	}
}
