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
// There are situations where the program should simply be quiet on the network: during
// maintenance, a measurement, or when someone else operates the
// devices. "Pause" is meant literally here — not only the UI stops
// asking, but the server too:
//
//	– no more status queries and no more images
//	– the MQTT connections are dropped
//	– go2rtc is stopped so no camera connections stay open
//	– the background loops pause
//
// Deliberately NOT persisted: after a restart the program runs again.
// A switch left on overnight that makes everything look dead the next
// morning would be a trap.

var netzPause atomic.Bool

func netPaused() bool { return netzPause.Load() }

// setNetworkPause pauses or resumes. The return value says what
// actually happened — the UI shows it.
func setNetworkPause(an bool) map[string]any {
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
	// go2rtc first, then the connections — otherwise tiles remain for which there
	// is no stream yet.
	go func() {
		startGo2rtc()
		time.Sleep(500 * time.Millisecond)
		connectAllMQTT()
	}()
	return map[string]any{"pausiert": false, "geaendert": true}
}

// stateAlsText is only for the test: it ensures the pause state
// does not accidentally end up in the configuration.
func stateAlsText() string {
	b, _ := json.Marshal(state)
	return string(b)
}

func handleNetworkPause(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]any{"pausiert": netPaused()})
	case http.MethodPost:
		var body struct {
			Pausiert bool `json:"pausiert"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, setNetworkPause(body.Pausiert))
	default:
		http.Error(w, "GET oder POST erwartet", http.StatusMethodNotAllowed)
	}
}
