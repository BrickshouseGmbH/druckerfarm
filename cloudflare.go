package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// ─── VEROEFFENTLICHUNG PER CLOUDFLARE-TUNNEL ─────────────────────────────────
//
// The broadcast page (/broadcast) can be exposed to the internet without port forwarding:
// cloudflared opens an OUTBOUND connection to Cloudflare and provides a
// public https address (….trycloudflare.com) pointing to a LOCAL
// Public-Server zeigt.
//
// IMPORTANT — security: the tunnel does NOT point to the control UI
// (appPort), but to a separate, strictly read-only public server
// (publicPort). It serves only the broadcast page and sanitized read data:
// NO control endpoints, NO printer access codes, NO camera
// credentials. So nobody outside can control anything and nothing confidential
// abgreifen.

const cloudflaredURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"

var (
	publicPort = 8766

	cfMu     sync.Mutex
	cfCmd    *exec.Cmd
	cfWanted bool
	cfURL    string // erkannte oeffentliche trycloudflare-Adresse
	cfErr    string
	cfGen    int

	pubSrv *http.Server

	cfURLRe = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)
)

func cloudflaredBinPath() string { return filepath.Join(appDir, "cloudflared"+exeSuffix()) }

func randToken() string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ─── PUBLIC SERVER (read-only, sanitized) ─────────────────────────────────────

// getOnly rejects anything but GET/HEAD — the public server has no
// schreibenden Vorgaenge.
func getOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "nur GET", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

// publicGate protects the public server in secret mode with a token
// (?k=… or cookie). In public mode it is freely reachable.
func publicGate(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		secret := state.TunnelMode == "secret"
		tok := state.TunnelToken
		mu.Unlock()
		if secret && tok != "" {
			k := r.URL.Query().Get("k")
			if k == "" {
				if c, e := r.Cookie("df_k"); e == nil {
					k = c.Value
				}
			}
			if k != tok {
				http.Error(w, "Zugang gesperrt (Token noetig)", http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "df_k", Value: tok, Path: "/", HttpOnly: true})
		}
		h(w, r)
	}
}

func startPublicServer() {
	cfMu.Lock()
	if pubSrv != nil {
		cfMu.Unlock()
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", publicGate(serveBroadcastPublic))
	mux.HandleFunc("/broadcast", publicGate(serveBroadcastPublic))
	mux.HandleFunc("/api/printers", publicGate(getOnly(handlePublicPrinters)))
	mux.HandleFunc("/api/status", publicGate(getOnly(handleStatus)))
	mux.HandleFunc("/api/settings", publicGate(getOnly(handlePublicSettings)))
	mux.HandleFunc("/api/cameras", publicGate(getOnly(handlePublicCameras)))
	mux.HandleFunc("/api/broadcast/config", publicGate(getOnly(handleBroadcastConfig)))
	mux.HandleFunc("/api/snapshot/", publicGate(getOnly(handleSnapshot)))
	mux.HandleFunc("/logo.svg", handleLogo)
	mux.HandleFunc("/splash.jpg", handleSplash)
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", publicPort), Handler: corsMiddleware(mux)}
	pubSrv = srv
	cfMu.Unlock()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Public-Server: %v", err)
		}
	}()
}

func stopPublicServer() {
	cfMu.Lock()
	srv := pubSrv
	pubSrv = nil
	cfMu.Unlock()
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

// serveBroadcastPublic serves the broadcast page over the tunnel — remote=true.
func serveBroadcastPublic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	page := replaceBroadcastOpts(pageWithVersion(), broadcastOptsJSONRemote(true))
	w.Write([]byte(page))
}

func handlePublicPrinters(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	out := make([]map[string]any, 0, len(state.Printers))
	for _, p := range state.Printers {
		// Deliberately WITHOUT access code and serial number.
		out = append(out, map[string]any{"ip": p.IP, "name": p.Name, "model": p.Model, "fav": p.Fav})
	}
	mu.Unlock()
	writeJSON(w, out)
}

func handlePublicCameras(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	out := make([]map[string]any, 0, len(state.Cameras))
	for _, c := range state.Cameras {
		// Bewusst OHNE Source (enthaelt Zugangsdaten).
		out = append(out, map[string]any{"id": c.ID, "name": c.Name, "stream": c.Stream, "kind": c.Kind})
	}
	mu.Unlock()
	writeJSON(w, out)
}

func handlePublicSettings(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	theme, lang, cols := state.Theme, state.Lang, state.Cols
	mu.Unlock()
	if theme == "" {
		theme = "light"
	}
	if lang == "" {
		lang = "de"
	}
	if cols == 0 {
		cols = 6
	}
	writeJSON(w, map[string]any{"theme": theme, "lang": lang, "cols": cols, "sprache_gewaehlt": true})
}

// ─── TUNNEL-PROZESS ───────────────────────────────────────────────────────────

func startTunnel() error {
	bin := cloudflaredBinPath()
	if !fileExists(bin) {
		return fmt.Errorf("cloudflared ist nicht installiert")
	}
	startPublicServer()

	cfMu.Lock()
	if cfCmd != nil {
		cfMu.Unlock()
		return nil // already running
	}
	cfURL = ""
	cfErr = ""
	target := fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	cmd := exec.Command(bin, "tunnel", "--no-autoupdate", "--url", target)
	hideProcessWindow(cmd)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		cfMu.Unlock()
		return err
	}
	cfCmd = cmd
	cfWanted = true
	cfGen++
	gen := cfGen
	cfMu.Unlock()

	go scanTunnelOutput(stdout)
	go scanTunnelOutput(stderr)
	go superviseTunnel(cmd, gen)
	log.Printf("✅ cloudflared gestartet (Ziel %s)", target)
	return nil
}

// scanTunnelOutput reads cloudflared's output line by line and captures the
// oeffentliche Adresse ab.
func scanTunnelOutput(rc io.ReadCloser) {
	if rc == nil {
		return
	}
	defer rc.Close()
	sc := bufio.NewScanner(rc)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if m := cfURLRe.FindString(sc.Text()); m != "" {
			cfMu.Lock()
			cfURL = m
			cfMu.Unlock()
			log.Printf("🌐 Tunnel-Adresse: %s", m)
		}
	}
}

func superviseTunnel(cmd *exec.Cmd, gen int) {
	err := cmd.Wait()
	cfMu.Lock()
	if cfGen == gen {
		cfCmd = nil
		if cfWanted {
			cfErr = "cloudflared wurde beendet"
			if err != nil {
				cfErr = err.Error()
			}
		}
	}
	cfMu.Unlock()
}

func stopTunnel() {
	cfMu.Lock()
	cfWanted = false
	cfGen++
	cmd := cfCmd
	cfCmd = nil
	cfURL = ""
	cfMu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	stopPublicServer()
}

// tunnelShareURL builds the shareable address (including the token in secret mode).
func tunnelShareURL() string {
	cfMu.Lock()
	base := cfURL
	cfMu.Unlock()
	if base == "" {
		return ""
	}
	mu.Lock()
	secret := state.TunnelMode == "secret"
	tok := state.TunnelToken
	mu.Unlock()
	u := base + "/broadcast"
	if secret && tok != "" {
		u += "?k=" + tok
	}
	return u
}

// ─── HTTP-API (Bedienoberflaeche) ─────────────────────────────────────────────

func handleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	cfMu.Lock()
	running := cfCmd != nil
	raw := cfURL
	errs := cfErr
	cfMu.Unlock()
	writeJSON(w, map[string]any{
		"installed": fileExists(cloudflaredBinPath()),
		"running":   running,
		"raw_url":   raw,
		"url":       tunnelShareURL(),
		"err":       errs,
	})
}

func handleTunnelStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	if err := startTunnel(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func handleTunnelStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	stopTunnel()
	writeJSON(w, map[string]any{"ok": true})
}

// handleTunnelInstall downloads cloudflared (a single .exe). Progress
// runs over the same channel as the other components (/api/components/progress).
func handleTunnelInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	go func() {
		setProgress(func(p *installProgress) {
			*p = installProgress{Active: true, Component: "cloudflared", Step: "download"}
		})
		tmp := cloudflaredBinPath() + ".tmp"
		if _, err := download(cloudflaredURL, tmp); err != nil {
			_ = os.Remove(tmp)
			setProgress(func(p *installProgress) { p.Step = "error"; p.Error = err.Error(); p.Active = false; p.Done = true })
			return
		}
		_ = os.Chmod(tmp, 0o755)
		if err := os.Rename(tmp, cloudflaredBinPath()); err != nil {
			setProgress(func(p *installProgress) { p.Step = "error"; p.Error = err.Error(); p.Active = false; p.Done = true })
			return
		}
		setProgress(func(p *installProgress) {
			p.Component = "cloudflared"
			p.Step = "done"
			p.Active = false
			p.Done = true
		})
		log.Printf("✅ cloudflared installiert: %s", cloudflaredBinPath())
	}()
	writeJSON(w, map[string]any{"ok": true})
}

func handleTunnelConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Mode  *string `json:"mode"`
			Click *bool   `json:"click"`
			Regen *bool   `json:"regen_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		if body.Mode != nil && (*body.Mode == "public" || *body.Mode == "secret") {
			state.TunnelMode = *body.Mode
			if state.TunnelMode == "secret" && state.TunnelToken == "" {
				state.TunnelToken = randToken()
			}
		}
		if body.Regen != nil && *body.Regen {
			state.TunnelToken = randToken()
		}
		if body.Click != nil {
			state.BroadcastClick = *body.Click
		}
		mu.Unlock()
		saveState()
	}

	mu.Lock()
	mode := state.TunnelMode
	hasTok := state.TunnelToken != ""
	click := state.BroadcastClick
	mu.Unlock()
	if mode == "" {
		mode = "public"
	}
	cfMu.Lock()
	running := cfCmd != nil
	cfMu.Unlock()
	writeJSON(w, map[string]any{
		"mode":      mode,
		"has_token": hasTok,
		"click":     click,
		"installed": fileExists(cloudflaredBinPath()),
		"running":   running,
		"url":       tunnelShareURL(),
	})
}
