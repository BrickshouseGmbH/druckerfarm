package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ─── EXTERNE KOMPONENTEN ──────────────────────────────────────────────────────
//
// go2rtc und ffmpeg werden nicht mehr ins Exe eingebettet, sondern bei Bedarf
// heruntergeladen. Die App läuft auch ohne beide: Drucker anlegen, MQTT-Status
// und Datei-Sync funktionieren, nur Video bzw. Snapshots fehlen dann.
//
//   go2rtc  → Video (WebRTC) und Snapshots
//   ffmpeg  → nur Snapshots; go2rtc braucht es, um H264 nach JPEG zu wandeln

const go2rtcVersion = "v1.9.14"

// Als Variablen, damit die Tests sie auf einen lokalen Server umbiegen koennen.
var (
	go2rtcURL = "https://github.com/AlexxIT/go2rtc/releases/download/" + go2rtcVersion + "/go2rtc_win64.zip"
	// Fester Release-Tag, deshalb ist der Hash pinnbar.
	go2rtcSHA256 = "dd4167d75cb04abe618855b7c71f8658bd009f60c1a71835d134d2c11c939907"

	// BtbN veroeffentlicht unter dem rollenden Tag "latest" staendig neu gebaute
	// Archive. Ein fester Hash wuerde nach jedem Rebuild brechen — stattdessen
	// wird die mitveroeffentlichte Pruefsummenliste geladen und daraus geprueft.
	ffmpegAsset        = "ffmpeg-master-latest-win64-lgpl-shared.zip"
	ffmpegURL          = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/" + ffmpegAsset
	ffmpegChecksumsURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/checksums.sha256"
)

type componentStatus struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Purpose    string `json:"purpose"`
	Installed  bool   `json:"installed"`
	Path       string `json:"path"`
	Version    string `json:"version"`
	DownloadMB int    `json:"download_mb"`
}

type installProgress struct {
	Active    bool   `json:"active"`
	Component string `json:"component"`
	Step      string `json:"step"` // download | extract | verify | done | error
	Received  int64  `json:"received"`
	Total     int64  `json:"total"`
	Done      bool   `json:"done"`
	Error     string `json:"error"`
}

var (
	dlMu       sync.Mutex
	dlProgress installProgress
	dlClient   = &http.Client{Timeout: 30 * time.Minute}
)

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func go2rtcBinPath() string { return filepath.Join(appDir, "go2rtc"+exeSuffix()) }
func ffmpegDir() string     { return filepath.Join(appDir, "ffmpeg") }
func ffmpegBinPath() string { return filepath.Join(ffmpegDir(), "ffmpeg"+exeSuffix()) }

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir() && st.Size() > 0
}

// toolVersion ruft "<bin> -version" auf und gibt die erste Zeile zurück. Dient
// zugleich als Funktionsprüfung nach dem Download.
func toolVersion(bin string, args ...string) string {
	if !fileExists(bin) {
		return ""
	}
	cmd := exec.Command(bin, args...)
	hideProcessWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if len(line) > 120 {
		line = line[:120]
	}
	return line
}

func componentList() []componentStatus {
	g2 := componentStatus{
		Key: "go2rtc", Label: "go2rtc " + go2rtcVersion,
		Purpose: "Video-Streams und Snapshots", DownloadMB: 7,
		Path: go2rtcBinPath(),
	}
	if fileExists(g2.Path) {
		g2.Installed = true
		g2.Version = toolVersion(g2.Path, "-version")
	}

	ff := componentStatus{
		Key: "ffmpeg", Label: "ffmpeg",
		Purpose: "Snapshots (H264 → JPEG)", DownloadMB: 85,
		Path: ffmpegBinPath(),
	}
	if fileExists(ff.Path) {
		ff.Installed = true
		ff.Version = toolVersion(ff.Path, "-version")
	} else if p, err := exec.LookPath("ffmpeg"); err == nil {
		// Schon systemweit vorhanden — dann kein Download nötig.
		ff.Installed = true
		ff.Path = p
		ff.Version = toolVersion(p, "-version")
	}

	return []componentStatus{g2, ff}
}

// envWithFFmpeg hängt das lokale ffmpeg-Verzeichnis vorne an den PATH des
// Kindprozesses, damit go2rtc es findet, ohne dass am System-PATH etwas
// geändert werden muss.
func envWithFFmpeg() []string {
	dir := ffmpegDir()
	if !fileExists(ffmpegBinPath()) {
		return os.Environ()
	}
	env := os.Environ()
	for i, kv := range env {
		if eq := strings.IndexByte(kv, '='); eq > 0 && strings.EqualFold(kv[:eq], "PATH") {
			env[i] = kv[:eq] + "=" + dir + string(os.PathListSeparator) + kv[eq+1:]
			return env
		}
	}
	return append(env, "PATH="+dir)
}

func setProgress(fn func(*installProgress)) {
	dlMu.Lock()
	fn(&dlProgress)
	dlMu.Unlock()
}

func progressSnapshot() installProgress {
	dlMu.Lock()
	defer dlMu.Unlock()
	return dlProgress
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	setProgress(func(pr *installProgress) { pr.Received = c.n })
	return n, err
}

func download(url, dest string) (string, error) {
	resp, err := dlClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d von %s", resp.StatusCode, url)
	}
	// Ohne Content-Length (chunked) lieber 0 melden als -1 — das UI zeigt dann
	// einen unbestimmten Balken statt einer negativen Prozentzahl.
	total := resp.ContentLength
	if total < 0 {
		total = 0
	}
	setProgress(func(pr *installProgress) { pr.Total = total; pr.Received = 0 })

	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	cr := &countingReader{r: resp.Body}
	if _, err := io.Copy(io.MultiWriter(f, h), cr); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// expectedFFmpegSHA holt die Prüfsummenliste des Releases und sucht den Eintrag
// für unser Archiv heraus.
func expectedFFmpegSHA() (string, error) {
	resp, err := dlClient.Get(ffmpegChecksumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Prüfsummen: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(body), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == ffmpegAsset {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", fmt.Errorf("kein Prüfsummen-Eintrag für %s", ffmpegAsset)
}

// unzipSelected entpackt genau die Einträge, für die pick einen Zieldateinamen
// liefert. Zip-Slip ist damit ausgeschlossen, weil der Zielname nie aus dem
// Archiv übernommen wird.
func unzipSelected(archive, destDir string, pick func(name string) (string, bool)) (int, error) {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return 0, err
	}
	defer zr.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}

	n := 0
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		out, ok := pick(f.Name)
		if !ok {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return n, err
		}
		dst := filepath.Join(destDir, filepath.Base(out))
		w, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			rc.Close()
			return n, err
		}
		_, err = io.Copy(w, io.LimitReader(rc, 1<<30))
		w.Close()
		rc.Close()
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func installGo2rtc(tmpDir string) error {
	setProgress(func(p *installProgress) { p.Component = "go2rtc"; p.Step = "download" })

	archive := filepath.Join(tmpDir, "go2rtc.zip")
	sum, err := download(go2rtcURL, archive)
	if err != nil {
		return err
	}
	setProgress(func(p *installProgress) { p.Step = "verify" })
	if sum != go2rtcSHA256 {
		return fmt.Errorf("Prüfsumme stimmt nicht (erwartet %s…, erhalten %s…)", go2rtcSHA256[:12], sum[:12])
	}

	setProgress(func(p *installProgress) { p.Step = "extract" })
	stopGo2rtc() // laufende Instanz beenden, sonst ist die Datei gesperrt
	staging := filepath.Join(tmpDir, "go2rtc-out")
	n, err := unzipSelected(archive, staging, func(name string) (string, bool) {
		base := strings.ToLower(filepath.Base(name))
		if base == "go2rtc.exe" || base == "go2rtc" {
			return "go2rtc" + exeSuffix(), true
		}
		return "", false
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("go2rtc-Binary im Archiv nicht gefunden")
	}
	if err := os.Rename(filepath.Join(staging, "go2rtc"+exeSuffix()), go2rtcBinPath()); err != nil {
		return err
	}

	setProgress(func(p *installProgress) { p.Step = "verify" })
	if v := toolVersion(go2rtcBinPath(), "-version"); v == "" {
		return fmt.Errorf("go2rtc lässt sich nicht starten")
	}
	rememberComponentHash("go2rtc", sum)
	return nil
}

func installFFmpeg(tmpDir string) error {
	setProgress(func(p *installProgress) { p.Component = "ffmpeg"; p.Step = "verify"; p.Received = 0; p.Total = 0 })
	want, err := expectedFFmpegSHA()
	if err != nil {
		return err
	}

	setProgress(func(p *installProgress) { p.Step = "download" })
	archive := filepath.Join(tmpDir, ffmpegAsset)
	sum, err := download(ffmpegURL, archive)
	if err != nil {
		return err
	}
	setProgress(func(p *installProgress) { p.Step = "verify" })
	if sum != want {
		return fmt.Errorf("Prüfsumme stimmt nicht (erwartet %s…, erhalten %s…)", want[:12], sum[:12])
	}

	setProgress(func(p *installProgress) { p.Step = "extract" })
	staging := filepath.Join(tmpDir, "ffmpeg-out")
	// Aus dem Shared-Build nur bin/ffmpeg.exe und die zugehörigen DLLs;
	// ffplay, ffprobe, Header und Import-Libs bleiben draußen.
	n, err := unzipSelected(archive, staging, func(name string) (string, bool) {
		p := strings.ToLower(filepath.ToSlash(name))
		if !strings.Contains(p, "/bin/") {
			return "", false
		}
		base := filepath.Base(p)
		if base == "ffmpeg"+exeSuffix() || strings.HasSuffix(base, ".dll") {
			return base, true
		}
		return "", false
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("ffmpeg-Binary im Archiv nicht gefunden")
	}

	if err := os.MkdirAll(ffmpegDir(), 0o755); err != nil {
		return err
	}
	entries, _ := os.ReadDir(staging)
	for _, e := range entries {
		if err := os.Rename(filepath.Join(staging, e.Name()), filepath.Join(ffmpegDir(), e.Name())); err != nil {
			return err
		}
	}

	setProgress(func(p *installProgress) { p.Step = "verify" })
	if v := toolVersion(ffmpegBinPath(), "-version"); v == "" {
		return fmt.Errorf("ffmpeg lässt sich nicht starten")
	}
	rememberComponentHash("ffmpeg", sum)
	return nil
}

func runInstall(keys []string) {
	tmpDir := filepath.Join(appDir, ".download")
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		setProgress(func(p *installProgress) { p.Active = false; p.Done = true; p.Step = "error"; p.Error = err.Error() })
		return
	}
	defer os.RemoveAll(tmpDir)

	var failed error
	for _, k := range keys {
		var err error
		switch k {
		case "go2rtc":
			err = installGo2rtc(tmpDir)
		case "ffmpeg":
			err = installFFmpeg(tmpDir)
		default:
			continue
		}
		if err != nil {
			failed = fmt.Errorf("%s: %v", k, err)
			break
		}
	}

	if failed != nil {
		log.Printf("❌ Komponenten-Download: %v", failed)
		setProgress(func(p *installProgress) {
			p.Active = false
			p.Done = true
			p.Step = "error"
			p.Error = failed.Error()
		})
		return
	}

	setProgress(func(p *installProgress) {
		p.Active = false
		p.Done = true
		p.Step = "done"
		p.Error = ""
	})
	log.Printf("✅ Komponenten installiert: %s", strings.Join(keys, ", "))

	// go2rtc (neu) starten, damit es das frische ffmpeg im PATH sieht
	go func() {
		stopGo2rtc()
		time.Sleep(500 * time.Millisecond)
		startGo2rtc()
	}()
}

// ─── HTTP ─────────────────────────────────────────────────────────────────────

func handleComponents(w http.ResponseWriter, r *http.Request) {
	comps := componentList()
	missing := []string{}
	for _, c := range comps {
		if !c.Installed {
			missing = append(missing, c.Key)
		}
	}
	mu.Lock()
	dismissed := state.SetupDismissed
	mu.Unlock()

	writeJSON(w, map[string]any{
		"components":   comps,
		"missing":      missing,
		"dismissed":    dismissed,
		"downloadable": runtime.GOOS == "windows" && runtime.GOARCH == "amd64",
		"dir":          appDir,
		"window_mode":  appWindowMode,
		"app_port":     appPort,
	})
}

func handleComponentsInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" {
		http.Error(w, "automatischer Download nur für Windows x64", http.StatusNotImplemented)
		return
	}

	var body struct {
		Components []string `json:"components"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if len(body.Components) == 0 {
		for _, c := range componentList() {
			if !c.Installed {
				body.Components = append(body.Components, c.Key)
			}
		}
	}
	if len(body.Components) == 0 {
		http.Error(w, "nichts zu installieren", http.StatusBadRequest)
		return
	}

	dlMu.Lock()
	if dlProgress.Active {
		dlMu.Unlock()
		http.Error(w, "Download läuft bereits", http.StatusConflict)
		return
	}
	dlProgress = installProgress{Active: true, Step: "download", Component: body.Components[0]}
	dlMu.Unlock()

	go runInstall(body.Components)
	writeJSON(w, map[string]any{"started": body.Components})
}

func handleComponentsProgress(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, progressSnapshot())
}

func handleComponentsDismiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Forever bool `json:"forever"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Forever {
		mu.Lock()
		state.SetupDismissed = true
		mu.Unlock()
		saveState()
	}
	writeJSON(w, map[string]any{"ok": true})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

// ─── AKTUALISIERUNG DER KOMPONENTEN ───────────────────────────────────────────
//
// go2rtc traegt eine Versionsnummer, die sich mit dem neuesten GitHub-Release
// vergleichen laesst. ffmpeg von BtbN nicht — dort wird unter einem rollenden
// Tag staendig neu gebaut. Deshalb wird beim Installieren die Pruefsumme des
// Archivs gemerkt und spaeter mit der veroeffentlichten verglichen.

type componentUpdate struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Installed string `json:"installed"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	Note      string `json:"note"`
}

func rememberComponentHash(key, sum string) {
	mu.Lock()
	if state.ComponentSHA == nil {
		state.ComponentSHA = map[string]string{}
	}
	state.ComponentSHA[key] = sum
	mu.Unlock()
	saveState()
}

func componentHash(key string) string {
	mu.Lock()
	defer mu.Unlock()
	return state.ComponentSHA[key]
}

// latestGo2rtcTag fragt das neueste Release ab.
func latestGo2rtcTag() (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/AlexxIT/go2rtc/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Druckerfarm")
	resp, err := dlClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub nicht erreichbar: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub HTTP %d", resp.StatusCode)
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rel); err != nil {
		return "", err
	}
	return strings.TrimSpace(rel.TagName), nil
}

// installedGo2rtcVersion liest die Version aus dem Programm selbst.
func installedGo2rtcVersion() string {
	return parseGo2rtcVersionLine(toolVersion(go2rtcBinPath(), "-version"))
}

// parseGo2rtcVersionLine zieht die Version aus "go2rtc version 1.9.14+dev.… windows/amd64".
func parseGo2rtcVersionLine(line string) string {
	f := strings.Fields(line)
	for i, w := range f {
		if strings.EqualFold(w, "version") && i+1 < len(f) {
			v := strings.SplitN(f[i+1], "+", 2)[0]
			if v == "" || !(v[0] >= '0' && v[0] <= '9') {
				return ""
			}
			return "v" + v
		}
	}
	return ""
}

func checkComponentUpdates() []componentUpdate {
	out := []componentUpdate{}

	// go2rtc
	g2 := componentUpdate{Key: "go2rtc", Label: "go2rtc"}
	if fileExists(go2rtcBinPath()) {
		g2.Installed = installedGo2rtcVersion()
		if tag, err := latestGo2rtcTag(); err == nil {
			g2.Latest = tag
			g2.Available = g2.Installed != "" && newerVersion(g2.Installed, tag)
		} else {
			g2.Note = err.Error()
		}
	} else {
		g2.Note = "nicht installiert"
	}
	out = append(out, g2)

	// ffmpeg — Vergleich ueber die Pruefsumme des Archivs
	ff := componentUpdate{Key: "ffmpeg", Label: "ffmpeg"}
	if fileExists(ffmpegBinPath()) {
		have := componentHash("ffmpeg")
		if have == "" {
			ff.Note = "Prüfsumme der Installation unbekannt — beim nächsten Nachladen wird sie gemerkt"
		} else if want, err := expectedFFmpegSHA(); err != nil {
			ff.Note = err.Error()
		} else {
			ff.Installed = have[:12]
			ff.Latest = want[:12]
			ff.Available = have != want
		}
	} else {
		ff.Note = "nicht installiert"
	}
	out = append(out, ff)

	return out
}

// installGo2rtcLatest holt die neueste veroeffentlichte Fassung. Anders als bei
// der Erstinstallation gibt es hier keine fest hinterlegte Pruefsumme — geprueft
// wird stattdessen, dass die Datei startet und die erwartete Version meldet.
func installGo2rtcLatest(tmpDir, tag string) error {
	setProgress(func(p *installProgress) { p.Component = "go2rtc"; p.Step = "download" })

	url := "https://github.com/AlexxIT/go2rtc/releases/download/" + tag + "/go2rtc_win64.zip"
	if exeSuffix() == "" {
		url = "https://github.com/AlexxIT/go2rtc/releases/download/" + tag + "/go2rtc_linux_amd64"
	}
	archive := filepath.Join(tmpDir, "go2rtc-latest")
	sum, err := download(url, archive)
	if err != nil {
		return err
	}

	setProgress(func(p *installProgress) { p.Step = "extract" })
	stopGo2rtc()
	staging := filepath.Join(tmpDir, "g2-latest-out")

	if strings.HasSuffix(url, ".zip") {
		n, err := unzipSelected(archive, staging, func(name string) (string, bool) {
			base := strings.ToLower(filepath.Base(name))
			if base == "go2rtc.exe" || base == "go2rtc" {
				return "go2rtc" + exeSuffix(), true
			}
			return "", false
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("go2rtc-Binary im Archiv nicht gefunden")
		}
	} else {
		os.MkdirAll(staging, 0o755)
		if err := copyFile(archive, filepath.Join(staging, "go2rtc")); err != nil {
			return err
		}
	}

	newBin := filepath.Join(staging, "go2rtc"+exeSuffix())
	os.Chmod(newBin, 0o755)

	setProgress(func(p *installProgress) { p.Step = "verify" })
	line := toolVersion(newBin, "-version")
	if line == "" {
		return fmt.Errorf("heruntergeladenes go2rtc startet nicht")
	}
	want := strings.TrimPrefix(tag, "v")
	if !strings.Contains(line, want) {
		return fmt.Errorf("gemeldete Version passt nicht zu %s: %q", tag, line)
	}

	if err := os.Rename(newBin, go2rtcBinPath()); err != nil {
		return err
	}
	rememberComponentHash("go2rtc", sum)
	log.Printf("go2rtc auf %s aktualisiert", tag)
	return nil
}

func runComponentUpdate(keys []string) {
	tmpDir := filepath.Join(appDir, ".update-components")
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		setProgress(func(p *installProgress) { p.Active = false; p.Done = true; p.Step = "error"; p.Error = err.Error() })
		return
	}
	defer os.RemoveAll(tmpDir)

	var failed error
	for _, k := range keys {
		var err error
		switch k {
		case "go2rtc":
			tag, terr := latestGo2rtcTag()
			if terr != nil {
				err = terr
			} else {
				err = installGo2rtcLatest(tmpDir, tag)
			}
		case "ffmpeg":
			err = installFFmpeg(tmpDir)
		}
		if err != nil {
			failed = fmt.Errorf("%s: %v", k, err)
			break
		}
	}

	if failed != nil {
		log.Printf("❌ Komponenten-Aktualisierung: %v", failed)
		setProgress(func(p *installProgress) {
			p.Active = false
			p.Done = true
			p.Step = "error"
			p.Error = failed.Error()
		})
	} else {
		setProgress(func(p *installProgress) { p.Active = false; p.Done = true; p.Step = "done"; p.Error = "" })
	}

	// In JEDEM Fall wieder starten. Zum Ersetzen der Datei wird go2rtc beendet —
	// brach die Aktualisierung danach ab, blieb es vorher fuer immer gestoppt,
	// und damit waren Videos und Snapshots weg.
	go func() {
		stopGo2rtc()
		time.Sleep(500 * time.Millisecond)
		startGo2rtc()
	}()
}

func handleComponentUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Components []string `json:"components"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Components) == 0 {
			for _, u := range checkComponentUpdates() {
				if u.Available {
					body.Components = append(body.Components, u.Key)
				}
			}
		}
		if len(body.Components) == 0 {
			http.Error(w, "nichts zu aktualisieren", http.StatusBadRequest)
			return
		}

		dlMu.Lock()
		if dlProgress.Active {
			dlMu.Unlock()
			http.Error(w, "Vorgang läuft bereits", http.StatusConflict)
			return
		}
		dlProgress = installProgress{Active: true, Step: "download", Component: body.Components[0]}
		dlMu.Unlock()

		go runComponentUpdate(body.Components)
		writeJSON(w, map[string]any{"started": body.Components})
		return
	}

	writeJSON(w, map[string]any{"updates": checkComponentUpdates()})
}
