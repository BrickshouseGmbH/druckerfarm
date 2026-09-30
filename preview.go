package main

import (
	"net/http"
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

type previewEntry struct {
	png []byte
	at  time.Time
}

var (
	previewMu    sync.Mutex
	previewCache = map[string]previewEntry{}
)

const previewTTL = 10 * time.Minute

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
	// Only .3mf carries an embedded plate preview.
	if !strings.HasSuffix(strings.ToLower(file), ".3mf") {
		http.Error(w, "keine Vorschau", http.StatusNotFound)
		return
	}

	// Resolve the path: a bare name is looked up under the configured SD path.
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
	previewMu.Lock()
	if e, ok := previewCache[key]; ok && time.Since(e.at) < previewTTL && len(e.png) > 0 {
		png := e.png
		previewMu.Unlock()
		servePNG(w, png)
		return
	}
	previewMu.Unlock()

	out, err := runPythonFTP([]string{"thumb", "--ip", ip, "--code", p.Code, "--path", path})
	if err != nil || len(out) == 0 {
		http.Error(w, "keine Vorschau", http.StatusNotFound)
		return
	}
	png := []byte(out)
	previewMu.Lock()
	previewCache[key] = previewEntry{png: png, at: time.Now()}
	previewMu.Unlock()
	servePNG(w, png)
}

func servePNG(w http.ResponseWriter, png []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=300")
	_, _ = w.Write(png)
}
