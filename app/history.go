package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── FEHLER-VERLAUF JE DRUCKER ────────────────────────────────────────────────
//
// Keeps a short per-printer log of error events — the moment a print goes into
// FAILED, a print_error appears, or new HMS codes show up. Each entry carries a
// timestamp, the state, the print_error code, the HMS codes and the file, so a
// printer's recent trouble can be looked up later (and helps diagnose the
// resume/cancel and upload-start issues without standing next to the machine).

type HistoryEntry struct {
	Time  int64    `json:"time"`  // unix seconds
	State string   `json:"state"` // gcode_state (e.g. FAILED)
	Error int      `json:"error"` // print_error code (0 = none)
	HMS   []string `json:"hms"`   // HMS error codes
	File  string   `json:"file"`  // print file at the time
}

var (
	historyMu   sync.Mutex
	historyLog  = map[string][]HistoryEntry{} // key: printer IP
	historyLast = map[string]string{}         // last recorded signature per IP
)

const historyMax = 80

func historyFile() string { return filepath.Join(appDir, "history.json") }

// errSignature builds a signature of the current error situation so the same
// ongoing error is not logged on every status tick.
func errSignature(s *PrinterStatus) string {
	st := strings.ToUpper(s.GcodeState)
	hms := strings.Join(s.HmsErrors, ",")
	return st + "|" + itoa(s.PrintError) + "|" + hms
}

// recordHistory appends an entry when a NEW error situation is seen. Called from
// the MQTT handler; uses its own lock (never holds historyMu while taking others).
func recordHistory(ip, file string, s *PrinterStatus) {
	st := strings.ToUpper(s.GcodeState)
	isError := st == "FAILED" || s.PrintError > 0 || len(s.HmsErrors) > 0
	if !isError {
		return
	}
	sig := errSignature(s)

	historyMu.Lock()
	if historyLast[ip] == sig {
		historyMu.Unlock()
		return // same error still standing — do not log again
	}
	historyLast[ip] = sig
	hms := append([]string(nil), s.HmsErrors...)
	entry := HistoryEntry{Time: time.Now().Unix(), State: st, Error: s.PrintError, HMS: hms, File: file}
	list := append(historyLog[ip], entry)
	if len(list) > historyMax {
		list = list[len(list)-historyMax:]
	}
	historyLog[ip] = list
	historyMu.Unlock()

	go saveHistory()
}

func saveHistory() {
	historyMu.Lock()
	data, err := json.Marshal(historyLog)
	historyMu.Unlock()
	if err != nil {
		return
	}
	_ = atomicWrite(historyFile(), data, 0o644)
}

func loadHistory() {
	data, err := os.ReadFile(historyFile())
	if err != nil {
		return
	}
	m := map[string][]HistoryEntry{}
	if json.Unmarshal(data, &m) == nil {
		historyMu.Lock()
		historyLog = m
		historyMu.Unlock()
	}
}

// handleHistory returns one printer's error history (newest first).
func handleHistory(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	if ip == "" {
		http.Error(w, "ip noetig", http.StatusBadRequest)
		return
	}
	historyMu.Lock()
	src := historyLog[ip]
	out := make([]HistoryEntry, len(src))
	for i, e := range src {
		out[len(src)-1-i] = e // reverse → newest first
	}
	historyMu.Unlock()
	writeJSON(w, map[string]any{"entries": out})
}
