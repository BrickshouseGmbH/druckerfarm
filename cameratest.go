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
// Welche Adressform ein Drucker versteht, laesst sich nicht zuverlaessig aus
// dem Modellnamen ableiten — die Firmware entscheidet, und die aendert sich.
// Statt weiter zu raten, probiert dieser Test die Formen an dem Geraet selbst
// durch und sagt, welche ein Bild liefert.
//
// go2rtc kann eine Adresse direkt pruefen: /api/frame.jpeg?src=<adresse> baut
// die Verbindung auf, holt ein Einzelbild und wirft sie wieder weg. Es braucht
// dafuer keinen Eintrag in der Konfiguration.

type kandidat struct {
	Schema  string `json:"schema"`
	Adresse string `json:"-"` // enthaelt den Zugangscode, gehoert nicht ins JSON
	Pfad    string `json:"pfad"`
}

// kameraKandidaten sind die Formen, die in freier Wildbahn vorkommen.
func kameraKandidaten(p Printer) []kandidat {
	auth := fmt.Sprintf("bblp:%s@%s:322", p.Code, p.IP)
	var out []kandidat
	for _, schema := range []string{"rtsps", "rtspx"} {
		for _, pfad := range []string{"/streaming/live/1", "/streaming/live/0"} {
			out = append(out, kandidat{
				Schema:  schema,
				Pfad:    pfad,
				Adresse: schema + "://" + auth + pfad,
			})
		}
	}
	return out
}

type kandidatErgebnis struct {
	Schema string `json:"schema"`
	Pfad   string `json:"pfad"`
	OK     bool   `json:"ok"`
	Bytes  int    `json:"bytes,omitempty"`
	Grund  string `json:"grund,omitempty"`
	MS     int64  `json:"ms"`
}

// probiereAdresse fragt go2rtc, ob es unter dieser Adresse ein Bild bekommt.
func probiereAdresse(adresse string, timeout time.Duration) kandidatErgebnis {
	res := kandidatErgebnis{}
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
	// go2rtc antwortet auch dann mit 200, wenn es kein Bild bekommen hat —
	// deshalb wird der Inhalt geprueft, nicht der Statuscode.
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

	// Erst die Erreichbarkeit — ohne offenen Port ist der Rest sinnlos und
	// dauert nur unnoetig lange.
	hafen := probePort(kopie.IP, 322, 3*time.Second)
	if !hafen.Offen {
		writeJSON(w, map[string]any{
			"ip": ip, "name": kopie.Name, "port322": false,
			"urteil": "Port 322 ist zu (" + hafen.Grund + ") — an der Adressform liegt es nicht. " +
				"LAN Mode Liveview am Gerät prüfen.",
			"ergebnisse": []kandidatErgebnis{},
		})
		return
	}

	var ergebnisse []kandidatErgebnis
	var sieger *kandidatErgebnis
	for _, k := range kameraKandidaten(kopie) {
		e := probiereAdresse(k.Adresse, 12*time.Second)
		e.Schema, e.Pfad = k.Schema, k.Pfad
		ergebnisse = append(ergebnisse, e)
		if e.OK && sieger == nil {
			kopieE := e
			sieger = &kopieE
			break // die erste Form, die ein Bild liefert, genuegt
		}
	}

	urteil := "Keine der geprüften Adressformen liefert ein Bild. Port 322 ist offen, " +
		"die Kamera antwortet aber nicht — meist ist LAN Mode Liveview am Gerät aus."
	if sieger != nil {
		urteil = "Funktioniert: " + sieger.Schema + "://…" + sieger.Pfad
		// Merken, damit die Konfiguration diese Form kuenftig zuerst nimmt.
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

// schemaAus liest die gemessene Form aus einer bereits kopierten Karte. Diese
// Funktion sperrt bewusst NICHTS — sie wird aus buildYaml gerufen, das die
// Sperre schon haelt.
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
