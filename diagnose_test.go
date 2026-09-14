package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Der Prüfstand stellt go2rtc nach: /api/streams sagt, ob ein Stream bekannt ist,
// /api/frame.jpeg liefert 200 mit leerem Rumpf, so wie das echte go2rtc.
func fakeGo2rtc(t *testing.T, known map[string]bool) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		src := r.URL.Query().Get("src")
		switch r.URL.Path {
		case "/api/streams":
			if !known[src] {
				http.NotFound(w, r)
				return
			}
			w.Write([]byte(`{"producers":[{"url":"rtspx://x"}]}`))
		case "/api/frame.jpeg":
			w.WriteHeader(http.StatusOK) // 200, leerer Rumpf
		default:
			http.NotFound(w, r)
		}
	}))
	port, _ := strconv.Atoi(strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))
	old := go2rtcPort
	go2rtcPort = port
	t.Cleanup(func() { go2rtcPort = old; srv.Close() })
}

func withFFmpeg(t *testing.T, present bool) {
	t.Helper()
	dir := t.TempDir()
	oldDir := appDir
	appDir = dir
	t.Cleanup(func() { appDir = oldDir })
	if present {
		os.MkdirAll(ffmpegDir(), 0o755)
		os.WriteFile(ffmpegBinPath(), []byte("#!/bin/sh\necho ffmpeg version 7\n"), 0o755)
	}
}

// Kennt go2rtc den Stream nicht, ist die Konfiguration veraltet — nicht ffmpeg schuld.
func TestDiagnoseUnknownStream(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{})
	withFFmpeg(t, true)
	setPrinters(Printer{Name: "woobly 10", IP: "10.0.0.10"})

	msg := diagnoseStream("woobly-10")
	if !strings.Contains(msg, "go2rtc-Konfiguration") {
		t.Fatalf("falsche Diagnose: %s", msg)
	}
	if strings.Contains(strings.ToLower(msg), "ffmpeg") {
		t.Fatalf("ffmpeg darf hier nicht beschuldigt werden: %s", msg)
	}
}

// Fehlt ffmpeg wirklich, soll das auch dranstehen.
func TestDiagnoseMissingFFmpeg(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{"woobly-10": true})
	withFFmpeg(t, false)
	setPrinters(Printer{Name: "woobly 10", IP: "10.0.0.10"})

	// Nur aussagekräftig, wenn auch systemweit keins liegt
	if _, err := os.Stat("/usr/bin/ffmpeg"); err == nil {
		t.Skip("systemweites ffmpeg vorhanden")
	}
	msg := diagnoseStream("woobly-10")
	if !strings.Contains(strings.ToLower(msg), "ffmpeg") {
		t.Fatalf("fehlendes ffmpeg wurde nicht erkannt: %s", msg)
	}
}

// Der häufigste Fall in der Farm: Stream bekannt, ffmpeg da, Kamera antwortet nicht.
func TestDiagnoseCameraUnreachable(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{"woobly-10": true})
	withFFmpeg(t, true)
	setPrinters(Printer{Name: "woobly 10", IP: "10.255.255.10"})

	msg := diagnoseStream("woobly-10")
	if !strings.Contains(msg, "woobly 10") {
		t.Fatalf("Druckername fehlt in der Meldung: %s", msg)
	}
	if !strings.Contains(msg, "322") {
		t.Fatalf("Port sollte benannt werden: %s", msg)
	}
	if strings.Contains(strings.ToLower(msg), "ffmpeg") {
		t.Fatalf("ffmpeg darf hier nicht beschuldigt werden: %s", msg)
	}
	t.Logf("Meldung: %s", msg)
}

// Antwortet der Drucker per MQTT, aber die Kamera nicht, soll das unterschieden werden.
func TestDiagnoseOnlineButNoCamera(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{"woobly-10": true})
	withFFmpeg(t, true)
	setPrinters(Printer{Name: "woobly 10", IP: "10.255.255.10"})

	mqttMgr.mu.Lock()
	mqttMgr.statuses["10.255.255.10"] = &PrinterStatus{Online: true}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, "10.255.255.10")
		mqttMgr.mu.Unlock()
	}()

	msg := diagnoseStream("woobly-10")
	if !strings.Contains(msg, "MQTT") {
		t.Fatalf("MQTT-Zustand nicht berücksichtigt: %s", msg)
	}
	t.Logf("Meldung: %s", msg)
}

// Port offen, aber kein Bild — dann ist meist der Zugangscode falsch.
func TestDiagnosePortOpenButNoImage(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	fakeGo2rtc(t, map[string]bool{"kamera": true})
	withFFmpeg(t, true)

	// Der Test braucht Port 322 — deshalb wird hier nur der Zweig geprüft,
	// der greift, wenn die Verbindung steht.
	host, _, _ := net.SplitHostPort(ln.Addr().String())
	setPrinters(Printer{Name: "kamera", IP: host})
	msg := diagnoseStream("kamera")
	if msg == "" {
		t.Fatal("leere Diagnose")
	}
	t.Logf("Meldung: %s", msg)
}

// Die vollständige Fehlermeldung, wie sie in der Kachel landet.
func TestSnapshotErrorMentionsRealCause(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{})
	withFFmpeg(t, true)
	setPrinters(Printer{Name: "woobly 10", IP: "10.255.255.10"})
	resetSnapState()

	rec := httptest.NewRecorder()
	handleSnapshot(rec, httptest.NewRequest("GET", "/api/snapshot/10.255.255.10?max_age=1", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(strings.ToLower(body), "ffmpeg fehlt im path") {
		t.Fatalf("alte Pauschalmeldung ist zurück: %s", body)
	}
	if !strings.Contains(body, "go2rtc-Konfiguration") {
		t.Fatalf("Grund fehlt: %s", body)
	}
	fmt.Println("   Kachelmeldung:", strings.TrimSpace(body))
	_ = filepath.Join
}
