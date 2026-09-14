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
// Neben den Druckerkameras kann das Tool eigenständige Videoquellen einbinden
// (z. B. eine Outdoor-Kamera). Sie hängen NICHT an einem Druckerobjekt und
// tauchen daher in keiner druckerbezogenen Auswertung auf: keine SSDP-Discovery,
// keine MQTT-Verbindung, kein FTP-Sync, keine Druckerzählung. Der einzige
// Berührungspunkt ist go2rtc — die Quelle wird als Stream in die go2rtc.yaml
// geschrieben, damit WebRTC und Snapshot wie bei den Druckern funktionieren.

type CameraCfg struct {
	ID     string `json:"id"`     // interne Kennung, z. B. "cw300"
	Name   string `json:"name"`   // Anzeigename, z. B. "Außenkamera Büro"
	Stream string `json:"stream"` // go2rtc-Streamname (meist == ID)
	// Source ist die vollständige go2rtc-Quelle inkl. Zugangsdaten
	// (z. B. xiaomi://user:pass@ip?did=…). Sie wird in die go2rtc.yaml
	// geschrieben — das übernimmt das Tool, der Anwender fasst go2rtc nicht an.
	// WICHTIG: niemals in Logs ausgeben.
	Source string `json:"source,omitempty"`
	Kind   string `json:"kind,omitempty"` // immer "standalone"
}

// cameraStreamsYaml erzeugt die go2rtc-Streamzeilen für die Standalone-Kameras.
// Wird an die Drucker-YAML angehängt, damit ein Neuschreiben der Datei die
// Kameras nicht verliert.
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

// kameraStreamName liefert den go2rtc-Streamnamen zu einer Kamera-ID (oder "").
func kameraStreamName(id string) string {
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

// handleCameras: GET listet die Kameras, POST legt an/aktualisiert.
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
		// Kein Logging der Source (enthält Zugangsdaten).
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

// slugID macht aus einem Namen eine schlanke ID (a–z, 0–9, Bindestrich).
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

// istKameraStream sagt, ob ein go2rtc-Streamname zu einer Standalone-Kamera
// gehört (Stream oder ID). Damit kann die Snapshot-Diagnose Kameras von
// Druckern trennen — die Kamera-Logik bleibt hier, nicht im Drucker-Teil.
func istKameraStream(stream string) bool {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.Cameras {
		if c.Stream == stream || c.ID == stream {
			return true
		}
	}
	return false
}

// kameraName liefert den Anzeigenamen zu einem Kamera-Stream/ID (oder "").
func kameraName(stream string) string {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.Cameras {
		if c.Stream == stream || c.ID == stream {
			return c.Name
		}
	}
	return ""
}

// kameraStreamDiagnose erklärt, warum eine Kamera kein Bild liefert — ohne jede
// Drucker-Annahme (keine MQTT-/Port-322-Prüfung, keinen Zugangscode). Die
// Xiaomi-Quelle verbindet per P2P; für den Sitzungsschlüssel wird kurz Internet
// gebraucht. Bewusst knapp und kameraspezifisch.
func kameraStreamDiagnose(stream string) string {
	name := kameraName(stream)
	if name == "" {
		name = stream
	}
	basis := `Kamera „` + name + `" liefert gerade kein Bild`
	if e := go2rtcProducerFehler(stream); e != "" {
		return basis + " — go2rtc meldet: " + e
	}
	return basis + ` — Kamera erreichbar? (Beim Verbindungsaufbau braucht go2rtc kurz Internet für den Xiaomi-Sitzungsschlüssel. Falls „xiaomi"/Login-Fehler: der Mi-Home-Login muss in genau dieser go2rtc-Instanz hinterlegt sein.)`
}

// go2rtcProducerFehler fragt go2rtc nach dem konkreten Fehler eines Streams
// (z. B. „xiaomi: login failed"). Best-effort — schlägt die Abfrage fehl,
// kommt "" zurück und die allgemeine Meldung greift.
func go2rtcProducerFehler(stream string) string {
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
	// Rekursiv nach "error"-Feldern suchen und den ersten nichtleeren Text nehmen.
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
// Top-Level-Abschnitte heraus, die NICHT von diesem Tool verwaltet werden
// (also nicht "api" und nicht "streams"). Genau dort legt go2rtc z. B. die
// Mi-Home-Zugangsdaten/Tokens ab. Sie werden beim Neuschreiben angehängt, damit
// ein einmal erfolgter Xiaomi-Login erhalten bleibt.
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
