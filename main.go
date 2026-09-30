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

	// UploadLog records what was uploaded where and when — persistently, so it
	// survives a restart.
	UploadLog []UploadEntry `json:"-"` // own file: upload_log.json

	// Runtime in seconds per printer, counted by this program.
	Laufzeit map[string]int64 `json:"-"` // own file: runtime.json

	// KameraSchema haelt fest, welche Adressform an einem Geraet tatsaechlich
	// returned an image — measured, not guessed from the model name.
	KameraSchema map[string]string `json:"kamera_schema,omitempty"`

	// FwUpdates: per printer the open firmware updates the printer
	// reported itself. Kept stored until it has installed them.
	FwUpdates map[string][]FwModuleUpdate `json:"-"` // own file: runtime.json

	// Reparatur: per printer (IP) the "in repair" marker. Source is the
	// maintenance.json on the SD; cached locally here.
	Reparatur map[string]RepairFlag `json:"-"` // own file: runtime.json

	// BlinkAus: per printer IP true when error blinking is disabled for THIS
	// printer (in addition to the global switch).
	BlinkAus map[string]bool `json:"blink_aus,omitempty"`

	// KameraAus: per printer IP true when the camera is permanently off
	// (tile shows "Private", no stream).
	KameraAus map[string]bool `json:"kamera_aus,omitempty"`

	// StartFilter: overview filter at start (all|online|offline). Default all.
	StartFilter string `json:"start_filter,omitempty"`

	// Broadcast page (/broadcast): decoupled read-only view for streaming/sharing.
	BroadcastView  string `json:"broadcast_view,omitempty"`  // all|online|offline
	BroadcastClick bool   `json:"broadcast_click,omitempty"` // Klicken erlaubt?
	BroadcastCols  int    `json:"broadcast_cols,omitempty"`  // Spaltenzahl im Raster

	// Cloudflare-Tunnel: public = frei erreichbar, secret = Token noetig.
	TunnelMode  string `json:"tunnel_mode,omitempty"`  // public|secret
	TunnelToken string `json:"tunnel_token,omitempty"` // Zugangstoken bei secret

	// YouTube-Live: Stream-Key + einstellbare Qualitaet.
	YTKey        string `json:"yt_key,omitempty"`
	YTBitrateK   int    `json:"yt_bitrate_k,omitempty"`  // Videobitrate in kbit/s
	YTResolution string `json:"yt_resolution,omitempty"` // z. B. 1920x1080
	YTFps        int    `json:"yt_fps,omitempty"`

	// AI assistant: API credentials (only local in config.json, gitignored,
	// never in the log, never returned to the UI).
	//
	// Legacy single-provider fields (kept for migration; new setups use
	// AIConnectors below). On load these are folded into one connector.
	AIProvider   string `json:"ai_provider,omitempty"` // openai | anthropic
	AIModel      string `json:"ai_model,omitempty"`
	OpenAIKey    string `json:"openai_key,omitempty"`
	AnthropicKey string `json:"anthropic_key,omitempty"`
	OllamaURL    string `json:"ollama_url,omitempty"` // lokales Modell (Ollama), Standard http://localhost:11434

	// AIConnectors: several named AI connections, each created like adding a
	// printer (first the connector, then model/key/endpoint). AIActiveID picks
	// the one currently used for explain/vision.
	AIConnectors []AIConnector `json:"ai_connectors,omitempty"`
	AIActiveID   string        `json:"ai_active_id,omitempty"`

	// Prompts: the instructions to the AI — editable in the tool. {lang} and
	// {error} are substituted. Empty = default.
	AIPromptExplain string `json:"ai_prompt_explain,omitempty"`
	AIPromptVision  string `json:"ai_prompt_vision,omitempty"`

	// PendingFwUpdate: per printer IP true when the firmware update should start
	// automatically once the running print is finished.
	PendingFwUpdate map[string]bool `json:"pending_fw_update,omitempty"`

	// JobPlan: the job planner — a list of design figures, each with a planned
	// quantity per printer MODEL (matrix figure × model). Shown on the Jobs page.
	JobPlan []PlanFigure `json:"job_plan,omitempty"`

	// Users / APITokens: login accounts (PBKDF2 password hash) and API tokens
	// (stored as sha256 hash only). Local in config.json, never logged.
	Users     []User     `json:"users,omitempty"`
	APITokens []APIToken `json:"api_tokens,omitempty"`

	// BoardTable: lookup table for the board planner. Per combination of
	// printer model and print file, figures per board and hours per
	// board — so these values no longer need to be typed into the planner by hand.
	BoardTable []BoardEntry `json:"board_table,omitempty"`

	// SpracheGewaehlt remembers whether the user chose their language
	// at first start. If false, the UI asks once.
	SpracheGewaehlt bool `json:"sprache_gewaehlt,omitempty"`

	// TimelapseAus: per printer IP true when the timelapse should stay
	// permanently OFF. The printer occasionally turns it back on by itself;
	// if this switch is set, the tool regularly re-sends "off".
	TimelapseAus map[string]bool `json:"timelapse_aus,omitempty"`
}

var (
	// mu protects state (AppState). RWMutex instead of Mutex: pure read paths
	// (status/health/error queries) take RLock and may thereby
	// run in parallel; only writes take the full lock. This
	// relieves the many concurrent status queries at large printer counts.
	mu            sync.RWMutex
	state         AppState
	dataFile      string
	uploadLogFile string
	runtimeFile   string
	appDir        string
	go2rtcCmd     *exec.Cmd
	go2rtcMu      sync.Mutex
	go2rtcPort    = 1984
	appPort       = 8765
)

func main() {
	hideConsoleOnWindows()

	exe, _ := os.Executable()
	exePath = exe
	exeDir := filepath.Dir(exe)

	// All data lives in %APPDATA%\Druckerfarm, no longer next to the exe.
	// This makes the exe replaceable (prerequisite for the update) and
	// credentials do not end up in a synced project folder.
	appDir = dataHome()
	if appDir == "" {
		appDir = exeDir
	}
	os.MkdirAll(appDir, 0o755)
	dataFile = filepath.Join(appDir, "config.json")
	uploadLogFile = filepath.Join(appDir, "upload_log.json")
	runtimeFile = filepath.Join(appDir, "runtime.json")
	setupLogFile()
	migrateFromExeDir(exeDir)
	cleanupOldExe()

	loadState()
	migrateAIConnectors() // fold legacy single-provider AI config into a connector
	initSync(appDir)
	initPraesenz(appDir) // local access log; parallel PC via SD file
	go repairLoop()      // read/reconcile repair markers from the SD

	// First clean up orphaned go2rtc processes. A crashed or
	// replaced run can leave some behind; without this they piled
	// up (seen 10x already) and a program update did not take while
	// an old process still held files and the port.
	if n, _ := killAllGo2rtc(); n > 0 {
		log.Printf("🧹 %d verwaiste(n) go2rtc-Prozess(e) beim Start beendet", n)
		time.Sleep(400 * time.Millisecond)
	}
	// start go2rtc if installed (otherwise the app runs without video)
	go startGo2rtc()
	// Connect MQTT to all printers
	go connectAllMQTT()
	go runtimeLoop()
	go startDiscoveryListener()
	// Request full status regularly (recovers a lost first response)
	go pushAllLoop()
	// Blink the chamber light on errors (models without a signal light)
	go errorLightLoop()
	go reconnectLoop()        // automatically reconnect offline printers
	go fwAutoLoop()           // trigger scheduled firmware updates after a print
	go timelapseEnforceLoop() // Zeitraffer bei markierten Druckern dauerhaft aus halten
	// Keep an eye on go2rtc: a crash or hang triggers a restart
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
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/media/list", handleMediaList)
	mux.HandleFunc("/api/media/download", handleMediaDownload)
	mux.HandleFunc("/api/media/delete", handleMediaDelete)
	mux.HandleFunc("/api/media/clear", handleMediaClear)
	mux.HandleFunc("/api/media/clear-all", handleMediaClearAll)
	mux.HandleFunc("/api/media/timelapse", handleMediaTimelapse)
	mux.HandleFunc("/api/media/timelapse-all", handleMediaTimelapseAll)
	mux.HandleFunc("/api/ftpcheck", handleFtpCheck)
	mux.HandleFunc("/api/errors", handleErrors)
	mux.HandleFunc("/api/debug", handleDebug)
	mux.HandleFunc("/api/sync/status", handleSyncStatus)
	mux.HandleFunc("/api/sync/config", handleSyncConfig)
	mux.HandleFunc("/api/sync/diff", handleSyncDiff)
	mux.HandleFunc("/api/sync/start", handleSyncControl)
	mux.HandleFunc("/api/sync/pause", handleSyncControl)
	mux.HandleFunc("/api/sync/stop", handleSyncControl)
	mux.HandleFunc("/api/sync/browse", handleSyncBrowse)
	mux.HandleFunc("/api/sync/download", handleSyncDownload)
	mux.HandleFunc("/api/sync/sdlist", handleSyncSDList)
	mux.HandleFunc("/api/sync/search/start", handleSearchStart)
	mux.HandleFunc("/api/sync/delete/start", handleDeleteStart)
	mux.HandleFunc("/api/jobs", handleJobsList)
	mux.HandleFunc("/api/plan", handlePlan)
	mux.HandleFunc("/api/preview", handlePreview)
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
	mux.HandleFunc("/broadcast", serveBroadcast)
	mux.HandleFunc("/api/broadcast/config", handleBroadcastConfig)
	mux.HandleFunc("/api/tunnel/status", handleTunnelStatus)
	mux.HandleFunc("/api/tunnel/start", handleTunnelStart)
	mux.HandleFunc("/api/tunnel/stop", handleTunnelStop)
	mux.HandleFunc("/api/tunnel/install", handleTunnelInstall)
	mux.HandleFunc("/api/tunnel/config", handleTunnelConfig)
	mux.HandleFunc("/api/youtube/status", handleYouTubeStatus)
	mux.HandleFunc("/api/youtube/start", handleYouTubeStart)
	mux.HandleFunc("/api/youtube/stop", handleYouTubeStop)
	mux.HandleFunc("/api/youtube/config", handleYouTubeConfig)
	mux.HandleFunc("/api/ai/config", handleAIConfig)
	mux.HandleFunc("/api/ai/explain", handleAIExplain)
	mux.HandleFunc("/api/ai/vision", handleAIVision)
	mux.HandleFunc("/api/ai/connectors", handleAIConnectors)
	mux.HandleFunc("/api/ai/connector", handleAIConnector)
	mux.HandleFunc("/api/ai/active", handleAIActive)
	// Auth: login, accounts, roles, API tokens.
	mux.HandleFunc("/api/needsetup", handleNeedSetup)
	mux.HandleFunc("/api/setup-admin", handleSetupAdmin)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/user", handleUser)
	mux.HandleFunc("/api/tokens", handleTokens)
	mux.HandleFunc("/api/token", handleToken)

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
	mux.HandleFunc("/api/network/pause", handleNetworkPause)
	mux.HandleFunc("/api/fwupdate/scan", handleFwUpdateScan)
	mux.HandleFunc("/api/fwupdate/start", handleFwUpdateStart)
	mux.HandleFunc("/api/fwupdate/schedule", handleFwSchedule)
	mux.HandleFunc("/api/xcam", handleXcam)
	mux.HandleFunc("/api/boardtable", handleBoardTable)
	mux.HandleFunc("/api/alive", handleAlive)
	mux.HandleFunc("/api/praesenz", handlePraesenz)
	mux.HandleFunc("/api/repair", handleRepair)
	mux.HandleFunc("/loading", serveLoading)
	mux.HandleFunc("/ready", handleReady)

	// Open window on loading page first — avoids white/transparent flash
	url := fmt.Sprintf("http://127.0.0.1:%d/loading", appPort)

	// Bind synchronously. Previously this ran in a goroutine and the error was
	// dropped — with a busy port the second instance started halfway and took
	// the first one down with it on exit.
	addr := fmt.Sprintf("127.0.0.1:%d", appPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// An instance is already running: focus its window and exit here.
		log.Printf("ℹ️  Port %d belegt — vermutlich laeuft die App schon. Oeffne ein Fenster darauf.\n", appPort)
		launchAppWindow(url)
		return
	}
	go http.Serve(ln, corsMiddleware(authMiddleware(mux)))
	waitForPort(appPort)

	// Open app window (Edge/Chrome in --app mode, no address bar)
	shutdown := make(chan struct{})
	var shutdownOnce sync.Once
	triggerShutdown := func() { shutdownOnce.Do(func() { close(shutdown) }) }

	if _, werr := launchAppWindow(url); werr != nil {
		log.Printf("⚠️  %v\n", werr)
	}

	// We detect shutdown via the page's open SSE connection: if it
	// drops (window closed) and no new one arrives within a few seconds (reload),
	// it shuts down. This is reliable — unlike the browser starter process.
	go uiPresenceWatcher(triggerShutdown)
	// Fallback if the page cannot open SSE at all (old browser etc.):
	// also shut down after long UI silence. Only applies if no
	// connection was ever made or no UI traffic runs anymore.
	go func() { waitUntilUIIdle(); triggerShutdown() }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
	case <-shutdown:
	}
	shutdownClean()
	os.Exit(0)
}

// shutdownClean tidies up on program close: the signal lights of the
// affected printers (X1E/X2D) go from "blinking" back to normal (on),
// and go2rtc is fully stopped — the own process as well as any
// orphans, including their child processes (ffmpeg). Without this a blinking
// light stayed on in the printer room and go2rtc sometimes kept running.
func shutdownClean() {
	// From now on nothing may restart go2rtc (neither watchdog nor
	// health loop nor a config restart), otherwise it would survive the shutdown.
	shuttingDown.Store(true)
	go2rtcMu.Lock()
	go2rtcWanted = false
	go2rtcMu.Unlock()

	// Force any debounced runtime.json write out before we exit.
	flushRuntimeSave()

	// FIRST stop go2rtc entirely — this should happen immediately. Previously
	// resetting the signal lights came first here (up to 6 s wait),
	// so go2rtc was only stopped afterwards and ran correspondingly long.
	stopTunnel()  // Cloudflare-Tunnel + Public-Server beenden
	stopYouTube() // laufenden YouTube-Push beenden
	stopGo2rtc()
	if n, _ := killAllGo2rtc(); n > 0 {
		log.Printf("🧹 %d go2rtc-Prozess(e) beim Beenden geschlossen", n)
	}

	stopPresence() // Stop-Zeile ins Zugriffs-Log, eigene Praesenz entfernen

	// THEN reset the lights, with a time limit: if a printer does not
	// answer, it continues after a few seconds anyway.
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Token")
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
// The logo is embedded in the program file — no adjacent file that
// can get lost.
//
//go:embed assets/logo.svg
var logoSVG []byte

func handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(logoSVG)
}

// The splash image is also embedded in the program file.
//
//go:embed assets/splash.jpg
var splashJPG []byte

// The third-party licenses are embedded in the program file so they can be
// linked and shown in the app — regardless of whether the
// markdown file is present next to it.
//
//go:embed THIRD_PARTY_LICENSES.md
var thirdPartyLicenses string

// The translations are provided as JSON per language and embedded firmly into
// Programmdatei eingebettet. Wer eigene Uebersetzungen will, bearbeitet
// lang/de.json or lang/en.json and rebuild the exe. When serving the
// page they are inserted at __LANGS_JSON__.
//
//go:embed lang/de.json
var langDE string

//go:embed lang/en.json
var langEN string

//go:embed lang/zh.json
var langZH string

//go:embed lang/es.json
var langES string

//go:embed lang/de-zh.json
var langDEZH string

//go:embed lang/de-es.json
var langDEES string

// de-en.json is the full word-for-word German->English mapping
// with which the UI is translated automatically in English —
// including dynamically generated text. Editable before building.
//
//go:embed lang/de-en.json
var langDEEN string

// handleOpenLicenses writes the embedded licenses as a file into the
// data directory and shows it in the system file explorer. This way no
// browser tab opens; the file sits tangibly in the folder.
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
		// /select shows the file selected in the explorer window.
		cmd = exec.Command("explorer.exe", "/select,"+pfad)
	case "darwin":
		cmd = exec.Command("open", "-R", pfad)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(pfad))
	}
	_ = cmd.Start() // explorer returns != 0 even on success
	writeJSON(w, map[string]any{"pfad": pfad})
}

func handleSplash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(splashJPG)
}

var readySince = time.Now()

// exePath is the path of the running program file — the update replaces it.
var exePath string

// dataHome returns the directory for configuration, tools and log.
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

// migrateFromExeDir moves an existing setup once from the
// program folder, so nobody re-types 42 printers after the move.
func migrateFromExeDir(exeDir string) {
	if exeDir == appDir {
		return
	}
	if _, err := os.Stat(dataFile); err == nil {
		return // already set up
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

// lastRequest records when the UI last fetched something. The
// dashboard polls every 5s, so "silent for X seconds" is the reliable
// signal that the window is really closed.
var lastRequest atomic.Int64

// As variables so tests can shrink them.
var (
	// Generous so the app does NOT exit prematurely: the msedge starter
	// leaves immediately if an Edge is already running — then the
	// idle detection takes over. On first start Windows (Defender/SmartScreen)
	// can slow the browser by dozens of seconds; with a tight deadline the
	// server shut down before the window had even loaded (ERR_CONNECTION_
	// REFUSED on all /api calls). Better a few seconds of zombie on
	// close than a dead server on start.
	uiIdleTimeout  = 60 * time.Second // this long without a UI request = window closed
	uiNeverGrace   = 10 * time.Minute // wait this long for the very first request
	uiIdlePollTick = 2 * time.Second
)

func noteRequest() { lastRequest.Store(time.Now().UnixNano()) }

// setupLogFile redirects log output to a file. With -H windowsgui there is
// no console — without this every message is invisible.
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

// waitUntilUIIdle returns as soon as the UI has been silent longer than
// uiIdleTimeout — or after uiNeverGrace if nobody ever connected.
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
	w.Write([]byte(strings.ReplaceAll(pageWithVersion(), "__BROADCAST_OPTS__", "null")))
}

// pageWithVersion inserts the built version number into the page — computed
// once, then from the cache. So the version already shows on the
// splash (gorilla loading page) without the UI having to fetch it first.
var (
	seiteEinmal sync.Once
	seiteCache  string
)

func pageWithVersion() string {
	seiteEinmal.Do(func() {
		langs := `{"de":` + strings.TrimSpace(langDE) + `,"en":` + strings.TrimSpace(langEN) +
			`,"zh":` + strings.TrimSpace(langZH) + `,"es":` + strings.TrimSpace(langES) + `}`
		overlays := `{"en":` + strings.TrimSpace(langDEEN) + `,"zh":` + strings.TrimSpace(langDEZH) +
			`,"es":` + strings.TrimSpace(langDEES) + `}`
		seite := strings.ReplaceAll(dashboardHTML, "__LANGS_JSON__", langs)
		seite = strings.ReplaceAll(seite, "__OVERLAYS_JSON__", overlays)
		seite = strings.ReplaceAll(seite, "__DEEN_JSON__", strings.TrimSpace(langDEEN))
		seite = strings.ReplaceAll(seite, "__APP_VERSION__", appVersion)
		seite = strings.ReplaceAll(seite, "__BUILD_ID__", buildID)
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
		var addFlags struct {
			Printer
			KameraAus bool `json:"kamera_aus"`
			BlinkAus  bool `json:"blink_aus"`
			InRepair  bool `json:"in_repair"`
		}
		if err := json.NewDecoder(r.Body).Decode(&addFlags); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		p := addFlags.Printer
		// The access code is no longer required: a printer may be
		// added without one and the code added later. Without it there is
		// no image and no file access, but the entry already exists
		// once — this is the usual flow when building up a farm.
		if p.Name == "" || p.IP == "" {
			http.Error(w, "name und ip sind erforderlich", 400)
			return
		}
		if p.Added == 0 {
			p.Added = time.Now().UnixMilli()
		}
		mu.Lock()
		// "Already present?" is decided by the serial number, NOT the IP. An
		// offline printer often still carries a stale IP; if a new
		// device is assigned the same IP, adding must not fail because of it —
		// the new device's live IP is the right one. Only without a serial
		// (manual, no value) does the IP remain the only clue.
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
		// Set optional flags from the +printer dialog right when creating.
		if addFlags.KameraAus {
			if state.KameraAus == nil {
				state.KameraAus = map[string]bool{}
			}
			state.KameraAus[p.IP] = true
		}
		if addFlags.BlinkAus {
			if state.BlinkAus == nil {
				state.BlinkAus = map[string]bool{}
			}
			state.BlinkAus[p.IP] = true
		}
		mu.Unlock()
		if addFlags.InRepair {
			setRepairCache(p.IP, RepairFlag{InRepair: true, By: eigenerName, Since: time.Now()})
		}
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
		go syncMQTT() // otherwise the new printer stays without status until restart
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
			mqttMgr.Disconnect(ip) // credentials may have changed
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
	// We detect duplicates primarily by serial number, only fallback by
	// IP (for rows without a serial). A stale IP of an
	// offline printer thus does not block adding a new device.
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
	// Kept short: a hanging printer otherwise occupies one of the four
	// transcode slots and slows down the healthy ones.
	snapFetchTimeout = 8 * time.Second
	snapMaxBackoff   = 2 * time.Minute

	// Every image request starts its own ffmpeg process in go2rtc. More than
	// that it cannot sustain — with 42 printers every 2 s that is 21
	// requests per second queuing before four processing slots.
	snapMaxRatePerSec = 3.0
)

type snapEntry struct {
	mu      sync.Mutex
	data    []byte
	ts      time.Time
	err     error
	fails   int       // aufeinanderfolgende Fehlschlaege
	nextTry time.Time // before this, no request is even made
}

// backoffFor grows with each failure: 15 s, 30 s, 60 s, then 2 min.
// So a permanently broken stream no longer occupies a slot.
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

// effectiveSnapInterval says how often a single printer can actually get a
// new image without overrunning go2rtc. The desired value
// is only honored as long as the overall rate allows.
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
		// go2rtc answers with 200 and an empty body in this case and keeps
		// the reason to itself — it is only in its own log. So
		// we look ourselves instead of blaming ffmpeg wholesale.
		return nil, fmt.Errorf("kein Bild: %s", diagnoseStream(stream))
	}
	return data, nil
}

// diagnoseStream finds the actual reason why go2rtc returns no image.
func diagnoseStream(stream string) string {
	// 1. Does go2rtc know the stream at all?
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

	// 2. Is ffmpeg present? Without it go2rtc cannot convert H264 to JPEG.
	if !fileExists(ffmpegBinPath()) {
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			return "ffmpeg fehlt — unter Einstellungen nachladen"
		}
	}

	// 3. Does the printer's camera answer?
	// Standalone cameras have no printer — their own camera-specific
	// Diagnose (sauber getrennt, siehe camera.go).
	if isCameraStream(stream) {
		return cameraStreamDiagnostics(stream)
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

	// 4. Port open but still no image — then usually the access code is wrong
	return fmt.Sprintf("%s nimmt Verbindungen an, liefert aber kein Bild — Zugangscode prüfen", name)
}

// printerForStream finds the printer whose name produced the stream name.
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
	if netPaused() {
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
		// Standalone camera? Then treat the id as a camera ID.
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
	// Raise it to what go2rtc can still handle at this printer count
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
		// Standalone cameras: do NOT return 503. The browser logs
		// every 4xx/5xx as a red console error — for a temporarily
		// unreachable outdoor camera that is just noise. Instead 200 with an empty
		// body and a hint header; the UI shows "unavailable"
		// without a spinner and without a console error and honors the pause.
		if isCameraStream(stream) {
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
	go2rtcMu.Lock()
	pid := 0
	if go2rtcCmd != nil && go2rtcCmd.Process != nil {
		pid = go2rtcCmd.Process.Pid
	}
	go2rtcMu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online":   checkGo2rtcRunning(),
		"port":     go2rtcPort,
		"prozesse": countGo2rtc(),
		"pid":      pid,
		"pids":     go2rtcPIDs(), // all running go2rtc PIDs, not only the managed one
	})
}

// handleG2KillAll is the emergency switch: stop all go2rtc processes — including
// orphans — and then start exactly one fresh. Helps when
// processes have piled up or a program update does not take.
func handleG2KillAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	stopGo2rtc() // cleanly deregister and stop our own process
	beendet, _ := killAllGo2rtc()
	time.Sleep(500 * time.Millisecond)
	// Start a fresh one if installed and not in the network pause.
	if !netPaused() && fileExists(go2rtcBinPath()) {
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
// restartGo2rtcAsync coalesces restarts instead of running each one.
//
// Previously each call started its own run. Adding ten printers
// in a row triggered ten overlapping stop-start sequences —
// go2rtc barely got running. Now each call just resets the deadline;
// it runs once, when things have settled.
//
// The precheck deliberately happens IMMEDIATELY, not in the goroutine:
// if no go2rtc is installed or network traffic is paused,
// nothing is left pending that could start up later.
var restartTimer *time.Timer

func restartGo2rtcAsync() {
	if netPaused() || !fileExists(go2rtcBinPath()) {
		return
	}
	go2rtcMu.Lock()
	defer go2rtcMu.Unlock()
	if restartTimer != nil {
		restartTimer.Stop()
	}
	restartTimer = time.AfterFunc(800*time.Millisecond, restartGo2rtc)
}

func handleG2Restart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	go restartGo2rtc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
}

// go2rtcWanted says whether go2rtc SHOULD run. Only this lets a crash be told
// einem gewollten Stopp unterscheiden.
var (
	go2rtcWanted bool
	go2rtcGen    int
	go2rtcFails  int
)

func startGo2rtc() {
	if shuttingDown.Load() {
		return // im Beenden nichts mehr hochfahren
	}
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
	cmd.Env = envWithFFmpeg() // so go2rtc finds our ffmpeg
	// Ausgabe mitschreiben. Vorher wurde sie verworfen — endete go2rtc, stand
	// from nowhere why, and every debugging was guesswork.
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

// verifyStreamsLoaded compares what was written to the configuration
// with what go2rtc made of it. If they differ, the file is faulty
// — go2rtc then runs but knows no or too few streams.
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

// superviseGo2rtc waits for the process to end. If it ends although it should
// run, it is restarted — previously it simply vanished and with it all
// videos and snapshots, until someone intervened by hand.
func superviseGo2rtc(cmd *exec.Cmd, gen int) {
	started := time.Now()
	err := cmd.Wait()

	go2rtcMu.Lock()
	superseded := gen != go2rtcGen
	wanted := go2rtcWanted
	if time.Since(started) > time.Minute {
		go2rtcFails = 0 // ran long enough, so no start problem
	}
	fails := go2rtcFails
	go2rtcMu.Unlock()

	if superseded || !wanted {
		return // intentionally stopped or long since replaced
	}
	// Waehrend einer Netzwerkpause bleibt go2rtc unten. Ohne diese Sperre
	// the supervisor would read the intended stop as a crash and pull it
	// back up seconds later — the pause would then be ineffective.
	if netPaused() {
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

// go2rtcHealthLoop catches the case where the process still exists
// but no longer responds.
func go2rtcHealthLoop() {
	// During the pause go2rtc deliberately stays down.
	for {
		time.Sleep(30 * time.Second)
		if netPaused() {
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
		// Second chance before a restart
		time.Sleep(3 * time.Second)
		if checkGo2rtcRunning() {
			continue
		}
		log.Printf("⚠️  go2rtc antwortet nicht auf Port %d — starte neu", go2rtcPort)
		restartGo2rtc()
	}
}

// restartGo2rtc restarts go2rtc thoroughly. First the self-managed
// process is stopped, then ALL still-running (including orphaned)
// go2rtc processes including children are removed — otherwise a stuck one
// keeps holding port 1984 and the fresh one cannot bind. This is exactly why
// the in-program restart failed: stopGo2rtc alone only stops the own process.
// neustartMu ensures only ONE restart runs at a time. Without it two
// simultaneous restarts (e.g. button + automatic trigger) could
// choke each other: one's killAllGo2rtc kills the freshly
// gestarteten des anderen.
var neustartMu sync.Mutex

func restartGo2rtc() {
	if shuttingDown.Load() {
		return // no restart during shutdown
	}
	if netPaused() {
		return // during the pause go2rtc deliberately stays down
	}
	if !neustartMu.TryLock() {
		return // a restart is already running
	}
	defer neustartMu.Unlock()

	stopGo2rtc()
	if n, _ := killAllGo2rtc(); n > 0 {
		log.Printf("🔁 Neustart: %d verbliebene(n) go2rtc-Prozess(e) beendet", n)
	}

	// Wait until port 1984 is really free — not blindly a fixed time.
	for i := 0; i < 25 && checkGo2rtcRunning(); i++ {
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond) // Logdatei/Handles freigeben lassen

	go2rtcMu.Lock()
	go2rtcWanted = true
	go2rtcMu.Unlock()
	startGo2rtc()

	// Check whether the fresh process actually comes up. If not,
	// clean up and start a second time — this catches stuck
	// ports and failed starts.
	if !waitForGo2rtc(8 * time.Second) {
		log.Printf("⚠️  go2rtc nach Neustart nicht erreichbar — zweiter Versuch")
		killAllGo2rtc()
		time.Sleep(700 * time.Millisecond)
		go2rtcMu.Lock()
		go2rtcWanted = true
		go2rtcMu.Unlock()
		startGo2rtc()
		if waitForGo2rtc(8 * time.Second) {
			log.Printf("✅ go2rtc nach zweitem Versuch erreichbar")
		} else {
			log.Printf("❌ go2rtc laesst sich nicht starten — siehe go2rtc.log")
		}
	} else {
		log.Printf("✅ go2rtc nach Neustart erreichbar")
	}
}

// waitForGo2rtc polls port 1984 until go2rtc answers or the deadline passes.
func waitForGo2rtc(frist time.Duration) bool {
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
	go2rtcGen++ // invalidate the running supervision
	if go2rtcCmd != nil && go2rtcCmd.Process != nil {
		killGo2rtcTree(go2rtcCmd) // including ffmpeg children — otherwise they hang ~10 s
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

// buildYaml builds the configuration. Duplicate stream names are fatal here:
// on a duplicate key go2rtc discards the ENTIRE streams section and
// then knows not a single stream — but keeps running as if
// nothing happened. So names are made unique here.
// cameraSources returns the addresses for a printer — in the order
// go2rtc should try them.
//
// The difference between the schemes is decisive: rtspx:// is a
// special form that works reliably on the X1 series. The H2 and
// P2 series respond to it with a redirect to rtsps:// and return no
// image — exactly the behaviour seen on all H devices. So
// steht dort rtsps:// vorn.
//
// Both are given so a device that behaves differently than its
// model name suggests still returns an image: go2rtc tries the
// entries in order.
func cameraSources(p Printer, gemessen map[string]string) []string {
	auth := fmt.Sprintf("bblp:%s@%s:322", p.Code, p.IP)

	// If it was measured on the device which form works, that takes precedence over
	// jeder Ableitung aus dem Modellnamen.
	//
	// The map is passed in and not read here: buildYaml runs
	// already under the lock; a second access to it hangs the program
	// up. Exactly that happened on the first attempt.
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
		for _, q := range cameraSources(p.printer, gemessen) {
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

// streamNameFor returns the actually used stream name of a printer,
// i.e. including the de-duplication.
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
	// IMPORTANT: on Mi-Home login go2rtc writes its own sections into this
	// file (account/token for the Xiaomi cloud key). When rewriting the
	// sections we manage (api/streams) these must NOT be
	// lost — otherwise go2rtc can no longer fetch the camera key. So
	// foreign top-level sections are carried over from the existing file.
	vorher, _ := os.ReadFile(pfad)
	fremd := erhalteFremdeSektionen(string(vorher))
	yaml := buildYaml(state.Printers, gemessen) + cameraStreamsYaml(cams) + fremd
	os.WriteFile(pfad, []byte(yaml), 0644)
}

// ─── PERSISTENCE ──────────────────────────────────────────────────────────────

func loadState() {
	data, _ := os.ReadFile(dataFile) // if missing, state stays empty — ok
	if len(data) > 0 {
		mu.Lock()
		json.Unmarshal(data, &state)
		mu.Unlock()
	}
	// Betriebsdaten liegen in eigenen Dateien; fehlen sie, einmalig aus einer
	// alten config.json uebernehmen (Migration).
	loadOrMigrateOps(data)
}

// saveState writes ONLY the settings to config.json. The four
// operational-data fields (upload history, runtime, repair, FwUpdates) are
// excluded with json:"-". Runtime/Reparatur/FwUpdates are small and
// are written along here (runtime.json); the large upload history
// stays out and is saved only on real changes via saveUploadLog()
// gesichert.
func saveState() {
	mu.RLock()
	data, _ := json.MarshalIndent(state, "", "  ")
	mu.RUnlock()
	backupConfig()
	atomicWrite(dataFile, data, 0644)
	scheduleRuntimeSave()
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

	// The self-counted runtime does not belong in the MQTT state — it
	// does not come from the device. It is placed alongside here.
	mu.RLock()
	laufzeit := map[string]int64{}
	for ip, sek := range state.Laufzeit {
		laufzeit[ip] = sek
	}
	mu.RUnlock()

	type erweitert struct {
		PrinterStatus
		LaufzeitSek  int64            `json:"laufzeit_sek,omitempty"`
		LaufzeitText string           `json:"laufzeit_text,omitempty"`
		FwUpdates    []FwModuleUpdate `json:"fw_updates,omitempty"`
		FwScheduled  bool             `json:"fw_scheduled,omitempty"`
		Reparatur    *RepairFlag      `json:"reparatur,omitempty"`
	}
	// Open firmware updates per printer — compared against the running
	// firmware and already-installed items drop out.
	fw := fwForDisplay()
	// Reparatur-Markierungen aus dem Zwischenspeicher.
	mu.RLock()
	rep := map[string]RepairFlag{}
	for ip, f := range state.Reparatur {
		rep[ip] = f
	}
	fwPlan := map[string]bool{}
	for ip, an := range state.PendingFwUpdate {
		fwPlan[ip] = an
	}
	mu.RUnlock()
	out := map[string]erweitert{}
	for ip, st := range statuses {
		e := erweitert{PrinterStatus: st, LaufzeitSek: laufzeit[ip], LaufzeitText: runtimeText(laufzeit[ip]), FwUpdates: fw[ip], FwScheduled: fwPlan[ip]}
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
	mu.RLock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.RUnlock()

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

// handlePrintCommand sends pause/resume/stop to one or more printers.
// Without "ips" nothing happens — an accidental call should not hit the whole
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

// handleSettings stores theme, language and column count. Previously these were in
// browser storage — and thus gone once the Edge profile was recreated.
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
		if body.Lang != nil && (*body.Lang == "de" || *body.Lang == "en" || *body.Lang == "zh" || *body.Lang == "es") {
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
			// Turning off also means: immediately restore normal light
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
	for k, v := range state.BlinkAus {
		if v {
			blinkAus[k] = true
		}
	}
	kameraAus := map[string]bool{}
	for k, v := range state.KameraAus {
		if v {
			kameraAus[k] = true
		}
	}
	timelapseAus := map[string]bool{}
	for k, v := range state.TimelapseAus {
		if v {
			timelapseAus[k] = true
		}
	}
	mu.Unlock()
	if startFilter == "" {
		startFilter = "all"
	}
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
		"timelapse_aus": timelapseAus, "sprache_gewaehlt": spracheGewaehlt})
}

// handleOpenFolder opens a folder in Explorer. A file:// link would
// be blocked by the browser from an http page, hence the detour.
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
	// Only open our own directories — no arbitrary path from outside
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
	// explorer.exe returns a non-zero value even on success
	_ = cmd.Start()
	writeJSON(w, map[string]any{"opened": clean})
}

// ─── DIAGNOSE ─────────────────────────────────────────────────────────────────

// handleDiagnostics collects everything needed to debug go2rtc
// — so one no longer has to guess what is going on.
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

// tailFile returns the last n lines of a file.
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
