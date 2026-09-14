package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── DRUCKER-FIRMWARE-UPDATES ─────────────────────────────────────────────────
//
// Woher der Drucker weiss, dass es ein Update gibt: "Nur LAN" trennt die
// Cloud-Bindung (Konto, Fernsteuerung), nicht das Netzwerk. Der Drucker prueft
// weiterhin ueber einen eigenen Kanal beim Hersteller, ob neue Firmware
// vorliegt, und legt das Ergebnis in seinen Statusbericht (upgrade_state /
// new_ver_list). Dieses Programm liest das nur mit — es loest KEIN Update aus.
//
// Gefundene Updates werden lokal gemerkt und bleiben als Marke hinter dem
// Druckernamen stehen, bis der Drucker die neue Fassung tatsaechlich
// installiert hat. Das Aufraeumen geschieht beim naechsten Statusabruf: meldet
// der Drucker die Baugruppe nicht mehr oder laeuft sie bereits auf der neuen
// Nummer, verschwindet die Marke.

// UpgradeMeldung ist ein einzelnes offenes Update, wie es aus dem Status kommt.
type UpgradeMeldung struct {
	Modul   string `json:"modul"`   // "ota", "ams/0" …
	Aktuell string `json:"aktuell"` // laufende Fassung
	Neu     string `json:"neu"`     // verfuegbare Fassung
}

// FwModulUpdate ist ein gemerktes Update samt Fundzeitpunkt.
type FwModulUpdate struct {
	Modul    string    `json:"modul"`
	Aktuell  string    `json:"aktuell"`
	Neu      string    `json:"neu"`
	Gefunden time.Time `json:"gefunden"`
}

// parseUpgradeState liest den upgrade_state-Block aus dem pushall.
func parseUpgradeState(v interface{}) []UpgradeMeldung {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	var out []UpgradeMeldung

	// Weg 1 — die ausfuehrliche Liste (neuere Firmware, P1/A1/H2 …). Sie nennt
	// je Baugruppe die laufende und die neue Nummer.
	if lst, ok := m["new_ver_list"].([]interface{}); ok {
		for _, e := range lst {
			em, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := em["name"].(string)
			cur, _ := em["cur_ver"].(string)
			neu, _ := em["new_ver"].(string)
			name = strings.TrimSpace(name)
			neu = strings.TrimSpace(neu)
			if name == "" || neu == "" {
				continue
			}
			// Nur echte Spruenge nach oben. Manche Firmware listet die
			// laufende Fassung auch dann, wenn nichts offen ist.
			if cur != "" && !newerVersion(cur, neu) {
				continue
			}
			out = append(out, UpgradeMeldung{Modul: name, Aktuell: cur, Neu: neu})
		}
	}
	if len(out) > 0 {
		return out
	}

	// Weg 2 — die X1-Reihe (auch X1E) kennt keine new_ver_list, sondern legt je
	// Baugruppe ein Einzelfeld an: ota_new_version_number, ams_new_version_number,
	// ahb_new_version_number, ext_new_version_number.
	//
	// Wichtig: new_version_state taugt NICHT als Schalter. Ein echter X1E meldet
	// new_version_state=1 UND trotzdem ota_new_version_number="01.03.00.00" — da
	// liegt ein Update vor. Maszgeblich ist allein: ist das Nummernfeld gefuellt
	// (und kein Platzhalter). Die laufende Fassung steht hier nicht dabei;
	// verglichen wird spaeter gegen get_version (nochOffen), damit bereits
	// Installiertes wieder verschwindet.
	feld := func(key, modul string) {
		val, _ := m[key].(string)
		val = strings.TrimSpace(val)
		if val == "" || istNullVersion(val) {
			return
		}
		out = append(out, UpgradeMeldung{Modul: modul, Neu: val})
	}
	feld("ota_new_version_number", "ota")
	feld("ams_new_version_number", "ams")
	feld("ahb_new_version_number", "ahb")
	feld("ext_new_version_number", "ext")
	return out
}

// istNullVersion erkennt Platzhalter wie "00.00.00.00".
func istNullVersion(v string) bool {
	for _, r := range v {
		if r != '0' && r != '.' {
			return false
		}
	}
	return true
}

// aktuelleVersionen liefert die laufenden Baugruppen-Staende aus get_version.
func aktuelleVersionen(st *PrinterStatus) map[string]string {
	m := map[string]string{}
	if st == nil || st.Info == nil {
		return m
	}
	if st.Info.Firmware != "" {
		m["ota"] = st.Info.Firmware
	}
	for _, a := range st.Info.AMS {
		m[strings.ToLower(a.Name)] = a.SW
		// Fuer die X1-Meldung "ams_new_version_number" ohne Index: der
		// niedrigste AMS-Stand entscheidet. Liegt irgendein AMS unter der
		// neuen Nummer, gilt das Update als offen.
		if cur, ok := m["ams"]; !ok || (a.SW != "" && newerVersion(a.SW, cur)) {
			m["ams"] = a.SW
		}
	}
	return m
}

// nochOffen streicht aus einer Update-Liste alles, was auf dem Drucker bereits
// installiert ist. Ist der Status unbekannt (offline), bleibt die Liste stehen.
func nochOffen(list []FwModulUpdate, st *PrinterStatus) []FwModulUpdate {
	if st == nil {
		return list
	}
	aktuell := aktuelleVersionen(st)
	out := []FwModulUpdate{}
	for _, u := range list {
		cur := aktuell[strings.ToLower(u.Modul)]
		if cur == "" {
			cur = u.Aktuell // notfalls die Fassung vom Fund
		}
		if newerVersion(cur, u.Neu) {
			out = append(out, u)
		}
	}
	return out
}

// fwAnzeige liefert je Drucker die offenen Updates fuer die Anzeige und raeumt
// dabei still installierte weg (persistiert nur, wenn sich etwas geaendert hat).
// Aufgerufen bei jedem Statusabruf — das ist das "beim Laden vergleichen".
func fwAnzeige() map[string][]FwModulUpdate {
	mu.Lock()
	defer mu.Unlock()
	res := map[string][]FwModulUpdate{}
	if state.FwUpdates == nil {
		return res
	}
	dirty := false
	for ip, list := range state.FwUpdates {
		st := mqttMgr.GetStatus(ip)
		offen := nochOffen(list, st)
		if st != nil && st.Online && len(offen) != len(list) {
			// Der Drucker hat (mindestens) eines der Updates installiert.
			if len(offen) == 0 {
				delete(state.FwUpdates, ip)
			} else {
				state.FwUpdates[ip] = offen
			}
			dirty = true
		}
		if len(offen) > 0 {
			res[ip] = offen
		}
	}
	if dirty {
		go saveState()
	}
	return res
}

// handleFwUpdateScan fragt jeden verbundenen Drucker frisch ab (pushall) und
// merkt sich, welche Baugruppe ein Update offen hat. Loest NICHTS aus.
func handleFwUpdateScan(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	angefragt := 0
	for _, p := range printers {
		// pushall bringt upgrade_state, get_version die laufende Firmware —
		// nur mit beiden laesst sich sauber vergleichen, was schon installiert
		// ist. get_version braucht die Seriennummer.
		ok := mqttMgr.RequestStatus(p)
		mqttMgr.RequestVersion(p)
		if ok {
			angefragt++
		}
	}
	// Kurz warten, bis die Antworten eingetroffen sind.
	time.Sleep(2500 * time.Millisecond)

	geprueft, mitUpdate := scanMergeFw(printers)
	saveState()

	// Rohe upgrade_state-Meldungen mitgeben — zur Diagnose, falls ein Modell
	// seine Update-Info anders aufbaut als erwartet.
	roh := map[string]string{}
	mqttMgr.mu.RLock()
	for _, p := range printers {
		if r, ok := mqttMgr.lastUpgrade[p.IP]; ok && r != "" {
			roh[p.IP] = r
		}
	}
	mqttMgr.mu.RUnlock()

	log.Printf("FW-UPDATE-SCAN: %d angefragt, %d geprueft, %d mit Update", angefragt, geprueft, mitUpdate)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"angefragt":  angefragt,
		"geprueft":   geprueft,
		"mit_update": mitUpdate,
		"roh":        roh,
	})
}

// handleFwUpdateStart stoesst am Drucker das Firmware-Update an. WICHTIG: Das
// ist bewusst risikobehaftet und nicht offiziell dokumentiert — es funktioniert
// nur bei aktiviertem Developer Mode und ist Best-Effort. Die Oberflaeche warnt
// vorher deutlich; hier wird nur der Befehl abgesetzt.
func handleFwUpdateStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP string `json:"ip"`
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

	// Bekannteste Form aus der Gemeinschaft: das anstehende Update bestaetigen.
	// Kein Modul/keine URL — der Drucker kennt das verfuegbare Update selbst.
	payload := `{"upgrade":{"sequence_id":"` + naechsteSeq() + `","command":"upgrade_confirm","src_id":1}}`
	if err := mqttMgr.PublishRaw(*ziel, payload); err != nil {
		log.Printf("FW-UPDATE-START %s fehlgeschlagen: %v", body.IP, err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	log.Printf("FW-UPDATE-START an %s gesendet (Best-Effort, Developer Mode noetig)", body.IP)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// scanMergeFw uebernimmt die frischen upgrade_state-Meldungen in den Speicher.
func scanMergeFw(printers []Printer) (geprueft, mitUpdate int) {
	mu.Lock()
	defer mu.Unlock()
	if state.FwUpdates == nil {
		state.FwUpdates = map[string][]FwModulUpdate{}
	}
	for _, p := range printers {
		st := mqttMgr.GetStatus(p.IP)
		if st == nil || !st.Online {
			continue // offline: alten Stand nicht anfassen
		}
		geprueft++
		frisch := []FwModulUpdate{}
		for _, u := range st.Upgrade {
			g := time.Now()
			for _, a := range state.FwUpdates[p.IP] {
				if a.Modul == u.Modul && a.Neu == u.Neu {
					g = a.Gefunden // Fundzeitpunkt bewahren
				}
			}
			frisch = append(frisch, FwModulUpdate{Modul: u.Modul, Aktuell: u.Aktuell, Neu: u.Neu, Gefunden: g})
		}
		frisch = nochOffen(frisch, st)
		if len(frisch) == 0 {
			delete(state.FwUpdates, p.IP)
		} else {
			state.FwUpdates[p.IP] = frisch
			mitUpdate++
		}
	}
	return
}
