package main

import (
	"strings"
	"testing"
)

func joinArgs(a []string) string { return strings.Join(a, " ") }

func TestBuildYouTubeArgs_Single(t *testing.T) {
	args := buildYouTubeArgs([]string{"rtsp://127.0.0.1:8554/a"}, ytOptions{Key: "KEY-123", Resolution: "1280x720", Fps: 10, BitrateK: 2000})
	s := joinArgs(args)
	if !strings.Contains(s, "rtmp://a.rtmp.youtube.com/live2/KEY-123") {
		t.Fatalf("RTMP-Ziel/Key fehlt: %s", s)
	}
	if strings.Contains(s, "xstack") {
		t.Fatalf("bei einem Stream darf kein xstack vorkommen: %s", s)
	}
	if !strings.Contains(s, "-map [v0]") {
		t.Fatalf("erwartete -map [v0]: %s", s)
	}
	if !strings.Contains(s, "-b:v 2000k") || !strings.Contains(s, "-r 10") {
		t.Fatalf("Bitrate/FPS nicht uebernommen: %s", s)
	}
	// No plaintext access code in the arguments (streams come from go2rtc locally).
	if strings.Contains(s, "bblp:") || strings.Contains(s, "@192.") {
		t.Fatalf("unerwartete Zugangsdaten in ffmpeg-Args: %s", s)
	}
}

func TestBuildYouTubeArgs_Grid(t *testing.T) {
	streams := []string{
		"rtsp://127.0.0.1:8554/a",
		"rtsp://127.0.0.1:8554/b",
		"rtsp://127.0.0.1:8554/c",
		"rtsp://127.0.0.1:8554/d",
	}
	args := buildYouTubeArgs(streams, ytOptions{Key: "K", Cols: 2, Resolution: "1920x1080", Fps: 15})
	s := joinArgs(args)
	if !strings.Contains(s, "xstack=inputs=4:layout=") {
		t.Fatalf("xstack fuer 4 Streams fehlt: %s", s)
	}
	// 4 Kacheln → 4 Layout-Positionen (3 Trennstriche) im xstack-Layout.
	li := strings.Index(s, "xstack=inputs=4:layout=")
	seg := s[li:]
	seg = seg[:strings.Index(seg, " ")] // up to the next space
	if strings.Count(seg, "|") != 3 {
		t.Fatalf("erwartete 4 Layout-Positionen: %s", seg)
	}
	if !strings.Contains(s, "-map [out]") {
		t.Fatalf("erwartete -map [out]: %s", s)
	}
	// Fuenf Eingaenge insgesamt (4 Video + 1 anullsrc) → Audio-Map auf 4:a.
	if !strings.Contains(s, "-map 4:a") {
		t.Fatalf("Audio-Map falsch: %s", s)
	}
}

func TestParseRes(t *testing.T) {
	if w, h := parseRes(""); w != 1920 || h != 1080 {
		t.Fatalf("Default falsch: %d x %d", w, h)
	}
	if w, h := parseRes("1280x720"); w != 1280 || h != 720 {
		t.Fatalf("Parsing falsch: %d x %d", w, h)
	}
	if w, h := parseRes("bloedsinn"); w != 1920 || h != 1080 {
		t.Fatalf("Fallback falsch: %d x %d", w, h)
	}
}

func TestTunnelURLRegex(t *testing.T) {
	line := "2024-01-01T00:00:00Z INF |  https://random-happy-tree-1234.trycloudflare.com    |"
	got := cfURLRe.FindString(line)
	want := "https://random-happy-tree-1234.trycloudflare.com"
	if got != want {
		t.Fatalf("Tunnel-URL nicht erkannt: %q", got)
	}
	if cfURLRe.FindString("keine url hier") != "" {
		t.Fatalf("Fehlalarm bei Zeile ohne URL")
	}
}
