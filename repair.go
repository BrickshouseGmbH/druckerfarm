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
// Ein Drucker kann als "in Reparatur" markiert werden. Damit alle PCs dieselbe
// Markierung sehen, liegt sie als Datei "maintenance.json" auf der SD-Karte des
// jeweiligen Druckers. Zusaetzlich wird sie lokal zwischengespeichert (config),
// damit die Anzeige sofort da ist und einen Offline-Drucker ueberdauert.

const maintenanceDatei = "maintenance.json"

type RepairFlag struct {
	InRepair bool      `json:"in_repair"`
	Note     string    `json:"note,omitempty"`
	By       string    `json:"by,omitempty"`
	Since    time.Time `json:"since"`
}

func maintenancePfad() string {
	base := strings.TrimRight(strings.TrimSpace(syncCfg.SDPath), "/")
	if base == "" {
		return "/" + maintenanceDatei
	}
	return base + "/" + maintenanceDatei
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

// schreibeReparaturSD legt die Markierung auf der SD des Druckers ab.
func schreibeReparaturSD(p Printer, f RepairFlag) error {
	fc := &printerFTP{ip: p.IP, code: p.Code}
	data, _ := json.MarshalIndent(f, "", "  ")
	return fc.stor(maintenancePfad(), bytes.NewReader(data))
}

// leseReparaturSD holt die Markierung von der SD (fehlt sie, gilt „nicht in
// Reparatur").
func leseReparaturSD(p Printer) (RepairFlag, bool) {
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

// syncReparatur gleicht lokalen Stand und SD ab (nur bei erreichbarem Drucker).
// Neuere Markierung (nach Zeitstempel) gewinnt.
func syncReparatur(p Printer) {
	if s := mqttMgr.GetStatus(p.IP); s == nil || !s.Online {
		return
	}
	sd, sdDa := leseReparaturSD(p)
	lokal, lokalDa := repairCache(p.IP)

	switch {
	case sdDa && (!lokalDa || sd.Since.After(lokal.Since)):
		setRepairCache(p.IP, sd) // anderer PC war neuer -> uebernehmen
	case lokalDa && (!sdDa || lokal.Since.After(sd.Since)):
		schreibeReparaturSD(p, lokal) // eigener Stand neuer -> auf SD schreiben
	}
}

// reparaturLoop liest die Markierungen der erreichbaren Drucker in Abstaenden,
// damit Aenderungen von anderen PCs ankommen. Bewusst langsam/gestaffelt.
func reparaturLoop() {
	time.Sleep(15 * time.Second)
	for {
		mu.Lock()
		printers := append([]Printer(nil), state.Printers...)
		mu.Unlock()
		for _, p := range printers {
			if netzPausiert() {
				break
			}
			if strings.TrimSpace(p.Serial) == "" || strings.TrimSpace(p.Code) == "" {
				continue
			}
			syncReparatur(p)
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
	setRepairCache(body.IP, flag) // lokal sofort, damit die Anzeige stimmt

	// Auf die SD schreiben, wenn erreichbar. Ist der Drucker offline (typisch bei
	// Reparatur), bleibt es lokal und wird beim naechsten Online-Abgleich auf die
	// SD nachgezogen.
	sdOK := true
	if s := mqttMgr.GetStatus(body.IP); s != nil && s.Online {
		if err := schreibeReparaturSD(*ziel, flag); err != nil {
			sdOK = false
			log.Printf("Reparatur-Flag %s: SD-Schreiben fehlgeschlagen: %v", body.IP, err)
		}
	} else {
		sdOK = false
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "auf_sd": sdOK, "flag": flag})
}
