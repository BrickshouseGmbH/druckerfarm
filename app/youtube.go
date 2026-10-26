package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── YOUTUBE-LIVE (ffmpeg → RTMP) ─────────────────────────────────────────────
//
// Builds a tile mosaic from the running camera streams and pushes it via
// RTMP to YouTube. ffmpeg is already present as a component (for snapshots).
//
// IMPORTANT — security: the inputs are the LOCAL go2rtc RTSP streams
// (rtsp://127.0.0.1:8554/<name>), NOT the printers directly. So the
// printer access codes are NOT in the ffmpeg command line, and go2rtc fetches each
// stream only once from the printer (no double access).

var go2rtcRTSPPort = 8554

var (
	ytMu      sync.Mutex
	ytCmd     *exec.Cmd
	ytWanted  bool
	ytErr     string
	ytGen     int
	ytStarted time.Time
	ytTiles   int
)

type ytOptions struct {
	Key        string
	BitrateK   int
	Resolution string
	Fps        int
	Cols       int
}

func ytFfmpegBin() string {
	if fileExists(ffmpegBinPath()) {
		return ffmpegBinPath()
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ffmpegBinPath()
}

// parseRes zerlegt "BxH" in gerade, sinnvolle Werte. Vorgabe 1920x1080.
func parseRes(s string) (int, int) {
	w, h := 1920, 1080
	parts := strings.Split(strings.ToLower(strings.TrimSpace(s)), "x")
	if len(parts) == 2 {
		if a, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil && a >= 320 && a <= 3840 {
			w = a
		}
		if b, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && b >= 240 && b <= 2160 {
			h = b
		}
	}
	return w &^ 1, h &^ 1
}

// buildYouTubeArgs builds the full ffmpeg argument list. Pure function —
// dadurch im Test pruefbar.
func buildYouTubeArgs(streams []string, o ytOptions) []string {
	W, H := parseRes(o.Resolution)
	fps := o.Fps
	if fps <= 0 {
		fps = 15
	}
	br := o.BitrateK
	if br <= 0 {
		br = 4500
	}
	n := len(streams)
	cols := o.Cols
	if cols < 1 {
		cols = int(math.Ceil(math.Sqrt(float64(n))))
	}
	if cols < 1 {
		cols = 1
	}
	if cols > n {
		cols = n
	}
	rows := (n + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	cellW := (W / cols) &^ 1
	cellH := (H / rows) &^ 1
	if cellW < 2 {
		cellW = 2
	}
	if cellH < 2 {
		cellH = 2
	}

	args := []string{"-hide_banner", "-loglevel", "warning"}
	for _, s := range streams {
		args = append(args, "-rtsp_transport", "tcp", "-i", s)
	}
	// Silent audio signal — YouTube Live requires an audio track.
	args = append(args, "-f", "lavfi", "-i", "anullsrc=channel_layout=stereo:sample_rate=44100")

	parts := make([]string, 0, n+1)
	for i := 0; i < n; i++ {
		parts = append(parts, fmt.Sprintf(
			"[%d:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=%d[v%d]",
			i, cellW, cellH, cellW, cellH, fps, i))
	}
	outLabel := "[v0]"
	if n > 1 {
		var labs strings.Builder
		layout := make([]string, n)
		for i := 0; i < n; i++ {
			labs.WriteString(fmt.Sprintf("[v%d]", i))
			col := i % cols
			row := i / cols
			layout[i] = fmt.Sprintf("%d_%d", col*cellW, row*cellH)
		}
		parts = append(parts, fmt.Sprintf("%sxstack=inputs=%d:layout=%s:fill=black[out]",
			labs.String(), n, strings.Join(layout, "|")))
		outLabel = "[out]"
	}
	args = append(args, "-filter_complex", strings.Join(parts, ";"))

	args = append(args,
		"-map", outLabel,
		"-map", fmt.Sprintf("%d:a", n),
		"-c:v", "libx264", "-preset", "veryfast", "-pix_fmt", "yuv420p",
		"-b:v", fmt.Sprintf("%dk", br), "-maxrate", fmt.Sprintf("%dk", br), "-bufsize", fmt.Sprintf("%dk", br*2),
		"-g", fmt.Sprintf("%d", fps*2), "-r", fmt.Sprintf("%d", fps),
		"-c:a", "aac", "-b:a", "128k", "-ar", "44100",
		"-f", "flv", "rtmp://a.rtmp.youtube.com/live2/"+o.Key,
	)
	return args
}

// ytStreams collects the local go2rtc RTSP addresses of the active tiles.
func ytStreams() []string {
	mu.Lock()
	printers := append([]Printer(nil), state.Printers...)
	kamAus := map[string]bool{}
	for k, v := range state.KameraAus {
		kamAus[k] = v
	}
	cams := append([]CameraCfg(nil), state.Cameras...)
	mu.Unlock()

	var out []string
	for _, e := range namedStreams(printers) {
		if kamAus[e.printer.IP] {
			continue
		}
		if s := mqttMgr.GetStatus(e.printer.IP); s == nil || !s.Online {
			continue
		}
		out = append(out, fmt.Sprintf("rtsp://127.0.0.1:%d/%s", go2rtcRTSPPort, e.stream))
	}
	for _, c := range cams {
		st := strings.TrimSpace(c.Stream)
		if st == "" {
			st = strings.TrimSpace(c.ID)
		}
		if st == "" {
			continue
		}
		out = append(out, fmt.Sprintf("rtsp://127.0.0.1:%d/%s", go2rtcRTSPPort, st))
	}
	return out
}

func startYouTube() error {
	mu.Lock()
	o := ytOptions{
		Key:        strings.TrimSpace(state.YTKey),
		BitrateK:   state.YTBitrateK,
		Resolution: state.YTResolution,
		Fps:        state.YTFps,
		Cols:       state.BroadcastCols,
	}
	mu.Unlock()
	if o.Key == "" {
		return fmt.Errorf("kein YouTube-Stream-Key hinterlegt")
	}
	bin := ytFfmpegBin()
	if !fileExists(bin) {
		return fmt.Errorf("ffmpeg ist nicht installiert")
	}
	streams := ytStreams()
	if len(streams) == 0 {
		return fmt.Errorf("keine aktiven Kamera-Streams vorhanden (go2rtc laeuft? Drucker online?)")
	}

	ytMu.Lock()
	if ytCmd != nil {
		ytMu.Unlock()
		return nil
	}
	args := buildYouTubeArgs(streams, o)
	cmd := exec.Command(bin, args...)
	cmd.Env = envWithFFmpeg()
	hideProcessWindow(cmd)
	if lf, err := os.OpenFile(filepath.Join(appDir, "youtube.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		cmd.Stdout = lf
		cmd.Stderr = lf
	}
	if err := cmd.Start(); err != nil {
		ytMu.Unlock()
		return err
	}
	ytCmd = cmd
	ytWanted = true
	ytErr = ""
	ytStarted = time.Now()
	ytTiles = len(streams)
	ytGen++
	gen := ytGen
	ytMu.Unlock()

	go superviseYouTube(cmd, gen)
	log.Printf("✅ YouTube-Push gestartet (%d Kacheln)", len(streams))
	return nil
}

func superviseYouTube(cmd *exec.Cmd, gen int) {
	err := cmd.Wait()
	ytMu.Lock()
	if ytGen == gen {
		ytCmd = nil
		if ytWanted {
			ytErr = "ffmpeg wurde beendet — siehe youtube.log"
			if err != nil {
				ytErr = err.Error()
			}
		}
	}
	ytMu.Unlock()
}

func stopYouTube() {
	ytMu.Lock()
	ytWanted = false
	ytGen++
	cmd := ytCmd
	ytCmd = nil
	ytMu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

// ─── HTTP-API ─────────────────────────────────────────────────────────────────

func handleYouTubeStatus(w http.ResponseWriter, r *http.Request) {
	ytMu.Lock()
	running := ytCmd != nil
	errs := ytErr
	tiles := ytTiles
	var secs int64
	if running {
		secs = int64(time.Since(ytStarted).Seconds())
	}
	ytMu.Unlock()
	mu.Lock()
	hasKey := strings.TrimSpace(state.YTKey) != ""
	mu.Unlock()
	writeJSON(w, map[string]any{
		"running":        running,
		"tiles":          tiles,
		"seconds":        secs,
		"err":            errs,
		"has_key":        hasKey,
		"ffmpeg_present": fileExists(ytFfmpegBin()),
		"go2rtc_present": fileExists(go2rtcBinPath()),
		"go2rtc_running": checkGo2rtcRunning(),
	})
}

func handleYouTubeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	if err := startYouTube(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func handleYouTubeStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	stopYouTube()
	writeJSON(w, map[string]any{"ok": true})
}

func handleYouTubeConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Key        *string `json:"key"`
			BitrateK   *int    `json:"bitrate_k"`
			Resolution *string `json:"resolution"`
			Fps        *int    `json:"fps"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		if body.Key != nil {
			state.YTKey = strings.TrimSpace(*body.Key)
		}
		if body.BitrateK != nil && *body.BitrateK >= 500 && *body.BitrateK <= 51000 {
			state.YTBitrateK = *body.BitrateK
		}
		if body.Resolution != nil {
			state.YTResolution = strings.TrimSpace(*body.Resolution)
		}
		if body.Fps != nil && *body.Fps >= 1 && *body.Fps <= 60 {
			state.YTFps = *body.Fps
		}
		mu.Unlock()
		saveState()
	}

	mu.Lock()
	hasKey := strings.TrimSpace(state.YTKey) != ""
	br := state.YTBitrateK
	res := state.YTResolution
	fps := state.YTFps
	mu.Unlock()
	if br == 0 {
		br = 4500
	}
	if res == "" {
		res = "1920x1080"
	}
	if fps == 0 {
		fps = 15
	}
	writeJSON(w, map[string]any{"has_key": hasKey, "bitrate_k": br, "resolution": res, "fps": fps})
}
