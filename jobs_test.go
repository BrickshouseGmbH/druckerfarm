package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func resetJobs() {
	jobsMu.Lock()
	jobs = map[string]*job{}
	jobsMu.Unlock()
}

func status(t *testing.T, id string, hits, failed int) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	handleJobStatus(rec, httptest.NewRequest("GET",
		fmt.Sprintf("/api/job/status?id=%s&hits=%d&failed=%d", id, hits, failed), nil))
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var d map[string]any
	json.Unmarshal(rec.Body.Bytes(), &d)
	return d
}

func num(v any) int {
	f, _ := v.(float64)
	return int(f)
}

// Der Zwischenstand liefert nur, was seit dem letzten Abruf dazugekommen ist —
// sonst müsste die Oberfläche die Liste bei jedem Abruf neu bauen.
func TestJobDeliversOnlyNewEntries(t *testing.T) {
	resetJobs()
	j := newJob("search", "test", 3)

	j.addHits([]sdSearchHit{{IP: "10.0.0.1", File: "a.3mf"}})
	d := status(t, j.id, 0, 0)
	if got := len(d["hits"].([]any)); got != 1 {
		t.Fatalf("want 1 hit, got %d", got)
	}
	if num(d["done"]) != 1 || num(d["total"]) != 3 {
		t.Fatalf("progress wrong: %v", d)
	}

	// nochmal mit demselben Stand abfragen → nichts Neues
	if got := len(status(t, j.id, 1, 0)["hits"].([]any)); got != 0 {
		t.Fatalf("already known hits were delivered again")
	}

	j.addHits([]sdSearchHit{{IP: "10.0.0.2", File: "b.3mf"}, {IP: "10.0.0.2", File: "c.3mf"}})
	d = status(t, j.id, 1, 0)
	if got := len(d["hits"].([]any)); got != 2 {
		t.Fatalf("want 2 new hits, got %d", got)
	}
	if num(d["total_hits"]) != 3 {
		t.Fatalf("total wrong: %v", d["total_hits"])
	}

	j.addFailure(jobFailure{IP: "10.0.0.3", Error: "Zeitüberschreitung"})
	d = status(t, j.id, 3, 0)
	if got := len(d["failed"].([]any)); got != 1 {
		t.Fatalf("failure was not delivered: %v", d)
	}
	if d["complete"] != false {
		t.Fatal("job should still be running")
	}
	j.finish()
	if status(t, j.id, 3, 1)["complete"] != true {
		t.Fatal("job should be complete")
	}
}

func TestUnknownJobIs404(t *testing.T) {
	resetJobs()
	rec := httptest.NewRecorder()
	handleJobStatus(rec, httptest.NewRequest("GET", "/api/job/status?id=gibtsnicht", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// Ein Abbruch muss laufende Arbeit stoppen, nicht nur ein Flag setzen.
func TestStopEndsTheWork(t *testing.T) {
	resetJobs()
	j := newJob("search", "x", 100)

	var processed int
	var mu2 sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer j.finish()
		for i := 0; i < 100; i++ {
			if j.cancelled() {
				return
			}
			mu2.Lock()
			processed++
			mu2.Unlock()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	time.Sleep(60 * time.Millisecond)
	rec := httptest.NewRecorder()
	handleJobStop(rec, httptest.NewRequest("POST", "/api/job/stop?id="+j.id, nil))
	if rec.Code != 200 {
		t.Fatalf("stop: %d", rec.Code)
	}
	wg.Wait()

	mu2.Lock()
	n := processed
	mu2.Unlock()
	if n >= 100 {
		t.Fatalf("work ran to completion despite the stop (%d)", n)
	}
	d := status(t, j.id, 0, 0)
	if d["stopped"] != true || d["complete"] != true {
		t.Fatalf("state after stop: %v", d)
	}
	// Zweiter Abbruch darf nicht in einen Panic laufen
	j.stop()
}

func TestSearchNeedsThreeCharacters(t *testing.T) {
	resetJobs()
	rec := httptest.NewRecorder()
	handleSearchStart(rec, httptest.NewRequest("POST", "/api/sync/search/start", strings.NewReader(`{"q":"ab"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestSearchWithoutPrintersInFilter(t *testing.T) {
	resetJobs()
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{IP: "10.0.0.1", Name: "A"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	rec := httptest.NewRecorder()
	handleSearchStart(rec, httptest.NewRequest("POST", "/api/sync/search/start",
		strings.NewReader(`{"q":"jimmy","ips":["10.9.9.9"]}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for an empty filter, got %d", rec.Code)
	}
}

// Ein Löschauftrag darf nur Drucker treffen, die wirklich in der Liste stehen.
func TestDeleteIgnoresUnknownPrinters(t *testing.T) {
	resetJobs()
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{IP: "10.0.0.1", Name: "A", Code: "x"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	rec := httptest.NewRecorder()
	handleDeleteStart(rec, httptest.NewRequest("POST", "/api/sync/delete/start",
		strings.NewReader(`{"items":[{"ip":"10.9.9.9","file":"fremd.3mf"}]}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("an unknown printer should be rejected, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	handleDeleteStart(rec2, httptest.NewRequest("POST", "/api/sync/delete/start", strings.NewReader(`{"items":[]}`)))
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("empty selection should be rejected, got %d", rec2.Code)
	}
}

// Ein Dateiname mit Pfadtrennern wäre ein Weg aus dem SD-Verzeichnis heraus.
func TestDeleteRejectsPathTraversal(t *testing.T) {
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{IP: "10.0.0.1", Name: "A", Code: "geheim"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	for _, bad := range []string{"../../etc/passwd", "unter/ordner.3mf", `..\windows\system32`} {
		if err := deleteSDFile("10.0.0.1", bad); err == nil || !strings.Contains(err.Error(), "ungültig") {
			t.Fatalf("%q was not rejected: %v", bad, err)
		}
	}
	if err := deleteSDFile("10.9.9.9", "datei.3mf"); err == nil {
		t.Fatal("unknown printer was not rejected")
	}
}

// Beendete Vorgänge dürfen sich nicht endlos ansammeln.
func TestFinishedJobsAreCleanedUp(t *testing.T) {
	resetJobs()
	oldJ := newJob("search", "alt", 1)
	oldJ.finish()
	oldJ.mu.Lock()
	oldJ.finished = time.Now().Add(-2 * jobKeepFor)
	oldJ.mu.Unlock()

	newJob("search", "neu", 1) // legt beim Anlegen den alten weg
	if getJob(oldJ.id) != nil {
		t.Fatal("expired job was not cleaned up")
	}
}

// ─── Gebündeltes Löschen ──────────────────────────────────────────────────────

// Die Dateien eines Druckers müssen in EINEN Aufruf gehen, nicht in einen pro
// Datei — das war der Flaschenhals bei über hundert Dateien.
func TestDeleteGroupsFilesPerPrinter(t *testing.T) {
	resetJobs()
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{
		{IP: "10.0.0.1", Name: "A", Code: "c1"},
		{IP: "10.0.0.2", Name: "B", Code: "c2"},
	}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	body := `{"items":[
		{"ip":"10.0.0.1","file":"a.3mf"},
		{"ip":"10.0.0.1","file":"b.3mf"},
		{"ip":"10.0.0.1","file":"c.3mf"},
		{"ip":"10.0.0.2","file":"a.3mf"}]}`
	rec := httptest.NewRecorder()
	handleDeleteStart(rec, httptest.NewRequest("POST", "/api/sync/delete/start", strings.NewReader(body)))
	if rec.Code != 200 {
		t.Fatalf("code %d: %s", rec.Code, rec.Body.String())
	}
	var d struct {
		Total    int `json:"total"`
		Printers int `json:"printers"`
	}
	json.Unmarshal(rec.Body.Bytes(), &d)
	if d.Total != 4 {
		t.Fatalf("want 4 files, got %d", d.Total)
	}
	if d.Printers != 2 {
		t.Fatalf("want 2 printers, got %d", d.Printers)
	}
}

func TestBundledDeleteRejectsBadNames(t *testing.T) {
	mu.Lock()
	old := state.Printers
	state.Printers = []Printer{{IP: "10.0.0.1", Name: "A", Code: "c"}}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers = old; mu.Unlock() }()

	// Nur unzulässige Namen: es darf kein FTP-Aufruf entstehen, das Ergebnis
	// meldet sie als abgelehnt.
	res, err := deleteSDFiles("10.0.0.1", []string{"../fremd.3mf", "unter/ordner.3mf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2 results, got %d", len(res))
	}
	for _, r := range res {
		if r.OK || !strings.Contains(r.Error, "ungültig") {
			t.Fatalf("bad name was not rejected: %+v", r)
		}
	}

	if _, err := deleteSDFiles("10.9.9.9", []string{"x.3mf"}); err == nil {
		t.Fatal("unknown printer must fail")
	}
}

// ─── Kammerbeleuchtung ────────────────────────────────────────────────────────

func TestErrorDetection(t *testing.T) {
	cases := []struct {
		name string
		s    *PrinterStatus
		want bool
	}{
		{"offline", &PrinterStatus{Online: false, PrintError: 5}, false},
		{"nil", nil, false},
		{"sauber", &PrinterStatus{Online: true, GcodeState: "RUNNING"}, false},
		{"print_error", &PrinterStatus{Online: true, PrintError: 117}, true},
		{"hms", &PrinterStatus{Online: true, HmsErrors: []string{"0300_0100"}}, true},
		{"failed", &PrinterStatus{Online: true, GcodeState: "FAILED"}, true},
	}
	for _, c := range cases {
		if got := printerHasError(c.s); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestBlinkModelDefaults(t *testing.T) {
	mu.Lock()
	old := state.BlinkModels
	state.BlinkModels = nil
	mu.Unlock()
	defer func() { mu.Lock(); state.BlinkModels = old; mu.Unlock() }()

	m := blinkModels()
	if !m["X1E"] || !m["X2D"] {
		t.Fatalf("defaults missing: %v", m)
	}
	if m["H2D"] || m["H2C"] {
		t.Fatalf("models with their own signal lamp must not blink: %v", m)
	}

	mu.Lock()
	state.BlinkModels = []string{" p1s ", "a1"}
	mu.Unlock()
	m = blinkModels()
	if !m["P1S"] || !m["A1"] {
		t.Fatalf("configured models not normalised: %v", m)
	}
	if m["X1E"] {
		t.Fatal("configured list must replace the defaults")
	}
}

// Ohne MQTT-Verbindung muss der Lichtbefehl einen klaren Fehler liefern statt
// still ins Leere zu laufen.
func TestChamberLightWithoutConnection(t *testing.T) {
	err := mqttMgr.SetChamberLight(Printer{IP: "10.9.9.9", Serial: "ABC"}, "flashing", 1000, 1000)
	if err == nil || !strings.Contains(err.Error(), "MQTT") {
		t.Fatalf("unhelpful error: %v", err)
	}
	if err := mqttMgr.SetChamberLight(Printer{IP: "10.9.9.9"}, "on", 0, 0); err == nil {
		t.Fatal("missing serial must fail")
	}
}

// ─── Blinken zuverlässig beenden ──────────────────────────────────────────────

// Der Drucker schickt das hms-Feld nicht bei jeder Nachricht mit. Ohne Aufräumen
// bliebe eine alte Störung ewig stehen und das Licht würde endlos blinken.
func TestResumeClearsStaleFaults(t *testing.T) {
	ip := "10.9.9.9"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{Online: true}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	// Störung tritt auf
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"PAUSE","print_error":117,"hms":[{"ecode":"0300_0100"}]}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("fault was not detected")
	}

	// Druck läuft wieder — ohne hms-Feld, so wie es der Drucker meist schickt
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":12}}`))
	s := mqttMgr.GetStatus(ip)
	if len(s.HmsErrors) != 0 || s.PrintError != 0 {
		t.Fatalf("stale fault survived the resume: %+v", s)
	}
	if printerHasError(s) {
		t.Fatal("printer still counts as faulty after resuming — the light would keep blinking")
	}
}

// Läuft der Druck durchgehend, darf ein Statusupdate die Störung nicht löschen.
func TestFaultDuringRunningIsKept(t *testing.T) {
	ip := "10.9.9.8"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{Online: true, GcodeState: "RUNNING"}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	mqttMgr.handleMessage(ip, []byte(`{"print":{"hms":[{"ecode":"0300_0100"}]}}`))
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":40}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("a fault during a running print must not be cleared")
	}
}

// Nach einem Neustart der App muss bekannt sein, wer noch blinkt.
func TestBlinkingListSurvivesRestart(t *testing.T) {
	mu.Lock()
	old := state.BlinkingIPs
	state.BlinkingIPs = []string{"10.0.0.5", "10.0.0.7"}
	mu.Unlock()
	defer func() { mu.Lock(); state.BlinkingIPs = old; mu.Unlock() }()

	mu.Lock()
	got := append([]string(nil), state.BlinkingIPs...)
	mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("list not kept: %v", got)
	}

	b, err := json.Marshal(AppState{BlinkingIPs: got})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"blinking_ips"`) {
		t.Fatalf("not persisted: %s", b)
	}
}

// Abschalten der Sonderbeleuchtung setzt die Liste zurück.
func TestDisablingBlinkResetsTheList(t *testing.T) {
	mu.Lock()
	oldP, oldB := state.Printers, state.BlinkingIPs
	state.Printers = nil // ohne MQTT-Verbindung passiert nichts weiter
	state.BlinkingIPs = []string{"10.0.0.5"}
	mu.Unlock()
	defer func() { mu.Lock(); state.Printers, state.BlinkingIPs = oldP, oldB; mu.Unlock() }()

	ResetChamberLights()

	mu.Lock()
	left := len(state.BlinkingIPs)
	mu.Unlock()
	if left != 0 {
		t.Fatalf("list was not cleared: %d entries left", left)
	}
}
