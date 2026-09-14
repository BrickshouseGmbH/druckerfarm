package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Printer struct {
	Model  string `json:"model"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Code   string `json:"code"`
	Serial string `json:"serial"`
	Fav    bool   `json:"fav"`
	Added  int64  `json:"added"`
}

type AppState struct {
	Printers       []Printer         `json:"printers"`
	Cameras        []CameraCfg       `json:"cameras,omitempty"`
	SetupDismissed bool              `json:"setup_dismissed"`
	UpdateRepo     string            `json:"update_repo"`
	ComponentSHA   map[string]string `json:"component_sha,omitempty"`
	BlinkModels    []string          `json:"blink_models,omitempty"`
	BlinkingIPs    []string          `json:"blinking_ips,omitempty"`
	ErrorBlink     *bool             `json:"error_blink,omitempty"`
	UpdateToken    string            `json:"update_token,omitempty"`
	Theme          string            `json:"theme,omitempty"`
	Lang           string            `json:"lang,omitempty"`
	Cols           int               `json:"cols,omitempty"`
	FilterPillen   map[string]bool   `json:"filter_pillen,omitempty"`

	// UploadLog haelt fest, was wann wohin geladen wurde — dauerhaft, damit es
	// einen Neustart uebersteht.
	UploadLog []UploadEintrag `json:"upload_log,omitempty"`

	// Laufzeit in Sekunden je Drucker, von diesem Programm mitgezaehlt.
	Laufzeit map[string]int64 `json:"laufzeit,omitempty"`

	// KameraSchema haelt fest, welche Adressform an einem Geraet tatsaechlich
	// ein Bild geliefert hat — gemessen, nicht aus dem Modellnamen geraten.
	KameraSchema map[string]string `json:"kamera_schema,omitempty"`

	// FwUpdates: je Drucker die offenen Firmware-Updates, die der Drucker
	// selbst gemeldet hat. Bleibt gespeichert, bis er sie installiert hat.
	FwUpdates map[string][]FwModulUpdate `json:"fw_updates,omitempty"`

	// Reparatur: je Drucker (IP) die „in Reparatur"-Markierung. Quelle ist die
	// maintenance.json auf der SD; hier lokal zwischengespeichert.
	Reparatur map[string]RepairFlag `json:"reparatur,omitempty"`

	// BlinkAus: je Drucker-IP true, wenn das Fehler-Blinken für DIESEN Drucker
	// abgeschaltet ist (zusätzlich zum globalen Schalter).
	BlinkAus map[string]bool `json:"blink_aus,omitempty"`

	// KameraAus: je Drucker-IP true, wenn die Kamera dauerhaft aus ist
	// (Kachel zeigt „Private", kein Stream).
	KameraAus map[string]bool `json:"kamera_aus,omitempty"`

	// StartFilter: Übersichts-Filter beim Start (all|online|offline). Standard all.
	StartFilter string `json:"start_filter,omitempty"`

	// SpracheGewaehlt merkt sich, ob der Anwender beim ersten Start seine
	// Sprache gewaehlt hat. Ist es false, fragt die Oberflaeche einmalig nach.
	SpracheGewaehlt bool `json:"sprache_gewaehlt,omitempty"`
}

var (
	mu         sync.Mutex
	state      AppState
	dataFile   string
	appDir     string
	go2rtcCmd  *exec.Cmd
	go2rtcMu   sync.Mutex
	go2rtcPort = 1984
	appPort    = 8765
)

func main() {
	hideConsoleOnWindows()

	exe, _ := os.Executable()
	exePath = exe
	exeDir := filepath.Dir(exe)

	// Alle Daten liegen in %APPDATA%\Druckerfarm, nicht mehr neben der Exe.
	// Damit ist die Exe austauschbar (Voraussetzung fuers Update) und
	// Zugangsdaten landen nicht in einem synchronisierten Projektordner.
	appDir = dataHome()
	if appDir == "" {
		appDir = exeDir
	}
	os.MkdirAll(appDir, 0o755)
	dataFile = filepath.Join(appDir, "config.json")
	setupLogFile()
	migrateFromExeDir(exeDir)
	cleanupOldExe()

	loadState()
	initSync(appDir)
	initPraesenz(appDir) // lokales Zugriffs-Log; Parallel-PC ueber SD-Datei
	go reparaturLoop()   // Reparatur-Markierungen von der SD einlesen/abgleichen

	// Zuerst verwaiste go2rtc-Prozesse aufraeumen. Ein abgestuerzter oder
	// ersetzter Programmlauf kann welche hinterlassen; ohne das haeuften sie
	// sich (schon 10-fach gesehen) und ein Programm-Update griff nicht, solange
	// noch ein alter Prozess Dateien und Port belegte.
	if n, _ := killAlleGo2rtc(); n > 0 {
		log.Printf("🧹 %d verwaiste(n) go2rtc-Prozess(e) beim Start beendet", n)
		time.Sleep(400 * time.Millisecond)
	}
	// go2rtc starten, falls installiert (sonst laeuft die App ohne Video)
	go startGo2rtc()
	// Connect MQTT to all printers
	go connectAllMQTT()
	go laufzeitLoop()
	go startDiscoveryListener()
	// Regelmaessig vollen Status anfordern (verlorene Erstantwort holt sich das zurueck)
	go pushAllLoop()
	// Kammerbeleuchtung bei Fehlern blinken lassen (Modelle ohne Signalleuchte)
	go errorLightLoop()
	go reconnectLoop() // offline Drucker automatisch neu verbinden
	// go2rtc im Blick behalten: Absturz oder Haenger fuehren zum Neustart
	go go2rtcHealthLoop()

	// HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveUI)
	mux.HandleFunc("/api/printers", handlePrinters)
	mux.HandleFunc("/api/printers/", handlePrinterByIP)
	mux.HandleFunc("/api/yaml", handleYaml)
	mux.HandleFunc("/api/export", handleExport)
	mux.HandleFunc("/api/import", handleImport)
	mux.HandleFunc("/api/go2rtc/status", handleG2Status)
	mux.HandleFunc("/api/go2rtc/restart", handleG2Restart)
	mux.HandleFunc("/api/go2rtc/killall", handleG2KillAll)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/errors", handleErrors)
	mux.HandleFunc("/api/debug", handleDebug)
	mux.HandleFunc("/api/sync/status", handleSyncStatus)
	mux.HandleFunc("/api/sync/config", handleSyncConfig)
	mux.HandleFunc("/api/sync/start", handleSyncControl)
	mux.HandleFunc("/api/sync/pause", handleSyncControl)
	mux.HandleFunc("/api/sync/stop", handleSyncControl)
	mux.HandleFunc("/api/sync/browse", handleSyncBrowse)
	mux.HandleFunc("/api/sync/download", handleSyncDownload)
	mux.HandleFunc("/api/sync/sdlist", handleSyncSDList)
	mux.HandleFunc("/api/sync/search/start", handleSearchStart)
	mux.HandleFunc("/api/sync/delete/start", handleDeleteStart)
	mux.HandleFunc("/api/jobs", handleJobsList)
	mux.HandleFunc("/api/job/status", handleJobStatus)
	mux.HandleFunc("/api/job/stop", handleJobStop)
	mux.HandleFunc("/api/open-folder", handleOpenFolder)
	mux.HandleFunc("/api/diagnostics", handleDiagnostics)
	mux.HandleFunc("/api/sync/upload", handleSyncUpload)
	mux.HandleFunc("/api/sync/delete", handleSyncDelete)
	mux.HandleFunc("/api/sync/diskspace", handleSyncDiskSpace)
	mux.HandleFunc("/api/camera/resolution", handleCameraResolution)
	mux.HandleFunc("/api/snapshot/", handleSnapshot)
	mux.HandleFunc("/api/components", handleComponents)
	mux.HandleFunc("/api/components/install", handleComponentsInstall)
	mux.HandleFunc("/api/components/progress", handleComponentsProgress)
	mux.HandleFunc("/api/components/dismiss", handleComponentsDismiss)
	mux.HandleFunc("/api/components/updates", handleComponentUpdates)
	mux.HandleFunc("/api/print/command", handlePrintCommand)
	mux.HandleFunc("/api/discover", handleDiscover)
	mux.HandleFunc("/api/update/check", handleUpdateCheck)
	mux.HandleFunc("/api/update/install", handleUpdateInstall)
	mux.HandleFunc("/api/update/progress", handleUpdateProgress)
	mux.HandleFunc("/api/update/config", handleUpdateConfig)
	mux.HandleFunc("/api/settings", handleSettings)
	mux.HandleFunc("/api/blink", handleBlink)
	mux.HandleFunc("/api/camera-off", handleCameraOff)

	// Add loading page route BEFORE starting server
	mux.HandleFunc("/logo.svg", handleLogo)
	mux.HandleFunc("/splash.jpg", handleSplash)
	mux.HandleFunc("/api/open-licenses", handleOpenLicenses)
	mux.HandleFunc("/api/reach", handleReach)
	mux.HandleFunc("/api/reconnect", handleReconnect)
	mux.HandleFunc("/api/cameras", handleCameras)
	mux.HandleFunc("/api/cameras/", handleCameraByID)
	mux.HandleFunc("/api/xiaomi-login", handleXiaomiProxy)
	mux.HandleFunc("/api/upload-log", handleUploadLog)
	mux.HandleFunc("/api/print/start", handlePrintStart)
	mux.HandleFunc("/api/print/temp", handleSetTemp)
	mux.HandleFunc("/api/ams/filament", handleSetFilament)
	mux.HandleFunc("/api/camera-test", handleCameraTest)
	mux.HandleFunc("/api/network/pause", handleNetzPause)
	mux.HandleFunc("/api/fwupdate/scan", handleFwUpdateScan)
	mux.HandleFunc("/api/fwupdate/start", handleFwUpdateStart)
	mux.HandleFunc("/api/praesenz", handlePraesenz)
	mux.HandleFunc("/api/repair", handleRepair)
	mux.HandleFunc("/loading", serveLoading)
	mux.HandleFunc("/ready", handleReady)

	// Open window on loading page first — avoids white/transparent flash
	url := fmt.Sprintf("http://127.0.0.1:%d/loading", appPort)

	// Synchron binden. Vorher lief das in einer Goroutine und der Fehler wurde
	// verworfen — bei belegtem Port startete die zweite Instanz halb und legte
	// beim Beenden die erste mit lahm.
	addr := fmt.Sprintf("127.0.0.1:%d", appPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Es laeuft bereits eine Instanz: Fenster darauf oeffnen und hier raus.
		log.Printf("ℹ️  Port %d belegt — vermutlich laeuft die App schon. Oeffne ein Fenster darauf.\n", appPort)
		launchAppWindow(url)
		return
	}
	go http.Serve(ln, corsMiddleware(mux))
	waitForPort(appPort)

	// Open app window (Edge/Chrome in --app mode, no address bar)
	shutdown := make(chan struct{})
	winCmd, werr := launchAppWindow(url)
	if werr != nil {
		log.Printf("⚠️  %v\n", werr)
	}

	// Wait for window close or signal
	go waitForWindow(winCmd, shutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		beendeSauber()
		return
	case <-shutdown:
	}

	// Der gestartete msedge.exe-Prozess ist oft nur ein Starter: bei kaltem
	// Profil startet Edge sich selbst neu, und wenn schon ein Edge dieses Profil
	// haelt, reicht er die Kommandozeile weiter und beendet sich sofort. Sein
	// Ende ist also nur ein Hinweis, kein Beweis. Frueher hat genau das hier den
	// Server abgeraeumt, waehrend das Fenster noch startete — Ergebnis war
	// ERR_CONNECTION_REFUSED beim ersten Start.
	waitUntilUIIdle()
	beendeSauber()
}

// beendeSauber raeumt beim Schliessen des Programms auf: die Signalleuchten der
// betroffenen Drucker (X1E/X2D) gehen von "blinkend" zurueck auf Normal (an),
// und go2rtc wird vollstaendig beendet — der eigene Prozess wie auch etwaige
// verwaiste, samt ihrer Kindprozesse (ffmpeg). Ohne das blieb ein blinkendes
// Lämpchen im Druckerraum stehen und go2rtc lief mitunter weiter.
func beendeSauber() {
	// ZUERST go2rtc restlos beenden — das soll sofort passieren. Frueher stand
	// hier das Zuruecksetzen der Signalleuchten davor (bis zu 6 s Wartezeit),
	// wodurch go2rtc erst danach beendet wurde und entsprechend lange lief.
	stopGo2rtc()
	if n, _ := killAlleGo2rtc(); n > 0 {
		log.Printf("🧹 %d go2rtc-Prozess(e) beim Beenden geschlossen", n)
	}

	beendePraesenz() // Stop-Zeile ins Zugriffs-Log, eigene Praesenz entfernen

	// DANACH die Lampen zuruecksetzen, mit Zeitbegrenzung: antwortet ein Drucker
	// nicht, geht es nach ein paar Sekunden trotzdem weiter.
	fertig := make(chan int, 1)
	go func() { fertig <- ResetChamberLights() }()
	select {
	case n := <-fertig:
		if n > 0 {
			log.Printf("🔦 %d Signalleuchte(n) auf Normal (an) zurueckgesetzt", n)
		}
	case <-time.After(4 * time.Second):
		log.Printf("🔦 Zuruecksetzen der Signalleuchten dauert — beende trotzdem")
	}
}

// ─── CORS ─────────────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		noteRequest()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ─── LOADING PAGE ─────────────────────────────────────────────────────────────

func serveLoading(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html><html>
<head><meta charset="UTF-8"><title>Druckerfarm</title>
<style>
*{margin:0;padding:0;box-sizing:border-box;}
body{background:#000 center/cover no-repeat;background-image:url(/splash.jpg);display:flex;align-items:center;justify-content:center;height:100vh;font-family:system-ui,-apple-system,sans-serif;}
.wrap{text-align:center;padding:26px 40px;background:rgba(0,0,0,0.5);border-radius:10px;backdrop-filter:blur(4px);}
h1{color:#fff;font-size:20px;font-weight:800;margin-bottom:6px;}
p{color:rgba(255,255,255,0.75);font-size:13px;margin-bottom:22px;font-family:ui-monospace,Consolas,monospace;}
.bar{width:220px;height:3px;background:rgba(255,255,255,0.2);border-radius:2px;margin:0 auto;overflow:hidden;}
.fill{height:100%;background:#00e5ff;border-radius:2px;animation:load 1.2s ease-in-out infinite;}
@keyframes load{0%{width:0;margin-left:0}50%{width:60%;margin-left:20%}100%{width:0;margin-left:100%}}
</style>
</head>
<body>
<div class="wrap">
  <h1>Druckerfarm</h1>
  <p>wird gestartet …</p>
  <div class="bar"><div class="fill"></div></div>
</div>
<script>
// Poll until the app is ready, then navigate
(function poll(){
  fetch('/ready').then(r=>{
    if(r.ok) window.location.replace('/');
    else setTimeout(poll, 300);
  }).catch(()=>setTimeout(poll, 300));
})();
</script>
</body></html>`))
}

// handleReady returns 200 once go2rtc has had a chance to start (1s grace period).
// Das Logo steckt fest in der Programmdatei — keine Datei daneben, die
// verlorengehen kann.
//
//go:embed assets/logo.svg
var logoSVG []byte

func handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(logoSVG)
}

// Das Startbild steckt ebenfalls fest in der Programmdatei.
//
//go:embed assets/splash.jpg
var splashJPG []byte

// Die Drittanbieter-Lizenzen stecken fest in der Programmdatei, damit sie in der
// App verlinkt und angezeigt werden koennen — unabhaengig davon, ob die
// Markdown-Datei daneben liegt.
//
//go:embed THIRD_PARTY_LICENSES.md
var thirdPartyLicenses string

// Die Uebersetzungen liegen als JSON je Sprache vor und werden fest in die
// Programmdatei eingebettet. Wer eigene Uebersetzungen will, bearbeitet
// lang/de.json bzw. lang/en.json und baut die Exe neu. Beim Ausliefern der
// Seite werden sie an die Stelle __LANGS_JSON__ eingesetzt.
//
//go:embed lang/de.json
var langDE string

//go:embed lang/en.json
var langEN string

// de-en.json ist die vollflaechige Wort-fuer-Wort-Zuordnung Deutsch->Englisch,
// mit der die Oberflaeche bei englischer Sprache automatisch uebersetzt wird —
// auch dynamisch erzeugter Text. Editierbar vor dem Bauen.
//
//go:embed lang/de-en.json
var langDEEN string

// handleOpenLicenses schreibt die eingebetteten Lizenzen als Datei ins
// Datenverzeichnis und zeigt sie im Datei-Explorer des Systems an. So oeffnet
// sich kein Browser-Tab, sondern die Datei liegt greifbar im Ordner.
func handleOpenLicenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	pfad := filepath.Join(appDir, "THIRD_PARTY_LICENSES.md")
	if err := os.WriteFile(pfad, []byte(thirdPartyLicenses), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// /select zeigt die Datei markiert im Explorer-Fenster.
		cmd = exec.Command("explorer.exe", "/select,"+pfad)
	case "darwin":
		cmd = exec.Command("open", "-R", pfad)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(pfad))
	}
	_ = cmd.Start() // explorer liefert auch bei Erfolg != 0
	writeJSON(w, map[string]any{"pfad": pfad})
}

func handleSplash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(splashJPG)
}

var readySince = time.Now()

// exePath ist der Pfad der laufenden Programmdatei — das Update ersetzt sie.
var exePath string

// dataHome liefert das Verzeichnis fuer Konfiguration, Werkzeuge und Log.
func dataHome() string {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		if h := os.Getenv("HOME"); h != "" {
			base = filepath.Join(h, "Library", "Application Support")
		}
	default:
		if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
			base = x
		} else if h := os.Getenv("HOME"); h != "" {
			base = filepath.Join(h, ".config")
		}
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "Druckerfarm")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// migrateFromExeDir holt eine vorhandene Einrichtung einmalig aus dem
// Programmordner herueber, damit nach dem Umzug niemand 42 Drucker neu eintippt.
func migrateFromExeDir(exeDir string) {
	if exeDir == appDir {
		return
	}
	if _, err := os.Stat(dataFile); err == nil {
		return // schon eingerichtet
	}
	old := filepath.Join(exeDir, "config.json")
	if _, err := os.Stat(old); err != nil {
		return // nichts zu uebernehmen
	}

	moved := []string{}
	for _, name := range []string{"config.json", "go2rtc.yaml", "go2rtc" + exeSuffix()} {
		if _, err := os.Stat(filepath.Join(exeDir, name)); err != nil {
			continue
		}
		if err := copyFile(filepath.Join(exeDir, name), filepath.Join(appDir, name)); err == nil {
			moved = append(moved, name)
		}
	}
	if entries, err := os.ReadDir(filepath.Join(exeDir, "ffmpeg")); err == nil {
		n := 0
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if err := copyFile(filepath.Join(exeDir, "ffmpeg", e.Name()), filepath.Join(appDir, "ffmpeg", e.Name())); err == nil {
				n++
			}
		}
		if n > 0 {
			moved = append(moved, fmt.Sprintf("ffmpeg (%d Dateien)", n))
		}
	}
	if len(moved) > 0 {
		log.Printf("Einrichtung aus %s uebernommen: %s", exeDir, strings.Join(moved, ", "))
	}
}

// lastRequest haelt fest, wann das UI zuletzt etwas abgerufen hat. Das
// Dashboard pollt alle 5s, deshalb ist "seit X Sekunden still" das verlaessliche
// Signal dafuer, dass das Fenster wirklich zu ist.
var lastRequest atomic.Int64

// Als Variablen, damit Tests sie verkleinern koennen.
var (
	uiIdleTimeout  = 15 * time.Second
	uiNeverGrace   = 90 * time.Second
	uiIdlePollTick = time.Second
)

func noteRequest() { lastRequest.Store(time.Now().UnixNano()) }

// setupLogFile leitet die Logausgabe in eine Datei um. Mit -H windowsgui gibt es
// keine Konsole — ohne das ist jede Meldung unsichtbar.
func setupLogFile() {
	path := filepath.Join(appDir, "druckerfarm.log")
	if st, err := os.Stat(path); err == nil && st.Size() > 1<<20 {
		os.Remove(path)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags)
	log.Printf("──── Start ────")
}

// waitUntilUIIdle kehrt zurueck, sobald das UI laenger als uiIdleTimeout still
// ist — oder nach uiNeverGrace, falls sich nie jemand verbunden hat.
func waitUntilUIIdle() {
	deadline := time.Now().Add(uiNeverGrace)
	for {
		time.Sleep(uiIdlePollTick)
		last := lastRequest.Load()
		if last == 0 {
			if time.Now().After(deadline) {
				return
			}
			continue
		}
		if time.Since(time.Unix(0, last)) > uiIdleTimeout {
			return
		}
	}
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if time.Since(readySince) > 1200*time.Millisecond {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}

// ─── UI ───────────────────────────────────────────────────────────────────────

func serveUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Set("Surrogate-Control", "no-store")
	w.Write([]byte(seiteMitVersion()))
}

// seiteMitVersion setzt die gebaute Versionsnummer in die Seite ein — einmal
// berechnet, danach aus dem Zwischenspeicher. So steht die Version schon auf dem
// Startbild (Gorilla-Ladeseite), ohne dass die Oberflaeche sie erst nachladen muss.
var (
	seiteEinmal sync.Once
	seiteCache  string
)

func seiteMitVersion() string {
	seiteEinmal.Do(func() {
		langs := `{"de":` + strings.TrimSpace(langDE) + `,"en":` + strings.TrimSpace(langEN) + `}`
		seite := strings.ReplaceAll(dashboardHTML, "__LANGS_JSON__", langs)
		seite = strings.ReplaceAll(seite, "__DEEN_JSON__", strings.TrimSpace(langDEEN))
		seite = strings.ReplaceAll(seite, "__APP_VERSION__", appVersion)
		seiteCache = seite
	})
	return seiteCache
}

// ─── PRINTER API ──────────────────────────────────────────────────────────────

func handlePrinters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode(state.Printers)

	case http.MethodPost:
		var p Printer
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// Der Zugangscode ist nicht mehr Bedingung: ein Drucker darf ohne
		// angelegt und der Code spaeter nachgetragen werden. Ohne ihn gibt es
		// kein Bild und keinen Dateizugriff, aber der Eintrag existiert schon
		// einmal — das ist beim Aufbau einer Farm der uebliche Ablauf.
		if p.Name == "" || p.IP == "" {
			http.Error(w, "name und ip sind erforderlich", 400)
			return
		}
		if p.Added == 0 {
			p.Added = time.Now().UnixMilli()
		}
		mu.Lock()
		// "Schon vorhanden?" entscheidet die Seriennummer, NICHT die IP. Ein
		// Offline-Drucker traegt oft noch eine veraltete IP; bekommt ein neues
		// Geraet dieselbe IP zugewiesen, darf das Anlegen daran nicht scheitern —
		// die Live-IP des neuen Geraets ist die richtige. Nur ohne Seriennummer
		// (manuell, ohne Angabe) bleibt die IP der einzige Anhaltspunkt.
		for _, e := range state.Printers {
			if p.Serial != "" && e.Serial != "" && strings.EqualFold(strings.TrimSpace(e.Serial), strings.TrimSpace(p.Serial)) {
				mu.Unlock()
				http.Error(w, "Drucker mit dieser Seriennummer ist bereits angelegt", 409)
				return
			}
			if p.Serial == "" && e.IP == p.IP {
				mu.Unlock()
				http.Error(w, "IP bereits vorhanden", 409)
				return
			}
		}
		state.Printers = append(state.Printers, p)
		mu.Unlock()
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
		go syncMQTT() // sonst bleibt der neue Drucker bis zum Neustart ohne Status
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handlePrinterByIP(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimPrefix(r.URL.Path, "/api/printers/")
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPut:
		var updated Printer
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		mu.Lock()
		found := false
		for i, p := range state.Printers {
			if p.IP == ip {
				updated.Added = p.Added // preserve original add time
				updated.Fav = p.Fav
				state.Printers[i] = updated
				found = true
				break
			}
		}
		mu.Unlock()
		if !found {
			http.Error(w, "nicht gefunden", 404)
			return
		}
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
		go func() {
			mqttMgr.Disconnect(ip) // Zugangsdaten koennen sich geaendert haben
			syncMQTT()
		}()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)

	case http.MethodDelete:
		mu.Lock()
		newList := []Printer{}
		found := false
		for _, p := range state.Printers {
			if p.IP == ip {
				found = true
				continue
			}
			newList = append(newList, p)
		}
		state.Printers = newList
		mu.Unlock()
		if !found {
			http.Error(w, "nicht gefunden", 404)
			return
		}
		saveState()
		writeGo2rtcYaml()
		mqttMgr.Disconnect(ip)
		w.WriteHeader(204)

	case http.MethodPatch:
		var body struct {
			Fav *bool `json:"fav"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		for i, p := range state.Printers {
			if p.IP == ip {
				if body.Fav != nil {
					state.Printers[i].Fav = *body.Fav
				}
				break
			}
		}
		mu.Unlock()
		saveState()
		w.WriteHeader(204)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

// ─── YAML / EXPORT / IMPORT ───────────────────────────────────────────────────

func handleYaml(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	ps := make([]Printer, len(state.Printers))
	copy(ps, state.Printers)
	mu.Unlock()
	mu.Lock()
	gemessen := map[string]string{}
	for k, v := range state.KameraSchema {
		gemessen[k] = v
	}
	cams := make([]CameraCfg, len(state.Cameras))
	copy(cams, state.Cameras)
	mu.Unlock()
	yaml := buildYaml(ps, gemessen) + cameraStreamsYaml(cams)
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", `attachment; filename="go2rtc.yaml"`)
		w.Header().Set("Content-Type", "text/yaml")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Write([]byte(yaml))
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	ps := make([]Printer, len(state.Printers))
	copy(ps, state.Printers)
	mu.Unlock()
	w.Header().Set("Content-Disposition", `attachment; filename="druckerfarm.csv"`)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Write([]byte("# Name,IP,AccessCode,Modell,Favorit\n"))
	for _, p := range ps {
		fav := "0"
		if p.Fav {
			fav = "1"
		}
		fmt.Fprintf(w, "%s,%s,%s,%s,%s\n", p.Name, p.IP, p.Code, p.Model, fav)
	}
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	added, skipped := 0, 0
	mu.Lock()
	// Duplikate erkennen wir vorrangig an der Seriennummer, nur ersatzweise an
	// der IP (fuer Zeilen ohne Seriennummer). Eine veraltete IP eines
	// Offline-Druckers blockiert so nicht das Anlegen eines neuen Geraets.
	existSerial := map[string]bool{}
	existIP := map[string]bool{}
	for _, p := range state.Printers {
		existIP[p.IP] = true
		if s := strings.TrimSpace(p.Serial); s != "" {
			existSerial[strings.ToLower(s)] = true
		}
	}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sep := ","
		if strings.Contains(line, ";") {
			sep = ";"
		}
		parts := strings.Split(line, sep)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) < 3 {
			skipped++
			continue
		}
		name, ip, code := parts[0], parts[1], parts[2]
		model := "UNBEKANNT"
		if len(parts) >= 4 && parts[3] != "" {
			model = strings.ToUpper(parts[3])
		}
		serial := ""
		if len(parts) >= 5 && parts[4] != "" {
			serial = parts[4]
		}
		schon := false
		if s := strings.ToLower(strings.TrimSpace(serial)); s != "" {
			schon = existSerial[s]
		} else {
			schon = existIP[ip]
		}
		if name == "" || ip == "" || code == "" || schon {
			skipped++
			continue
		}
		state.Printers = append(state.Printers, Printer{
			Model: model, Name: name, IP: ip, Code: code, Serial: serial,
			Added: time.Now().UnixMilli(),
		})
		existIP[ip] = true
		if s := strings.ToLower(strings.TrimSpace(serial)); s != "" {
			existSerial[s] = true
		}
		added++
	}
	mu.Unlock()
	if added > 0 {
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"added": added, "skipped": skipped})
}

// ─── SNAPSHOTS ────────────────────────────────────────────────────────────────
//
// The UI used to call go2rtc's /api/frame.jpeg directly. Two problems with that:
// it had to rebuild the go2rtc stream name in JavaScript (and got it wrong), and
// go2rtc answers 200 with an empty body when it cannot produce a frame, so every
// failure looked like a success. On top of that each snapshot spawns an ffmpeg
// transcode — 36 printers polling in parallel is enough to kill go2rtc.
//
// So snapshots go through here: the stream name comes from streamName(), frames
// are cached per stream, at most snapMaxConcurrent transcodes run at once, and a
// missing frame turns into a real HTTP error with a readable reason.

const (
	snapMaxConcurrent = 4
	snapMinInterval   = 1500 * time.Millisecond
	// Kurz gehalten: ein haengender Drucker belegt sonst einen der vier
	// Transcode-Plaetze und bremst die gesunden aus.
	snapFetchTimeout = 8 * time.Second
	snapMaxBackoff   = 2 * time.Minute

	// Jede Bildanfrage startet in go2rtc einen eigenen ffmpeg-Vorgang. Mehr als
	// das haelt es auf Dauer nicht aus — bei 42 Druckern alle 2 s waeren es 21
	// Anfragen pro Sekunde, die sich vor vier Bearbeitungsplaetzen stauen.
	snapMaxRatePerSec = 3.0
)

type snapEntry struct {
	mu      sync.Mutex
	data    []byte
	ts      time.Time
	err     error
	fails   int       // aufeinanderfolgende Fehlschlaege
	nextTry time.Time // vorher wird gar nicht erst angefragt
}

// backoffFor waechst mit jedem Fehlschlag: 15 s, 30 s, 60 s, dann 2 min.
// Damit belegt ein dauerhaft kaputter Stream keinen Platz mehr.
func backoffFor(fails int) time.Duration {
	d := time.Duration(1<<uint(min(fails, 4))) * 15 * time.Second / 2
	if d > snapMaxBackoff {
		d = snapMaxBackoff
	}
	if d < 15*time.Second {
		d = 15 * time.Second
	}
	return d
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	snapMu     sync.Mutex
	snapCache  = map[string]*snapEntry{}
	snapSem    = make(chan struct{}, snapMaxConcurrent)
	snapClient = &http.Client{Timeout: snapFetchTimeout}
)

// effectiveSnapInterval sagt, wie oft ein einzelner Drucker tatsaechlich ein
// neues Bild bekommen kann, ohne go2rtc zu ueberfahren. Der gewuenschte Wert
// wird nur eingehalten, solange die Gesamtrate das hergibt.
func effectiveSnapInterval(requested time.Duration) time.Duration {
	mu.Lock()
	n := len(state.Printers)
	mu.Unlock()
	if n == 0 {
		n = 1
	}

	rate := snapMaxRatePerSec
	// Ist go2rtc in letzter Zeit mehrfach abgestuerzt, weiter zurueckfahren.
	go2rtcMu.Lock()
	fails := go2rtcFails
	go2rtcMu.Unlock()
	if fails > 0 {
		rate = rate / float64(1+minInt(fails, 4))
	}

	floor := time.Duration(float64(n)/rate*1000) * time.Millisecond
	if floor < snapMinInterval {
		floor = snapMinInterval
	}
	if requested < floor {
		return floor
	}
	return requested
}

func snapEntryFor(stream string) *snapEntry {
	snapMu.Lock()
	defer snapMu.Unlock()
	e := snapCache[stream]
	if e == nil {
		e = &snapEntry{}
		snapCache[stream] = e
	}
	return e
}

// fetchFrame asks go2rtc for a single JPEG. An empty or non-JPEG body is
// reported as an error instead of being passed on as a valid image.
func fetchFrame(stream string) ([]byte, error) {
	snapSem <- struct{}{}
	defer func() { <-snapSem }()

	u := fmt.Sprintf("http://127.0.0.1:%d/api/frame.jpeg?src=%s", go2rtcPort, url.QueryEscape(stream))
	resp, err := snapClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("go2rtc nicht erreichbar: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("go2rtc HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		// go2rtc antwortet in diesem Fall mit 200 und leerem Rumpf und behaelt
		// den Grund fuer sich — der steht nur in seinem eigenen Log. Deshalb
		// wird hier selbst nachgesehen, statt pauschal ffmpeg zu verdaechtigen.
		return nil, fmt.Errorf("kein Bild: %s", diagnoseStream(stream))
	}
	return data, nil
}

// diagnoseStream sucht den tatsaechlichen Grund, warum go2rtc kein Bild liefert.
func diagnoseStream(stream string) string {
	// 1. Kennt go2rtc den Stream ueberhaupt?
	u := fmt.Sprintf("http://127.0.0.1:%d/api/streams?src=%s", go2rtcPort, url.QueryEscape(stream))
	resp, err := snapClient.Get(u)
	if err != nil {
		return "go2rtc antwortet nicht"
	}
	code := resp.StatusCode
	resp.Body.Close()
	if code == http.StatusNotFound {
		return fmt.Sprintf("Stream %q steht nicht in der go2rtc-Konfiguration — go2rtc neu starten", stream)
	}

	// 2. Ist ffmpeg da? Ohne kann go2rtc H264 nicht nach JPEG wandeln.
	if !fileExists(ffmpegBinPath()) {
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			return "ffmpeg fehlt — unter Einstellungen nachladen"
		}
	}

	// 3. Antwortet die Kamera des Druckers?
	// Standalone-Kameras haben keinen Drucker — eigene, kameraspezifische
	// Diagnose (sauber getrennt, siehe camera.go).
	if istKameraStream(stream) {
		return kameraStreamDiagnose(stream)
	}
	ip, name, online := printerForStream(stream)
	if ip == "" {
		return fmt.Sprintf("kein Drucker zum Stream %q gefunden", stream)
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "322"), 2*time.Second)
	if err != nil {
		if online {
			return fmt.Sprintf("%s antwortet per MQTT, die Kamera auf Port 322 aber nicht — Kamera eingeschaltet?", name)
		}
		return fmt.Sprintf("%s ist nicht erreichbar (Port 322 zu, Gerät vermutlich aus)", name)
	}
	conn.Close()

	// 4. Port offen, trotzdem kein Bild — dann stimmt meist der Zugangscode nicht
	return fmt.Sprintf("%s nimmt Verbindungen an, liefert aber kein Bild — Zugangscode prüfen", name)
}

// printerForStream findet den Drucker, aus dessen Namen der Streamname entstand.
func printerForStream(stream string) (ip, name string, online bool) {
	mu.Lock()
	var found *Printer
	for i := range state.Printers {
		if streamName(state.Printers[i]) == stream {
			found = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if found == nil {
		return "", "", false
	}
	s := mqttMgr.GetStatus(found.IP)
	return found.IP, found.Name, s != nil && s.Online
}

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if netzPausiert() {
		http.Error(w, "Netzwerkverkehr ist angehalten", http.StatusServiceUnavailable)
		return
	}
	ip := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/snapshot/"), "/")
	if ip == "" {
		http.Error(w, "IP fehlt", http.StatusBadRequest)
		return
	}

	mu.Lock()
	stream := ""
	for _, p := range state.Printers {
		if p.IP == ip {
			stream = streamName(p)
			break
		}
	}
	if stream == "" {
		// Standalone-Kamera? Dann die Kennung als Kamera-ID auffassen.
		for _, c := range state.Cameras {
			if c.ID == ip {
				if s := strings.TrimSpace(c.Stream); s != "" {
					stream = s
				} else {
					stream = c.ID
				}
				break
			}
		}
	}
	mu.Unlock()
	if stream == "" {
		http.Error(w, "unbekannt: "+ip, http.StatusNotFound)
		return
	}

	maxAge := snapMinInterval
	if v := r.URL.Query().Get("max_age"); v != "" {
		if secs, err := strconv.ParseFloat(v, 64); err == nil && secs > 0 {
			maxAge = time.Duration(secs * float64(time.Second))
		}
	}
	// Auf das anheben, was go2rtc bei dieser Druckerzahl noch verkraftet
	effective := effectiveSnapInterval(maxAge)
	throttled := effective > maxAge
	maxAge = effective

	// Holding the per-stream lock across the fetch collapses concurrent requests
	// for the same printer into a single transcode.
	e := snapEntryFor(stream)
	e.mu.Lock()
	stale := e.ts.IsZero() || time.Since(e.ts) > maxAge
	blocked := time.Now().Before(e.nextTry)
	if stale && !blocked {
		data, err := fetchFrame(stream)
		e.ts = time.Now()
		e.err = err
		if err != nil {
			e.fails++
			e.nextTry = time.Now().Add(backoffFor(e.fails))
			if e.fails == 1 || e.fails%10 == 0 {
				log.Printf("Snapshot %s: %v (Pause %s)", stream, err, backoffFor(e.fails))
			}
		} else {
			e.data = data
			e.fails = 0
			e.nextTry = time.Time{}
		}
	}
	data, cachedErr, age := e.data, e.err, time.Since(e.ts)
	retryIn := time.Until(e.nextTry)
	fails := e.fails
	e.mu.Unlock()

	if len(data) == 0 {
		msg := "kein Frame verfügbar"
		if cachedErr != nil {
			msg = cachedErr.Error()
		}
		if retryIn > 0 {
			msg = fmt.Sprintf("%s — pausiert, nächster Versuch in %.0f s", msg, retryIn.Seconds())
		}
		w.Header().Set("X-Snapshot-Interval", fmt.Sprintf("%.0f", maxAge.Seconds()))
		if throttled {
			w.Header().Set("X-Snapshot-Throttled", "1")
		}
		w.Header().Set("X-Snapshot-Retry-In", fmt.Sprintf("%.0f", retryIn.Seconds()))
		w.Header().Set("X-Snapshot-Fails", fmt.Sprintf("%d", fails))
		// Standalone-Kameras: KEIN 503 zurückgeben. Der Browser protokolliert
		// jeden 4xx/5xx als roten Konsolenfehler — bei einer zeitweise nicht
		// erreichbaren Außenkamera ist das nur Lärm. Stattdessen 200 mit leerem
		// Rumpf und einem Hinweis-Header; die Oberfläche zeigt „nicht verfügbar"
		// ohne Ladekreis und ohne Konsolenfehler und hält die Pause ein.
		if istKameraStream(stream) {
			w.Header().Set("X-Snapshot-Unavailable", "1")
			w.Header().Set("X-Snapshot-Reason", msg)
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, msg, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Snapshot-Age", fmt.Sprintf("%.1f", age.Seconds()))
	w.Header().Set("X-Snapshot-Interval", fmt.Sprintf("%.0f", maxAge.Seconds()))
	if throttled {
		w.Header().Set("X-Snapshot-Throttled", "1")
	}
	if cachedErr != nil {
		w.Header().Set("X-Snapshot-Stale", cachedErr.Error())
	}
	w.Write(data)
}

// ─── go2rtc ───────────────────────────────────────────────────────────────────

func handleG2Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online":   checkGo2rtcRunning(),
		"port":     go2rtcPort,
		"prozesse": zaehleGo2rtc(),
	})
}

// handleG2KillAll ist der Notausschalter: alle go2rtc-Prozesse beenden — auch
// verwaiste — und danach genau einen frischen starten. Hilft, wenn sich
// Prozesse gehaeuft haben oder ein Programm-Update nicht greift.
func handleG2KillAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	stopGo2rtc() // eigenen Prozess sauber abmelden und beenden
	beendet, _ := killAlleGo2rtc()
	time.Sleep(500 * time.Millisecond)
	// Einen frischen starten, sofern installiert und nicht in der Netzpause.
	if !netzPausiert() && fileExists(go2rtcBinPath()) {
		go2rtcMu.Lock()
		go2rtcWanted = true
		go2rtcMu.Unlock()
		startGo2rtc()
	}
	log.Printf("🧯 Killswitch: %d go2rtc-Prozess(e) beendet, einer neu gestartet", beendet)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"beendet": beendet})
}

// restartGo2rtcAsync restarts go2rtc in the background after a short delay.
// Called automatically when printers are added/removed.
// restartGo2rtcAsync sammelt Neustarts ein, statt jeden einzeln auszufuehren.
//
// Vorher startete jeder Aufruf einen eigenen Ablauf. Wer zehn Drucker
// nacheinander anlegt, loeste damit zehn ueberlappende Stopp-Start-Folgen aus —
// go2rtc kam dabei kaum zum Laufen. Jetzt setzt jeder Aufruf nur die Frist neu;
// ausgefuehrt wird einmal, wenn Ruhe eingekehrt ist.
//
// Die Vorpruefung geschieht bewusst SOFORT und nicht erst in der Goroutine:
// ist gar kein go2rtc installiert oder ist der Netzverkehr angehalten, bleibt
// gar nichts liegen, was spaeter noch loslaufen koennte.
var restartTimer *time.Timer

func restartGo2rtcAsync() {
	if netzPausiert() || !fileExists(go2rtcBinPath()) {
		return
	}
	go2rtcMu.Lock()
	defer go2rtcMu.Unlock()
	if restartTimer != nil {
		restartTimer.Stop()
	}
	restartTimer = time.AfterFunc(800*time.Millisecond, neustartGo2rtc)
}

func handleG2Restart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	go neustartGo2rtc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
}

// go2rtcWanted sagt, ob go2rtc laufen SOLL. Nur so laesst sich ein Absturz von
// einem gewollten Stopp unterscheiden.
var (
	go2rtcWanted bool
	go2rtcGen    int
	go2rtcFails  int
)

func startGo2rtc() {
	go2rtcMu.Lock()

	binaryPath := go2rtcBinPath()
	if !fileExists(binaryPath) {
		go2rtcMu.Unlock()
		log.Printf("ℹ️  go2rtc nicht installiert — Video/Snapshots aus (%s)\n", binaryPath)
		return
	}
	writeGo2rtcYaml()

	yamlPath := filepath.Join(appDir, "go2rtc.yaml")
	cmd := exec.Command(binaryPath, "-config", yamlPath)
	cmd.Env = envWithFFmpeg() // damit go2rtc unser ffmpeg findet
	// Ausgabe mitschreiben. Vorher wurde sie verworfen — endete go2rtc, stand
	// nirgends warum, und jede Fehlersuche war Raten.
	if lf, err := os.OpenFile(filepath.Join(appDir, "go2rtc.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		cmd.Stdout = lf
		cmd.Stderr = lf
	}
	hideProcessWindow(cmd) // Windows: CREATE_NO_WINDOW — no black terminal
	if err := cmd.Start(); err != nil {
		go2rtcMu.Unlock()
		log.Printf("❌ go2rtc Start-Fehler: %v\n", err)
		return
	}
	go2rtcCmd = cmd
	go2rtcWanted = true
	go2rtcGen++
	gen := go2rtcGen
	go2rtcMu.Unlock()

	log.Printf("✅ go2rtc PID %d\n", cmd.Process.Pid)
	go superviseGo2rtc(cmd, gen)
	go verifyStreamsLoaded()
}

// verifyStreamsLoaded vergleicht, was in die Konfiguration geschrieben wurde,
// mit dem, was go2rtc daraus gemacht hat. Weicht es ab, ist die Datei fehlerhaft
// — go2rtc laeuft dann zwar, kennt aber keine oder zu wenige Streams.
func verifyStreamsLoaded() {
	for i := 0; i < 15; i++ {
		time.Sleep(time.Second)
		if checkGo2rtcRunning() {
			break
		}
		if i == 14 {
			log.Printf("⚠️  go2rtc antwortet nach 15 s nicht auf Port %d", go2rtcPort)
			return
		}
	}

	mu.Lock()
	want := len(state.Printers)
	mu.Unlock()

	resp, err := snapClient.Get(fmt.Sprintf("http://127.0.0.1:%d/api/streams", go2rtcPort))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var streams map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&streams); err != nil {
		return
	}

	got := len(streams)
	setStreamCount(got, want)
	switch {
	case want > 0 && got == 0:
		log.Printf("❌ go2rtc kennt KEINEN Stream, obwohl %d Drucker eingetragen sind — go2rtc.yaml fehlerhaft, siehe go2rtc.log", want)
	case got < want:
		log.Printf("⚠️  go2rtc kennt nur %d von %d Streams — siehe go2rtc.log", got, want)
	default:
		log.Printf("✅ go2rtc hat %d Stream(s) geladen", got)
	}
}

var streamCount struct {
	sync.Mutex
	got, want int
	checked   bool
}

func setStreamCount(got, want int) {
	streamCount.Lock()
	streamCount.got, streamCount.want, streamCount.checked = got, want, true
	streamCount.Unlock()
}

// superviseGo2rtc wartet auf das Ende des Prozesses. Endet er, obwohl er laufen
// soll, wird er neu gestartet — vorher blieb er einfach weg und mit ihm alle
// Videos und Snapshots, bis jemand von Hand eingriff.
func superviseGo2rtc(cmd *exec.Cmd, gen int) {
	started := time.Now()
	err := cmd.Wait()

	go2rtcMu.Lock()
	superseded := gen != go2rtcGen
	wanted := go2rtcWanted
	if time.Since(started) > time.Minute {
		go2rtcFails = 0 // lief lange genug, also kein Startproblem
	}
	fails := go2rtcFails
	go2rtcMu.Unlock()

	if superseded || !wanted {
		return // gewollt beendet oder laengst ersetzt
	}
	// Waehrend einer Netzwerkpause bleibt go2rtc unten. Ohne diese Sperre
	// deutet der Aufpasser das gewollte Beenden als Absturz und zieht es
	// Sekunden spaeter wieder hoch — die Pause waere damit wirkungslos.
	if netzPausiert() {
		log.Printf("go2rtc beendet — Netzwerkverkehr ist angehalten, kein Neustart")
		return
	}

	wait := time.Duration(1<<uint(minInt(fails, 5))) * time.Second
	if wait > time.Minute {
		wait = time.Minute
	}
	log.Printf("⚠️  go2rtc unerwartet beendet (%v) — Neustart in %s", err, wait)

	go2rtcMu.Lock()
	go2rtcFails++
	go2rtcMu.Unlock()

	time.Sleep(wait)

	go2rtcMu.Lock()
	stillWanted := go2rtcWanted && gen == go2rtcGen
	go2rtcMu.Unlock()
	if stillWanted {
		startGo2rtc()
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// go2rtcHealthLoop faengt den Fall ab, dass der Prozess zwar noch existiert,
// aber nicht mehr antwortet.
func go2rtcHealthLoop() {
	// Waehrend der Pause bleibt go2rtc bewusst unten.
	for {
		time.Sleep(30 * time.Second)
		if netzPausiert() {
			continue
		}

		go2rtcMu.Lock()
		wanted := go2rtcWanted
		go2rtcMu.Unlock()
		if !wanted || !fileExists(go2rtcBinPath()) {
			continue
		}
		if checkGo2rtcRunning() {
			continue
		}
		// Zweite Chance, bevor neu gestartet wird
		time.Sleep(3 * time.Second)
		if checkGo2rtcRunning() {
			continue
		}
		log.Printf("⚠️  go2rtc antwortet nicht auf Port %d — starte neu", go2rtcPort)
		neustartGo2rtc()
	}
}

// neustartGo2rtc startet go2rtc gruendlich neu. Erst wird der selbst verwaltete
// Prozess beendet, dann werden ALLE noch laufenden (auch verwaiste)
// go2rtc-Prozesse samt Kindern beseitigt — sonst haelt ein haengengebliebener
// weiterhin Port 1984 und der frische kann nicht binden. Genau daran scheiterte
// der Neustart im Programm: stopGo2rtc allein beendet nur den eigenen Prozess.
// neustartMu sorgt dafuer, dass immer nur EIN Neustart laeuft. Ohne das konnten
// sich zwei gleichzeitige Neustarts (z. B. Knopf + automatischer Anlass)
// gegenseitig abwuergen: der killAlleGo2rtc des einen beendet den frisch
// gestarteten des anderen.
var neustartMu sync.Mutex

func neustartGo2rtc() {
	if netzPausiert() {
		return // waehrend der Pause bleibt go2rtc bewusst unten
	}
	if !neustartMu.TryLock() {
		return // es laeuft bereits ein Neustart
	}
	defer neustartMu.Unlock()

	stopGo2rtc()
	if n, _ := killAlleGo2rtc(); n > 0 {
		log.Printf("🔁 Neustart: %d verbliebene(n) go2rtc-Prozess(e) beendet", n)
	}

	// Warten, bis Port 1984 wirklich frei ist — nicht blind eine feste Zeit.
	for i := 0; i < 25 && checkGo2rtcRunning(); i++ {
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond) // Logdatei/Handles freigeben lassen

	go2rtcMu.Lock()
	go2rtcWanted = true
	go2rtcMu.Unlock()
	startGo2rtc()

	// Nachsehen, ob der frische Prozess wirklich hochkommt. Wenn nicht,
	// aufraeumen und ein zweites Mal starten — das faengt haengengebliebene
	// Ports und Fehlstarts ab.
	if !warteAufGo2rtc(8 * time.Second) {
		log.Printf("⚠️  go2rtc nach Neustart nicht erreichbar — zweiter Versuch")
		killAlleGo2rtc()
		time.Sleep(700 * time.Millisecond)
		go2rtcMu.Lock()
		go2rtcWanted = true
		go2rtcMu.Unlock()
		startGo2rtc()
		if warteAufGo2rtc(8 * time.Second) {
			log.Printf("✅ go2rtc nach zweitem Versuch erreichbar")
		} else {
			log.Printf("❌ go2rtc laesst sich nicht starten — siehe go2rtc.log")
		}
	} else {
		log.Printf("✅ go2rtc nach Neustart erreichbar")
	}
}

// warteAufGo2rtc pollt Port 1984, bis go2rtc antwortet oder die Frist ablaeuft.
func warteAufGo2rtc(frist time.Duration) bool {
	ende := time.Now().Add(frist)
	for time.Now().Before(ende) {
		if checkGo2rtcRunning() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return checkGo2rtcRunning()
}

func stopGo2rtc() {
	go2rtcMu.Lock()
	defer go2rtcMu.Unlock()
	go2rtcWanted = false
	go2rtcGen++ // laufende Ueberwachung fuer ungueltig erklaeren
	if go2rtcCmd != nil && go2rtcCmd.Process != nil {
		beendeGo2rtcBaum(go2rtcCmd) // samt ffmpeg-Kindern — sonst haengen die ~10 s
		go2rtcCmd = nil
	}
}

func checkGo2rtcRunning() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", go2rtcPort), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ─── YAML ─────────────────────────────────────────────────────────────────────

// buildYaml erzeugt die Konfiguration. Doppelte Streamnamen sind dabei toedlich:
// go2rtc verwirft bei einem doppelten Schluessel die GESAMTE streams-Sektion und
// kennt anschliessend keinen einzigen Stream — laeuft aber weiter, als waere
// nichts. Deshalb werden Namen hier eindeutig gemacht.
// kameraQuellen liefert die Adressen fuer einen Drucker — in der Reihenfolge,
// in der go2rtc sie ausprobieren soll.
//
// Der Unterschied zwischen den Schemata ist entscheidend: rtspx:// ist eine
// Sonderform, die bei der X1-Reihe zuverlaessig funktioniert. Die H2- und
// P2-Reihe antwortet darauf mit einer Umleitung auf rtsps:// und liefert kein
// Bild — genau das Verhalten, das bei allen H-Geraeten zu sehen war. Deshalb
// steht dort rtsps:// vorn.
//
// Angegeben werden beide, damit ein Geraet, das sich anders verhaelt als sein
// Modellname vermuten laesst, trotzdem ein Bild liefert: go2rtc probiert die
// Eintraege der Reihe nach durch.
func kameraQuellen(p Printer, gemessen map[string]string) []string {
	auth := fmt.Sprintf("bblp:%s@%s:322", p.Code, p.IP)

	// Wurde am Geraet gemessen, welche Form funktioniert, hat die Vorrang vor
	// jeder Ableitung aus dem Modellnamen.
	//
	// Die Karte wird hereingereicht und nicht hier gelesen: buildYaml laeuft
	// bereits unter der Sperre, ein zweiter Zugriff darauf haengt das Programm
	// auf. Genau das ist beim ersten Anlauf passiert.
	if schema, pfad, ok := schemaAus(gemessen, p.IP); ok {
		gemessen := schema + "://" + auth + pfad
		rest := []string{}
		for _, s2 := range []string{"rtsps", "rtspx"} {
			for _, p2 := range []string{"/streaming/live/1", "/streaming/live/0"} {
				a := s2 + "://" + auth + p2
				if a != gemessen {
					rest = append(rest, a)
				}
			}
		}
		return append([]string{gemessen}, rest...)
	}

	pfad := auth + "/streaming/live/1"
	sicher := "rtsps://" + pfad
	sonder := "rtspx://" + pfad

	m := strings.ToUpper(strings.TrimSpace(p.Model))
	switch {
	case strings.HasPrefix(m, "H2"), strings.HasPrefix(m, "P2"), strings.HasPrefix(m, "H1"):
		return []string{sicher, sonder}
	default:
		return []string{sonder, sicher}
	}
}

func buildYaml(printers []Printer, gemessen map[string]string) string {
	var sb strings.Builder
	sb.WriteString("# go2rtc.yaml\n\napi:\n  origin: '*'\n\nstreams:\n")
	for _, p := range namedStreams(printers) {
		sb.WriteString("  " + p.stream + ":\n")
		for _, q := range kameraQuellen(p.printer, gemessen) {
			sb.WriteString("    - " + q + "\n")
		}
	}
	return sb.String()
}

type streamEntry struct {
	stream  string
	printer Printer
}

func namedStreams(printers []Printer) []streamEntry {
	used := map[string]bool{}
	out := make([]streamEntry, 0, len(printers))
	for _, p := range printers {
		name := streamName(p)
		if used[name] {
			base := name
			for i := 2; ; i++ {
				name = fmt.Sprintf("%s-%d", base, i)
				if !used[name] {
					break
				}
			}
			log.Printf("⚠️  Streamname %q kommt mehrfach vor (%s) — verwende %q", base, p.IP, name)
		}
		used[name] = true
		out = append(out, streamEntry{stream: name, printer: p})
	}
	return out
}

// streamNameFor liefert den tatsaechlich verwendeten Streamnamen eines Druckers,
// also inklusive der Entdopplung.
func streamNameFor(ip string) string {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()
	for _, e := range namedStreams(printers) {
		if e.printer.IP == ip {
			return e.stream
		}
	}
	return ""
}

func streamName(p Printer) string {
	var sb strings.Builder
	prev := false
	for _, c := range strings.ToLower(p.Name) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			sb.WriteRune(c)
			prev = false
		} else if !prev {
			sb.WriteRune('-')
			prev = true
		}
	}
	s := strings.Trim(sb.String(), "-")
	if s == "" {
		s = "printer-" + strings.ReplaceAll(p.IP, ".", "-")
	}
	return s
}

func writeGo2rtcYaml() {
	mu.Lock()
	gemessen := map[string]string{}
	for k, v := range state.KameraSchema {
		gemessen[k] = v
	}
	cams := make([]CameraCfg, len(state.Cameras))
	copy(cams, state.Cameras)
	mu.Unlock()
	pfad := filepath.Join(appDir, "go2rtc.yaml")
	// WICHTIG: go2rtc schreibt beim Mi-Home-Login eigene Abschnitte in diese
	// Datei (Konto/Token für den Xiaomi-Cloud-Schlüssel). Beim Neuschreiben der
	// von uns verwalteten Abschnitte (api/streams) dürfen diese NICHT verloren
	// gehen — sonst kann go2rtc den Kameraschlüssel nicht mehr holen. Deshalb
	// werden fremde Top-Level-Abschnitte aus der bestehenden Datei übernommen.
	vorher, _ := os.ReadFile(pfad)
	fremd := erhalteFremdeSektionen(string(vorher))
	yaml := buildYaml(state.Printers, gemessen) + cameraStreamsYaml(cams) + fremd
	os.WriteFile(pfad, []byte(yaml), 0644)
}

// ─── PERSISTENCE ──────────────────────────────────────────────────────────────

func loadState() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	json.Unmarshal(data, &state)
}

func saveState() {
	mu.Lock()
	data, _ := json.MarshalIndent(state, "", "  ")
	mu.Unlock()
	os.WriteFile(dataFile, data, 0644)
}

// ─── HELPERS ──────────────────────────────────────────────────────────────────

func waitForPort(port int) {
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// ─── STATUS API ───────────────────────────────────────────────────────────────

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	statuses := mqttMgr.AllStatuses()

	// Die selbst gezaehlte Laufzeit gehoert nicht in den MQTT-Zustand — sie
	// kommt nicht vom Geraet. Sie wird hier daneben gestellt.
	mu.Lock()
	laufzeit := map[string]int64{}
	for ip, sek := range state.Laufzeit {
		laufzeit[ip] = sek
	}
	mu.Unlock()

	type erweitert struct {
		PrinterStatus
		LaufzeitSek  int64           `json:"laufzeit_sek,omitempty"`
		LaufzeitText string          `json:"laufzeit_text,omitempty"`
		FwUpdates    []FwModulUpdate `json:"fw_updates,omitempty"`
		Reparatur    *RepairFlag     `json:"reparatur,omitempty"`
	}
	// Offene Firmware-Updates je Drucker — dabei wird gegen die laufende
	// Firmware abgeglichen und bereits Installiertes faellt weg.
	fw := fwAnzeige()
	// Reparatur-Markierungen aus dem Zwischenspeicher.
	mu.Lock()
	rep := map[string]RepairFlag{}
	for ip, f := range state.Reparatur {
		rep[ip] = f
	}
	mu.Unlock()
	out := map[string]erweitert{}
	for ip, st := range statuses {
		e := erweitert{PrinterStatus: st, LaufzeitSek: laufzeit[ip], LaufzeitText: laufzeitText(laufzeit[ip]), FwUpdates: fw[ip]}
		if f, ok := rep[ip]; ok && f.InRepair {
			cp := f
			e.Reparatur = &cp
		}
		out[ip] = e
	}
	json.NewEncoder(w).Encode(out)
}

// ─── ERRORS API ───────────────────────────────────────────────────────────────

type PrinterError struct {
	IP       string   `json:"ip"`
	Name     string   `json:"name"`
	Messages []string `json:"messages"`
}

func handleErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	statuses := mqttMgr.AllStatuses()
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	var errors []PrinterError
	for _, p := range printers {
		s, ok := statuses[p.IP]
		if !ok || !s.Online {
			continue
		}
		var msgs []string
		if s.PrintError != 0 {
			msgs = append(msgs, PrintErrorText(s.PrintError))
		}
		for _, e := range s.HmsErrors {
			msgs = append(msgs, HMSErrorText(e))
		}
		if len(msgs) > 0 {
			errors = append(errors, PrinterError{
				IP:       p.IP,
				Name:     p.Name,
				Messages: msgs,
			})
		}
	}
	json.NewEncoder(w).Encode(errors)
}

// ─── DEBUG API ────────────────────────────────────────────────────────────────

func handleDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	type DebugEntry struct {
		Name        string         `json:"name"`
		IP          string         `json:"ip"`
		Serial      string         `json:"serial"`
		Connected   bool           `json:"connected"`
		Status      *PrinterStatus `json:"status"`
		LastPayload string         `json:"last_payload,omitempty"`
	}

	var entries []DebugEntry
	mqttMgr.mu.RLock()
	for _, p := range printers {
		connected := false
		if c, ok := mqttMgr.clients[p.IP]; ok {
			connected = c.IsConnected()
		}
		var st *PrinterStatus
		if s, ok := mqttMgr.statuses[p.IP]; ok {
			cp := *s
			st = &cp
		}
		lastPayload := mqttMgr.lastPayload[p.IP]
		entries = append(entries, DebugEntry{
			Name: p.Name, IP: p.IP, Serial: p.Serial,
			Connected: connected, Status: st,
			LastPayload: lastPayload,
		})
	}
	mqttMgr.mu.RUnlock()

	json.NewEncoder(w).Encode(entries)
}

func handleCameraResolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()
	sent := mqttMgr.SetCameraResolution(printers, body.Resolution)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"sent": sent})
}

// ─── DRUCKSTEUERUNG ───────────────────────────────────────────────────────────

// handlePrintCommand schickt pause/resume/stop an einen oder mehrere Drucker.
// Ohne "ips" passiert nichts — ein versehentlicher Aufruf soll nicht die ganze
// Farm treffen.
func handlePrintCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IPs     []string `json:"ips"`
		Command string   `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := printCommands[body.Command]; !ok {
		http.Error(w, "unbekanntes Kommando: "+body.Command, http.StatusBadRequest)
		return
	}
	if len(body.IPs) == 0 {
		http.Error(w, "keine Drucker angegeben", http.StatusBadRequest)
		return
	}

	wanted := map[string]bool{}
	for _, ip := range body.IPs {
		wanted[ip] = true
	}
	mu.Lock()
	var targets []Printer
	for _, p := range state.Printers {
		if wanted[p.IP] {
			targets = append(targets, p)
		}
	}
	mu.Unlock()

	sent := 0
	failed := []map[string]string{}
	for _, p := range targets {
		if err := mqttMgr.SendPrintCommand(p, body.Command); err != nil {
			failed = append(failed, map[string]string{"ip": p.IP, "name": p.Name, "error": err.Error()})
			continue
		}
		sent++
	}

	writeJSON(w, map[string]any{"sent": sent, "failed": failed, "command": body.Command})
}

// ─── EINSTELLUNGEN ────────────────────────────────────────────────────────────

// handleSettings haelt Theme, Sprache und Spaltenzahl fest. Bisher lagen die im
// Browserspeicher — und waren damit weg, sobald das Edge-Profil neu angelegt wurde.
func handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Theme           *string         `json:"theme"`
			Lang            *string         `json:"lang"`
			Cols            *int            `json:"cols"`
			FilterPillen    map[string]bool `json:"filter_pillen"`
			ErrorBlink      *bool           `json:"error_blink"`
			StartFilter     *string         `json:"start_filter"`
			SpracheGewaehlt *bool           `json:"sprache_gewaehlt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		if body.Theme != nil && (*body.Theme == "dark" || *body.Theme == "light") {
			state.Theme = *body.Theme
		}
		if body.Lang != nil && (*body.Lang == "de" || *body.Lang == "en") {
			state.Lang = *body.Lang
		}
		if body.SpracheGewaehlt != nil {
			state.SpracheGewaehlt = *body.SpracheGewaehlt
		}
		if body.Cols != nil && *body.Cols >= 1 && *body.Cols <= 12 {
			state.Cols = *body.Cols
		}
		if body.FilterPillen != nil {
			state.FilterPillen = body.FilterPillen
		}
		if body.StartFilter != nil {
			sf := *body.StartFilter
			if sf == "all" || sf == "online" || sf == "offline" {
				state.StartFilter = sf
			}
		}
		turnedOff := false
		if body.ErrorBlink != nil {
			was := state.ErrorBlink == nil || *state.ErrorBlink
			v := *body.ErrorBlink
			state.ErrorBlink = &v
			turnedOff = was && !v
		}
		mu.Unlock()
		saveState()
		if turnedOff {
			// Abschalten heisst auch: sofort wieder normales Licht herstellen
			go ResetChamberLights()
		}
	}

	mu.Lock()
	theme, lang, cols := state.Theme, state.Lang, state.Cols
	filterPillen := map[string]bool{}
	for k, v := range state.FilterPillen {
		filterPillen[k] = v
	}
	blink := state.ErrorBlink == nil || *state.ErrorBlink
	blinkModelList := append([]string(nil), state.BlinkModels...)
	spracheGewaehlt := state.SpracheGewaehlt
	startFilter := state.StartFilter
	blinkAus := map[string]bool{}
	for k, v := range state.BlinkAus { if v { blinkAus[k] = true } }
	kameraAus := map[string]bool{}
	for k, v := range state.KameraAus { if v { kameraAus[k] = true } }
	mu.Unlock()
	if startFilter == "" { startFilter = "all" }
	if len(blinkModelList) == 0 {
		blinkModelList = defaultBlinkModels
	}
	if theme == "" {
		theme = "light"
	}
	if lang == "" {
		lang = "de"
	}
	if cols == 0 {
		cols = 6
	}
	writeJSON(w, map[string]any{"theme": theme, "lang": lang, "cols": cols,
		"error_blink": blink, "blink_models": blinkModelList, "filter_pillen": filterPillen,
		"start_filter": startFilter, "blink_aus": blinkAus, "kamera_aus": kameraAus,
		"sprache_gewaehlt": spracheGewaehlt})
}

// handleOpenFolder oeffnet einen Ordner im Explorer. Ein Link auf file:// wuerde
// aus einer http-Seite heraus vom Browser blockiert, deshalb der Umweg.
func handleOpenFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	target := body.Path
	if target == "" {
		target = appDir
	}
	// Nur eigene Verzeichnisse oeffnen — kein beliebiger Pfad von aussen
	clean := filepath.Clean(target)
	if !strings.HasPrefix(clean, filepath.Clean(appDir)) {
		http.Error(w, "Pfad liegt ausserhalb des Datenverzeichnisses", http.StatusBadRequest)
		return
	}
	if st, err := os.Stat(clean); err == nil && !st.IsDir() {
		clean = filepath.Dir(clean)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", clean)
	case "darwin":
		cmd = exec.Command("open", clean)
	default:
		cmd = exec.Command("xdg-open", clean)
	}
	// explorer.exe liefert auch bei Erfolg einen Rueckgabewert ungleich 0
	_ = cmd.Start()
	writeJSON(w, map[string]any{"opened": clean})
}

// ─── DIAGNOSE ─────────────────────────────────────────────────────────────────

// handleDiagnostics fasst alles zusammen, was zur Fehlersuche bei go2rtc noetig
// ist — damit nicht mehr geraten werden muss, was gerade los ist.
func handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	go2rtcMu.Lock()
	wanted := go2rtcWanted
	pid := 0
	if go2rtcCmd != nil && go2rtcCmd.Process != nil {
		pid = go2rtcCmd.Process.Pid
	}
	fails := go2rtcFails
	go2rtcMu.Unlock()

	streamCount.Lock()
	got, want, checked := streamCount.got, streamCount.want, streamCount.checked
	streamCount.Unlock()

	// Was meldet go2rtc gerade?
	live := map[string]any{}
	liveErr := ""
	if resp, err := snapClient.Get(fmt.Sprintf("http://127.0.0.1:%d/api/streams", go2rtcPort)); err == nil {
		json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&live)
		resp.Body.Close()
	} else {
		liveErr = err.Error()
	}

	// Doppelte Streamnamen aufspueren
	seen := map[string]string{}
	var collisions []string
	for _, e := range namedStreams(printers) {
		base := streamName(e.printer)
		if other, ok := seen[base]; ok {
			collisions = append(collisions, fmt.Sprintf("%s + %s → %q", other, e.printer.Name, base))
		}
		seen[base] = e.printer.Name
	}

	writeJSON(w, map[string]any{
		"version":          appVersion,
		"data_dir":         appDir,
		"go2rtc_binary":    go2rtcBinPath(),
		"go2rtc_present":   fileExists(go2rtcBinPath()),
		"go2rtc_version":   toolVersion(go2rtcBinPath(), "-version"),
		"ffmpeg_binary":    ffmpegBinPath(),
		"ffmpeg_present":   fileExists(ffmpegBinPath()),
		"ffmpeg_version":   toolVersion(ffmpegBinPath(), "-version"),
		"should_run":       wanted,
		"pid":              pid,
		"restarts":         fails,
		"port":             go2rtcPort,
		"port_reachable":   checkGo2rtcRunning(),
		"printers":         len(printers),
		"streams_loaded":   len(live),
		"streams_error":    liveErr,
		"checked":          checked,
		"streams_at_start": map[string]int{"geladen": got, "erwartet": want},
		"snapshot_interval_2s": fmt.Sprintf("%.0f s (gewünscht 2 s, %d Drucker, höchstens %.0f Anfragen/s)",
			effectiveSnapInterval(2*time.Second).Seconds(), len(printers), snapMaxRatePerSec),
		"name_collisions": collisions,
		"go2rtc_log":      tailFile(filepath.Join(appDir, "go2rtc.log"), 40),
		"app_log":         tailFile(filepath.Join(appDir, "druckerfarm.log"), 25),
	})
}

// tailFile liefert die letzten n Zeilen einer Datei.
func tailFile(path string, n int) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return []string{"(" + err.Error() + ")"}
	}
	if len(b) > 256<<10 {
		b = b[len(b)-(256<<10):]
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
