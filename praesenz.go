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

// ─── MEHR-PC-NUTZUNG ÜBER DIE SD-KARTE DES DRUCKERS ───────────────────────────
//
// Statt eines geteilten Ordners auf dem PC nutzen wir die SD-Karte EINES Druckers
// als gemeinsame Pinnwand. Dort liegt "multi-access-controll.json" mit einer
// kurzen Liste aktiver Sitzungen. Jede Instanz frischt in Abstaenden ihren
// eigenen Eintrag auf (den „Cookie": ist er schon da mit unserer Kennung,
// aendert sich nichts) und liest die der anderen. Trägt ein anderer PC einen
// frischen Eintrag, laeuft er parallel — die Oberflaeche zeigt oben ein Band.
//
// Bewusst nur EIN Drucker (der erste erreichbare) als Pinnwand, damit nicht auf
// allen Geraeten FTP-Verkehr entsteht. Faellt er aus, uebernimmt automatisch der
// naechste. Alles best-effort: klappt FTP nicht (Drucker offline, kein Python),
// bleibt die Erkennung still — ohne Fehler.

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

type multiAccessDatei struct {
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

func beendePraesenz() { logZugriff("Stop") }

// mischeSessions raeumt veraltete Sitzungen weg, frischt den eigenen Eintrag auf
// (Schluessel ist der Hostname = ein Eintrag je PC) und liefert die anderen
// aktiven PCs zurueck. Reine Logik — dadurch testbar.
func mischeSessions(alt []accessSession, host, user, instanz string, jetzt time.Time) (neu []accessSession, andere []string) {
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

// waehleZielDrucker sucht den ersten erreichbaren Drucker mit Zugangscode und
// Seriennummer — er dient als Pinnwand.
func waehleZielDrucker() (Printer, bool) {
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

// aktualisiereMultiAccess liest die Pinnwand-Datei vom Drucker, frischt den
// eigenen Eintrag auf, ermittelt andere PCs und schreibt die Datei zurueck.
func aktualisiereMultiAccess() {
	ziel, ok := waehleZielDrucker()
	if !ok {
		return
	}
	fc := &printerFTP{ip: ziel.IP, code: ziel.Code}
	pfad := multiSDPfad()

	var datei multiAccessDatei
	if rc, err := fc.retr(pfad); err == nil {
		b, _ := io.ReadAll(rc)
		rc.Close()
		json.Unmarshal(b, &datei) // Fehler = leere Datei, ist ok
	}

	neu, andere := mischeSessions(datei.Sessions, eigenerHost, winUser(), instanzID, time.Now())
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
		if !netzPausiert() {
			aktualisiereMultiAccess()
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
