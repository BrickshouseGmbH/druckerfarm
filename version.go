package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ─── GERAETEVERSIONEN ─────────────────────────────────────────────────────────
//
// Der Drucker verraet seine Firmware nur auf Nachfrage: info.get_version liefert
// eine Liste von Baugruppen mit Namen, Hardware- und Softwarestand. Darin steckt
// die Druckerfirmware (Baugruppe "ota") und, sofern angeschlossen, je ein
// Eintrag fuer die AMS-Einheiten.
//
// Zu den Betriebsstunden, ehrlich: die stehen in der oertlichen Schnittstelle
// NICHT. Weder get_version noch die Statusmeldung enthalten einen Zaehler. Was
// hier angezeigt wird, zaehlt dieses Programm deshalb selbst mit — ab dem Tag,
// an dem der Drucker eingetragen wurde. Das ist als solches gekennzeichnet und
// nicht mit dem Zaehler im Geraet zu verwechseln.

type ModulVersion struct {
	Name string `json:"name"`
	HW   string `json:"hw_ver,omitempty"`
	SW   string `json:"sw_ver,omitempty"`
	SN   string `json:"sn,omitempty"`
}

type GeraeteInfo struct {
	Firmware  string         `json:"firmware,omitempty"` // Baugruppe "ota"
	AMS       []ModulVersion `json:"ams,omitempty"`      // je AMS eine Zeile
	Module    []ModulVersion `json:"module,omitempty"`   // alles, fuer die Diagnose
	Abgefragt time.Time      `json:"abgefragt,omitempty"`
}

// parseVersionReport liest die Antwort auf get_version.
func parseVersionReport(payload []byte) (*GeraeteInfo, bool) {
	var w struct {
		Info struct {
			Command string         `json:"command"`
			Module  []ModulVersion `json:"module"`
		} `json:"info"`
	}
	if err := json.Unmarshal(payload, &w); err != nil {
		return nil, false
	}
	if w.Info.Command != "get_version" || len(w.Info.Module) == 0 {
		return nil, false
	}

	info := &GeraeteInfo{Abgefragt: time.Now(), Module: w.Info.Module}
	for _, m := range w.Info.Module {
		name := strings.ToLower(m.Name)
		switch {
		case name == "ota":
			info.Firmware = m.SW
		case strings.HasPrefix(name, "ams"):
			info.AMS = append(info.AMS, m)
		}
	}
	// Nach Namen sortieren, damit AMS 1 vor AMS 2 steht.
	sort.Slice(info.AMS, func(a, b int) bool { return info.AMS[a].Name < info.AMS[b].Name })
	return info, true
}

// amsBezeichnung macht aus "ams/0" ein "AMS 1".
func amsBezeichnung(modulName string) string {
	n := strings.ToLower(strings.TrimSpace(modulName))
	n = strings.TrimPrefix(n, "ams")
	n = strings.TrimPrefix(n, "/")
	n = strings.TrimPrefix(n, "_")
	if n == "" {
		return "AMS"
	}
	var zahl int
	if _, err := fmt.Sscanf(n, "%d", &zahl); err == nil {
		return fmt.Sprintf("AMS %d", zahl+1)
	}
	return "AMS " + strings.ToUpper(n)
}

// RequestVersion fragt die Baugruppenliste an.
func (m *MQTTManager) RequestVersion(p Printer) bool {
	if p.Serial == "" {
		return false
	}
	m.mu.RLock()
	client, ok := m.clients[p.IP]
	m.mu.RUnlock()
	if !ok || !client.IsConnected() {
		return false
	}
	payload := `{"info":{"sequence_id":"` + naechsteSeq() + `","command":"get_version"}}`
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	return tok.WaitTimeout(3*time.Second) && tok.Error() == nil
}

// ─── BETRIEBSZEIT ─────────────────────────────────────────────────────────────
//
// Selbst mitgezaehlt, weil das Geraet den eigenen Zaehler nicht herausgibt.
// Erhoeht wird nur, solange wirklich gedruckt wird.

const laufzeitTakt = 60 * time.Second

func laufzeitLoop() {
	for {
		time.Sleep(laufzeitTakt)
		if netzPausiert() {
			continue
		}
		mu.Lock()
		if state.Laufzeit == nil {
			state.Laufzeit = map[string]int64{}
		}
		printers := make([]Printer, len(state.Printers))
		copy(printers, state.Printers)
		mu.Unlock()

		geaendert := false
		for _, p := range printers {
			s := mqttMgr.GetStatus(p.IP)
			if s == nil || !s.Online {
				continue
			}
			// Betriebsstunden = eingeschaltet und erreichbar. Frueher wurde nur
			// die reine Druckzeit gezaehlt; gemeint sind aber die Stunden, die
			// das Geraet ueberhaupt laeuft.
			mu.Lock()
			state.Laufzeit[p.IP] += int64(laufzeitTakt / time.Second)
			mu.Unlock()
			geaendert = true
		}
		if geaendert {
			saveState()
		}
	}
}

// laufzeitText macht aus Sekunden eine lesbare Angabe.
func laufzeitText(sek int64) string {
	if sek <= 0 {
		return ""
	}
	std := sek / 3600
	if std < 1 {
		return fmt.Sprintf("%d min", sek/60)
	}
	if std < 100 {
		return fmt.Sprintf("%d h %d min", std, (sek%3600)/60)
	}
	return fmt.Sprintf("%d h", std)
}
