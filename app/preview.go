package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── DRUCK-VORSCHAU (3MF-Thumbnail) ───────────────────────────────────────────
//
// A sliced .3mf embeds a colored plate render at Metadata/plate_1.png. This
// endpoint fetches that file from the printer's SD over FTP (via the Python
// helper's "thumb" action), extracts the PNG and returns it — so the detail
// modal can show what is being printed. Results are cached briefly so opening
// the modal repeatedly does not re-download the (possibly large) 3MF each time.

const previewTTL = 10 * time.Minute

// The preview cache lives on the PC (appDir/previewcache), never on the SD card
// and never in the user's print-file folder. A .gcode.3mf is a ZIP; it is
// downloaded ONCE per file-version and every plate render (Metadata/plate_N.png)
// is extracted into a hidden cache dir. Cache validity is keyed by FTP SIZE +
// MDTM, so re-uploading a file under the same name (new content) is detected and
// the preview refreshed.

var extractMu sync.Mutex // serialize 3MF downloads (also eases FTP/421 load)

func previewCacheDir() string { return filepath.Join(appDir, "previewcache") }

// resolveSDPath turns a bare file name into a full SD path.
func resolveSDPath(file string) string {
	if strings.HasPrefix(file, "/") {
		return file
	}
	base := syncCfg.SDPath
	if base == "" {
		base = "/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + file
}

// previewToken asks the printer for SIZE+MDTM (no download) and returns a short,
// filesystem-safe version token. Empty string means the file could not be stat'd.
func previewToken(ip, code, path string) string {
	out, err := runPythonFTP([]string{"stat", "--ip", ip, "--code", code, "--path", path})
	if err != nil {
		return ""
	}
	f := strings.Fields(strings.TrimSpace(out))
	if len(f) == 0 {
		return ""
	}
	tok := f[0]
	if len(f) > 1 {
		tok += "_" + f[1]
	}
	tok = strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '-'
	}, tok)
	return tok
}

func previewVerDir(ip, path, token string) string {
	sum := sha1.Sum([]byte(ip + "|" + path))
	return filepath.Join(previewCacheDir(), hex.EncodeToString(sum[:]), token)
}

func listCachedPlates(verDir string) []int {
	ents, err := os.ReadDir(verDir)
	if err != nil {
		return nil
	}
	var nums []int
	for _, e := range ents {
		name := e.Name()
		if !strings.HasSuffix(name, ".png") {
			continue
		}
		if v, err := strconv.Atoi(strings.TrimSuffix(name, ".png")); err == nil {
			nums = append(nums, v)
		}
	}
	sort.Ints(nums)
	return nums
}

// ensureExtracted downloads the 3MF once (unless already cached for this
// version) and extracts all plate PNGs into verDir. Returns the plate numbers.
func ensureExtracted(ip, code, path, verDir string) ([]int, error) {
	if plates := listCachedPlates(verDir); len(plates) > 0 {
		return plates, nil
	}
	extractMu.Lock()
	defer extractMu.Unlock()
	if plates := listCachedPlates(verDir); len(plates) > 0 { // re-check under lock
		return plates, nil
	}
	if err := os.MkdirAll(verDir, 0o755); err != nil {
		return nil, err
	}
	// Prune older versions of the same file (other tokens) so an overwritten
	// file's stale previews don't pile up — keep only the current version.
	if parent := filepath.Dir(verDir); parent != "" {
		if ents, err := os.ReadDir(parent); err == nil {
			for _, e := range ents {
				if e.Name() != filepath.Base(verDir) {
					_ = os.RemoveAll(filepath.Join(parent, e.Name()))
				}
			}
		}
	}
	out, err := runPythonFTP([]string{"extract", "--ip", ip, "--code", code, "--path", path, "--outdir", verDir})
	if err != nil {
		return nil, err
	}
	var res struct {
		Plates []int `json:"plates"`
	}
	_ = json.Unmarshal([]byte(strings.TrimSpace(out)), &res)
	return res.Plates, nil
}

// handlePlates returns the available plate numbers for a file plus a version
// token. The UI calls this first, then requests each plate via /api/preview.
func handlePlates(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	file := strings.TrimSpace(r.URL.Query().Get("file"))
	if ip == "" || file == "" {
		http.Error(w, "ip und file noetig", http.StatusBadRequest)
		return
	}
	p := printerByIP(ip)
	if p == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	if !strings.HasSuffix(strings.ToLower(file), ".3mf") { // also matches .gcode.3mf
		writeJSON(w, map[string]any{"plates": []int{}, "token": ""})
		return
	}
	path := resolveSDPath(file)
	token := previewToken(ip, p.Code, path)
	if token == "" {
		token = "x"
	}
	plates, err := ensureExtracted(ip, p.Code, path, previewVerDir(ip, path, token))
	if err != nil {
		writeJSON(w, map[string]any{"plates": []int{}, "token": token})
		return
	}
	writeJSON(w, map[string]any{"plates": plates, "token": token})
}

// handlePreview serves one plate PNG from the PC cache (downloading+extracting
// once if needed). Query: ip, file, plate (default 1), token (from /api/plates).
func handlePreview(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	file := strings.TrimSpace(r.URL.Query().Get("file"))
	if ip == "" || file == "" {
		http.Error(w, "ip und file noetig", http.StatusBadRequest)
		return
	}
	p := printerByIP(ip)
	if p == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	if !strings.HasSuffix(strings.ToLower(file), ".3mf") {
		http.Error(w, "keine Vorschau", http.StatusNotFound)
		return
	}
	plate := strings.TrimSpace(r.URL.Query().Get("plate"))
	if _, err := strconv.Atoi(plate); err != nil {
		plate = "1" // default to the first plate; guards against path traversal
	}
	path := resolveSDPath(file)
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		if token = previewToken(ip, p.Code, path); token == "" {
			token = "x"
		}
	}
	verDir := previewVerDir(ip, path, token)
	pngPath := filepath.Join(verDir, plate+".png")
	if _, err := os.Stat(pngPath); err != nil {
		if _, err := ensureExtracted(ip, p.Code, path, verDir); err != nil {
			http.Error(w, "keine Vorschau", http.StatusNotFound)
			return
		}
	}
	data, err := os.ReadFile(pngPath)
	if err != nil || len(data) == 0 {
		http.Error(w, "keine Vorschau", http.StatusNotFound)
		return
	}
	servePNG(w, data)
}

func servePNG(w http.ResponseWriter, png []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=300")
	_, _ = w.Write(png)
}

var (
	colorsMu    sync.Mutex
	colorsCache = map[string]struct {
		json []byte
		at   time.Time
	}{}
)

// handleFileColors returns the filament colours a .3mf needs, as a JSON array of
// hex strings — for the file↔AMS colour comparison in the detail modal.
func handleFileColors(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	file := strings.TrimSpace(r.URL.Query().Get("file"))
	if ip == "" || file == "" {
		http.Error(w, "ip und file noetig", http.StatusBadRequest)
		return
	}
	p := printerByIP(ip)
	if p == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}
	if !strings.HasSuffix(strings.ToLower(file), ".3mf") {
		writeJSON(w, map[string]any{"colors": []string{}})
		return
	}
	path := file
	if !strings.HasPrefix(path, "/") {
		base := syncCfg.SDPath
		if base == "" {
			base = "/"
		}
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		path = base + file
	}
	key := ip + "|" + path
	colorsMu.Lock()
	if e, ok := colorsCache[key]; ok && time.Since(e.at) < previewTTL && len(e.json) > 0 {
		j := e.json
		colorsMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(j)
		return
	}
	colorsMu.Unlock()

	out, err := runPythonFTP([]string{"colors", "--ip", ip, "--code", p.Code, "--path", path})
	if err != nil {
		writeJSON(w, map[string]any{"colors": []string{}})
		return
	}
	var cols []string
	_ = json.Unmarshal([]byte(strings.TrimSpace(out)), &cols)
	body, _ := json.Marshal(map[string]any{"colors": cols})
	colorsMu.Lock()
	colorsCache[key] = struct {
		json []byte
		at   time.Time
	}{json: body, at: time.Now()}
	colorsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}
