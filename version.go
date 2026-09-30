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
// The printer reveals its firmware only on request: info.get_version returns
// a list of modules with name, hardware and software level. It contains
// the printer firmware (module "ota") and, if connected, one
// entry per AMS unit.
//
// About the operating hours, honestly: they are NOT in the local interface.
// Neither get_version nor the status message contains a counter. What
// is shown here is therefore counted by this program itself — from the day
// the printer was added. This is marked as such and
// must not be confused with the counter in the device.

type ModuleVersion struct {
	Name string `json:"name"`
	HW   string `json:"hw_ver,omitempty"`
	SW   string `json:"sw_ver,omitempty"`
	SN   string `json:"sn,omitempty"`
}

type DeviceInfo struct {
	Firmware  string          `json:"firmware,omitempty"` // Baugruppe "ota"
	AMS       []ModuleVersion `json:"ams,omitempty"`      // one line per AMS
	Module    []ModuleVersion `json:"module,omitempty"`   // everything, for diagnostics
	Abgefragt time.Time       `json:"abgefragt,omitempty"`
}

// parseVersionReport reads the response to get_version.
func parseVersionReport(payload []byte) (*DeviceInfo, bool) {
	var w struct {
		Info struct {
			Command string          `json:"command"`
			Module  []ModuleVersion `json:"module"`
		} `json:"info"`
	}
	if err := json.Unmarshal(payload, &w); err != nil {
		return nil, false
	}
	if w.Info.Command != "get_version" || len(w.Info.Module) == 0 {
		return nil, false
	}

	info := &DeviceInfo{Abgefragt: time.Now(), Module: w.Info.Module}
	for _, m := range w.Info.Module {
		name := strings.ToLower(m.Name)
		switch {
		case name == "ota":
			info.Firmware = m.SW
		case strings.HasPrefix(name, "ams"):
			info.AMS = append(info.AMS, m)
		}
	}
	// Sort by name so AMS 1 comes before AMS 2.
	sort.Slice(info.AMS, func(a, b int) bool { return info.AMS[a].Name < info.AMS[b].Name })
	return info, true
}

// amsBezeichnung turns "ams/0" into "AMS 1".
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

// RequestVersion requests the module list.
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
	payload := `{"info":{"sequence_id":"` + nextSeq() + `","command":"get_version"}}`
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	return tok.WaitTimeout(3*time.Second) && tok.Error() == nil
}

// ─── BETRIEBSZEIT ─────────────────────────────────────────────────────────────
//
// Counted by us, because the device does not expose its own counter.
// Incremented only while actually printing.

const runtimeTick = 60 * time.Second

func runtimeLoop() {
	for {
		time.Sleep(runtimeTick)
		if netPaused() {
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
			// Operating hours = powered on and reachable. Previously only
			// the pure print time was counted; but what is meant are the hours the
			// device is running at all.
			mu.Lock()
			state.Laufzeit[p.IP] += int64(runtimeTick / time.Second)
			mu.Unlock()
			geaendert = true
		}
		if geaendert {
			saveState()
		}
	}
}

// runtimeText turns seconds into a readable value.
func runtimeText(sek int64) string {
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
