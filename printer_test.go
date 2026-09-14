package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func parsePrint(t *testing.T, body string) *PrinterStatus {
	t.Helper()
	ip := "10.9.9.9"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{}
	mqttMgr.mu.Unlock()
	t.Cleanup(func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	})
	mqttMgr.handleMessage(ip, []byte(body))
	return mqttMgr.GetStatus(ip)
}

// the printer sends partial updates: an absent key means "unchanged", a present key
// with value 0 means really 0. The old code dropped every zero.
func TestZeroValuesAreStored(t *testing.T) {
	s := parsePrint(t, `{"print":{"mc_percent":0,"mc_remaining_time":0,"layer_num":0,"gcode_state":"IDLE"}}`)
	if s.Progress != 0 || s.RemainTime != 0 {
		t.Fatalf("zeros not stored: %+v", s)
	}

	s2 := parsePrint(t, `{"print":{"mc_percent":42,"mc_remaining_time":90}}`)
	if s2.Progress != 42 || s2.RemainTime != 90 {
		t.Fatalf("values not stored: %+v", s2)
	}
}

// A finished job reporting 0% must not keep showing the previous job's value.
func TestProgressResetsToZero(t *testing.T) {
	ip := "10.9.9.9"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{Progress: 87, RemainTime: 40}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	mqttMgr.handleMessage(ip, []byte(`{"print":{"mc_percent":0,"mc_remaining_time":0}}`))
	s := mqttMgr.GetStatus(ip)
	if s.Progress != 0 || s.RemainTime != 0 {
		t.Fatalf("stale values kept: %+v", s)
	}
}

// Keys that are absent must not clobber what we already know.
func TestAbsentKeysKeepPreviousValue(t *testing.T) {
	ip := "10.9.9.9"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{Progress: 55, SubtaskName: "teil.3mf"}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	mqttMgr.handleMessage(ip, []byte(`{"print":{"nozzle_temper":215.0}}`))
	s := mqttMgr.GetStatus(ip)
	if s.Progress != 55 || s.SubtaskName != "teil.3mf" {
		t.Fatalf("previous values lost: %+v", s)
	}
	if s.NozzleTemp != 215 {
		t.Fatalf("new value missing: %+v", s)
	}
}

const amsPayload = `{"print":{"ams":{"ams":[{"id":"0","humidity":"4","temp":"28.5","tray":[
  {"id":"0","tray_type":"PLA","tray_color":"FF6A13FF","tray_sub_brands":"PLA Basic","remain":72},
  {"id":"1","tray_type":"PETG","tray_color":"0A5EB0FF","remain":30},
  {"id":"2"},
  {"id":"3","tray_type":"","tray_color":"00000000"}]},
  {"id":"1","humidity":"5","temp":"27.0","tray":[{"id":"0","tray_type":"ABS","tray_color":"1A1A1AFF","remain":100}]}],
  "tray_now":"1","ams_exist_bits":"3"},
  "vt_tray":{"id":"254","tray_type":"PLA","tray_color":"FFFFFFFF","tray_sub_brands":"PLA Matte"}}}`

func TestAMSParsing(t *testing.T) {
	s := parsePrint(t, amsPayload)

	if len(s.AMS) != 2 {
		t.Fatalf("want 2 AMS units, got %d", len(s.AMS))
	}
	if s.TrayNow != "1" {
		t.Fatalf("tray_now = %q", s.TrayNow)
	}

	u0 := s.AMS[0]
	if u0.Humidity != "4" || u0.Temp != "28.5" {
		t.Fatalf("unit meta wrong: %+v", u0)
	}
	if len(u0.Trays) != 4 {
		t.Fatalf("want 4 trays, got %d", len(u0.Trays))
	}
	if u0.Trays[0].Type != "PLA" || u0.Trays[0].Color != "FF6A13FF" || u0.Trays[0].SubBrand != "PLA Basic" || u0.Trays[0].Remain != 72 {
		t.Fatalf("tray 0 wrong: %+v", u0.Trays[0])
	}
	// empty slot without any fields
	if u0.Trays[2].Type != "" || u0.Trays[2].Color != "" {
		t.Fatalf("empty tray should stay empty: %+v", u0.Trays[2])
	}
	// 00000000 is the colour of an empty slot and must not become black
	if u0.Trays[3].Color != "" {
		t.Fatalf("00000000 should be treated as unknown: %+v", u0.Trays[3])
	}
	if u0.Trays[1].Remain != 30 {
		t.Fatalf("remain wrong: %+v", u0.Trays[1])
	}
	if s.AMS[1].Trays[0].Type != "ABS" {
		t.Fatalf("second unit wrong: %+v", s.AMS[1])
	}
	if s.ExtSpool == nil || s.ExtSpool.SubBrand != "PLA Matte" {
		t.Fatalf("external spool missing: %+v", s.ExtSpool)
	}
}

// AMS data only arrives now and then — a message without it must not wipe it.
func TestAMSSurvivesMessagesWithoutIt(t *testing.T) {
	ip := "10.9.9.9"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	mqttMgr.handleMessage(ip, []byte(amsPayload))
	mqttMgr.handleMessage(ip, []byte(`{"print":{"mc_percent":10}}`))

	s := mqttMgr.GetStatus(ip)
	if len(s.AMS) != 2 {
		t.Fatalf("AMS was wiped by an unrelated message: %+v", s.AMS)
	}
}

func TestAMSJSONReachesTheUI(t *testing.T) {
	s := parsePrint(t, amsPayload)
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"ams"`, `"color":"FF6A13FF"`, `"sub_brand":"PLA Basic"`, `"tray_now":"1"`, `"ext_spool"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s missing from JSON: %s", want, b)
		}
	}
}

// ─── Drucksteuerung ───────────────────────────────────────────────────────────

func postCmd(body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handlePrintCommand(rec, httptest.NewRequest("POST", "/api/print/command", strings.NewReader(body)))
	return rec
}

func TestPrintCommandRejectsBadInput(t *testing.T) {
	if rec := postCmd(`{"ips":["1.2.3.4"],"command":"selbstzerstoerung"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown command should be 400, got %d", rec.Code)
	}
	// no targets must never fan out to the whole farm
	if rec := postCmd(`{"command":"stop"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty ip list should be 400, got %d", rec.Code)
	}
	if rec := postCmd(`{"ips":["1.2.3.4"],"command":"stop"}`); rec.Code != 200 {
		t.Fatalf("valid request should be 200, got %d", rec.Code)
	}
}

func TestPrintCommandReportsUnreachablePrinters(t *testing.T) {
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{Name: "Woobly 1", IP: "10.9.9.1", Serial: "ABC"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	rec := postCmd(`{"ips":["10.9.9.1"],"command":"pause"}`)
	var d struct {
		Sent   int                 `json:"sent"`
		Failed []map[string]string `json:"failed"`
	}
	json.Unmarshal(rec.Body.Bytes(), &d)
	if d.Sent != 0 || len(d.Failed) != 1 {
		t.Fatalf("expected one failure, got %+v", d)
	}
	if !strings.Contains(d.Failed[0]["error"], "MQTT") {
		t.Fatalf("unhelpful error: %v", d.Failed[0])
	}
}

func TestPrintCommandPayloads(t *testing.T) {
	for _, c := range []string{"pause", "resume", "stop"} {
		p, ok := printCommands[c]
		if !ok {
			t.Fatalf("%s missing", c)
		}
		if !strings.Contains(p, `"command":"`+c+`"`) || !strings.Contains(p, `"print"`) {
			t.Fatalf("%s payload looks wrong: %s", c, p)
		}
		// "param" gehoert laut Protokoll zwingend dazu, auch leer. Fehlt es,
		// nimmt der Drucker den Befehl an und tut nichts.
		if !strings.Contains(p, `"param":""`) {
			t.Fatalf("%s: Pflichtfeld param fehlt: %s", c, p)
		}
		// Die Vorlage bekommt die sequence_id erst beim Senden.
		fertig := fmt.Sprintf(p, "9001")
		if !json.Valid([]byte(fertig)) {
			t.Fatalf("%s: kein gueltiges JSON: %s", c, fertig)
		}
		if !strings.Contains(fertig, `"sequence_id":"9001"`) {
			t.Fatalf("%s: sequence_id nicht eingesetzt: %s", c, fertig)
		}
	}
}

// Gegen einen echten Payload aus druckerfarm.log — nicht gegen meine Annahme,
// wie der Drucker die Daten schickt.
func TestAMSAgainstRealPayload(t *testing.T) {
	raw, err := os.ReadFile("testdata/ams_real.json")
	if err != nil {
		t.Skip("keine echten Testdaten hinterlegt")
	}
	s := parsePrint(t, string(raw))

	if len(s.AMS) != 1 {
		t.Fatalf("want 1 AMS unit, got %d", len(s.AMS))
	}
	u := s.AMS[0]
	if u.Humidity == "" || u.Temp == "" {
		t.Fatalf("unit meta missing: %+v", u)
	}
	if len(u.Trays) != 4 {
		t.Fatalf("want 4 trays, got %d", len(u.Trays))
	}

	loaded := 0
	for _, tr := range u.Trays {
		if tr.Type == "" {
			continue // leeres Fach
		}
		loaded++
		if tr.Color == "" {
			t.Fatalf("loaded tray without colour: %+v", tr)
		}
		if len(tr.Color) != 8 {
			t.Fatalf("colour is not RRGGBBAA: %q", tr.Color)
		}
		if len(tr.Colors) == 0 {
			t.Fatalf("cols not parsed: %+v", tr)
		}
	}
	if loaded != 3 {
		t.Fatalf("want 3 loaded trays, got %d", loaded)
	}
	if s.TrayNow != "0" {
		t.Fatalf("tray_now = %q", s.TrayNow)
	}
	// Die externe Spule ist in diesen Daten leer und darf deshalb nicht als
	// bestueckt durchgereicht werden.
	if s.ExtSpool != nil && s.ExtSpool.Type != "" {
		t.Fatalf("empty external spool reported as loaded: %+v", s.ExtSpool)
	}
	t.Logf("gelesen: %d Faecher, aktiv=%s, Farben=%v", loaded, s.TrayNow,
		[]string{u.Trays[0].Color, u.Trays[1].Color, u.Trays[3].Color})
}

// Ein Drucker darf ohne Zugangscode angelegt werden — beim Aufbau einer Farm
// traegt man erst die Geraete ein und die Codes spaeter nach. Name und Adresse
// bleiben Bedingung, ohne die ist der Eintrag sinnlos.
func TestAnlegenOhneZugangscode(t *testing.T) {
	faelle := []struct {
		name    string
		koerper string
		willOK  bool
	}{
		{"ohne Code", `{"model":"X1E","name":"a","ip":"10.0.0.9","code":"","serial":"S1"}`, true},
		{"mit Code", `{"model":"X1E","name":"a","ip":"10.0.0.9","code":"abc","serial":"S1"}`, true},
		{"ohne Name", `{"model":"X1E","name":"","ip":"10.0.0.9","code":"abc"}`, false},
		{"ohne IP", `{"model":"X1E","name":"a","ip":"","code":"abc"}`, false},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			mu.Lock()
			alt := state.Printers
			state.Printers = nil
			mu.Unlock()
			defer func() { mu.Lock(); state.Printers = alt; mu.Unlock() }()

			req := httptest.NewRequest(http.MethodPost, "/api/printers", strings.NewReader(f.koerper))
			rec := httptest.NewRecorder()
			handlePrinters(rec, req)
			ok := rec.Code < 300
			if ok != f.willOK {
				t.Fatalf("HTTP %d (%s), erwartet ok=%v", rec.Code, strings.TrimSpace(rec.Body.String()), f.willOK)
			}
		})
	}
}

// Duplikate werden an der Seriennummer erkannt, nicht an der IP: ein neues
// Geraet mit einer IP, die ein bestehender (Offline-)Drucker noch traegt, muss
// sich anlegen lassen. Dieselbe Seriennummer dagegen ist ein Duplikat.
func TestAddDuplikatUeberSeriennummer(t *testing.T) {
	mu.Lock()
	altP := state.Printers
	state.Printers = []Printer{{Name: "Alt", IP: "192.168.0.50", Serial: "S-ALT", Code: "x"}}
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); state.Printers = altP; mu.Unlock() })

	post := func(body string) int {
		rec := httptest.NewRecorder()
		handlePrinters(rec, httptest.NewRequest("POST", "/api/printers", strings.NewReader(body)))
		return rec.Code
	}

	// Gleiche IP, ANDERE Seriennummer -> erlaubt (der Alte ist offline/umgezogen).
	if code := post(`{"name":"Neu","ip":"192.168.0.50","serial":"S-NEU","code":"y"}`); code != 201 {
		t.Fatalf("gleiche IP, andere Serial sollte anlegen (201), bekam %d", code)
	}
	// Gleiche Seriennummer -> Duplikat (409).
	if code := post(`{"name":"NochMal","ip":"192.168.0.99","serial":"S-ALT","code":"z"}`); code != 409 {
		t.Fatalf("gleiche Seriennummer sollte 409 sein, bekam %d", code)
	}
}
