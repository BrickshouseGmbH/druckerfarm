package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Die Nachricht muss genau die Form haben, die das Protokoll verlangt —
// sonst passiert dasselbe wie bei pause: der Drucker nimmt sie und tut nichts.
func TestProjectFilePayloadForm(t *testing.T) {
	p, err := projectFilePayload(printStartReq{IP: "10.0.0.1", Datei: "teil.3mf", UseAMS: true, Fach: 2}, "9001")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid([]byte(p)) {
		t.Fatalf("kein gueltiges JSON: %s", p)
	}
	var w struct {
		Print map[string]any `json:"print"`
	}
	if err := json.Unmarshal([]byte(p), &w); err != nil {
		t.Fatal(err)
	}
	for _, feld := range []string{"sequence_id", "command", "param", "project_id", "profile_id",
		"task_id", "subtask_id", "url", "ams_mapping", "use_ams", "bed_type"} {
		if _, ok := w.Print[feld]; !ok {
			t.Fatalf("Pflichtfeld %q fehlt", feld)
		}
	}
	if w.Print["command"] != "project_file" {
		t.Fatalf("falscher Befehl: %v", w.Print["command"])
	}
	// Fuer oertliche Auftraege sind die Kennungen laut Protokoll "0".
	for _, feld := range []string{"project_id", "profile_id", "task_id", "subtask_id"} {
		if w.Print[feld] != "0" {
			t.Fatalf("%s muss \"0\" sein, ist %v", feld, w.Print[feld])
		}
	}
	if w.Print["url"] != "ftp:///teil.3mf" {
		t.Fatalf("falsche URL: %v", w.Print["url"])
	}
}

// Die Farbzuordnung wird von hinten gefuellt — bei einer Farbe steht das Fach
// ganz am Ende. Steht es vorn, faengt der Drucker gar nicht erst an.
func TestAmsMappingVonHinten(t *testing.T) {
	cases := map[int]string{
		0:  "[-1,-1,-1,-1,0]",
		2:  "[-1,-1,-1,-1,2]",
		-1: "[-1,-1,-1,-1,-1]",
	}
	for fach, will := range cases {
		if got := amsMapping(fach); got != will {
			t.Fatalf("Fach %d: bekommen %s, erwartet %s", fach, got, will)
		}
	}
}

// Ohne AMS darf keine Zuordnung mitgeschickt werden.
func TestOhneAmsKeineZuordnung(t *testing.T) {
	p, err := projectFilePayload(printStartReq{Datei: "a.3mf", UseAMS: false}, "9002")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p, `"ams_mapping":[]`) {
		t.Fatalf("Zuordnung sollte leer sein: %s", p)
	}
	if !strings.Contains(p, `"use_ams":false`) {
		t.Fatalf("use_ams sollte false sein: %s", p)
	}
}

// Dateinamen aus der Liste, nicht aus der Fantasie: Pfadtrenner und
// Rueckwaertsschritte werden abgewiesen, bevor irgendetwas gesendet wird.
func TestPrintStartWeistUnsinnAb(t *testing.T) {
	schlecht := []string{"", "   ", "../etc/passwd", "ordner/teil.3mf", `ordner\teil.3mf`,
		"teil.txt", "teil.stl"}
	for _, d := range schlecht {
		if _, err := projectFilePayload(printStartReq{Datei: d}, "9003"); err == nil {
			t.Fatalf("%q wurde durchgelassen", d)
		}
	}
	for _, d := range []string{"teil.3mf", "TEIL.3MF", "modell.gcode"} {
		if _, err := projectFilePayload(printStartReq{Datei: d}, "9004"); err != nil {
			t.Fatalf("%q wurde faelschlich abgewiesen: %v", d, err)
		}
	}
}

// Der Fehler 0x07FF8012 kam daher, dass eine rohe .gcode-Datei mit dem
// project_file-Befehl und dem 3mf-internen Pfad gestartet wurde. Fuer .gcode
// muss es gcode_file mit dem Dateinamen sein.
func TestGcodeNutztGcodeFile(t *testing.T) {
	p, err := projectFilePayload(printStartReq{Datei: "modell.gcode"}, "9001")
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Print map[string]any `json:"print"`
	}
	if err := json.Unmarshal([]byte(p), &w); err != nil {
		t.Fatal(err)
	}
	if w.Print["command"] != "gcode_file" {
		t.Fatalf("erwartet gcode_file, bekommen %v", w.Print["command"])
	}
	if w.Print["param"] != "modell.gcode" {
		t.Fatalf("falscher Pfad: %v", w.Print["param"])
	}
	if _, hat := w.Print["ams_mapping"]; hat {
		t.Fatal("eine rohe .gcode braucht keine ams_mapping")
	}
}

// Eine .3mf bleibt beim project_file-Befehl.
func Test3mfNutztProjectFile(t *testing.T) {
	p, _ := projectFilePayload(printStartReq{Datei: "teil.3mf"}, "9002")
	var w struct {
		Print map[string]any `json:"print"`
	}
	json.Unmarshal([]byte(p), &w)
	if w.Print["command"] != "project_file" {
		t.Fatalf("erwartet project_file, bekommen %v", w.Print["command"])
	}
	if w.Print["url"] != "ftp:///teil.3mf" {
		t.Fatalf("falsche URL: %v", w.Print["url"])
	}
}

// Mehrfarbige Zuordnung: die Farben werden der Reihe nach den Faechern
// zugeordnet, "nicht zuordnen" wird zu -1.
func TestMehrfarbZuordnung(t *testing.T) {
	p, err := projectFilePayload(printStartReq{Datei: "bunt.3mf", Mapping: []int{4, -1, 6}}, "9010")
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Print map[string]any `json:"print"`
	}
	json.Unmarshal([]byte(p), &w)
	roh, _ := json.Marshal(w.Print["ams_mapping"])
	if string(roh) != "[4,-1,6]" {
		t.Fatalf("Zuordnung falsch: %s", roh)
	}
	if w.Print["use_ams"] != true {
		t.Fatalf("use_ams sollte true sein: %v", w.Print["use_ams"])
	}
}
