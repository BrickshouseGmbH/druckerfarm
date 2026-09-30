package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ─── KAMERA-TEST ──────────────────────────────────────────────────────────────
//
// Which address form a printer understands cannot be reliably derived from
// the model name — the firmware decides, and that changes.
// Instead of guessing on, this test tries the forms on the device itself
// and reports which one returns an image.
//
// go2rtc can test an address directly: /api/frame.jpeg?src=<address> opens
// the connection, grabs a single frame and drops it again. It needs
// no configuration entry for that.

type candidate struct {
	Schema  string `json:"schema"`
	Adresse string `json:"-"` // contains the access code, does not belong in JSON
	Pfad    string `json:"pfad"`
}

// cameraCandidates are the forms that occur in the wild.
func cameraCandidates(p Printer) []candidate {
	auth := fmt.Sprintf("bblp:%s@%s:322", p.Code, p.IP)
	var out []candidate
	for _, schema := range []string{"rtsps", "rtspx"} {
		for _, pfad := range []string{"/streaming/live/1", "/streaming/live/0"} {
			out = append(out, candidate{
				Schema:  schema,
				Pfad:    pfad,
				Adresse: schema + "://" + auth + pfad,
			})
		}
	}
	return out
}

type candidateResult struct {
	Schema string `json:"schema"`
	Pfad   string `json:"pfad"`
	OK     bool   `json:"ok"`
	Bytes  int    `json:"bytes,omitempty"`
	Grund  string `json:"grund,omitempty"`
	MS     int64  `json:"ms"`
}

// probiereAdresse asks go2rtc whether it gets an image at this address.
func probiereAdresse(adresse string, timeout time.Duration) candidateResult {
	res := candidateResult{}
	start := time.Now()
	u := fmt.Sprintf("http://127.0.0.1:%d/api/frame.jpeg?src=%s", go2rtcPort, url.QueryEscape(adresse))
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(u)
	res.MS = time.Since(start).Milliseconds()
	if err != nil {
		res.Grund = "go2rtc antwortet nicht: " + err.Error()
		return res
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		res.Grund = fmt.Sprintf("go2rtc HTTP %d", resp.StatusCode)
		return res
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	res.Bytes = len(data)
	res.MS = time.Since(start).Milliseconds()
	// go2rtc answers with 200 even when it got no image —
	// so the content is checked, not the status code.
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		res.Grund = "kein Bild (leere Antwort)"
		return res
	}
	res.OK = true
	return res
}

func handleCameraTest(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "ip fehlt", http.StatusBadRequest)
		return
	}
	mu.Lock()
	var p *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == ip {
			p = &state.Printers[i]
			break
		}
	}
	var kopie Printer
	if p != nil {
		kopie = *p
	}
	mu.Unlock()
	if p == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}

	// Reachability first — without an open port the rest is pointless and
	// only takes unnecessarily long.
	hafen := probePort(kopie.IP, 322, 3*time.Second)
	if !hafen.Offen {
		writeJSON(w, map[string]any{
			"ip": ip, "name": kopie.Name, "port322": false,
			"urteil": "Port 322 ist zu (" + hafen.Grund + ") — an der Adressform liegt es nicht. " +
				"LAN Mode Liveview am Gerät prüfen.",
			"ergebnisse": []candidateResult{},
		})
		return
	}

	var ergebnisse []candidateResult
	var sieger *candidateResult
	for _, k := range cameraCandidates(kopie) {
		e := probiereAdresse(k.Adresse, 12*time.Second)
		e.Schema, e.Pfad = k.Schema, k.Pfad
		ergebnisse = append(ergebnisse, e)
		if e.OK && sieger == nil {
			kopieE := e
			sieger = &kopieE
			break // the first form that returns an image is enough
		}
	}

	urteil := "Keine der geprüften Adressformen liefert ein Bild. Port 322 ist offen, " +
		"die Kamera antwortet aber nicht — meist ist LAN Mode Liveview am Gerät aus."
	if sieger != nil {
		urteil = "Funktioniert: " + sieger.Schema + "://…" + sieger.Pfad
		// Remember it so the configuration tries this form first next time.
		mu.Lock()
		if state.KameraSchema == nil {
			state.KameraSchema = map[string]string{}
		}
		state.KameraSchema[ip] = sieger.Schema + "|" + sieger.Pfad
		mu.Unlock()
		saveState()
		writeGo2rtcYaml()
	}

	writeJSON(w, map[string]any{
		"ip": ip, "name": kopie.Name, "port322": true,
		"ergebnisse": ergebnisse, "urteil": urteil,
		"gemerkt": sieger != nil,
	})
}

// schemaAus reads the measured form from an already-copied map. This
// function deliberately locks NOTHING — it is called from buildYaml, which
// already holds the lock.
func schemaAus(gemessen map[string]string, ip string) (schema, pfad string, ok bool) {
	v := gemessen[ip]
	if v == "" {
		return "", "", false
	}
	teile := strings.SplitN(v, "|", 2)
	if len(teile) != 2 {
		return "", "", false
	}
	return teile[0], teile[1], true
}

var _ = json.Marshal
