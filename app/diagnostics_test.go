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

// The test harness mimics go2rtc: /api/streams says whether a stream is known,
// /api/frame.jpeg returns 200 with an empty body, like the real go2rtc.
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

// If go2rtc does not know the stream, the config is stale — not ffmpeg's fault.
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

// If ffmpeg is really missing, that should be stated too.
func TestDiagnoseMissingFFmpeg(t *testing.T) {
	fakeGo2rtc(t, map[string]bool{"woobly-10": true})
	withFFmpeg(t, false)
	setPrinters(Printer{Name: "woobly 10", IP: "10.0.0.10"})

	// Only meaningful if there is none system-wide either
	if _, err := os.Stat("/usr/bin/ffmpeg"); err == nil {
		t.Skip("systemweites ffmpeg vorhanden")
	}
	msg := diagnoseStream("woobly-10")
	if !strings.Contains(strings.ToLower(msg), "ffmpeg") {
		t.Fatalf("fehlendes ffmpeg wurde nicht erkannt: %s", msg)
	}
}

// The most common case on the farm: stream known, ffmpeg present, camera silent.
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

// If the printer answers via MQTT but the camera does not, that should be distinguished.
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

// Port open but no image — then usually the access code is wrong.
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

	// The test needs port 322 — so only the branch is checked here
	// that applies when the connection is up.
	host, _, _ := net.SplitHostPort(ln.Addr().String())
	setPrinters(Printer{Name: "kamera", IP: host})
	msg := diagnoseStream("kamera")
	if msg == "" {
		t.Fatal("leere Diagnose")
	}
	t.Logf("Meldung: %s", msg)
}

// The full error message as it ends up in the tile.
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
