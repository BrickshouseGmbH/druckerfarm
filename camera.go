package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ─── STANDALONE-KAMERAS ───────────────────────────────────────────────────────
//
// Besides the printer cameras, the tool can embed standalone video sources
// (e.g. an outdoor camera). They are NOT tied to a printer object and have
// tauchen daher in keiner druckerbezogenen Auswertung auf: keine SSDP-Discovery,
// no MQTT connection, no FTP sync, no printer count. The only
// touch point is go2rtc — the source is written as a stream into go2rtc.yaml
// so WebRTC and snapshot work just like for the printers.

type CameraCfg struct {
	ID     string `json:"id"`     // interne Kennung, z. B. "cw300"
	Name   string `json:"name"`   // display name, e.g. "Outdoor camera office"
	Stream string `json:"stream"` // go2rtc-Streamname (meist == ID)
	// Source is the full go2rtc source including credentials
	// (e.g. xiaomi://user:pass@ip?did=…). It is written into go2rtc.yaml
	// — the tool does this; the user does not touch go2rtc.
	// WICHTIG: niemals in Logs ausgeben.
	Source string `json:"source,omitempty"`
	Kind   string `json:"kind,omitempty"` // immer "standalone"
	Fav    bool   `json:"fav,omitempty"`  // als Favorit markiert
}

// cameraStreamsYaml generates the go2rtc stream lines for the standalone cameras.
// Appended to the printer YAML so rewriting the file does not lose the
// cameras.
func cameraStreamsYaml(cams []CameraCfg) string {
	var sb strings.Builder
	for _, c := range cams {
		stream := strings.TrimSpace(c.Stream)
		src := strings.TrimSpace(c.Source)
		if stream == "" || src == "" {
			continue
		}
		sb.WriteString("  " + stream + ":\n")
		sb.WriteString("    - " + src + "\n")
	}
	return sb.String()
}

// cameraStreamName returns the go2rtc stream name for a camera ID (or "").
func cameraStreamName(id string) string {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.Cameras {
		if c.ID == id {
			if s := strings.TrimSpace(c.Stream); s != "" {
				return s
			}
			return c.ID
		}
	}
	return ""
}

// handleCameras: GET lists the cameras, POST creates/updates.
func handleCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		cams := make([]CameraCfg, len(state.Cameras))
		copy(cams, state.Cameras)
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(cams)

	case http.MethodPost:
		var c CameraCfg
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.ID = strings.TrimSpace(c.ID)
		c.Name = strings.TrimSpace(c.Name)
		c.Stream = strings.TrimSpace(c.Stream)
		c.Source = strings.TrimSpace(c.Source)
		c.Kind = "standalone"
		if c.ID == "" {
			c.ID = slugID(c.Name)
		}
		if c.Stream == "" {
			c.Stream = c.ID
		}
		if c.ID == "" || c.Name == "" || c.Source == "" {
			http.Error(w, "id/name/source erforderlich", http.StatusBadRequest)
			return
		}
		mu.Lock()
		ersetzt := false
		for i := range state.Cameras {
			if state.Cameras[i].ID == c.ID {
				c.Fav = state.Cameras[i].Fav // keep the favorite marker when editing
				state.Cameras[i] = c
				ersetzt = true
				break
			}
		}
		if !ersetzt {
			state.Cameras = append(state.Cameras, c)
		}
		mu.Unlock()
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
		// No logging of the source (contains credentials).
		ausgabe := c
		_ = json.NewEncoder(w).Encode(ausgabe)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCameraByID: DELETE /api/cameras/{id}
func handleCameraByID(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/cameras/"), "/")
	if id == "" {
		http.Error(w, "id fehlt", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		// Only toggle the favorite marker — without a go2rtc restart.
		var body struct {
			Fav *bool `json:"fav"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		found := false
		for i := range state.Cameras {
			if state.Cameras[i].ID == id {
				if body.Fav != nil {
					state.Cameras[i].Fav = *body.Fav
				}
				found = true
				break
			}
		}
		mu.Unlock()
		if !found {
			http.Error(w, "unbekannte Kamera", http.StatusNotFound)
			return
		}
		saveState()
		writeJSON(w, map[string]any{"ok": true})
	case http.MethodDelete:
		mu.Lock()
		neu := state.Cameras[:0:0]
		gefunden := false
		for _, c := range state.Cameras {
			if c.ID == id {
				gefunden = true
				continue
			}
			neu = append(neu, c)
		}
		state.Cameras = neu
		mu.Unlock()
		if !gefunden {
			http.Error(w, "unbekannte Kamera", http.StatusNotFound)
			return
		}
		saveState()
		writeGo2rtcYaml()
		restartGo2rtcAsync()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// slugID turns a name into a slim ID (a–z, 0–9, hyphen).
func slugID(name string) string {
	var sb strings.Builder
	prev := false
	for _, c := range strings.ToLower(name) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			sb.WriteRune(c)
			prev = false
		} else if !prev {
			sb.WriteRune('-')
			prev = true
		}
	}
	return strings.Trim(sb.String(), "-")
}

// isCameraStream reports whether a go2rtc stream name belongs to a standalone
// camera (stream or ID). This lets the snapshot diagnostics separate cameras
// from printers — the camera logic stays here, not in the printer part.
func isCameraStream(stream string) bool {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.Cameras {
		if c.Stream == stream || c.ID == stream {
			return true
		}
	}
	return false
}

// cameraName returns the display name for a camera stream/ID (or "").
func cameraName(stream string) string {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.Cameras {
		if c.Stream == stream || c.ID == stream {
			return c.Name
		}
	}
	return ""
}

// cameraStreamDiagnostics explains why a camera returns no image — without any
// printer assumptions (no MQTT/port-322 check, no access code). The
// Xiaomi source connects via P2P; the session key briefly needs internet.
// Deliberately terse and camera-specific.
func cameraStreamDiagnostics(stream string) string {
	name := cameraName(stream)
	if name == "" {
		name = stream
	}
	basis := `Kamera „` + name + `" liefert gerade kein Bild`
	if e := go2rtcProducerError(stream); e != "" {
		return basis + " — go2rtc meldet: " + e
	}
	return basis + ` — Kamera erreichbar? (Beim Verbindungsaufbau braucht go2rtc kurz Internet für den Xiaomi-Sitzungsschlüssel. Falls „xiaomi"/Login-Fehler: der Mi-Home-Login muss in genau dieser go2rtc-Instanz hinterlegt sein.)`
}

// go2rtcProducerError asks go2rtc for the concrete error of a stream
// (e.g. "xiaomi: login failed"). Best-effort — if the query fails,
// "" is returned and the generic message applies.
func go2rtcProducerError(stream string) string {
	u := fmt.Sprintf("http://127.0.0.1:%d/api/streams?src=%s", go2rtcPort, url.QueryEscape(stream))
	resp, err := snapClient.Get(u)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var v any
	if json.NewDecoder(resp.Body).Decode(&v) != nil {
		return ""
	}
	// Search recursively for "error" fields and take the first non-empty text.
	var suche func(any) string
	suche = func(n any) string {
		switch t := n.(type) {
		case map[string]any:
			if e, ok := t["error"].(string); ok && strings.TrimSpace(e) != "" {
				return strings.TrimSpace(e)
			}
			for _, val := range t {
				if r := suche(val); r != "" {
					return r
				}
			}
		case []any:
			for _, val := range t {
				if r := suche(val); r != "" {
					return r
				}
			}
		}
		return ""
	}
	return suche(v)
}

// erhalteFremdeSektionen liest aus einer bestehenden go2rtc.yaml alle
// top-level sections that are NOT managed by this tool
// (i.e. not "api" and not "streams"). That is exactly where go2rtc stores e.g. the
// Mi-Home credentials/tokens. They are appended when rewriting so that
// a completed Xiaomi login is preserved.
func erhalteFremdeSektionen(vorhanden string) string {
	if strings.TrimSpace(vorhanden) == "" {
		return ""
	}
	verwaltet := map[string]bool{"api": true, "streams": true}
	var out []string
	behalten := false
	for _, ln := range strings.Split(vorhanden, "\n") {
		obenLinks := len(ln) > 0 && ln[0] != ' ' && ln[0] != '\t'
		if obenLinks {
			t := strings.TrimSpace(ln)
			if t == "" || strings.HasPrefix(t, "#") {
				behalten = false
				continue
			}
			key := t
			if i := strings.IndexByte(t, ':'); i >= 0 {
				key = strings.TrimSpace(t[:i])
			}
			behalten = !verwaltet[key]
		}
		if behalten {
			out = append(out, ln)
		}
	}
	body := strings.TrimRight(strings.Join(out, "\n"), "\n")
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return "\n" + body + "\n"
}
