package main

import "testing"

func TestParseBblpURL(t *testing.T) {
	ip, code, ok := parseBblpURL("    - rtsps://bblp:abcd1234@192.168.1.50:322/streaming/live/1")
	if !ok || ip != "192.168.1.50" || code != "abcd1234" {
		t.Fatalf("got ip=%q code=%q ok=%v", ip, code, ok)
	}
	if _, _, ok := parseBblpURL("  somestream:"); ok {
		t.Fatalf("stream key line should not parse as URL")
	}
}

func TestParseYamlRows(t *testing.T) {
	y := "# go2rtc.yaml\n\napi:\n  origin: '*'\n\nstreams:\n" +
		"  drucker_a:\n    - rtsps://bblp:CODE1@192.168.1.10:322/streaming/live/1\n" +
		"    - rtspx://bblp:CODE1@192.168.1.10:322/streaming/live/0\n" +
		"  drucker_b:\n    - rtspx://bblp:CODE2@192.168.1.11:322/streaming/live/1\n"
	rows := parseImportRows(y)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d: %+v", len(rows), rows)
	}
	if rows[0].IP != "192.168.1.10" || rows[0].Code != "CODE1" || rows[0].Name != "drucker_a" {
		t.Errorf("row0 wrong: %+v", rows[0])
	}
	if rows[1].IP != "192.168.1.11" || rows[1].Code != "CODE2" {
		t.Errorf("row1 wrong: %+v", rows[1])
	}
}

func TestParseCSVRowsSkipsHeader(t *testing.T) {
	csv := "Name,IP,Code,Model,Serial\n" +
		"Drucker 1,192.168.1.20,pass0001,H2D,01P00ABC\n" +
		"Drucker 2;192.168.1.21;pass0002;X1C;01S00XYZ\n"
	rows := parseImportRows(csv)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows (header skipped), got %d: %+v", len(rows), rows)
	}
	if rows[0].IP != "192.168.1.20" || rows[0].Code != "pass0001" || rows[0].Model != "H2D" {
		t.Errorf("row0 wrong: %+v", rows[0])
	}
	if rows[1].IP != "192.168.1.21" || rows[1].Serial != "01S00XYZ" {
		t.Errorf("row1 (semicolon) wrong: %+v", rows[1])
	}
}
