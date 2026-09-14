package main

import (
	"bytes"
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

// ─── SYNC CONFIG ──────────────────────────────────────────────────────────────

type SyncConfig struct {
	RefPath      string `json:"ref_path"` // local or UNC path
	AutoSync     bool   `json:"auto_sync"`
	AutoInterval int    `json:"auto_interval"` // minutes, default 60
	SDPath       string `json:"sd_path"`       // path on SD card, default "/"
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
	os.WriteFile(syncCfgFile, data, 0644)
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

	// Check ref path is accessible
	refPath := syncCfg.RefPath
	if refPath == "" {
		syncProgress = SyncProgress{
			Status: SyncIdle,
			Errors: []string{"Kein Referenz-Ordner konfiguriert"},
			Log:    []string{"Fehler: Kein Referenz-Ordner konfiguriert"},
		}
		syncMu.Unlock()
		return
	}

	if _, err := os.Stat(refPath); err != nil {
		syncProgress = SyncProgress{
			Status: SyncIdle,
			Errors: []string{"Referenz-Ordner nicht erreichbar: " + refPath},
			Log:    []string{"Fehler: Referenz-Ordner nicht erreichbar — Sync abgebrochen"},
		}
		syncMu.Unlock()
		return
	}

	// Load local gcode files
	localFiles, err := listLocalGcode(refPath)
	if err != nil || len(localFiles) == 0 {
		msg := "Keine .gcode Dateien im Referenz-Ordner gefunden"
		if err != nil {
			msg = "Fehler beim Lesen des Referenz-Ordners: " + err.Error()
		}
		syncProgress = SyncProgress{Status: SyncIdle, Errors: []string{msg}, Log: []string{msg}}
		syncMu.Unlock()
		return
	}

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
		Log:    []string{fmt.Sprintf("Sync gestartet — %d Drucker, %d lokale Dateien", len(printers), len(localFiles))},
	}
	syncMu.Unlock()

	logSync(fmt.Sprintf("Referenz-Ordner: %s", refPath))

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

		if err := syncPrinter(p, localFiles); err != nil {
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
					json.NewEncoder(w).Encode(SDCardInfo{IP: ip, Error: "Timeout nach 20s"})
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
				results[idx] = SDCardInfo{IP: pr.IP, Name: pr.Name, Error: "Timeout"}
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

// Python FTP script with SSL session reuse (works with dem vsftpd des Druckers)
const ftpPyScript = `
import sys, ssl, ftplib, os, json, socket

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

    ftp.quit()

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

// friendlyFTPError uebersetzt das, was das Python-Hilfsskript im Fehlerfall auf
// stderr schreibt, in einen Satz, den man lesen kann. Vorher landete der
// vollstaendige Python-Traceback in der Oberflaeche — sechs Zeilen Technik, aus
// denen niemand ablesen konnte, dass schlicht der Drucker nicht antwortet.
func friendlyFTPError(raw string) string {
	t := strings.TrimSpace(raw)
	if t == "" {
		return "Drucker antwortet nicht"
	}
	// Nur die letzte Zeile des Tracebacks traegt die eigentliche Meldung.
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
	case strings.Contains(low, "timed out"), strings.Contains(low, "timeout"),
		strings.Contains(low, "10060"), strings.Contains(low, "etimedout"):
		return "Zeitüberschreitung — Drucker antwortet nicht auf Port 990"
	case strings.Contains(low, "refused"), strings.Contains(low, "10061"):
		return "Verbindung abgelehnt — FTP am Drucker aus oder falscher Port"
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

	cmd := exec.Command(findPython(), append([]string{script}, args...)...)
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s", friendlyFTPError(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("%s", friendlyFTPError(err.Error()))
	}
	return string(out), nil
}

func runPythonFTPWithInput(args []string, r io.Reader) error {
	script, err := writePyScript()
	if err != nil {
		return err
	}
	defer os.Remove(script)

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
		return info
	}
	info.Files = files
	for _, f := range files {
		info.FileCount++
		info.TotalSize += f.Size
	}
	return info
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

	// Jeder Versuch wird festgehalten — auch der gescheiterte. Gerade der ist
	// spaeter interessant, wenn jemand fragt, warum eine Datei fehlt.
	start := time.Now()
	eintrag := UploadEintrag{IP: ip, Name: printer.Name, Datei: header.Filename, Bytes: header.Size}

	fc := &printerFTP{ip: ip, code: printer.Code}
	if err := fc.stor(sdPath+header.Filename, f2); err != nil {
		eintrag.Fehler = friendlyFTPError(err.Error())
		eintrag.Dauer = time.Since(start).Milliseconds()
		merkeUpload(eintrag)
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	eintrag.Erfolg = true
	eintrag.Dauer = time.Since(start).Milliseconds()
	merkeUpload(eintrag)

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

// deleteSDFile entfernt eine Datei von der SD-Karte. Ausgelagert, damit der
// Mehrfach-Löschvorgang denselben Weg nimmt wie das Löschen einzelner Dateien.
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
	// Pfadtrennzeichen im Dateinamen wären ein Weg aus dem SD-Verzeichnis heraus
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

// ─── GEBÜNDELTES LÖSCHEN UND SPEICHER-ÜBERSICHT ───────────────────────────────

type deleteResult struct {
	Path  string `json:"path"`
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// deleteSDFiles entfernt mehrere Dateien eines Druckers in einer einzigen
// FTP-Sitzung. Vorher lief je Datei ein eigener Python-Prozess mit eigener
// Anmeldung — bei über hundert Dateien war das der Flaschenhals.
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
	// Vollen Pfad wieder auf den Dateinamen zurückführen
	for i := range res {
		res[i].Path = strings.TrimPrefix(res[i].Path, sdPath)
	}
	return append(rejected, res...), nil
}

func printerCode(ip string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
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

// runPythonFTPOut wie runPythonFTP, reicht aber zusätzlich Eingaben durch.
func runPythonFTPOut(args []string, stdin io.Reader) (string, error) {
	script, err := writePyScript()
	if err != nil {
		return "", err
	}
	defer os.Remove(script)

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
