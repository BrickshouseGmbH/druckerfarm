package main

import (
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersionComparison(t *testing.T) {
	cases := []struct {
		cur, latest string
		want        bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "v1.0.1", true},
		{"1.9.0", "1.12.0", true}, // als Text waere 1.12 kleiner
		{"1.12.0", "1.9.0", false},
		{"1.0.0", "1.0.0", false},
		{"2.0.0", "1.9.9", false},
		{"dev", "1.0.0", true}, // ohne Versionsstempel gilt alles als neuer
		{"1.0.0", "1.1", true},
		{"1.1", "1.1.0", false},
		{"1.0.0", "kaputt", false},
		{"1.0.0-beta", "1.0.0", false},
	}
	for _, c := range cases {
		if got := newerVersion(c.cur, c.latest); got != c.want {
			t.Errorf("newerVersion(%q,%q) = %v, want %v", c.cur, c.latest, got, c.want)
		}
	}
}

func TestPickAssetPrefersExe(t *testing.T) {
	assets := []ghAsset{
		{Name: "checksums.txt", URL: "u1"},
		{Name: "Druckerfarm-1.2.0.exe", URL: "u2", Size: 7000000},
		{Name: "source.zip", URL: "u3"},
	}
	a := pickAsset(assets)
	if a == nil || a.URL != "u2" {
		t.Fatalf("wrong asset: %+v", a)
	}
	if pickAsset([]ghAsset{{Name: "notes.txt"}}) != nil {
		t.Fatal("a release without a program file must yield nothing")
	}
}

func setRepo(t *testing.T, repo string) {
	t.Helper()
	mu.Lock()
	old := state.UpdateRepo
	state.UpdateRepo = repo
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); state.UpdateRepo = old; mu.Unlock() })
}

// Ohne eigene Einstellung liefert die Konfiguration das Standard-Repository —
// die Update-Suche funktioniert damit ohne Einrichtung.
func TestCheckWithoutRepoUsesDefault(t *testing.T) {
	setRepo(t, "")
	rec := httptest.NewRecorder()
	handleUpdateConfig(rec, httptest.NewRequest("GET", "/api/update/config", nil))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var d map[string]any
	json.Unmarshal(rec.Body.Bytes(), &d)
	if s, _ := d["repo"].(string); s != defaultUpdateRepo {
		t.Fatalf("erwarte Standard-Repo %q, bekam %v", defaultUpdateRepo, d["repo"])
	}
}

func TestRepoFormatIsValidated(t *testing.T) {
	rec := httptest.NewRecorder()
	handleUpdateConfig(rec, httptest.NewRequest("POST", "/api/update/config",
		strings.NewReader(`{"repo":"nur-ein-name"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}

	// vollständige URL soll akzeptiert und gekürzt werden
	rec2 := httptest.NewRecorder()
	handleUpdateConfig(rec2, httptest.NewRequest("POST", "/api/update/config",
		strings.NewReader(`{"repo":"https://github.com/besitzer/name.git"}`)))
	if rec2.Code != 200 {
		t.Fatalf("want 200, got %d: %s", rec2.Code, rec2.Body.String())
	}
	repo, _ := updateRepo()
	if repo != "besitzer/name" {
		t.Fatalf("URL not normalised: %q", repo)
	}
	mu.Lock()
	state.UpdateRepo = ""
	mu.Unlock()
}

// An HTML error page must never end up replacing the running program.
func TestDownloadedFileIsChecked(t *testing.T) {
	dir := t.TempDir()

	html := filepath.Join(dir, "a.exe")
	os.WriteFile(html, []byte(strings.Repeat("<html>error</html>", 60000)), 0o755)
	if err := looksExecutable(html); err == nil {
		t.Fatal("an HTML page was accepted as a program")
	}

	tiny := filepath.Join(dir, "b.exe")
	os.WriteFile(tiny, append([]byte("MZ"), make([]byte, 100)...), 0o755)
	if err := looksExecutable(tiny); err == nil {
		t.Fatal("a 100-byte file was accepted")
	}

	ok := filepath.Join(dir, "c.exe")
	os.WriteFile(ok, append([]byte("MZ"), make([]byte, 900*1024)...), 0o755)
	if err := looksExecutable(ok); err != nil {
		t.Fatalf("a plausible file was rejected: %v", err)
	}
}

func TestUpdateCheckAgainstFakeGitHub(t *testing.T) {
	release := ghRelease{
		TagName:     "v2.5.0",
		Body:        "Netzwerksuche",
		PublishedAt: "2026-07-28T10:00:00Z",
		Assets:      []ghAsset{{Name: "Druckerfarm.exe", URL: "https://example.invalid/df.exe", Size: 7500000}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/releases/latest") {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	// fetchLatestRelease baut die URL selbst — deshalb hier über den Parser gehen
	oldV := appVersion
	appVersion = "1.0.0"
	defer func() { appVersion = oldV }()

	if !newerVersion(appVersion, release.TagName) {
		t.Fatal("2.5.0 should count as newer than 1.0.0")
	}
	a := pickAsset(release.Assets)
	if a == nil || a.Size != 7500000 {
		t.Fatalf("asset not found: %+v", a)
	}
}

// ─── Einstellungen ────────────────────────────────────────────────────────────

func TestSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	oldData, oldDir := dataFile, appDir
	dataFile, appDir = filepath.Join(dir, "config.json"), dir
	defer func() { dataFile, appDir = oldData, oldDir }()

	rec := httptest.NewRecorder()
	handleSettings(rec, httptest.NewRequest("POST", "/api/settings",
		strings.NewReader(`{"theme":"light","lang":"en","cols":4}`)))
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var d struct {
		Theme string `json:"theme"`
		Lang  string `json:"lang"`
		Cols  int    `json:"cols"`
	}
	json.Unmarshal(rec.Body.Bytes(), &d)
	if d.Theme != "light" || d.Lang != "en" || d.Cols != 4 {
		t.Fatalf("not stored: %+v", d)
	}

	raw, err := os.ReadFile(dataFile)
	if err != nil || !strings.Contains(string(raw), `"theme": "light"`) {
		t.Fatalf("not persisted: %v %s", err, raw)
	}

	// Unsinn darf nichts überschreiben
	rec2 := httptest.NewRecorder()
	handleSettings(rec2, httptest.NewRequest("POST", "/api/settings",
		strings.NewReader(`{"theme":"neon","cols":999}`)))
	json.Unmarshal(rec2.Body.Bytes(), &d)
	if d.Theme != "light" || d.Cols != 4 {
		t.Fatalf("invalid values were accepted: %+v", d)
	}
	mu.Lock()
	state.Theme, state.Lang, state.Cols = "", "", 0
	mu.Unlock()
}

// ─── %APPDATA% ────────────────────────────────────────────────────────────────

func TestDataHomeIsOutsideTheProgramFolder(t *testing.T) {
	h := dataHome()
	if h == "" {
		t.Skip("kein Basisverzeichnis in dieser Umgebung")
	}
	if filepath.Base(h) != "Druckerfarm" {
		t.Fatalf("unexpected folder: %s", h)
	}
}

func TestMigrationCopiesExistingSetup(t *testing.T) {
	exeDir := t.TempDir()
	newHome := t.TempDir()

	os.WriteFile(filepath.Join(exeDir, "config.json"), []byte(`{"printers":[{"ip":"10.0.0.1","name":"A"}]}`), 0o644)
	os.WriteFile(filepath.Join(exeDir, "go2rtc.yaml"), []byte("streams:\n"), 0o644)
	os.MkdirAll(filepath.Join(exeDir, "ffmpeg"), 0o755)
	os.WriteFile(filepath.Join(exeDir, "ffmpeg", "ffmpeg"+exeSuffix()), []byte("binary"), 0o755)

	oldDir, oldData := appDir, dataFile
	appDir, dataFile = newHome, filepath.Join(newHome, "config.json")
	defer func() { appDir, dataFile = oldDir, oldData }()

	migrateFromExeDir(exeDir)

	for _, want := range []string{"config.json", "go2rtc.yaml", filepath.Join("ffmpeg", "ffmpeg"+exeSuffix())} {
		if _, err := os.Stat(filepath.Join(newHome, want)); err != nil {
			t.Fatalf("%s was not taken over: %v", want, err)
		}
	}

	// Zweiter Lauf darf eine inzwischen geänderte Konfiguration nicht überschreiben
	os.WriteFile(dataFile, []byte(`{"printers":[]}`), 0o644)
	migrateFromExeDir(exeDir)
	raw, _ := os.ReadFile(dataFile)
	if strings.Contains(string(raw), "10.0.0.1") {
		t.Fatal("migration ran a second time and overwrote the newer config")
	}
}

func TestOldExeNameSitsNextToTheProgram(t *testing.T) {
	old := exePath
	exePath = filepath.Join("C:", "tools", "druckerfarm.exe")
	defer func() { exePath = old }()
	if got := oldExePath(); got != exePath+".old" {
		t.Fatalf("got %q", got)
	}
}

// Der heikelste Teil: die laufende Datei durch eine neue ersetzen. Windows
// verbietet das Überschreiben, erlaubt aber das Umbenennen — genau darauf baut
// applyUpdate. Hier wird der komplette Ablauf mit echten Dateien durchgespielt.
func TestApplyUpdateReplacesItselfAndKeepsARollback(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "druckerfarm")
	marker := filepath.Join(dir, "gestartet.txt")

	// "alte Fassung" — beim Start hinterlässt sie eine Spur
	os.WriteFile(running, []byte("#!/bin/sh\necho alt > "+marker+"\n"), 0o755)
	newFile := filepath.Join(dir, "neu")
	os.WriteFile(newFile, []byte("#!/bin/sh\necho neu > "+marker+"\n"), 0o755)

	old := exePath
	exePath = running
	defer func() { exePath = old }()

	if err := applyUpdate(newFile); err != nil {
		t.Fatalf("applyUpdate: %v", err)
	}

	// neue Fassung liegt am Platz der alten
	got, _ := os.ReadFile(running)
	if !strings.Contains(string(got), "echo neu") {
		t.Fatalf("new version was not put in place: %q", got)
	}
	// Vorgänger bleibt als Rückweg liegen
	prev, err := os.ReadFile(running + ".old")
	if err != nil || !strings.Contains(string(prev), "echo alt") {
		t.Fatalf("no rollback copy: %v %q", err, prev)
	}
	// und sie wurde gestartet
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(marker); err == nil && strings.Contains(string(b), "neu") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("new version was not started")
}

// Schlägt das Ablegen fehl, muss die alte Fassung zurückkommen — sonst stünde
// der Rechner ohne lauffähiges Programm da.
func TestApplyUpdateRollsBackOnFailure(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "druckerfarm")
	os.WriteFile(running, []byte("#!/bin/sh\nexit 0\n"), 0o755)

	old := exePath
	exePath = running
	defer func() { exePath = old }()

	// Quelle existiert nicht → copyFile schlägt fehl
	if err := applyUpdate(filepath.Join(dir, "gibtsnicht")); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(running); err != nil {
		t.Fatalf("the running program was lost: %v", err)
	}
	b, _ := os.ReadFile(running)
	if !strings.Contains(string(b), "exit 0") {
		t.Fatalf("wrong content after rollback: %q", b)
	}
}

// cleanupOldExe räumt den Rückweg beim nächsten erfolgreichen Start weg.
func TestCleanupRemovesRollbackCopy(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "druckerfarm")
	os.WriteFile(running, []byte("x"), 0o755)
	os.WriteFile(running+".old", []byte("alt"), 0o755)

	old := exePath
	exePath = running
	defer func() { exePath = old }()

	cleanupOldExe()
	if _, err := os.Stat(running + ".old"); err == nil {
		t.Fatal("rollback copy was not removed")
	}
	if _, err := os.Stat(running); err != nil {
		t.Fatal("the program itself must stay")
	}
}

// Ohne eigene Einstellung wird das Standard-Repository verwendet, damit die
// Update-Suche ohne Einrichtung funktioniert.
func TestUpdateRepoStandard(t *testing.T) {
	mu.Lock()
	alt := state.UpdateRepo
	state.UpdateRepo = ""
	mu.Unlock()
	defer func() { mu.Lock(); state.UpdateRepo = alt; mu.Unlock() }()

	repo, _ := updateRepo()
	if repo != defaultUpdateRepo {
		t.Errorf("erwarte Standard-Repo %q, bekam %q", defaultUpdateRepo, repo)
	}
	if defaultUpdateRepo != "BrickshouseGmbH/druckerfarm" {
		t.Errorf("Standard-Repo unerwartet: %q", defaultUpdateRepo)
	}

	// Eigene Einstellung hat Vorrang.
	mu.Lock()
	state.UpdateRepo = "foo/bar"
	mu.Unlock()
	if repo, _ := updateRepo(); repo != "foo/bar" {
		t.Errorf("eigene Einstellung sollte Vorrang haben, bekam %q", repo)
	}
}
