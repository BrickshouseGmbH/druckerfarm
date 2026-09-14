package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── SELBSTAKTUALISIERUNG ─────────────────────────────────────────────────────
//
// Die Version wird beim Bauen eingestempelt:
//   go build -ldflags "-X main.appVersion=1.4.0" ...
//
// Bezugsquelle sind GitHub-Releases. Das Repository steht in der Konfiguration
// und laesst sich in den Einstellungen setzen, damit kein neuer Build noetig ist.
//
// Zum Ersetzen: Windows verbietet das Ueberschreiben einer laufenden Datei,
// erlaubt aber das Umbenennen. Deshalb benennt sich die App selbst in
// *.old.exe um, schreibt die neue Fassung an den frei gewordenen Pfad, startet
// sie und beendet sich. Beim naechsten Start verschwindet die alte Datei — bis
// dahin ist sie der Rueckweg.

var appVersion = "dev"

type updateInfo struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	URL       string `json:"url"`
	Notes     string `json:"notes"`
	Repo      string `json:"repo"`
	Published string `json:"published"`
	Size      int64  `json:"size"`
}

type updateProgress struct {
	Active   bool   `json:"active"`
	Step     string `json:"step"` // download | verify | replace | done | error
	Received int64  `json:"received"`
	Total    int64  `json:"total"`
	Done     bool   `json:"done"`
	Error    string `json:"error"`
}

var (
	updMu       sync.Mutex
	updProgress updateProgress
	updClient   = &http.Client{Timeout: 15 * time.Minute}
)

func setUpdProgress(fn func(*updateProgress)) {
	updMu.Lock()
	fn(&updProgress)
	updMu.Unlock()
}

// ─── Versionsvergleich ────────────────────────────────────────────────────────

var versionPart = regexp.MustCompile(`\d+`)

// parseVersion macht aus "v1.12.3-beta" die Zahlenfolge [1 12 3]. Vergleich
// erfolgt zahlenweise, damit 1.12 groesser ist als 1.9 (als Text waere es das nicht).
func parseVersion(s string) []int {
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "v"))
	if i := strings.IndexAny(s, "-+ "); i > 0 {
		s = s[:i]
	}
	out := []int{}
	for _, m := range versionPart.FindAllString(s, 4) {
		n, err := strconv.Atoi(m)
		if err != nil {
			break
		}
		out = append(out, n)
	}
	return out
}

// newerVersion sagt, ob b neuer ist als a.
func newerVersion(a, b string) bool {
	va, vb := parseVersion(a), parseVersion(b)
	if len(vb) == 0 {
		return false
	}
	if len(va) == 0 {
		return true // laufende Fassung ohne Versionsstempel ("dev")
	}
	for i := 0; i < len(va) || i < len(vb); i++ {
		x, y := 0, 0
		if i < len(va) {
			x = va[i]
		}
		if i < len(vb) {
			y = vb[i]
		}
		if y != x {
			return y > x
		}
	}
	return false
}

// ─── GitHub ───────────────────────────────────────────────────────────────────

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

type ghRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt string    `json:"published_at"`
	Assets      []ghAsset `json:"assets"`
}

// defaultUpdateRepo ist die Bezugsquelle fuer Programm-Updates, wenn der Anwender
// nichts eigenes hinterlegt hat. So funktioniert die Update-Suche ohne
// Einrichtung; ueber die Einstellungen laesst sie sich weiterhin ueberschreiben.
const defaultUpdateRepo = "BrickshouseGmbH/druckerfarm"

func updateRepo() (repo, token string) {
	mu.Lock()
	defer mu.Unlock()
	repo = strings.TrimSpace(state.UpdateRepo)
	if repo == "" {
		repo = defaultUpdateRepo
	}
	return repo, strings.TrimSpace(state.UpdateToken)
}

// pickAsset waehlt die Programmdatei aus den Release-Anhaengen. Bewusst
// unabhaengig vom System, auf dem der Code gerade laeuft: veroeffentlicht wird
// eine Windows-Datei, und danach wird gesucht — sonst faende ein Testlauf unter
// Linux die .exe nicht.
func pickAsset(assets []ghAsset) *ghAsset {
	for i := range assets {
		if strings.HasSuffix(strings.ToLower(assets[i].Name), ".exe") {
			return &assets[i]
		}
	}
	for i := range assets {
		if !strings.Contains(assets[i].Name, ".") {
			return &assets[i]
		}
	}
	return nil
}

func fetchLatestRelease() (*ghRelease, error) {
	repo, token := updateRepo()
	if repo == "" {
		return nil, fmt.Errorf("kein Repository hinterlegt")
	}
	if strings.Count(repo, "/") != 1 {
		return nil, fmt.Errorf("Repository muss als besitzer/name angegeben werden, nicht %q", repo)
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Druckerfarm/"+appVersion)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := updClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub nicht erreichbar: %v", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("Repository %s hat keine Releases (oder ist privat und der Token fehlt)", repo)
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("GitHub verweigert den Zugriff (HTTP %d) — Token pruefen", resp.StatusCode)
	default:
		return nil, fmt.Errorf("GitHub HTTP %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func checkUpdate() (updateInfo, error) {
	repo, _ := updateRepo()
	info := updateInfo{Current: appVersion, Repo: repo}

	rel, err := fetchLatestRelease()
	if err != nil {
		return info, err
	}
	info.Latest = strings.TrimSpace(rel.TagName)
	info.Notes = rel.Body
	info.Published = rel.PublishedAt

	if a := pickAsset(rel.Assets); a != nil {
		info.URL = a.URL
		info.Size = a.Size
	}
	info.Available = info.URL != "" && newerVersion(appVersion, info.Latest)
	return info, nil
}

// ─── Herunterladen und ersetzen ───────────────────────────────────────────────

func downloadUpdate(url, dest string) error {
	repo, token := updateRepo()
	_ = repo
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Druckerfarm/"+appVersion)
	req.Header.Set("Accept", "application/octet-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := updClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength
	if total < 0 {
		total = 0
	}
	setUpdProgress(func(p *updateProgress) { p.Total = total; p.Received = 0; p.Step = "download" })

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 256*1024)
	var got int64
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			got += int64(n)
			setUpdProgress(func(p *updateProgress) { p.Received = got })
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	return nil
}

// looksExecutable prueft, ob die geladene Datei plausibel ist. Ohne diese
// Pruefung wuerde eine HTML-Fehlerseite als Programmdatei an den Platz der
// laufenden App wandern.
func looksExecutable(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.Size() < 500*1024 {
		return fmt.Errorf("Datei ist mit %d Bytes zu klein fuer ein Programm", st.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	head := make([]byte, 2)
	if _, err := io.ReadFull(f, head); err != nil {
		return err
	}
	// "MZ" steht am Anfang jeder Windows-Programmdatei. Ohne diese Pruefung
	// koennte eine HTML-Fehlerseite passender Groesse durchrutschen.
	if string(head) != "MZ" {
		return fmt.Errorf("keine Windows-Programmdatei (Kennung %q)", string(head))
	}
	return nil
}

func oldExePath() string { return exePath + ".old" }

// cleanupOldExe raeumt die Vorgaengerversion nach einem geglueckten Start weg.
func cleanupOldExe() {
	if exePath == "" {
		return
	}
	if err := os.Remove(oldExePath()); err == nil {
		log.Printf("Vorgaengerversion entfernt: %s", filepath.Base(oldExePath()))
	}
}

// applyUpdate ersetzt die laufende Programmdatei und startet sie neu.
func applyUpdate(newFile string) error {
	if exePath == "" {
		return fmt.Errorf("eigener Programmpfad unbekannt")
	}
	setUpdProgress(func(p *updateProgress) { p.Step = "replace" })

	// Reste eines frueheren Updates wegraeumen, sonst schlaegt das Umbenennen fehl
	os.Remove(oldExePath())

	// Umbenennen der laufenden Datei ist erlaubt, Ueberschreiben nicht
	if err := os.Rename(exePath, oldExePath()); err != nil {
		return fmt.Errorf("konnte laufende Datei nicht beiseitelegen: %v", err)
	}
	if err := copyFile(newFile, exePath); err != nil {
		os.Rename(oldExePath(), exePath) // zurueck auf Anfang
		return fmt.Errorf("neue Fassung konnte nicht abgelegt werden: %v", err)
	}

	cmd := exec.Command(exePath)
	cmd.Dir = filepath.Dir(exePath)
	if err := cmd.Start(); err != nil {
		os.Remove(exePath)
		os.Rename(oldExePath(), exePath)
		return fmt.Errorf("neue Fassung startet nicht: %v", err)
	}
	return nil
}

func runUpdate(info updateInfo) {
	tmp := filepath.Join(appDir, ".update")
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)

	dest := filepath.Join(tmp, "druckerfarm-new"+exeSuffix())
	if err := downloadUpdate(info.URL, dest); err != nil {
		failUpdate(err)
		return
	}
	setUpdProgress(func(p *updateProgress) { p.Step = "verify" })
	if err := looksExecutable(dest); err != nil {
		failUpdate(err)
		return
	}
	if err := applyUpdate(dest); err != nil {
		failUpdate(err)
		return
	}

	setUpdProgress(func(p *updateProgress) {
		p.Active = false
		p.Done = true
		p.Step = "done"
	})
	log.Printf("Update auf %s eingespielt, starte neu", info.Latest)

	// Der Nachfolger laeuft bereits — hier ordentlich beenden.
	go func() {
		time.Sleep(1500 * time.Millisecond)
		stopGo2rtc()
		os.Exit(0)
	}()
}

func failUpdate(err error) {
	log.Printf("Update fehlgeschlagen: %v", err)
	setUpdProgress(func(p *updateProgress) {
		p.Active = false
		p.Done = true
		p.Step = "error"
		p.Error = err.Error()
	})
}

// ─── HTTP ─────────────────────────────────────────────────────────────────────

func handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	info, err := checkUpdate()
	if err != nil {
		writeJSON(w, map[string]any{
			"current": appVersion,
			"repo":    info.Repo,
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, info)
}

func handleUpdateInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	info, err := checkUpdate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if !info.Available {
		http.Error(w, "keine neuere Fassung verfuegbar", http.StatusBadRequest)
		return
	}

	updMu.Lock()
	if updProgress.Active {
		updMu.Unlock()
		http.Error(w, "Update laeuft bereits", http.StatusConflict)
		return
	}
	updProgress = updateProgress{Active: true, Step: "download"}
	updMu.Unlock()

	go runUpdate(info)
	writeJSON(w, map[string]any{"started": info.Latest})
}

func handleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	updMu.Lock()
	p := updProgress
	updMu.Unlock()
	writeJSON(w, p)
}

func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Repo  string `json:"repo"`
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		repo := strings.TrimSpace(body.Repo)
		repo = strings.TrimPrefix(repo, "https://github.com/")
		repo = strings.TrimSuffix(strings.TrimSuffix(repo, "/"), ".git")
		if repo != "" && strings.Count(repo, "/") != 1 {
			http.Error(w, "Format ist besitzer/name", http.StatusBadRequest)
			return
		}
		mu.Lock()
		state.UpdateRepo = repo
		state.UpdateToken = strings.TrimSpace(body.Token)
		mu.Unlock()
		saveState()
	}

	repo, token := updateRepo()
	writeJSON(w, map[string]any{
		"repo":      repo,
		"has_token": token != "",
		"version":   appVersion,
		"exe":       exePath,
		"data_dir":  appDir,
	})
}
