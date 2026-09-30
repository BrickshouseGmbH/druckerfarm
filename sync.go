package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ftpProcTimeout caps how long a single Python FTP call may run. It is longer
// than the script's own 20s socket timeout, and acts as a hard backstop: if the
// process still hangs, it is killed so it cannot keep an FTP session open on the
// printer (which would pile up toward the "421 too many connections" limit).
const ftpProcTimeout = 45 * time.Second

// ftpCommand builds the Python FTP command bound to a context that kills the
// process when the timeout elapses.
func ftpCommand(ctx context.Context, script string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, findPython(), append([]string{script}, args...)...)
	hideWindow(cmd)
	return cmd
}

// ─── SYNC CONFIG ──────────────────────────────────────────────────────────────

type SyncConfig struct {
	RefPath      string `json:"ref_path"` // local or UNC path
	AutoSync     bool   `json:"auto_sync"`
	AutoInterval int    `json:"auto_interval"` // minutes, default 60
	SDPath       string `json:"sd_path"`       // path on SD card, default "/"
	Mode         string `json:"mode"`          // manual | auto | both
	// ModelFolders: a separate local folder per printer model. Fits when
	// each model needs different gcode files. Empty -> RefPath as default.
	ModelFolders map[string]string `json:"model_folders,omitempty"`
}

// syncFolderFor returns the local folder for a model: the model mapping
// first, otherwise the global default folder (RefPath).
func syncFolderFor(model string) string {
	m := strings.ToUpper(strings.TrimSpace(model))
	if syncCfg.ModelFolders != nil {
		if p := strings.TrimSpace(syncCfg.ModelFolders[m]); p != "" {
			return p
		}
	}
	return strings.TrimSpace(syncCfg.RefPath)
}

// ─── SYNC STATE ───────────────────────────────────────────────────────────────

type SyncStatus string

const (
	SyncIdle    SyncStatus = "idle"
	SyncRunning SyncStatus = "running"
	SyncPaused  SyncStatus = "paused"
)

type SyncProgress struct {
	Status      SyncStatus `json:"status"`
	Current     string     `json:"current_printer"`
	CurrentFile string     `json:"current_file"`
	Done        int        `json:"done"`
	Total       int        `json:"total"`
	Uploaded    int        `json:"uploaded"`
	Skipped     int        `json:"skipped"`
	Errors      []string   `json:"errors"`
	LastSync    string     `json:"last_sync"`
	Log         []string   `json:"log"`
}

var (
	syncCfg      SyncConfig
	syncCfgFile  string
	syncMu       sync.Mutex
	syncProgress = SyncProgress{Status: SyncIdle, Errors: []string{}, Log: []string{}}
	syncStop     = make(chan struct{}, 1)
	syncPause    chan struct{}
	syncResume   chan struct{}
	syncAutoTick *time.Ticker
)

// ─── INIT ─────────────────────────────────────────────────────────────────────

func initSync(dir string) {
	syncCfgFile = filepath.Join(dir, "sync.json")
	loadSyncConfig()
	if syncCfg.AutoInterval == 0 {
		syncCfg.AutoInterval = 60
	}
	if syncCfg.SDPath == "" {
		syncCfg.SDPath = "/"
	}
	if syncCfg.AutoSync {
		startAutoSync()
	}
}

func loadSyncConfig() {
	data, err := os.ReadFile(syncCfgFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &syncCfg)
}

func saveSyncConfig() {
	data, _ := json.MarshalIndent(syncCfg, "", "  ")
	atomicWrite(syncCfgFile, data, 0644)
}

func startAutoSync() {
	if syncAutoTick != nil {
		syncAutoTick.Stop()
	}
	d := time.Duration(syncCfg.AutoInterval) * time.Minute
	syncAutoTick = time.NewTicker(d)
	go func() {
		for range syncAutoTick.C {
			syncMu.Lock()
			status := syncProgress.Status
			syncMu.Unlock()
			if status == SyncIdle {
				go runSync()
			}
		}
	}()
}

// ─── SYNC RUNNER ──────────────────────────────────────────────────────────────

func runSync() {
	syncMu.Lock()
	if syncProgress.Status == SyncRunning {
		syncMu.Unlock()
		return
	}

	// At least one folder must be configured — global or per model.
	if strings.TrimSpace(syncCfg.RefPath) == "" && len(syncCfg.ModelFolders) == 0 {
		syncProgress = SyncProgress{
			Status: SyncIdle,
			Errors: []string{"Kein Ordner konfiguriert (global oder je Modell)"},
			Log:    []string{"Fehler: Kein Ordner konfiguriert"},
		}
		syncMu.Unlock()
		return
	}
	// Folder contents are read only once per folder.
	folderCache := map[string]map[string]string{}

	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	syncPause = make(chan struct{}, 1)
	syncResume = make(chan struct{}, 1)
	// Drain stop channel
	select {
	case <-syncStop:
	default:
	}

	syncProgress = SyncProgress{
		Status: SyncRunning,
		Total:  len(printers),
		Errors: []string{},
		Log:    []string{fmt.Sprintf("Sync gestartet — %d Drucker", len(printers))},
	}
	syncMu.Unlock()

	for i, p := range printers {
		// Check stop
		select {
		case <-syncStop:
			logSync("Sync gestoppt")
			setSyncStatus(SyncIdle)
			return
		default:
		}

		// Check pause
		syncMu.Lock()
		isPaused := syncProgress.Status == SyncPaused
		syncMu.Unlock()
		if isPaused {
			logSync("Sync pausiert — warte auf Fortsetzen...")
			select {
			case <-syncResume:
				logSync("Sync fortgesetzt")
			case <-syncStop:
				logSync("Sync gestoppt")
				setSyncStatus(SyncIdle)
				return
			}
		}

		syncMu.Lock()
		syncProgress.Current = p.Name
		syncProgress.Done = i
		syncMu.Unlock()

		logSync(fmt.Sprintf("[%d/%d] %s (%s)", i+1, len(printers), p.Name, p.IP))

		folder := syncFolderFor(p.Model)
		if strings.TrimSpace(folder) == "" {
			logSync("  ⚠ kein Ordner fuer Modell " + p.Model + " — uebersprungen")
			continue
		}
		lf, ok := folderCache[folder]
		if !ok {
			var lerr error
			lf, lerr = listLocalGcode(folder)
			if lerr != nil {
				lf = map[string]string{}
				logSync("  ⚠ Ordner nicht lesbar: " + folder)
			}
			folderCache[folder] = lf
		}
		if len(lf) == 0 {
			logSync("  ⚠ keine Dateien im Ordner " + folder + " — uebersprungen")
			continue
		}

		if err := syncPrinter(p, lf); err != nil {
			logSync(fmt.Sprintf("  ✗ Fehler: %v", err))
			syncMu.Lock()
			syncProgress.Errors = append(syncProgress.Errors, p.Name+": "+err.Error())
			syncMu.Unlock()
		}
	}

	syncMu.Lock()
	syncProgress.Status = SyncIdle
	syncProgress.Current = ""
	syncProgress.CurrentFile = ""
	syncProgress.Done = syncProgress.Total
	syncProgress.LastSync = time.Now().Format("02.01.2006 15:04")
	syncProgress.Log = append(syncProgress.Log, fmt.Sprintf("✓ Sync abgeschlossen — %d hochgeladen, %d bereits vorhanden",
		syncProgress.Uploaded, syncProgress.Skipped))
	syncMu.Unlock()
}

func syncPrinter(p Printer, localFiles map[string]string) error {
	fc, err := printerFTPConnect(p.IP, p.Code)
	if err != nil {
		return fmt.Errorf("FTP Verbindung fehlgeschlagen: %v", err)
	}
	defer fc.quit()

	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}

	existing, err := fc.list(sdPath)
	if err != nil {
		return fmt.Errorf("SD-Karte nicht lesbar: %v", err)
	}

	onSD := make(map[string]bool)
	for _, e := range existing {
		onSD[strings.ToLower(e.Name)] = true
	}

	for name, localPath := range localFiles {
		select {
		case <-syncStop:
			return fmt.Errorf("gestoppt")
		default:
		}
		if onSD[strings.ToLower(name)] {
			syncMu.Lock()
			syncProgress.Skipped++
			syncMu.Unlock()
			continue
		}
		syncMu.Lock()
		syncProgress.CurrentFile = name
		syncMu.Unlock()
		logSync(fmt.Sprintf("  ↑ %s", name))
		f, err := os.Open(localPath)
		if err != nil {
			logSync(fmt.Sprintf("  ✗ Kann Datei nicht öffnen: %v", err))
			continue
		}
		remotePath := sdPath
		if !strings.HasSuffix(remotePath, "/") {
			remotePath += "/"
		}
		remotePath += name
		if err := fc.stor(remotePath, f); err != nil {
			f.Close()
			logSync(fmt.Sprintf("  ✗ Upload fehlgeschlagen: %v", err))
			continue
		}
		f.Close()
		syncMu.Lock()
		syncProgress.Uploaded++
		syncMu.Unlock()
		logSync(fmt.Sprintf("  ✓ %s hochgeladen", name))
	}
	return nil
}

func listLocalGcode(dir string) (map[string]string, error) {
	files := make(map[string]string)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".gcode") {
			files[e.Name()] = filepath.Join(dir, e.Name())
		}
	}
	return files, nil
}

func logSync(msg string) {
	log.Println("[SYNC]", msg)
	syncMu.Lock()
	syncProgress.Log = append(syncProgress.Log, msg)
	// Keep last 200 log lines
	if len(syncProgress.Log) > 200 {
		syncProgress.Log = syncProgress.Log[len(syncProgress.Log)-200:]
	}
	syncMu.Unlock()
}

func setSyncStatus(s SyncStatus) {
	syncMu.Lock()
	syncProgress.Status = s
	syncMu.Unlock()
}

// ─── HTTP API ─────────────────────────────────────────────────────────────────

func handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	syncMu.Lock()
	defer syncMu.Unlock()
	json.NewEncoder(w).Encode(syncProgress)
}

func handleSyncConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		syncMu.Lock()
		defer syncMu.Unlock()
		json.NewEncoder(w).Encode(syncCfg)
	case http.MethodPost:
		var cfg SyncConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		syncMu.Lock()
		syncCfg = cfg
		if syncCfg.AutoInterval == 0 {
			syncCfg.AutoInterval = 60
		}
		if syncCfg.SDPath == "" {
			syncCfg.SDPath = "/"
		}
		syncMu.Unlock()
		saveSyncConfig()
		if cfg.AutoSync {
			startAutoSync()
		} else if syncAutoTick != nil {
			syncAutoTick.Stop()
		}
		json.NewEncoder(w).Encode(syncCfg)
	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleSyncControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/sync/")

	switch action {
	case "start":
		syncMu.Lock()
		status := syncProgress.Status
		syncMu.Unlock()
		if status == SyncIdle {
			go runSync()
			w.Write([]byte(`{"ok":true}`))
		} else if status == SyncPaused {
			setSyncStatus(SyncRunning)
			select {
			case syncResume <- struct{}{}:
			default:
			}
			w.Write([]byte(`{"ok":true}`))
		} else {
			http.Error(w, "already running", 409)
		}

	case "pause":
		syncMu.Lock()
		if syncProgress.Status == SyncRunning {
			syncProgress.Status = SyncPaused
		}
		syncMu.Unlock()
		w.Write([]byte(`{"ok":true}`))

	case "stop":
		select {
		case syncStop <- struct{}{}:
		default:
		}
		// Also unblock pause if paused
		select {
		case syncResume <- struct{}{}:
		default:
		}
		w.Write([]byte(`{"ok":true}`))

	default:
		http.Error(w, "unknown action", 400)
	}
}

// handleSyncLocalFiles lists .gcode files from local path (for browser file picker)
func handleSyncBrowse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := r.URL.Query().Get("path")
	if path == "" {
		path = syncCfg.RefPath
	}
	if path == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"files": []string{}, "error": "kein Pfad"})
		return
	}
	files, err := listLocalGcode(path)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"files": []string{}, "error": err.Error()})
		return
	}
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"files": names, "count": len(names), "path": path})
}

// handleSyncDownload downloads a single gcode file from a printer's SD card
func handleSyncDownload(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	file := r.URL.Query().Get("file")
	if ip == "" || file == "" {
		http.Error(w, "ip und file erforderlich", 400)
		return
	}

	mu.Lock()
	var printer *Printer
	for i, p := range state.Printers {
		if p.IP == ip {
			printer = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if printer == nil {
		http.Error(w, "Drucker nicht gefunden", 404)
		return
	}

	fc, err := printerFTPConnect(ip, printer.Code)
	if err != nil {
		http.Error(w, "FTP Verbindung fehlgeschlagen: "+err.Error(), 500)
		return
	}
	defer fc.quit()

	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	if !strings.HasSuffix(sdPath, "/") {
		sdPath += "/"
	}
	remotePath := sdPath + file

	resp, err := fc.retr(remotePath)
	if err != nil {
		http.Error(w, "Datei nicht gefunden: "+err.Error(), 404)
		return
	}
	defer resp.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+file+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, resp)
}

// ─── SD CARD LISTING ──────────────────────────────────────────────────────────

type SDFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type SDCardInfo struct {
	IP        string       `json:"ip"`
	Name      string       `json:"name"`
	FileCount int          `json:"file_count"`
	TotalSize int64        `json:"total_size"` // bytes used by gcode files
	Error     string       `json:"error,omitempty"`
	ErrorCode string       `json:"error_code,omitempty"` // stable key for the UI translation
	Files     []SDFileInfo `json:"files,omitempty"`
}

func handleSyncSDList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")

	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	if ip != "" {
		// Return file list for one printer with timeout
		done := make(chan SDCardInfo, 1)
		for _, p := range printers {
			if p.IP == ip {
				pr := p
				go func() { done <- fetchSDFiles(pr) }()
				select {
				case info := <-done:
					json.NewEncoder(w).Encode(info)
				case <-time.After(20 * time.Second):
					json.NewEncoder(w).Encode(sdErr(ip, "", "Timeout nach 20s", "ftpTimeout"))
				}
				return
			}
		}
		http.Error(w, "nicht gefunden", 404)
		return
	}

	// Single printer summary with timeout
	results := make([]SDCardInfo, len(printers))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for i, p := range printers {
		wg.Add(1)
		go func(idx int, pr Printer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			done := make(chan SDCardInfo, 1)
			go func() { done <- fetchSDSummary(pr) }()
			select {
			case info := <-done:
				results[idx] = info
			case <-time.After(12 * time.Second):
				results[idx] = sdErr(pr.IP, pr.Name, "Timeout", "ftpTimeout")
			}
		}(i, p)
	}
	wg.Wait()
	json.NewEncoder(w).Encode(results)
}

// ftpConnect kept for upload in syncPrinter
func ftpConnect(p Printer) (*printerFTP, error) {
	return printerFTPConnect(p.IP, p.Code)
}

// printerFTP uses curl for FTPS — handles vsftpd session reuse via OpenSSL.
type printerFTP struct {
	ip   string
	code string
}

func printerFTPConnect(ip, code string) (*printerFTP, error) {
	return &printerFTP{ip: ip, code: code}, nil
}

func (f *printerFTP) baseURL() string {
	return fmt.Sprintf("ftps://%s:990", f.ip)
}

// Python FTP script with SSL session reuse (works with the printer's vsftpd)
const ftpPyScript = `
import sys, ssl, ftplib, os, json, socket

# Never let a stuck operation hang forever — the printer would keep the FTP
# session open (half-open) and, with vsftpd's per-IP limit, that leads to
# "421 too many connections" on later attempts.
socket.setdefaulttimeout(20)

class ImplicitFTP_TLS(ftplib.FTP_TLS):
    """Implicit FTPS client (port 990) with SSL session reuse for dem vsftpd des Druckers."""

    def connect(self, host, port=990, timeout=10, source_address=None):
        # For implicit FTPS, wrap the socket in TLS immediately
        self.host = host
        self.port = port
        self.timeout = timeout
        sock = socket.create_connection((host, port), timeout)
        self.sock = self.context.wrap_socket(sock,
            server_hostname=host,
            server_side=False)
        self.af = sock.family
        self.file = self.sock.makefile('r', encoding=self.encoding)
        self.welcome = self.getresp()
        return self.welcome
    
    def ntransfercmd(self, cmd, rest=None):
        conn, size = ftplib.FTP.ntransfercmd(self, cmd, rest)
        if self._prot_p:
            # Reuse the control channel TLS session — required by vsftpd
            ctrl_session = self.sock.session
            conn = self.context.wrap_socket(conn,
                server_side=False,
                server_hostname=self.host,
                session=ctrl_session)
        return conn, size

def main():
    import argparse
    p = argparse.ArgumentParser()
    p.add_argument('action')  # list, retr, stor
    p.add_argument('--ip')
    p.add_argument('--code')
    p.add_argument('--path', default='/')
    p.add_argument('--out', default='')
    p.add_argument('--depth', default='1')
    args = p.parse_args()

    ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE

    ftp = ImplicitFTP_TLS(context=ctx)
    ftp.connect(args.ip, 990, timeout=15)
    ftp.prot_p()
    ftp.login('bblp', args.code)

    # Everything below runs inside a try/finally so the control connection is
    # ALWAYS closed cleanly (QUIT, or CLOSE if QUIT fails) — even on an error.
    # A session left open would count against the printer's per-IP FTP limit.
    try:
      run_action(ftp, args)
    finally:
      try:
        ftp.quit()
      except Exception:
        try:
          ftp.close()
        except Exception:
          pass

def run_action(ftp, args):
    if args.action == 'list':
        path = args.path if args.path else '/'
        files = []
        def cb(line):
            parts = line.split(None, 8)
            if len(parts) >= 9:
                name = parts[8].strip()
                size = int(parts[4]) if parts[4].isdigit() else 0
                if name.lower().endswith('.gcode') or name.lower().endswith('.3mf'):
                    files.append({'name': name, 'size': size})
            elif len(parts) >= 1:
                # Some servers return just filenames
                name = parts[-1].strip()
                if name.lower().endswith('.gcode') or name.lower().endswith('.3mf'):
                    files.append({'name': name, 'size': 0})
        try:
            ftp.retrlines('LIST ' + path, cb)
        except Exception as e:
            sys.stderr.write('LIST error: ' + str(e) + '\n')
        # Also try MLSD if LIST returned nothing
        if not files:
            try:
                for name, facts in ftp.mlsd(path):
                    if name.lower().endswith('.gcode') or name.lower().endswith('.3mf'):
                        size = int(facts.get('size', 0))
                        files.append({'name': name, 'size': size})
            except Exception:
                pass
        print(json.dumps(files))

    elif args.action == 'retr':
        if args.out:
            with open(args.out, 'wb') as fout:
                ftp.retrbinary('RETR ' + args.path, fout.write)
        else:
            ftp.retrbinary('RETR ' + args.path, sys.stdout.buffer.write)

    elif args.action == 'delete':
        # Try DELE command directly on control channel
        resp = ftp.voidcmd('DELE ' + args.path)
        print(resp)

    elif args.action == 'deletemany':
        # Mehrere Dateien in EINER Sitzung. Vorher wurde je Datei ein eigener
        # Python-Prozess mit eigener FTP-Anmeldung gestartet — bei 126 Dateien
        # also 126 Anmeldungen.
        paths = json.loads(sys.stdin.read() or '[]')
        results = []
        for path in paths:
            try:
                ftp.voidcmd('DELE ' + path)
                results.append({'path': path, 'ok': True})
            except Exception as e:
                results.append({'path': path, 'ok': False, 'error': str(e)})
        print(json.dumps(results))

    elif args.action == 'scan':
        # Ordner und Dateien auflisten, ohne Filter auf Dateiendungen — dient
        # dazu, erst einmal zu sehen was auf dem Speicher liegt.
        maxdepth = int(args.depth or '1')
        out = []

        def entries_of(path):
            found = []
            try:
                for name, facts in ftp.mlsd(path):
                    found.append((name, facts.get('type', ''), int(facts.get('size', 0) or 0)))
                if found:
                    return found
            except Exception:
                pass
            lines = []
            try:
                ftp.retrlines('LIST ' + path, lines.append)
            except Exception as e:
                sys.stderr.write('LIST ' + path + ': ' + str(e) + chr(10))
                return found
            for line in lines:
                parts = line.split(None, 8)
                if len(parts) < 9:
                    continue
                name = parts[8].strip()
                typ = 'dir' if line[:1] == 'd' else 'file'
                size = int(parts[4]) if parts[4].isdigit() else 0
                found.append((name, typ, size))
            return found

        def walk(path, depth):
            for name, typ, size in entries_of(path):
                if name in ('.', '..') or not name:
                    continue
                full = path.rstrip('/') + '/' + name
                if typ == 'dir':
                    out.append({'path': full, 'dir': True, 'size': 0})
                    if depth < maxdepth:
                        walk(full, depth + 1)
                else:
                    out.append({'path': full, 'dir': False, 'size': size})

        walk(args.path or '/', 0)
        print(json.dumps(out))

    elif args.action == 'diskspace':
        result = {}
        try:
            avail = ftp.sendcmd('SITE AVAIL')
            import re
            m = re.search(r'(\d+)', avail)
            if m: result['free'] = int(m.group(1))
        except: pass
        try:
            total_resp = ftp.sendcmd('SITE TOTAL')
            import re
            m = re.search(r'(\d+)', total_resp)
            if m: result['total'] = int(m.group(1))
        except: pass
        print(json.dumps(result))

    elif args.action == 'stor':
        ftp.storbinary('STOR ' + args.path, sys.stdin.buffer)

    elif args.action == 'thumb':
        # Download the .3mf into memory and pull out the plate preview PNG
        # (Metadata/plate_1.png). Sliced 3MF files embed a colored render there.
        import io, zipfile
        buf = io.BytesIO()
        ftp.retrbinary('RETR ' + args.path, buf.write)
        buf.seek(0)
        data = b''
        try:
            with zipfile.ZipFile(buf) as z:
                names = z.namelist()
                def score(n):
                    nl = n.lower()
                    if not (nl.startswith('metadata/') and nl.endswith('.png')):
                        return -1
                    if 'plate_1.png' in nl: return 100
                    if 'plate_' in nl and 'small' not in nl: return 80
                    if 'plate_' in nl: return 50
                    if 'thumbnail' in nl or 'top' in nl: return 40
                    return 10
                cand = sorted([n for n in names if score(n) >= 0], key=score, reverse=True)
                if cand:
                    data = z.read(cand[0])
        except Exception as e:
            sys.stderr.write('thumb: ' + str(e) + chr(10))
        if args.out:
            with open(args.out, 'wb') as f:
                f.write(data)
        else:
            sys.stdout.buffer.write(data)

main()
`

func writePyScript() (string, error) {
	tmp, err := os.CreateTemp("", "printerfarm_ftp_*.py")
	if err != nil {
		return "", err
	}
	tmp.WriteString(ftpPyScript)
	tmp.Close()
	return tmp.Name(), nil
}

// friendlyFTPError translates what the Python helper writes to stderr on
// failure into a readable sentence. Previously the full
// Python traceback ended up in the UI — six lines of tech from
// which nobody could tell that the printer simply does not answer.
// ftpErrKey maps a (raw or already-processed) error message to a
// stable key. The UI translates this key via t() into
// the chosen language — so e.g. the port-990 message appears in English
// when the app is set to English, instead of hard-coded German. An empty return
// means "no known category" (then the plain text stays).
func ftpErrKey(raw string) string {
	low := strings.ToLower(raw)
	switch {
	case is421(raw):
		return "ftpBusy"
	case strings.Contains(low, "990") && (strings.Contains(low, "zeitüber") || strings.Contains(low, "timed out") || strings.Contains(low, "timeout")):
		return "ftpTimeout990"
	case strings.Contains(low, "abgelehnt") || strings.Contains(low, "refused") || strings.Contains(low, "10061"):
		return "ftpRefused990"
	case strings.Contains(low, "nicht erreichbar") || strings.Contains(low, "no route") || strings.Contains(low, "unreachable"):
		return "ftpNoRoute"
	case strings.Contains(low, "zugangscode") || strings.Contains(low, "login") || strings.Contains(low, "530"):
		return "ftpBadCode"
	case strings.Contains(low, "verschlüssel") || strings.Contains(low, "ssl") || strings.Contains(low, "certificate"):
		return "ftpSSL"
	case strings.Contains(low, "nicht vorhanden") || strings.Contains(low, "no such") || strings.Contains(low, "550"):
		return "ftpNoSuchFile"
	case strings.Contains(low, "python"):
		return "ftpNoPython"
	case strings.Contains(low, "timeout") || strings.Contains(low, "zeitüber"):
		return "ftpTimeout"
	}
	return ""
}

func friendlyFTPError(raw string) string {
	t := strings.TrimSpace(raw)
	if t == "" {
		return "Drucker antwortet nicht"
	}
	// Only the last line of the traceback carries the actual message.
	lines := strings.Split(strings.ReplaceAll(t, "\r\n", "\n"), "\n")
	last := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			last = l
			break
		}
	}
	low := strings.ToLower(t)
	switch {
	case is421(t):
		return "Drucker meldet zu viele FTP-Verbindungen (421) — zu viele gleichzeitige Zugriffe von diesem PC. Andere FTP-Programme schließen und kurz warten; das Tool versucht es automatisch erneut."
	case strings.Contains(low, "timed out"), strings.Contains(low, "timeout"),
		strings.Contains(low, "10060"), strings.Contains(low, "etimedout"):
		// Bambu printers offer file access exclusively via implicit
		// FTPS on port 990 — there is no other port. If 990 does not
		// answer, it is almost always the printer: LAN/developer mode off,
		// printer in cloud-only mode, or old firmware. So a hint here
		// instead of just "timeout".
		return "Zeitüberschreitung auf Port 990 — am Drucker den LAN-/Entwicklermodus einschalten (Einstellungen › Allgemein) und sicherstellen, dass er nicht nur im Cloud-Modus läuft. Einen anderen FTP-Port gibt es beim Drucker nicht."
	case strings.Contains(low, "refused"), strings.Contains(low, "10061"):
		return "Verbindung auf Port 990 abgelehnt — FTP/LAN-Modus am Drucker ist aus. Einen anderen Port bietet der Drucker nicht."
	case strings.Contains(low, "no route to host"), strings.Contains(low, "unreachable"),
		strings.Contains(low, "10065"):
		return "Drucker nicht erreichbar — im Netz nicht auffindbar"
	case strings.Contains(low, "530"), strings.Contains(low, "login"),
		strings.Contains(low, "authentication"):
		return "Zugangscode wird nicht angenommen"
	case strings.Contains(low, "ssl"), strings.Contains(low, "certificate"):
		return "Verschlüsselte Verbindung fehlgeschlagen"
	case strings.Contains(low, "no such file"), strings.Contains(low, "550"):
		return "Ordner oder Datei nicht vorhanden"
	case strings.Contains(low, "python"), strings.Contains(low, "not found"):
		return "Python nicht gefunden — für den Dateizugriff wird Python benötigt"
	}
	if len(last) > 160 {
		last = last[:160] + " …"
	}
	return last
}

func runPythonFTP(args []string) (string, error) {
	script, err := writePyScript()
	if err != nil {
		return "", err
	}
	defer os.Remove(script)

	// Only one FTP session per printer at a time (see ftpgate.go) — avoids the
	// vsftpd per-IP connection limit that answers with "421 too many connections".
	unlock := lockFTP(argIP(args))
	defer unlock()

	// On a 421 the printer is momentarily out of connection slots; wait briefly
	// and retry instead of failing. This call has no stdin, so retrying is safe.
	for attempt := 0; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), ftpProcTimeout)
		cmd := ftpCommand(ctx, script, args)
		out, err := cmd.Output()
		cancel()
		if err == nil {
			return string(out), nil
		}
		raw := err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			raw = string(exitErr.Stderr)
		}
		if is421(raw) && attempt < len(ftpRetryDelays) {
			time.Sleep(ftpRetryDelays[attempt])
			continue
		}
		return "", fmt.Errorf("%s", friendlyFTPError(raw))
	}
}

func runPythonFTPWithInput(args []string, r io.Reader) error {
	script, err := writePyScript()
	if err != nil {
		return err
	}
	defer os.Remove(script)

	unlock := lockFTP(argIP(args))
	defer unlock()

	cmd := exec.Command(findPython(), append([]string{script}, args...)...)
	hideWindow(cmd)
	cmd.Stdin = r
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Upload fehlgeschlagen: %s", friendlyFTPError(string(out)))
	}
	return nil
}

func (f *printerFTP) list(path string) ([]SDFileInfo, error) {
	if path == "" {
		path = "/"
	}
	out, err := runPythonFTP([]string{"list", "--ip", f.ip, "--code", f.code, "--path", path})
	if err != nil {
		return nil, err
	}

	var items []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &items); err != nil {
		return nil, fmt.Errorf("parse: %v (output: %s)", err, out)
	}
	files := make([]SDFileInfo, len(items))
	for i, it := range items {
		files[i] = SDFileInfo{Name: it.Name, Size: it.Size}
	}
	return files, nil
}

type tempFileReader struct {
	*os.File
	path string
}

func (t *tempFileReader) Close() error {
	err := t.File.Close()
	os.Remove(t.path)
	return err
}

func (f *printerFTP) retr(path string) (io.ReadCloser, error) {
	// Download to temp file via --out, then serve
	outFile, err := os.CreateTemp("", "printerfarm_dl_*")
	if err != nil {
		return nil, err
	}
	outName := outFile.Name()
	outFile.Close()

	_, err = runPythonFTP([]string{"retr", "--ip", f.ip, "--code", f.code, "--path", path, "--out", outName})
	if err != nil {
		os.Remove(outName)
		return nil, err
	}

	f2, err := os.Open(outName)
	if err != nil {
		os.Remove(outName)
		return nil, err
	}
	return &tempFileReader{File: f2, path: outName}, nil
}

func (f *printerFTP) stor(remotePath string, r io.Reader) error {
	return runPythonFTPWithInput([]string{"stor", "--ip", f.ip, "--code", f.code, "--path", remotePath}, r)
}

func findPython() string {
	// Try PATH first (linux/mac or properly configured Windows)
	for _, name := range []string{"pythonw", "pythonw3", "python3", "python"} {
		if path, err := exec.LookPath(name); err == nil {
			// On Windows, skip the Microsoft Store stub (returns exit code 9009)
			cmd := exec.Command(path, "--version")
			hideWindow(cmd)
			if err := cmd.Run(); err == nil {
				return path
			}
		}
	}
	// Search common Windows Python/Anaconda locations
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		homeDir + `naconda3\python.exe`,
		homeDir + `\miniconda3\python.exe`,
		homeDir + `\AppData\Local\Programs\Python\Python312\python.exe`,
		homeDir + `\AppData\Local\Programs\Python\Python311\python.exe`,
		homeDir + `\AppData\Local\Programs\Python\Python310\python.exe`,
		homeDir + `\AppData\Local\Programs\Python\Python39\python.exe`,
		`C:\Python312\python.exe`,
		`C:\Python311\python.exe`,
		`C:\Python310\python.exe`,
		`C:\Python39\python.exe`,
		`C:naconda3\python.exe`,
		`C:\ProgramDatanaconda3\python.exe`,
		`C:\miniconda3\python.exe`,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "python"
}

func (f *printerFTP) quit() {}

func listSD(p Printer) ([]SDFileInfo, error) {
	f, err := printerFTPConnect(p.IP, p.Code)
	if err != nil {
		return nil, err
	}
	defer f.quit()
	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	return f.list(sdPath)
}

func fetchSDSummary(p Printer) SDCardInfo {
	info := SDCardInfo{IP: p.IP, Name: p.Name}
	files, err := listSD(p)
	if err != nil {
		info.Error = err.Error()
		info.ErrorCode = ftpErrKey(info.Error)
		// A "timeout"/TLS-looking failure on port 990 is often really the
		// printer turning us away with a plaintext "421 too many connections"
		// (its per-IP FTP limit). Knock once more and read the greeting so the
		// real cause is shown instead of a misleading LAN-mode hint.
		if info.ErrorCode != "ftpBusy" {
			if fp := probeFTP990(p.IP); fp.Busy421 {
				info.Error = "Drucker meldet zu viele FTP-Verbindungen (421) — andere FTP-Clients schließen oder Drucker kurz neu starten"
				info.ErrorCode = "ftpBusy"
			}
		}
		return info
	}
	for _, f := range files {
		info.FileCount++
		info.TotalSize += f.Size
	}
	return info
}

func fetchSDFiles(p Printer) SDCardInfo {
	info := SDCardInfo{IP: p.IP, Name: p.Name}
	files, err := listSD(p)
	if err != nil {
		info.Error = err.Error()
		info.ErrorCode = ftpErrKey(info.Error)
		// A "timeout"/TLS-looking failure on port 990 is often really the
		// printer turning us away with a plaintext "421 too many connections"
		// (its per-IP FTP limit). Knock once more and read the greeting so the
		// real cause is shown instead of a misleading LAN-mode hint.
		if info.ErrorCode != "ftpBusy" {
			if fp := probeFTP990(p.IP); fp.Busy421 {
				info.Error = "Drucker meldet zu viele FTP-Verbindungen (421) — andere FTP-Clients schließen oder Drucker kurz neu starten"
				info.ErrorCode = "ftpBusy"
			}
		}
		return info
	}
	info.Files = files
	for _, f := range files {
		info.FileCount++
		info.TotalSize += f.Size
	}
	return info
}

// sdErr builds an SDCardInfo error response with a matching translation key.
func sdErr(ip, name, msg, code string) SDCardInfo {
	return SDCardInfo{IP: ip, Name: name, Error: msg, ErrorCode: code}
}

func handleSyncUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, `{"error":"ip required"}`, 400)
		return
	}
	mu.Lock()
	var printer *Printer
	for i, p := range state.Printers {
		if p.IP == ip {
			printer = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if printer == nil {
		http.Error(w, `{"error":"printer not found"}`, 404)
		return
	}
	r.ParseMultipartForm(200 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 400)
		return
	}
	defer file.Close()
	// Accept only print files — otherwise junk ends up on the SD.
	lname := strings.ToLower(header.Filename)
	if !strings.HasSuffix(lname, ".gcode") && !strings.HasSuffix(lname, ".3mf") {
		http.Error(w, `{"error":"Falsches Dateiformat: nur .gcode / .3mf erlaubt"}`, 400)
		return
	}
	tmp, err := os.CreateTemp("", "printerfarm_upload_*_"+header.Filename)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	tmp.Close()
	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	if !strings.HasSuffix(sdPath, "/") {
		sdPath += "/"
	}
	f2, err := os.Open(tmp.Name())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	defer f2.Close()

	// Every attempt is recorded — even the failed one. That one is
	// interesting later when someone asks why a file is missing.
	start := time.Now()
	eintrag := UploadEntry{IP: ip, Name: printer.Name, Datei: header.Filename, Bytes: header.Size}

	fc := &printerFTP{ip: ip, code: printer.Code}
	if err := fc.stor(sdPath+header.Filename, f2); err != nil {
		friendly := friendlyFTPError(err.Error())
		// Schreiben schlaegt fehl, obwohl FTP grundsaetzlich antwortet? Dann steckt
		// usually no SD/USB card in the printer — that is the most common cause.
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "550") || strings.Contains(low, "no such") ||
			strings.Contains(low, "not found") || strings.Contains(low, "permission") ||
			strings.Contains(low, "denied") || strings.Contains(low, "write") ||
			strings.Contains(low, "storage") || strings.Contains(low, "space") {
			friendly = "Upload fehlgeschlagen — steckt eine SD-Karte/USB im Drucker? Der Speicher lässt sich nicht beschreiben."
		}
		eintrag.Fehler = friendly
		eintrag.Dauer = time.Since(start).Milliseconds()
		noteUpload(eintrag)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]any{"error": friendly})
		return
	}
	eintrag.Erfolg = true
	eintrag.Dauer = time.Since(start).Milliseconds()
	noteUpload(eintrag)

	w.Write([]byte(`{"ok":true,"file":"` + header.Filename + `"}`))
}

func handleSyncDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE only", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	file := r.URL.Query().Get("file")
	if ip == "" || file == "" {
		http.Error(w, `{"error":"ip and file required"}`, 400)
		return
	}
	mu.Lock()
	var printer *Printer
	for i, p := range state.Printers {
		if p.IP == ip {
			printer = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if printer == nil {
		http.Error(w, `{"error":"not found"}`, 404)
		return
	}
	if err := deleteSDFile(ip, file); err != nil {
		http.Error(w, `{"error":"`+strings.TrimSpace(err.Error())+`"}`, 500)
		return
	}
	w.Write([]byte(`{"ok":true}`))
}

// deleteSDFile removes a file from the SD card. Extracted so the
// bulk delete takes the same path as deleting single files.
func deleteSDFile(ip, file string) error {
	mu.Lock()
	code := ""
	for _, p := range state.Printers {
		if p.IP == ip {
			code = p.Code
			break
		}
	}
	mu.Unlock()
	if code == "" {
		return fmt.Errorf("unbekannter Drucker %s", ip)
	}

	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	if !strings.HasSuffix(sdPath, "/") {
		sdPath += "/"
	}
	// Path separators in the file name would be a way out of the SD directory
	if strings.ContainsAny(file, "/\\") {
		return fmt.Errorf("ungültiger Dateiname %q", file)
	}
	_, err := runPythonFTP([]string{"delete", "--ip", ip, "--code", code, "--path", sdPath + file})
	return err
}

func handleSyncDiskSpace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, `{"error":"ip required"}`, 400)
		return
	}
	mu.Lock()
	var printer *Printer
	for i, p := range state.Printers {
		if p.IP == ip {
			printer = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if printer == nil {
		w.Write([]byte(`{}`))
		return
	}
	out, err := runPythonFTP([]string{"diskspace", "--ip", ip, "--code", printer.Code, "--path", "/"})
	if err != nil {
		w.Write([]byte(`{}`))
		return
	}
	w.Write([]byte(strings.TrimSpace(out)))
}

// ─── BULK DELETE AND STORAGE OVERVIEW ─────────────────────────────────────────

type deleteResult struct {
	Path  string `json:"path"`
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// deleteSDFiles entfernt mehrere Dateien eines Druckers in einer einzigen
// FTP session. Previously each file ran its own Python process with its own
// login — with over a hundred files that was the bottleneck.
func deleteSDFiles(ip string, files []string) ([]deleteResult, error) {
	code, err := printerCode(ip)
	if err != nil {
		return nil, err
	}

	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	if !strings.HasSuffix(sdPath, "/") {
		sdPath += "/"
	}

	var paths []string
	var rejected []deleteResult
	for _, f := range files {
		if strings.ContainsAny(f, "/\\") {
			rejected = append(rejected, deleteResult{Path: f, Error: "ungültiger Dateiname"})
			continue
		}
		paths = append(paths, sdPath+f)
	}
	if len(paths) == 0 {
		return rejected, nil
	}

	payload, _ := json.Marshal(paths)
	out, err := runPythonFTPOut([]string{"deletemany", "--ip", ip, "--code", code}, bytes.NewReader(payload))
	if err != nil {
		return rejected, err
	}

	var res []deleteResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &res); err != nil {
		return rejected, fmt.Errorf("unerwartete Antwort: %v", err)
	}
	// Map the full path back to the file name
	for i := range res {
		res[i].Path = strings.TrimPrefix(res[i].Path, sdPath)
	}
	return append(rejected, res...), nil
}

func printerCode(ip string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	for _, p := range state.Printers {
		if p.IP == ip {
			if p.Code == "" {
				return "", fmt.Errorf("%s: kein Zugangscode hinterlegt", ip)
			}
			return p.Code, nil
		}
	}
	return "", fmt.Errorf("unbekannter Drucker %s", ip)
}

// runPythonFTPOut like runPythonFTP, but also passes input through.
func runPythonFTPOut(args []string, stdin io.Reader) (string, error) {
	script, err := writePyScript()
	if err != nil {
		return "", err
	}
	defer os.Remove(script)

	unlock := lockFTP(argIP(args))
	defer unlock()

	cmd := exec.Command(findPython(), append([]string{script}, args...)...)
	hideWindow(cmd)
	cmd.Stdin = stdin
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s", friendlyFTPError(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("%s", friendlyFTPError(err.Error()))
	}
	return string(out), nil
}

type storageEntry struct {
	Path string `json:"path"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`
}

// handleSyncDiff reports per printer which files on the SD are NOT
// present in the assigned (model) folder. Basis for the hint in
// the UI. Only reachable printers with an access code are checked.
func handleSyncDiff(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	printers := append([]Printer(nil), state.Printers...)
	mu.Unlock()
	sdPath := syncCfg.SDPath
	if sdPath == "" {
		sdPath = "/"
	}
	folderCache := map[string]map[string]string{}
	type diffRow struct {
		IP     string   `json:"ip"`
		Name   string   `json:"name"`
		Model  string   `json:"model"`
		Folder string   `json:"folder"`
		Extra  []string `json:"extra"`
		Err    string   `json:"err,omitempty"`
	}
	out := []diffRow{}
	for _, p := range printers {
		if strings.TrimSpace(p.Code) == "" {
			continue
		}
		if s := mqttMgr.GetStatus(p.IP); s == nil || !s.Online {
			continue
		}
		folder := syncFolderFor(p.Model)
		row := diffRow{IP: p.IP, Name: p.Name, Model: p.Model, Folder: folder}
		lf, ok := folderCache[folder]
		if !ok {
			l, _ := listLocalGcode(folder)
			if l == nil {
				l = map[string]string{}
			}
			lf = l
			folderCache[folder] = lf
		}
		inFolder := map[string]bool{}
		for name := range lf {
			inFolder[strings.ToLower(name)] = true
		}
		fc, err := printerFTPConnect(p.IP, p.Code)
		if err != nil {
			row.Err = "FTP: " + err.Error()
			out = append(out, row)
			continue
		}
		existing, err := fc.list(sdPath)
		fc.quit()
		if err != nil {
			row.Err = "SD: " + err.Error()
			out = append(out, row)
			continue
		}
		for _, e := range existing {
			low := strings.ToLower(e.Name)
			if !strings.HasSuffix(low, ".gcode") && !strings.HasSuffix(low, ".3mf") {
				continue
			}
			if !inFolder[low] {
				row.Extra = append(row.Extra, e.Name)
			}
		}
		if len(row.Extra) > 0 || row.Err != "" {
			out = append(out, row)
		}
	}
	writeJSON(w, out)
}
