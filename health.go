package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// ─── HEALTH / DIAGNOSTICS ─────────────────────────────────────────────────────
//
// /api/health returns a single compact snapshot of how the whole farm and the
// tool itself are doing: how many printers are configured, online and reporting
// errors; whether go2rtc is running and with how many processes; the tool's
// uptime, version and build; and whether a UI window is currently connected.
// It is meant for a small diagnostics panel in Settings and for quick checks
// ("is everything healthy right now?") without opening the modal for each
// printer.

// healthPrinterErr is one printer that is online but reporting error messages.
type healthPrinterErr struct {
	IP       string   `json:"ip"`
	Name     string   `json:"name"`
	Messages []string `json:"messages"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	statuses := mqttMgr.AllStatuses()
	mu.RLock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	camCount := len(state.Cameras)
	mu.RUnlock()

	online := 0
	connected := 0
	var errs []healthPrinterErr
	for _, p := range printers {
		if mqttMgr.IsConnected(p.IP) {
			connected++
		}
		s, ok := statuses[p.IP]
		if !ok || !s.Online {
			continue
		}
		online++
		var msgs []string
		if s.PrintError != 0 {
			msgs = append(msgs, PrintErrorText(s.PrintError))
		}
		for _, e := range s.HmsErrors {
			msgs = append(msgs, HMSErrorText(e))
		}
		if len(msgs) > 0 {
			errs = append(errs, healthPrinterErr{IP: p.IP, Name: p.Name, Messages: msgs})
		}
	}

	uptime := int64(time.Since(readySince).Seconds())

	resp := map[string]any{
		"version":         appVersion,
		"build":           buildID,
		"uptime_seconds":  uptime,
		"printers_total":  len(printers),
		"printers_online": online,
		"mqtt_connected":  connected,
		"cameras":         camCount,
		"go2rtc_running":  checkGo2rtcRunning(),
		"go2rtc_count":    countGo2rtc(),
		"go2rtc_pids":     go2rtcPIDs(),
		"ui_connected":    uiConns.Load() > 0,
		"errors":          errs,
		"error_count":     len(errs),
	}
	json.NewEncoder(w).Encode(resp)
}
