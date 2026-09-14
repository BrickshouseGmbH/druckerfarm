package main

import (
	"strings"
	"testing"
)

// Antwort auf get_version, wie sie das Protokoll beschreibt.
func TestParseVersionReport(t *testing.T) {
	payload := []byte(`{"info":{"command":"get_version","sequence_id":"9001","module":[
		{"name":"ota","hw_ver":"","sw_ver":"01.08.02.00","sn":""},
		{"name":"rv1126","hw_ver":"AP05","sw_ver":"00.00.28.55","sn":"X"},
		{"name":"ams/0","hw_ver":"AMS08","sw_ver":"00.00.06.49","sn":"A1"},
		{"name":"ams/1","hw_ver":"AMS08","sw_ver":"00.00.06.49","sn":"A2"}]}}`)
	info, ok := parseVersionReport(payload)
	if !ok {
		t.Fatal("Antwort nicht erkannt")
	}
	if info.Firmware != "01.08.02.00" {
		t.Fatalf("Firmware falsch: %q", info.Firmware)
	}
	if len(info.AMS) != 2 {
		t.Fatalf("erwartet 2 AMS, bekommen %d", len(info.AMS))
	}
	if info.AMS[0].Name != "ams/0" || info.AMS[0].SW != "00.00.06.49" {
		t.Fatalf("erstes AMS falsch: %+v", info.AMS[0])
	}
	if len(info.Module) != 4 {
		t.Fatalf("alle Baugruppen sollten erhalten bleiben, sind %d", len(info.Module))
	}
}

// Statusmeldungen duerfen nicht als Versionsantwort durchgehen.
func TestParseVersionIgnoriertStatus(t *testing.T) {
	for _, p := range []string{
		`{"print":{"gcode_state":"RUNNING"}}`,
		`{"info":{"command":"push_status"}}`,
		`{"info":{"command":"get_version","module":[]}}`,
		`kein json`,
	} {
		if _, ok := parseVersionReport([]byte(p)); ok {
			t.Fatalf("faelschlich erkannt: %s", p)
		}
	}
}

// "ams/0" ist das erste AMS und heisst fuer den Anwender "AMS 1".
func TestAmsBezeichnung(t *testing.T) {
	f := map[string]string{"ams/0": "AMS 1", "ams/1": "AMS 2", "ams_2": "AMS 3", "ams": "AMS"}
	for ein, will := range f {
		if got := amsBezeichnung(ein); got != will {
			t.Fatalf("%q -> %q, erwartet %q", ein, got, will)
		}
	}
}

func TestLaufzeitText(t *testing.T) {
	f := map[int64]string{0: "", 90: "1 min", 3600: "1 h 0 min", 5400: "1 h 30 min", 3600 * 4000: "4000 h"}
	for sek, will := range f {
		if got := laufzeitText(sek); got != will {
			t.Fatalf("%d s -> %q, erwartet %q", sek, got, will)
		}
	}
	if !strings.Contains(laufzeitText(3600*150), "150 h") {
		t.Fatal("ab 100 Stunden ohne Minuten")
	}
}
