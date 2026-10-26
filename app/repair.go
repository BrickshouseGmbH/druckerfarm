package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── REPARATUR-FLAG AUF DER DRUCKER-SD ────────────────────────────────────────
//
// A printer can be marked "in repair". So all PCs see the same
// marker, it is stored as "maintenance.json" on the SD card of the
// respective printer. It is also cached locally (config),
// so the display is immediate and survives an offline printer.

const maintenanceFile = "maintenance.json"

type RepairFlag struct {
	InRepair bool      `json:"in_repair"`
	Note     string    `json:"note,omitempty"`
	By       string    `json:"by,omitempty"`
	Since    time.Time `json:"since"`
}

func maintenancePfad() string {
	base := strings.TrimRight(strings.TrimSpace(syncCfg.SDPath), "/")
	if base == "" {
		return "/" + maintenanceFile
	}
	return base + "/" + maintenanceFile
}

func repairCache(ip string) (RepairFlag, bool) {
	mu.Lock()
	defer mu.Unlock()
	if state.Reparatur == nil {
		return RepairFlag{}, false
	}
	f, ok := state.Reparatur[ip]
	return f, ok
}

func setRepairCache(ip string, f RepairFlag) {
	mu.Lock()
	if state.Reparatur == nil {
		state.Reparatur = map[string]RepairFlag{}
	}
	state.Reparatur[ip] = f
	mu.Unlock()
	saveState()
}

// writeRepairSD stores the marker on the printer's SD card.
func writeRepairSD(p Printer, f RepairFlag) error {
	fc := &printerFTP{ip: p.IP, code: p.Code}
	data, _ := json.MarshalIndent(f, "", "  ")
	return fc.stor(maintenancePfad(), bytes.NewReader(data))
}

// readRepairSD reads the marker from the SD (if missing, counts as "not in
// Reparatur").
func readRepairSD(p Printer) (RepairFlag, bool) {
	fc := &printerFTP{ip: p.IP, code: p.Code}
	rc, err := fc.retr(maintenancePfad())
	if err != nil {
		return RepairFlag{}, false
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	var f RepairFlag
	if json.Unmarshal(b, &f) != nil {
		return RepairFlag{}, false
	}
	return f, true
}

// syncRepair reconciles local state and SD (only for a reachable printer).
// The newer marker (by timestamp) wins.
func syncRepair(p Printer) {
	if s := mqttMgr.GetStatus(p.IP); s == nil || !s.Online {
		return
	}
	sd, sdDa := readRepairSD(p)
	lokal, lokalDa := repairCache(p.IP)

	switch {
	case sdDa && (!lokalDa || sd.Since.After(lokal.Since)):
		setRepairCache(p.IP, sd) // anderer PC war neuer -> uebernehmen
	case lokalDa && (!sdDa || lokal.Since.After(sd.Since)):
		writeRepairSD(p, lokal) // our state is newer -> write to SD
	}
}

// repairLoop reads the markers of reachable printers at intervals,
// so changes from other PCs arrive. Deliberately slow/staggered.
func repairLoop() {
	time.Sleep(15 * time.Second)
	for {
		mu.Lock()
		printers := append([]Printer(nil), state.Printers...)
		mu.Unlock()
		for _, p := range printers {
			if netPaused() {
				break
			}
			if strings.TrimSpace(p.Serial) == "" || strings.TrimSpace(p.Code) == "" {
				continue
			}
			syncRepair(p)
			time.Sleep(4 * time.Second) // staffeln, um FTP-Last gering zu halten
		}
		time.Sleep(3 * time.Minute)
	}
}

func handleRepair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP       string `json:"ip"`
		InRepair bool   `json:"in_repair"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	var ziel *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == body.IP {
			ziel = &state.Printers[i]
			break
		}
	}
	mu.Unlock()
	if ziel == nil {
		http.Error(w, "Drucker nicht gefunden", http.StatusNotFound)
		return
	}

	flag := RepairFlag{InRepair: body.InRepair, Note: strings.TrimSpace(body.Note), By: eigenerName, Since: time.Now()}
	setRepairCache(body.IP, flag) // locally at once so the display is correct

	// Write to the SD when reachable. If the printer is offline (typical during
	// repair), it stays local and is written to the SD at the next online sync.
	// SD nachgezogen.
	sdOK := true
	if s := mqttMgr.GetStatus(body.IP); s != nil && s.Online {
		if err := writeRepairSD(*ziel, flag); err != nil {
			sdOK = false
			log.Printf("Reparatur-Flag %s: SD-Schreiben fehlgeschlagen: %v", body.IP, err)
		}
	} else {
		sdOK = false
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "auf_sd": sdOK, "flag": flag})
}
