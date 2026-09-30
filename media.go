package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── MEDIA MANAGEMENT (timelapse videos & snapshots) ──────────────────────────
//
// On the same FTPS storage (port 990) that holds the print files, the printer
// also stores timelapse videos and camera snapshots. This file exposes them for
// viewing, downloading, single and bulk deletion, and toggles the timelapse via
// MQTT — including an option to keep it permanently off (otherwise the printer
// occasionally turns it back on by itself).
//
// The folders differ slightly by firmware/model, so each kind is searched under
// several common roots:
//   video    → /timelapse
//   snapshot → /image, /ipcam
// Only known media extensions are accepted, so a print file (.gcode/.3mf) is
// never touched here.

var mediaRoots = map[string][]string{
	"video":    {"/timelapse"},
	"snapshot": {"/image", "/ipcam"},
}

var mediaExt = map[string][]string{
	"video":    {".mp4", ".avi", ".mov", ".mkv"},
	"snapshot": {".jpg", ".jpeg", ".png", ".bmp"},
}

type mediaFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// hasMediaExt reports whether name ends with an allowed extension for kind.
func hasMediaExt(kind, name string) bool {
	low := strings.ToLower(name)
	for _, e := range mediaExt[kind] {
		if strings.HasSuffix(low, e) {
			return true
		}
	}
	return false
}

// pathInMediaRoots makes sure a full path really lies within one of the media
// root directories and cannot escape via ".." — protection against deleting or
// reading arbitrary files through these endpoints.
func pathInMediaRoots(kind, full string) bool {
	if strings.Contains(full, "..") {
		return false
	}
	for _, root := range mediaRoots[kind] {
		if full == root || strings.HasPrefix(full, root+"/") {
			return true
		}
	}
	return false
}

// mediaScan searches the roots of a kind for media files.
func mediaScan(ip, code, kind string) ([]mediaFile, error) {
	roots, ok := mediaRoots[kind]
	if !ok {
		return nil, fmt.Errorf("unbekannte Medienart %q", kind)
	}
	var files []mediaFile
	var lastErr error
	for _, root := range roots {
		out, err := runPythonFTP([]string{"scan", "--ip", ip, "--code", code, "--path", root, "--depth", "2"})
		if err != nil {
			lastErr = err
			continue // root may not exist — try the next one
		}
		var items []struct {
			Path string `json:"path"`
			Dir  bool   `json:"dir"`
			Size int64  `json:"size"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &items) != nil {
			continue
		}
		for _, it := range items {
			if it.Dir || !hasMediaExt(kind, it.Path) {
				continue
			}
			files = append(files, mediaFile{Path: it.Path, Name: path.Base(it.Path), Size: it.Size})
		}
	}
	if files == nil && lastErr != nil {
		return nil, lastErr
	}
	return files, nil
}

// handleMediaList: GET ?ip=&kind=video|snapshot — list plus free space.
func handleMediaList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "video"
	}
	code, err := printerCode(ip)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	if _, ok := mediaRoots[kind]; !ok {
		http.Error(w, `{"error":"bad kind"}`, 400)
		return
	}

	type resp struct {
		Files     []mediaFile `json:"files"`
		Free      int64       `json:"free"`
		Total     int64       `json:"total"`
		Error     string      `json:"error,omitempty"`
		ErrorCode string      `json:"error_code,omitempty"`
	}
	done := make(chan resp, 1)
	go func() {
		var out resp
		files, err := mediaScan(ip, code, kind)
		if err != nil {
			out.Error = friendlyFTPError(err.Error())
			out.ErrorCode = ftpErrKey(out.Error)
		}
		out.Files = files
		if ds, err := runPythonFTP([]string{"diskspace", "--ip", ip, "--code", code, "--path", "/"}); err == nil {
			var d struct {
				Free  int64 `json:"free"`
				Total int64 `json:"total"`
			}
			if json.Unmarshal([]byte(strings.TrimSpace(ds)), &d) == nil {
				out.Free, out.Total = d.Free, d.Total
			}
		}
		done <- out
	}()
	select {
	case out := <-done:
		json.NewEncoder(w).Encode(out)
	case <-time.After(30 * time.Second):
		json.NewEncoder(w).Encode(resp{Error: "Timeout nach 30s"})
	}
}

// handleMediaDownload: GET ?ip=&kind=&path= — stream a file (view/save).
func handleMediaDownload(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	kind := r.URL.Query().Get("kind")
	full := r.URL.Query().Get("path")
	if !pathInMediaRoots(kind, full) {
		http.Error(w, "ungültiger Pfad", 400)
		return
	}
	code, err := printerCode(ip)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	fc := &printerFTP{ip: ip, code: code}
	rc, err := fc.retr(full)
	if err != nil {
		http.Error(w, friendlyFTPError(err.Error()), 500)
		return
	}
	defer rc.Close()
	name := path.Base(full)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, rc)
}

// handleMediaDelete: DELETE ?ip=&kind=&path= — delete one media file.
func handleMediaDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	kind := r.URL.Query().Get("kind")
	full := r.URL.Query().Get("path")
	if !pathInMediaRoots(kind, full) {
		http.Error(w, `{"error":"ungültiger Pfad"}`, 400)
		return
	}
	code, err := printerCode(ip)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 404)
		return
	}
	if _, err := runPythonFTP([]string{"delete", "--ip", ip, "--code", code, "--path", full}); err != nil {
		http.Error(w, `{"error":"`+strings.TrimSpace(friendlyFTPError(err.Error()))+`"}`, 500)
		return
	}
	w.Write([]byte(`{"ok":true}`))
}

// handleMediaClear: POST ?ip=&kind= — delete ALL media of one kind
// ("Clear video data" / "Clear snapshot data"). Returns the per-file results.
func handleMediaClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	kind := r.URL.Query().Get("kind")
	code, err := printerCode(ip)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	files, err := mediaScan(ip, code, kind)
	if err != nil {
		fe := friendlyFTPError(err.Error())
		json.NewEncoder(w).Encode(map[string]any{"error": fe, "error_code": ftpErrKey(fe)})
		return
	}
	var paths []string
	for _, f := range files {
		if pathInMediaRoots(kind, f.Path) {
			paths = append(paths, f.Path)
		}
	}
	if len(paths) == 0 {
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "deleted": 0})
		return
	}
	payload, _ := json.Marshal(paths)
	out, err := runPythonFTPOut([]string{"deletemany", "--ip", ip, "--code", code}, bytes.NewReader(payload))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": friendlyFTPError(err.Error())})
		return
	}
	var res []deleteResult
	json.Unmarshal([]byte(strings.TrimSpace(out)), &res)
	ok := 0
	for _, rr := range res {
		if rr.OK {
			ok++
		}
	}
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "deleted": ok, "total": len(paths), "results": res})
}

// handleMediaTimelapse: POST ?ip=&on=0|1&enforce=0|1 — toggle timelapse and
// optionally keep it permanently off.
func handleMediaTimelapse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	on := r.URL.Query().Get("on") == "1"
	enforce := r.URL.Query().Get("enforce") == "1"

	mu.Lock()
	var pr *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == ip {
			pr = &state.Printers[i]
			break
		}
	}
	if pr != nil {
		if state.TimelapseAus == nil {
			state.TimelapseAus = map[string]bool{}
		}
		// Enforce only makes sense when switching off.
		if enforce && !on {
			state.TimelapseAus[ip] = true
		} else {
			delete(state.TimelapseAus, ip)
		}
	}
	var p Printer
	if pr != nil {
		p = *pr
	}
	mu.Unlock()
	if pr == nil {
		http.Error(w, `{"error":"nicht gefunden"}`, 404)
		return
	}
	saveState()
	if err := mqttMgr.SetTimelapse(p, on); err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "on": on, "enforce": enforce && !on})
}

// ─── FARM-WIDE MEDIA ACTIONS ──────────────────────────────────────────────────
//
// The same actions as per-printer, but across many printers at once — callable
// from the File Sync view. An optional body {"ips":[...]} narrows the set; if it
// is missing, the action applies to all configured printers.

// mediaTargets returns the target printers: the IPs named in the body, else all.
func mediaTargets(r *http.Request) []Printer {
	var body struct {
		IPs []string `json:"ips"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}
	want := map[string]bool{}
	for _, ip := range body.IPs {
		want[ip] = true
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []Printer
	for _, p := range state.Printers {
		if len(want) == 0 || want[p.IP] {
			out = append(out, p)
		}
	}
	return out
}

type mediaFarmResult struct {
	IP      string `json:"ip"`
	Name    string `json:"name"`
	Deleted int    `json:"deleted,omitempty"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

// handleMediaClearAll deletes all media of one kind across multiple printers.
func handleMediaClearAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	kind := r.URL.Query().Get("kind")
	if _, ok := mediaRoots[kind]; !ok {
		http.Error(w, `{"error":"bad kind"}`, 400)
		return
	}
	targets := mediaTargets(r)
	results := make([]mediaFarmResult, len(targets))
	sem := make(chan struct{}, 4) // at most four printers at a time
	var wg sync.WaitGroup
	for i, p := range targets {
		wg.Add(1)
		go func(idx int, pr Printer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			res := mediaFarmResult{IP: pr.IP, Name: pr.Name}
			if pr.Code == "" {
				res.Error = "kein Zugangscode"
				results[idx] = res
				return
			}
			files, err := mediaScan(pr.IP, pr.Code, kind)
			if err != nil {
				res.Error = friendlyFTPError(err.Error())
				results[idx] = res
				return
			}
			var paths []string
			for _, f := range files {
				if pathInMediaRoots(kind, f.Path) {
					paths = append(paths, f.Path)
				}
			}
			if len(paths) == 0 {
				res.OK = true
				results[idx] = res
				return
			}
			payload, _ := json.Marshal(paths)
			out, err := runPythonFTPOut([]string{"deletemany", "--ip", pr.IP, "--code", pr.Code}, bytes.NewReader(payload))
			if err != nil {
				res.Error = friendlyFTPError(err.Error())
				results[idx] = res
				return
			}
			var dr []deleteResult
			json.Unmarshal([]byte(strings.TrimSpace(out)), &dr)
			for _, d := range dr {
				if d.OK {
					res.Deleted++
				}
			}
			res.OK = true
			results[idx] = res
		}(i, p)
	}
	wg.Wait()
	json.NewEncoder(w).Encode(map[string]any{"results": results})
}

// handleMediaTimelapseAll toggles the timelapse across multiple printers.
func handleMediaTimelapseAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	on := r.URL.Query().Get("on") == "1"
	enforce := r.URL.Query().Get("enforce") == "1"
	targets := mediaTargets(r)

	// Set/remove the enforce markers and persist.
	mu.Lock()
	if state.TimelapseAus == nil {
		state.TimelapseAus = map[string]bool{}
	}
	for _, p := range targets {
		if enforce && !on {
			state.TimelapseAus[p.IP] = true
		} else {
			delete(state.TimelapseAus, p.IP)
		}
	}
	mu.Unlock()
	saveState()

	results := make([]mediaFarmResult, len(targets))
	for i, p := range targets {
		res := mediaFarmResult{IP: p.IP, Name: p.Name}
		if err := mqttMgr.SetTimelapse(p, on); err != nil {
			res.Error = err.Error()
		} else {
			res.OK = true
		}
		results[i] = res
	}
	json.NewEncoder(w).Encode(map[string]any{"results": results, "on": on, "enforce": enforce && !on})
}

// timelapseEnforceLoop keeps the timelapse permanently off for printers whose
// TimelapseAus flag is set: if the printer reports "enable", "disable" is sent
// again. This explains and fixes the observed "the printer ignores my off
// command" — it re-enables it on its own.
func timelapseEnforceLoop() {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()
	for range tick.C {
		if shuttingDown.Load() {
			return
		}
		mu.Lock()
		var targets []Printer
		for _, p := range state.Printers {
			if state.TimelapseAus[p.IP] {
				targets = append(targets, p)
			}
		}
		mu.Unlock()
		for _, p := range targets {
			st := mqttMgr.GetStatus(p.IP)
			if st == nil || !st.Online {
				continue
			}
			// Only step in when the printer reports the timelapse as "enable".
			if st.TimelapseKnown && st.Timelapse == "enable" {
				if err := mqttMgr.SetTimelapse(p, false); err == nil {
					logSync("Zeitraffer bei " + p.Name + " war wieder an — erneut ausgeschaltet")
				}
			}
		}
	}
}

// mediaSizeText formats bytes human-readably (for logs/responses if needed).
func mediaSizeText(b int64) string {
	const u = 1024
	if b < u {
		return strconv.FormatInt(b, 10) + " B"
	}
	div, exp := int64(u), 0
	for n := b / u; n >= u; n /= u {
		div *= u
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
