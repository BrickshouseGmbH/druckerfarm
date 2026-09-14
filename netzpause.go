package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// ─── NETZWERKVERKEHR ANHALTEN ─────────────────────────────────────────────────
//
// Es gibt Lagen, in denen das Programm im Netz schlicht still sein soll: waehrend
// einer Wartung, bei einer Messung, oder wenn jemand anderes die Geraete
// bedient. "Pausieren" heisst hier woertlich — nicht nur die Oberflaeche hoert
// auf zu fragen, sondern auch der Server:
//
//	– keine Statusabfragen und keine Bilder mehr
//	– die MQTT-Verbindungen werden getrennt
//	– go2rtc wird beendet, damit keine Kameraverbindungen offen bleiben
//	– die Schleifen im Hintergrund setzen aus
//
// Absichtlich NICHT gespeichert: nach einem Neustart laeuft das Programm wieder.
// Ein Schalter, der ueber Nacht stehen bleibt und am naechsten Morgen alles tot
// aussehen laesst, waere eine Falle.

var netzPause atomic.Bool

func netzPausiert() bool { return netzPause.Load() }

// setzeNetzPause haelt an oder laesst wieder los. Der Rueckgabewert sagt, was
// dabei tatsaechlich passiert ist — die Oberflaeche zeigt das an.
func setzeNetzPause(an bool) map[string]any {
	if netzPause.Load() == an {
		return map[string]any{"pausiert": an, "geaendert": false}
	}
	netzPause.Store(an)

	if an {
		log.Printf("⏸  Netzwerkverkehr angehalten — MQTT getrennt, go2rtc beendet")
		mu.Lock()
		printers := make([]Printer, len(state.Printers))
		copy(printers, state.Printers)
		mu.Unlock()
		for _, p := range printers {
			mqttMgr.Disconnect(p.IP)
		}
		stopGo2rtc()
		return map[string]any{"pausiert": true, "geaendert": true,
			"getrennt": len(printers), "go2rtc": "beendet"}
	}

	log.Printf("▶  Netzwerkverkehr wieder aufgenommen")
	// Erst go2rtc, dann die Verbindungen — sonst stehen Kacheln da, fuer die es
	// noch keinen Stream gibt.
	go func() {
		startGo2rtc()
		time.Sleep(500 * time.Millisecond)
		connectAllMQTT()
	}()
	return map[string]any{"pausiert": false, "geaendert": true}
}

// stateAlsText dient nur dem Test: er stellt sicher, dass der Pausezustand
// nicht versehentlich in der Konfiguration landet.
func stateAlsText() string {
	b, _ := json.Marshal(state)
	return string(b)
}

func handleNetzPause(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]any{"pausiert": netzPausiert()})
	case http.MethodPost:
		var body struct {
			Pausiert bool `json:"pausiert"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, setzeNetzPause(body.Pausiert))
	default:
		http.Error(w, "GET oder POST erwartet", http.StatusMethodNotAllowed)
	}
}
