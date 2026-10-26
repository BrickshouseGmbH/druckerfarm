package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ─── MULTI-PC USE VIA THE PRINTER'S SD CARD ───────────────────────────────────
//
// Instead of a shared folder on the PC we use ONE printer's SD card
// as a shared pinboard. It holds "multi-access-controll.json" with an
// kurzen Liste aktiver Sitzungen. Jede Instanz frischt in Abstaenden ihren
// own entry (the "cookie": if it is already there with our id,
// nothing changes) and reads the others'. If another PC adds a
// fresh entry, it runs in parallel — the UI shows a banner on top.
//
// Deliberately only ONE printer (the first reachable) as pinboard, so FTP
// traffic does not hit every device. If it fails, the next one automatically
// takes over. All best-effort: if FTP fails (printer offline, no Python),
// detection stays silent — without error.

var (
	instanzID   string
	eigenerHost string
	eigenerName string
	zugriffLog  string

	multiMu     sync.Mutex
	andereAktiv []string
)

const (
	multiDatei = "multi-access-controll.json"
	multiAlter = 4 * time.Minute // aeltere Sitzungen gelten als beendet
)

type accessSession struct {
	Host    string    `json:"host"`
	User    string    `json:"user"`
	Instanz string    `json:"instanz"`
	Zuletzt time.Time `json:"zuletzt"`
}

type multiAccessFile struct {
	Sessions []accessSession `json:"sessions"`
}

func init() { instanzID = "DF" + zufallHex(3) }

func zufallHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%06x", time.Now().UnixNano()&0xffffff)
	}
	return hex.EncodeToString(b)
}

func initPraesenz(appDir string) {
	eigenerHost = hostName()
	eigenerName = winUser() + "@" + eigenerHost
	if appDir != "" {
		zugriffLog = filepath.Join(appDir, "zugriff.log")
		logZugriff("Start")
	}
	go multiAccessLoop()
}

func hostName() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "unbekannt"
}

func winUser() string {
	for _, k := range []string{"USERNAME", "USER"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return "?"
}

func logZugriff(was string) {
	if zugriffLog == "" {
		return
	}
	zeile := fmt.Sprintf("%s  %s  %s (PID %d)\r\n",
		time.Now().Format("2006-01-02 15:04:05"), eigenerName, was, os.Getpid())
	if fh, err := os.OpenFile(zugriffLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		fh.WriteString(zeile)
		fh.Close()
	}
}

func stopPresence() { logZugriff("Stop") }

// mergeSessions clears stale sessions, refreshes our own entry
// (key is the hostname = one entry per PC) and returns the others
// aktiven PCs zurueck. Reine Logik — dadurch testbar.
func mergeSessions(alt []accessSession, host, user, instanz string, jetzt time.Time) (neu []accessSession, andere []string) {
	neu = []accessSession{}
	eigenerDa := false
	for _, s := range alt {
		if jetzt.Sub(s.Zuletzt) > multiAlter {
			continue // gilt als beendet
		}
		if strings.EqualFold(s.Host, host) {
			s.User, s.Instanz, s.Zuletzt = user, instanz, jetzt // eigenen auffrischen
			eigenerDa = true
		}
		neu = append(neu, s)
	}
	if !eigenerDa {
		neu = append(neu, accessSession{Host: host, User: user, Instanz: instanz, Zuletzt: jetzt})
	}
	for _, s := range neu {
		if !strings.EqualFold(s.Host, host) {
			andere = append(andere, s.User+"@"+s.Host)
		}
	}
	sort.Strings(andere)
	return neu, andere
}

// pickTargetPrinter finds the first reachable printer with an access code and
// Seriennummer — er dient als Pinnwand.
func pickTargetPrinter() (Printer, bool) {
	mu.Lock()
	printers := append([]Printer(nil), state.Printers...)
	mu.Unlock()
	for _, p := range printers {
		if strings.TrimSpace(p.Serial) == "" || strings.TrimSpace(p.Code) == "" {
			continue
		}
		if s := mqttMgr.GetStatus(p.IP); s != nil && s.Online {
			return p, true
		}
	}
	return Printer{}, false
}

func multiSDPfad() string {
	base := strings.TrimRight(strings.TrimSpace(syncCfg.SDPath), "/")
	if base == "" {
		return "/" + multiDatei
	}
	return base + "/" + multiDatei
}

// refreshMultiAccess reads the pinboard file from the printer, refreshes
// our own entry, detects other PCs and writes the file back.
func refreshMultiAccess() {
	ziel, ok := pickTargetPrinter()
	if !ok {
		return
	}
	fc := &printerFTP{ip: ziel.IP, code: ziel.Code}
	pfad := multiSDPfad()

	var datei multiAccessFile
	if rc, err := fc.retr(pfad); err == nil {
		b, _ := io.ReadAll(rc)
		rc.Close()
		json.Unmarshal(b, &datei) // error = empty file, that's ok
	}

	neu, andere := mergeSessions(datei.Sessions, eigenerHost, winUser(), instanzID, time.Now())
	datei.Sessions = neu

	multiMu.Lock()
	andereAktiv = andere
	multiMu.Unlock()

	if b, err := json.MarshalIndent(datei, "", "  "); err == nil {
		fc.stor(pfad, bytes.NewReader(b))
	}
}

func multiAccessLoop() {
	time.Sleep(25 * time.Second) // dem Start Zeit lassen
	for {
		if !netPaused() {
			refreshMultiAccess()
		}
		time.Sleep(90 * time.Second)
	}
}

func handlePraesenz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	multiMu.Lock()
	andere := append([]string(nil), andereAktiv...)
	multiMu.Unlock()
	json.NewEncoder(w).Encode(map[string]any{"selbst": eigenerName, "andere": andere})
}
