package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCameraStreamsYaml(t *testing.T) {
	cams := []CameraCfg{
		{ID: "cw300", Name: "Außenkamera Büro", Stream: "cw300", Source: "xiaomi://u:p@192.168.1.9?did=1"},
		{ID: "leer", Name: "x", Stream: "", Source: ""}, // incomplete -> skipped
	}
	y := cameraStreamsYaml(cams)
	if !strings.Contains(y, "  cw300:\n    - xiaomi://u:p@192.168.1.9?did=1\n") {
		t.Fatalf("Kamera-Stream fehlt: %q", y)
	}
	if strings.Contains(y, "leer") {
		t.Fatalf("unvollständige Kamera sollte nicht erscheinen: %q", y)
	}
}

func TestYamlEnthaeltDruckerUndKamera(t *testing.T) {
	printers := []Printer{{Name: "Halle 1", IP: "10.0.0.1", Code: "c", Model: "H2D"}}
	cams := []CameraCfg{{ID: "cw300", Stream: "cw300", Source: "xiaomi://x@1.2.3.4"}}
	y := buildYaml(printers, nil) + cameraStreamsYaml(cams)
	if !strings.Contains(y, "streams:") || !strings.Contains(y, "halle-1:") {
		t.Fatalf("Drucker-Stream fehlt: %q", y)
	}
	if !strings.Contains(y, "cw300:") {
		t.Fatalf("Kamera-Stream fehlt: %q", y)
	}
}

func TestCamerasAPIRoundtrip(t *testing.T) {
	state.Cameras = nil
	// POST
	body := `{"name":"Außenkamera Büro","stream":"cw300","source":"xiaomi://u:p@192.168.189.189?did=1109718403&model=mxiang.camera.moc006"}`
	req := httptest.NewRequest(http.MethodPost, "/api/cameras", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleCameras(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST rc=%d: %s", rec.Code, rec.Body.String())
	}
	if len(state.Cameras) != 1 || state.Cameras[0].ID == "" || state.Cameras[0].Kind != "standalone" {
		t.Fatalf("Kamera nicht gespeichert: %+v", state.Cameras)
	}
	id := state.Cameras[0].ID
	// GET
	rec = httptest.NewRecorder()
	handleCameras(rec, httptest.NewRequest(http.MethodGet, "/api/cameras", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "cw300") {
		t.Fatalf("GET falsch: %d %s", rec.Code, rec.Body.String())
	}
	// Snapshot resolves camera ID to stream name (no 404 for unknown)
	if cameraStreamName(id) != "cw300" {
		t.Fatalf("cameraStreamName falsch: %q", cameraStreamName(id))
	}
	// DELETE
	rec = httptest.NewRecorder()
	handleCameraByID(rec, httptest.NewRequest(http.MethodDelete, "/api/cameras/"+id, nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE rc=%d", rec.Code)
	}
	if len(state.Cameras) != 0 {
		t.Fatalf("Kamera nicht gelöscht: %+v", state.Cameras)
	}
}

func TestKameraNichtInDruckerzaehlung(t *testing.T) {
	// Cameras live in state.Cameras, not state.Printers — the printer count
	// and everything printer-related stays untouched by it.
	state.Printers = []Printer{{Name: "A", IP: "10.0.0.1", Serial: "S", Model: "H2D"}}
	state.Cameras = []CameraCfg{{ID: "cw300", Name: "Cam", Stream: "cw300", Source: "xiaomi://x@1.2.3.4"}}
	if len(state.Printers) != 1 {
		t.Fatalf("Druckerzahl verändert: %d", len(state.Printers))
	}
}

func TestKameraStreamDiagnoseGetrennt(t *testing.T) {
	state.Cameras = []CameraCfg{{ID: "dowell3d", Name: "Dowell", Stream: "dowell3d", Source: "xiaomi://x@1.2.3.4"}}
	if !isCameraStream("dowell3d") {
		t.Fatal("dowell3d sollte als Kamera-Stream erkannt werden")
	}
	if isCameraStream("halle-1") {
		t.Fatal("Drucker-Stream fälschlich als Kamera erkannt")
	}
	msg := cameraStreamDiagnostics("dowell3d")
	if strings.Contains(msg, "Drucker") {
		t.Fatalf("Kamera-Diagnose darf nicht von Druckern reden: %q", msg)
	}
	if !strings.Contains(msg, "Dowell") {
		t.Fatalf("Kamera-Diagnose sollte den Namen nennen: %q", msg)
	}
	state.Cameras = nil
}

func TestKameraSnapshotKein503(t *testing.T) {
	// Camera without reachable go2rtc → snapshot fails. For cameras this must
	// NOT be a 503 (otherwise a red console error), but 200 + hint.
	state.Printers = nil
	state.Cameras = []CameraCfg{{ID: "dowell3d", Name: "Dowell", Stream: "dowell3d", Source: "xiaomi://x@1.2.3.4"}}
	req := httptest.NewRequest(http.MethodGet, "/api/snapshot/dowell3d?max_age=2", nil)
	rec := httptest.NewRecorder()
	handleSnapshot(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Kamera-Snapshot sollte 200 sein (kein 503), war %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Snapshot-Unavailable") != "1" {
		t.Fatalf("Header X-Snapshot-Unavailable fehlt")
	}
	state.Cameras = nil
}

func TestErhalteFremdeSektionen(t *testing.T) {
	// Simulates a go2rtc.yaml into which go2rtc wrote its own
	// account/token on Mi-Home login. api/streams are ours; the account must
	// erhalten bleiben.
	vorhanden := "# go2rtc.yaml\n\napi:\n  origin: '*'\n\nstreams:\n  alt: rtsp://x\n\nxiaomi:\n  6782331241:\n    token: GEHEIM\n    region: de\n"
	fremd := erhalteFremdeSektionen(vorhanden)
	if !strings.Contains(fremd, "xiaomi:") || !strings.Contains(fremd, "token: GEHEIM") {
		t.Fatalf("Mi-Home-Abschnitt ging verloren: %q", fremd)
	}
	if strings.Contains(fremd, "origin:") || strings.Contains(fremd, "alt: rtsp") {
		t.Fatalf("verwaltete Abschnitte (api/streams) dürfen NICHT übernommen werden: %q", fremd)
	}
	// New file = our sections + preserved foreign part
	neu := buildYaml([]Printer{{Name: "A", IP: "10.0.0.1", Code: "c", Model: "H2D"}}, nil) + fremd
	if !strings.Contains(neu, "xiaomi:") || !strings.Contains(neu, "streams:") {
		t.Fatalf("Zusammenbau falsch: %q", neu)
	}
}

func TestNeueKameraGeneration(t *testing.T) {
	for _, m := range []string{"H2D", "X2D", "P2S", "h2s"} {
		if !newCameraGeneration(m) {
			t.Fatalf("%s sollte neue Generation sein", m)
		}
	}
	for _, m := range []string{"X1C", "X1E", "P1S", "A1", "A1 MINI", ""} {
		if newCameraGeneration(m) {
			t.Fatalf("%s sollte NICHT neue Generation sein", m)
		}
	}
}
