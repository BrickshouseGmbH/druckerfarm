package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// a tiny script that behaves like a tool answering "-version"
const fakeTool = "#!/bin/sh\necho \"fake tool version 9.9\"\n"

func zipWith(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(0o755)
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func useTempAppDir(t *testing.T) string {
	t.Helper()
	old := appDir
	appDir = t.TempDir()
	t.Cleanup(func() { appDir = old })
	return appDir
}

func resetProgress() { dlMu.Lock(); dlProgress = installProgress{}; dlMu.Unlock() }

// ─── unzipSelected ────────────────────────────────────────────────────────────

func TestUnzipSelectedPicksOnlyWantedEntries(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.zip")
	os.WriteFile(archive, zipWith(t, map[string]string{
		"build/bin/ffmpeg":      fakeTool,
		"build/bin/avcodec.dll": "dll",
		"build/bin/ffprobe":     "nope",
		"build/include/x.h":     "nope",
		"build/lib/x.lib":       "nope",
	}), 0o644)

	out := filepath.Join(dir, "out")
	n, err := unzipSelected(archive, out, func(name string) (string, bool) {
		p := strings.ToLower(filepath.ToSlash(name))
		if !strings.Contains(p, "/bin/") {
			return "", false
		}
		base := filepath.Base(p)
		if base == "ffmpeg" || strings.HasSuffix(base, ".dll") {
			return base, true
		}
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 extracted files, got %d", n)
	}
	for _, want := range []string{"ffmpeg", "avcodec.dll"} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Fatalf("%s missing: %v", want, err)
		}
	}
	for _, unwanted := range []string{"ffprobe", "x.h", "x.lib"} {
		if _, err := os.Stat(filepath.Join(out, unwanted)); err == nil {
			t.Fatalf("%s should not have been extracted", unwanted)
		}
	}
}

// A malicious archive must not be able to write outside the destination.
func TestUnzipSelectedIsZipSlipSafe(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "evil.zip")
	os.WriteFile(archive, zipWith(t, map[string]string{
		"bin/../../../../../../tmp/pwned": "x",
	}), 0o644)

	out := filepath.Join(dir, "out")
	n, err := unzipSelected(archive, out, func(name string) (string, bool) { return name, true })
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1, got %d", n)
	}
	// landed inside out/, flattened to its base name
	if _, err := os.Stat(filepath.Join(out, "pwned")); err != nil {
		t.Fatalf("expected out/pwned: %v", err)
	}
	entries, _ := os.ReadDir(out)
	if len(entries) != 1 {
		t.Fatalf("unexpected extra entries: %v", entries)
	}
}

// ─── go2rtc ───────────────────────────────────────────────────────────────────

func TestInstallGo2rtcHappyPath(t *testing.T) {
	dir := useTempAppDir(t)
	resetProgress()

	payload := zipWith(t, map[string]string{"go2rtc": fakeTool})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	oldURL, oldSum := go2rtcURL, go2rtcSHA256
	go2rtcURL, go2rtcSHA256 = srv.URL+"/go2rtc_win64.zip", sha(payload)
	defer func() { go2rtcURL, go2rtcSHA256 = oldURL, oldSum }()

	if err := installGo2rtc(t.TempDir()); err != nil {
		t.Fatalf("install: %v", err)
	}
	if !fileExists(go2rtcBinPath()) {
		t.Fatalf("binary missing at %s", go2rtcBinPath())
	}
	comps := componentList()
	if !comps[0].Installed {
		t.Fatalf("go2rtc should report installed")
	}
	if !strings.Contains(comps[0].Version, "fake tool") {
		t.Fatalf("version not read back: %q", comps[0].Version)
	}
	_ = dir
}

// A tampered download must be rejected and nothing installed.
func TestInstallGo2rtcRejectsBadChecksum(t *testing.T) {
	useTempAppDir(t)
	resetProgress()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipWith(t, map[string]string{"go2rtc": "tampered"}))
	}))
	defer srv.Close()

	oldURL, oldSum := go2rtcURL, go2rtcSHA256
	go2rtcURL = srv.URL + "/x.zip"
	go2rtcSHA256 = sha([]byte("something else entirely"))
	defer func() { go2rtcURL, go2rtcSHA256 = oldURL, oldSum }()

	err := installGo2rtc(t.TempDir())
	if err == nil {
		t.Fatal("expected a checksum error")
	}
	if !strings.Contains(err.Error(), "Prüfsumme") {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileExists(go2rtcBinPath()) {
		t.Fatal("binary was installed despite the bad checksum")
	}
}

func TestInstallGo2rtcHandlesHTTPError(t *testing.T) {
	useTempAppDir(t)
	resetProgress()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer srv.Close()

	old := go2rtcURL
	go2rtcURL = srv.URL + "/missing.zip"
	defer func() { go2rtcURL = old }()

	if err := installGo2rtc(t.TempDir()); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("want a 404 error, got %v", err)
	}
}

// ─── ffmpeg ───────────────────────────────────────────────────────────────────

func TestInstallFFmpegUsesPublishedChecksums(t *testing.T) {
	useTempAppDir(t)
	resetProgress()

	payload := zipWith(t, map[string]string{
		"ffmpeg-build/bin/ffmpeg":      fakeTool,
		"ffmpeg-build/bin/avcodec.dll": "dll-bytes",
		"ffmpeg-build/bin/ffprobe":     "skip me",
		"ffmpeg-build/doc/readme.txt":  "skip me",
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.sha256", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "deadbeef  some-other-file.zip\n%s  %s\n", sha(payload), ffmpegAsset)
	})
	mux.HandleFunc("/"+ffmpegAsset, func(w http.ResponseWriter, r *http.Request) { w.Write(payload) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldURL, oldSums := ffmpegURL, ffmpegChecksumsURL
	ffmpegURL, ffmpegChecksumsURL = srv.URL+"/"+ffmpegAsset, srv.URL+"/checksums.sha256"
	defer func() { ffmpegURL, ffmpegChecksumsURL = oldURL, oldSums }()

	if err := installFFmpeg(t.TempDir()); err != nil {
		t.Fatalf("install: %v", err)
	}
	if !fileExists(ffmpegBinPath()) {
		t.Fatalf("ffmpeg missing at %s", ffmpegBinPath())
	}
	if !fileExists(filepath.Join(ffmpegDir(), "avcodec.dll")) {
		t.Fatal("companion dll was not extracted")
	}
	if fileExists(filepath.Join(ffmpegDir(), "ffprobe")) {
		t.Fatal("ffprobe should have been skipped")
	}
}

func TestInstallFFmpegRejectsMismatch(t *testing.T) {
	useTempAppDir(t)
	resetProgress()

	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.sha256", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", sha([]byte("not what is served")), ffmpegAsset)
	})
	mux.HandleFunc("/"+ffmpegAsset, func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipWith(t, map[string]string{"b/bin/ffmpeg": fakeTool}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldURL, oldSums := ffmpegURL, ffmpegChecksumsURL
	ffmpegURL, ffmpegChecksumsURL = srv.URL+"/"+ffmpegAsset, srv.URL+"/checksums.sha256"
	defer func() { ffmpegURL, ffmpegChecksumsURL = oldURL, oldSums }()

	if err := installFFmpeg(t.TempDir()); err == nil || !strings.Contains(err.Error(), "Prüfsumme") {
		t.Fatalf("want checksum error, got %v", err)
	}
	if fileExists(ffmpegBinPath()) {
		t.Fatal("ffmpeg installed despite mismatch")
	}
}

func TestExpectedFFmpegSHAMissingEntry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "abc  unrelated.zip\n")
	}))
	defer srv.Close()
	old := ffmpegChecksumsURL
	ffmpegChecksumsURL = srv.URL
	defer func() { ffmpegChecksumsURL = old }()

	if _, err := expectedFFmpegSHA(); err == nil {
		t.Fatal("expected an error for the missing entry")
	}
}

// ─── Fortschritt ──────────────────────────────────────────────────────────────

func TestProgressReportsBytes(t *testing.T) {
	useTempAppDir(t)
	resetProgress()

	payload := bytes.Repeat([]byte("x"), 512*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprint(len(payload)))
		w.Write(payload)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "f.bin")
	sum, err := download(srv.URL, dest)
	if err != nil {
		t.Fatal(err)
	}
	if sum != sha(payload) {
		t.Fatal("hash mismatch")
	}
	p := progressSnapshot()
	if p.Total != int64(len(payload)) || p.Received != int64(len(payload)) {
		t.Fatalf("progress not tracked: %+v", p)
	}
}

// ─── HTTP-Endpunkte ───────────────────────────────────────────────────────────

func TestComponentsEndpointReportsMissing(t *testing.T) {
	useTempAppDir(t)
	mu.Lock()
	state.SetupDismissed = false
	mu.Unlock()

	rec := httptest.NewRecorder()
	handleComponents(rec, httptest.NewRequest("GET", "/api/components", nil))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var d struct {
		Missing   []string `json:"missing"`
		Dismissed bool     `json:"dismissed"`
	}
	json.Unmarshal(rec.Body.Bytes(), &d)
	found := false
	for _, m := range d.Missing {
		if m == "go2rtc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("go2rtc should be reported missing, got %v", d.Missing)
	}
	if d.Dismissed {
		t.Fatal("should not be dismissed yet")
	}
}

func TestDismissIsPersisted(t *testing.T) {
	dir := useTempAppDir(t)
	oldData := dataFile
	dataFile = filepath.Join(dir, "config.json")
	defer func() { dataFile = oldData }()

	mu.Lock()
	state.SetupDismissed = false
	mu.Unlock()

	rec := httptest.NewRecorder()
	handleComponentsDismiss(rec, httptest.NewRequest("POST", "/api/components/dismiss",
		strings.NewReader(`{"forever":true}`)))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}

	raw, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	if !strings.Contains(string(raw), `"setup_dismissed": true`) {
		t.Fatalf("flag not persisted: %s", raw)
	}

	rec2 := httptest.NewRecorder()
	handleComponents(rec2, httptest.NewRequest("GET", "/api/components", nil))
	if !strings.Contains(rec2.Body.String(), `"dismissed":true`) {
		t.Fatalf("dismissed not reported: %s", rec2.Body.String())
	}

	mu.Lock()
	state.SetupDismissed = false
	mu.Unlock()
}

// A second install request while one is running must be refused.
func TestInstallIsSingleFlight(t *testing.T) {
	useTempAppDir(t)
	dlMu.Lock()
	dlProgress = installProgress{Active: true, Step: "download"}
	dlMu.Unlock()
	defer resetProgress()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/components/install", strings.NewReader(`{"components":["go2rtc"]}`))
	handleComponentsInstall(rec, req)
	// on linux the platform guard fires first; both outcomes mean "not started"
	if rec.Code != http.StatusConflict && rec.Code != http.StatusNotImplemented {
		t.Fatalf("want 409 or 501, got %d", rec.Code)
	}
}

func TestEnvWithFFmpegPrependsPath(t *testing.T) {
	dir := useTempAppDir(t)
	os.MkdirAll(ffmpegDir(), 0o755)
	os.WriteFile(ffmpegBinPath(), []byte(fakeTool), 0o755)

	var got string
	for _, kv := range envWithFFmpeg() {
		if strings.HasPrefix(strings.ToUpper(kv), "PATH=") {
			got = kv
		}
	}
	if !strings.HasPrefix(got[5:], ffmpegDir()) {
		t.Fatalf("ffmpeg dir not prepended: %q", got)
	}
	if strings.Count(got, "PATH=") != 1 {
		t.Fatalf("PATH duplicated: %q", got)
	}
	_ = dir
	_ = time.Now
}
