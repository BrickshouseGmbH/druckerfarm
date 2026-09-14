package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ─── PRINTER STATUS ───────────────────────────────────────────────────────────

type PrinterStatus struct {
	Online     bool      `json:"online"`
	LastSeen   time.Time `json:"last_seen"`
	ConnectErr string    `json:"connect_err,omitempty"`

	GcodeState  string `json:"gcode_state"`
	Progress    int    `json:"progress"`
	Layer       int    `json:"layer"`
	TotalLayers int    `json:"total_layers"`
	RemainTime  int    `json:"remain_time"`
	SubtaskName string `json:"subtask_name"`

	NozzleTemp   float64 `json:"nozzle_temp"`
	NozzleTarget float64 `json:"nozzle_target"`
	BedTemp      float64 `json:"bed_temp"`
	BedTarget    float64 `json:"bed_target"`
	ChamberTemp  float64 `json:"chamber_temp"`

	FanSpeed   int      `json:"fan_speed"`
	PrintError int      `json:"print_error"`
	HmsErrors  []string `json:"hms_errors"`

	// AckedHms sind Stoerungsmeldungen, die beim Anlaufen des Drucks bereits
	// anstanden. Der X1 schickt bei jeder Meldung die vollstaendige HMS-Liste
	// mit — auch dann noch, wenn der Druck laengst weiterlaeuft und die
	// Meldung am Geraet nur nicht quittiert wurde. Ohne dieses Gedaechtnis
	// fuellte sich HmsErrors sofort nach dem Leeren wieder und das Kammerlicht
	// blinkte endlos weiter.
	AckedHms map[string]bool `json:"-"`

	// AMS und Filament
	AMS      []AMSUnit `json:"ams"`
	TrayNow  string    `json:"tray_now"`
	ExtSpool *AMSTray  `json:"ext_spool,omitempty"`

	// Geraeteinfo aus info.get_version — Firmware und AMS-Staende
	Info *GeraeteInfo `json:"info,omitempty"`

	// Offene Firmware-Updates, wie der Drucker sie selbst meldet (upgrade_state /
	// new_ver_list im pushall). Rein lesend — dieses Programm loest nichts aus.
	Upgrade []UpgradeMeldung `json:"upgrade,omitempty"`

	// Debug
	MsgCount  int    `json:"msg_count"`
	LastTopic string `json:"last_topic,omitempty"`
}

// AMSTray ist ein einzelnes Fach. Color kommt als RRGGBBAA vom Drucker.
type AMSTray struct {
	Slot     string   `json:"slot"`
	Type     string   `json:"type"`
	Color    string   `json:"color"`
	Colors   []string `json:"colors,omitempty"` // mehrfarbiges Filament
	SubBrand string   `json:"sub_brand"`
	Remain   int      `json:"remain"`
}

type AMSUnit struct {
	ID       string    `json:"id"`
	Humidity string    `json:"humidity"`
	Temp     string    `json:"temp"`
	Trays    []AMSTray `json:"trays"`
}

// mqttDebug schaltet das Mitschreiben kompletter MQTT-Payloads ein. Standardmaessig
// aus: bei 36 Druckern schreibt das sonst Megabyte pro Minute ins Logfile.
var mqttDebug = os.Getenv("DRUCKERFARM_MQTT_DEBUG") != ""

// ─── MQTT MANAGER ─────────────────────────────────────────────────────────────

type MQTTManager struct {
	mu          sync.RWMutex
	clients     map[string]mqtt.Client
	statuses    map[string]*PrinterStatus
	lastPayload map[string]string // raw payload for debug
	lastUpgrade map[string]string // roher upgrade_state je IP, für Diagnose
}

var mqttMgr = &MQTTManager{
	clients:     make(map[string]mqtt.Client),
	statuses:    make(map[string]*PrinterStatus),
	lastPayload: make(map[string]string),
	lastUpgrade: make(map[string]string),
}

func (m *MQTTManager) Connect(p Printer) {
	if p.Serial == "" {
		return
	}

	m.mu.Lock()
	if c, ok := m.clients[p.IP]; ok && c.IsConnected() {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	log.Printf("MQTT connecting to %s (serial=%s)\n", p.IP, p.Serial)

	opts := mqtt.NewClientOptions()

	// MQTT laeuft ueber TLS auf Port 8883
	opts.AddBroker(fmt.Sprintf("ssl://%s:8883", p.IP))

	// Der Drucker verlangt eine bestimmte Client-ID-Format
	// STABILE Client-ID (ohne Zeitstempel): der Drucker erlaubt nur EINE lokale
	// MQTT-Verbindung. Mit wechselnder ID kann eine neue Verbindung eine
	// hängengebliebene alte nicht übernehmen — der Platz bleibt von einem „Zombie"
	// belegt und der Drucker bleibt offline. Mit fester ID kickt der Broker beim
	// erneuten CONNECT die alte Sitzung automatisch weg (MQTT-Standard), und der
	// Reconnect greift zuverlässig.
	opts.SetClientID("druckerfarm-" + p.Serial)
	opts.SetUsername("bblp")
	opts.SetPassword(p.Code)

	// Der Drucker nutzt ein selbstsigniertes Zertifikat
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	opts.SetConnectTimeout(8 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetCleanSession(true)
	opts.SetOrderMatters(false)

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("MQTT %s lost: %v\n", p.IP, err)
		m.mu.Lock()
		if s, ok := m.statuses[p.IP]; ok {
			s.Online = false
			s.ConnectErr = err.Error()
		}
		m.mu.Unlock()
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("MQTT %s connected, subscribing...\n", p.IP)

		// Topic-Format des Druckers
		topic := fmt.Sprintf("device/%s/report", p.Serial)
		log.Printf("MQTT subscribing to: %s\n", topic)

		token := c.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
			m.handleMessage(p.IP, msg.Payload())
		})
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			log.Printf("MQTT subscribe error %s: %v\n", p.IP, token.Error())
		} else {
			log.Printf("MQTT subscribed OK: %s\n", topic)
		}

		m.mu.Lock()
		if _, ok := m.statuses[p.IP]; !ok {
			m.statuses[p.IP] = &PrinterStatus{}
		}
		m.statuses[p.IP].Online = true
		m.statuses[p.IP].ConnectErr = ""
		m.statuses[p.IP].LastSeen = time.Now()
		m.mu.Unlock()

		// Request full status — pushall-Format des Druckers
		reqTopic := fmt.Sprintf("device/%s/request", p.Serial)
		payload := `{"pushing": {"sequence_id": "0", "command": "pushall", "version": 1, "push_target": 1}}`
		tok := c.Publish(reqTopic, 1, false, payload)
		tok.Wait()
		log.Printf("MQTT pushall an %s, err=%v\n", p.IP, tok.Error())
	})

	client := mqtt.NewClient(opts)
	m.mu.Lock()
	m.clients[p.IP] = client
	m.statuses[p.IP] = &PrinterStatus{}
	m.mu.Unlock()

	go func() {
		token := client.Connect()
		if token.WaitTimeout(10 * time.Second) {
			if token.Error() != nil {
				log.Printf("MQTT connect error %s: %v\n", p.IP, token.Error())
				m.mu.Lock()
				if s, ok := m.statuses[p.IP]; ok {
					s.ConnectErr = token.Error().Error()
				}
				m.mu.Unlock()
			}
		} else {
			log.Printf("MQTT connect timeout %s\n", p.IP)
			m.mu.Lock()
			if s, ok := m.statuses[p.IP]; ok {
				s.ConnectErr = "connection timeout"
			}
			m.mu.Unlock()
		}
	}()
}

func (m *MQTTManager) Disconnect(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.clients[ip]; ok {
		c.Disconnect(500)
		delete(m.clients, ip)
	}
	delete(m.statuses, ip)
}

func (m *MQTTManager) GetStatus(ip string) *PrinterStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.statuses[ip]; ok {
		copy := *s
		return &copy
	}
	return nil
}

func (m *MQTTManager) AllStatuses() map[string]PrinterStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]PrinterStatus, len(m.statuses))
	for ip, s := range m.statuses {
		result[ip] = *s
	}
	return result
}

func (m *MQTTManager) handleMessage(ip string, payload []byte) {
	m.mu.Lock()
	if s, ok := m.statuses[ip]; ok {
		s.MsgCount++
		s.LastSeen = time.Now()
	}
	// Store last 500 chars of raw payload for debug
	raw := string(payload)
	if len(raw) > 5000 {
		raw = raw[:5000]
	}
	m.lastPayload[ip] = raw
	m.mu.Unlock()

	if mqttDebug {
		log.Printf("MQTT msg from %s (%d bytes)\nPAYLOAD: %s\n", ip, len(payload), string(payload))
	}

	// Antwortet der Drucker auf einen Steuerbefehl, wartet dort jemand darauf.
	pruefeQuittung(ip, payload)

	// Die Baugruppenliste kommt nur auf Nachfrage und in einer eigenen Huelle.
	if info, ok := parseVersionReport(payload); ok {
		m.mu.Lock()
		if st, vorhanden := m.statuses[ip]; vorhanden {
			st.Info = info
		}
		m.mu.Unlock()
		log.Printf("MQTT %s: Firmware %s, %d AMS-Baugruppe(n)", ip, info.Firmware, len(info.AMS))
		return
	}

	// Der Drucker verpackt den Status in {"print": {...}}
	var wrapper struct {
		Print json.RawMessage `json:"print"`
	}
	if err := json.Unmarshal(payload, &wrapper); err != nil || wrapper.Print == nil {
		// Try direct parse (some firmware versions)
		if mqttDebug {
			log.Printf("MQTT no 'print' key from %s, raw: %.200s\n", ip, string(payload))
		}
		return
	}

	// Use a generic map to handle all fields flexibly
	var raw2 map[string]interface{}
	if err := json.Unmarshal(wrapper.Print, &raw2); err != nil {
		log.Printf("MQTT parse error %s: %v\n", ip, err)
		return
	}

	// Helper to extract float
	getFloat := func(key string) float64 {
		v, ok := raw2[key]
		if !ok {
			return 0
		}
		switch x := v.(type) {
		case float64:
			return x
		case string:
			var f float64
			fmt.Sscanf(x, "%f", &f)
			return f
		}
		return 0
	}
	getInt := func(key string) int {
		v, ok := raw2[key]
		if !ok {
			return -1
		}
		switch x := v.(type) {
		case float64:
			return int(x)
		case int:
			return x
		}
		return -1
	}
	getString := func(key string) string {
		v, ok := raw2[key]
		if !ok {
			return ""
		}
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}

	if mqttDebug {
		keys := make([]string, 0, len(raw2))
		for k := range raw2 {
			keys = append(keys, k)
		}
		log.Printf("MQTT keys from %s: %v\n", ip, keys)
	}

	m.mu.Lock()
	s, ok := m.statuses[ip]
	if !ok {
		s = &PrinterStatus{}
		m.statuses[ip] = s
	}
	s.Online = true
	s.LastSeen = time.Now()

	// Der Drucker schickt Teil-Updates: ein fehlender Schluessel heisst "unveraendert",
	// ein vorhandener mit Wert 0 heisst wirklich 0. Frueher stand hier ueberall
	// "> 0" — ein Drucker bei 0 % bekam damit nie einen Fortschritt gesetzt und
	// behielt stattdessen den Wert des vorigen Auftrags.
	has := func(key string) bool { _, ok := raw2[key]; return ok }

	if has("gcode_state") {
		prev := strings.ToUpper(s.GcodeState)
		next := strings.ToUpper(getString("gcode_state"))
		s.GcodeState = getString("gcode_state")
		// Wechselt der Drucker von Pause oder Fehler zurueck auf Laufen, gilt die
		// Stoerung als quittiert. Ohne das haengen alte HMS-Meldungen ewig im
		// Speicher — der Drucker schickt das Feld naemlich nicht bei jeder
		// Nachricht mit, und dann blinkt das Licht endlos weiter.
		if next == "RUNNING" && prev != "RUNNING" && prev != "" {
			if len(s.HmsErrors) > 0 || s.PrintError > 0 {
				log.Printf("MQTT %s: Druck laeuft wieder — alte Stoerungsmeldungen verworfen", ip)
			}
			// Was jetzt noch ansteht, gilt als quittiert. Kommt es gleich per
			// pushall wieder herein, loest es kein Blinken mehr aus. Etwas
			// wirklich Neues waehrend des Drucks aber sehr wohl.
			s.AckedHms = map[string]bool{}
			for _, e := range s.HmsErrors {
				s.AckedHms[e] = true
			}
			s.HmsErrors = nil
			s.PrintError = 0
		}
		if next != "RUNNING" && prev == "RUNNING" {
			// Druck vorbei: das Gedaechtnis wird geleert, sonst bliebe eine
			// alte Meldung fuer alle Zeiten stumm.
			s.AckedHms = nil
		}
	}
	if has("mc_percent") {
		s.Progress = getInt("mc_percent")
	}
	if has("mc_remaining_time") {
		s.RemainTime = getInt("mc_remaining_time")
	}
	if has("layer_num") {
		s.Layer = getInt("layer_num")
	}
	if has("total_layer_num") {
		s.TotalLayers = getInt("total_layer_num")
	}
	if has("subtask_name") {
		s.SubtaskName = getString("subtask_name")
	}
	if has("nozzle_temper") {
		s.NozzleTemp = getFloat("nozzle_temper")
	}
	if has("nozzle_target_temper") {
		s.NozzleTarget = getFloat("nozzle_target_temper")
	}
	if has("bed_temper") {
		s.BedTemp = getFloat("bed_temper")
	}
	if has("bed_target_temper") {
		s.BedTarget = getFloat("bed_target_temper")
	}
	if has("chamber_temper") {
		s.ChamberTemp = getFloat("chamber_temper")
	}
	if has("cooling_fan_speed") {
		s.FanSpeed = getInt("cooling_fan_speed")
	}
	if has("print_error") {
		s.PrintError = getInt("print_error")
	}

	// AMS — nur ersetzen, wenn die Nachricht wirklich AMS-Daten enthaelt
	if units, trayNow, ok := parseAMS(raw2); ok {
		s.AMS = units
		s.TrayNow = trayNow
	}
	if vt, ok := raw2["vt_tray"]; ok {
		if tray, ok2 := parseTray(vt); ok2 {
			s.ExtSpool = &tray
		}
	}

	// Firmware-Stand: der Drucker legt in upgrade_state ab, ob und welche
	// Baugruppe ein Update offen hat. Kommt der Block mit, gilt er als
	// massgeblich — auch eine leere Liste heisst dann "nichts offen".
	if usRaw, ok := raw2["upgrade_state"]; ok {
		s.Upgrade = parseUpgradeState(usRaw)
		if b, err := json.MarshalIndent(usRaw, "", "  "); err == nil {
			m.lastUpgrade[ip] = string(b)
		}
	}

	// HMS errors — always update (empty list = no errors)
	if hmsRaw, ok := raw2["hms"]; ok {
		s.HmsErrors = nil
		if hmsList, ok := hmsRaw.([]interface{}); ok {
			for _, h := range hmsList {
				if hm, ok := h.(map[string]interface{}); ok {
					if e, ok := hm["ecode"].(string); ok && e != "" {
						s.HmsErrors = append(s.HmsErrors, e)
					}
				}
			}
		}
	}
	m.mu.Unlock()
}

func connectAllMQTT() {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()
	for _, p := range printers {
		if p.Serial != "" {
			go mqttMgr.Connect(p)
		}
	}
}

// ─── HMS ERROR TABLE ──────────────────────────────────────────────────────────

var hmsErrorMap = map[string]string{
	"07FF0001": "Filament-Jam erkannt",
	"07FF0002": "Filament leer",
	"07FF0003": "Filament nicht erkannt",
	"07FF0004": "Filament zu schwach",
	"0500C001": "Extruder-Motor-Fehler",
	"0500C002": "Extruder überhitzt",
	"05008001": "Nozzle-Verstopfung",
	"05008002": "Nozzle-Temperatur außer Bereich",
	"0300C001": "Heizbett-Temperatur außer Bereich",
	"0300C002": "Heizbett-Überhitzung",
	"03008001": "Heizbett-Sensor-Fehler",
	"0C00C001": "X-Achse blockiert",
	"0C00C002": "Y-Achse blockiert",
	"0C00C003": "Z-Achse blockiert",
	"0C008001": "Homing-Fehler X",
	"0C008002": "Homing-Fehler Y",
	"0C008003": "Homing-Fehler Z",
	"1200C001": "AMS Kommunikationsfehler",
	"1200C002": "AMS Filament-Sensor-Fehler",
	"12008001": "AMS Filament-Jam",
	"12008002": "AMS Filament leer",
	"12008003": "AMS Filament nicht erkannt",
	"12008004": "AMS Motor-Fehler",
	"0400C001": "Kammer-Temperatur zu hoch",
	"04008001": "Lüfter-Fehler",
	"0700C001": "Druckplatte nicht erkannt",
	"0700C002": "Druckplatte falsch eingelegt",
	"0B00C001": "Kamera-Fehler",
	"0B00C002": "Lidar-Fehler (AI-Erkennung)",
}

func HMSErrorText(ecode string) string {
	normalized := ""
	for _, c := range ecode {
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') {
			normalized += string(c)
		}
	}
	if len(normalized) >= 8 {
		key := normalized[:8]
		if msg, ok := hmsErrorMap[key]; ok {
			return msg
		}
	}
	return "Fehler: " + ecode
}

func PrintErrorText(code int) string {
	switch code {
	case 0:
		return ""
	case 50331904:
		return "Filament-Jam"
	case 50331905:
		return "Filament leer"
	case 50397184:
		return "Nozzle-Verstopfung"
	case 83886080:
		return "Heizbett-Fehler"
	case 117440512:
		return "AMS-Fehler"
	default:
		return fmt.Sprintf("Fehler #%d", code)
	}
}

func (m *MQTTManager) SetCameraResolution(printers []Printer, resolution string) int {
	if resolution != "720p" && resolution != "1080p" {
		return 0
	}
	payload := fmt.Sprintf(`{"camera":{"sequence_id":"0","command":"ipcam_resolution_set","resolution":"%s"}}`, resolution)
	sent := 0
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range printers {
		if p.Serial == "" {
			continue
		}
		// Die neue Generation (H2/X2/P2) kennt das alte 720p/1080p-Umschalten
		// per ipcam_resolution_set nicht. Schickt man es trotzdem, kann der
		// Drucker seine Kamera/Liveview abschalten (X2D „geht von selbst wieder
		// aus"). Deshalb wird der Befehl für diese Modelle NICHT gesendet — der
		// Livestream läuft unabhängig davon weiter.
		if neueKameraGeneration(p.Model) {
			continue
		}
		client, ok := m.clients[p.IP]
		if !ok || !client.IsConnected() {
			continue
		}
		topic := fmt.Sprintf("device/%s/request", p.Serial)
		tok := client.Publish(topic, 1, false, payload)
		tok.Wait()
		if tok.Error() == nil {
			sent++
			log.Printf("CAM %s -> %s\n", p.IP, resolution)
		}
	}
	return sent
}

// ─── AMS ──────────────────────────────────────────────────────────────────────

func asString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%g", x)
	}
	return ""
}

// parseTray liest ein einzelnes Fach. Leere Faecher liefert es mit leerem Typ
// zurueck, damit die Oberflaeche "leer" von "nicht vorhanden" unterscheiden kann.
func parseTray(v interface{}) (AMSTray, bool) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return AMSTray{}, false
	}
	t := AMSTray{
		Slot:     asString(m["id"]),
		Type:     strings.TrimSpace(asString(m["tray_type"])),
		Color:    strings.ToUpper(strings.TrimSpace(asString(m["tray_color"]))),
		SubBrand: strings.TrimSpace(asString(m["tray_sub_brands"])),
		Remain:   -1,
	}
	if r, ok := m["remain"].(float64); ok {
		t.Remain = int(r)
	}
	// "cols" listet bei mehrfarbigem Filament alle Farben auf. Werkseigene
	// Rollen fuellen es immer, Fremdfilament haeufig nur mit einem Eintrag.
	if cols, ok := m["cols"].([]interface{}); ok {
		for _, c := range cols {
			if hex := strings.ToUpper(strings.TrimSpace(asString(c))); hex != "" && hex != "00000000" {
				t.Colors = append(t.Colors, hex)
			}
		}
	}
	if t.Color == "" && len(t.Colors) > 0 {
		t.Color = t.Colors[0]
	}
	// "00000000" ist die Farbe eines leeren Fachs — als unbekannt behandeln
	if t.Color == "00000000" {
		t.Color = ""
	}
	return t, true
}

// parseAMS zieht die AMS-Einheiten aus einer Statusnachricht. Das dritte
// Rueckgabefeld sagt, ob die Nachricht ueberhaupt AMS-Daten enthielt — der
// Drucker schickt sie nur gelegentlich mit.
func parseAMS(raw map[string]interface{}) ([]AMSUnit, string, bool) {
	outer, ok := raw["ams"].(map[string]interface{})
	if !ok {
		return nil, "", false
	}
	list, ok := outer["ams"].([]interface{})
	if !ok {
		return nil, "", false
	}

	trayNow := asString(outer["tray_now"])
	units := make([]AMSUnit, 0, len(list))
	for _, item := range list {
		um, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		u := AMSUnit{
			ID:       asString(um["id"]),
			Humidity: asString(um["humidity"]),
			Temp:     asString(um["temp"]),
		}
		if trays, ok := um["tray"].([]interface{}); ok {
			for _, tv := range trays {
				if t, ok := parseTray(tv); ok {
					u.Trays = append(u.Trays, t)
				}
			}
		}
		units = append(units, u)
	}
	return units, trayNow, true
}

// ─── KOMMANDOS ────────────────────────────────────────────────────────────────

// Das Feld "param" gehoert laut Protokoll zwingend dazu, auch wenn es immer
// leer bleibt. Ohne das Feld nimmt der Drucker den Befehl entgegen und tut
// nichts — genau das Verhalten, das gemeldet wurde: kein Fehler, keine Wirkung.
// Das Feld "param" gehoert laut Protokoll zwingend dazu, auch wenn es immer
// leer bleibt. Die sequence_id wird je Befehl vergeben, damit die Antwort des
// Druckers eindeutig zugeordnet werden kann.
var printCommands = map[string]string{
	"pause":  `{"print":{"sequence_id":"%s","command":"pause","param":""}}`,
	"resume": `{"print":{"sequence_id":"%s","command":"resume","param":""}}`,
	"stop":   `{"print":{"sequence_id":"%s","command":"stop","param":""}}`,
}

// SendPrintCommand schickt pause/resume/stop an einen Drucker.
func (m *MQTTManager) SendPrintCommand(p Printer, cmd string) error {
	vorlage, ok := printCommands[cmd]
	if !ok {
		return fmt.Errorf("unbekanntes Kommando %q", cmd)
	}
	if p.Serial == "" {
		return fmt.Errorf("%s: keine Seriennummer hinterlegt", p.IP)
	}

	m.mu.RLock()
	client, connected := m.clients[p.IP]
	m.mu.RUnlock()
	if !connected || !client.IsConnected() {
		return fmt.Errorf("%s: keine MQTT-Verbindung", p.IP)
	}

	seq := naechsteSeq()
	payload := fmt.Sprintf(vorlage, seq)

	// Erst zuhoeren, dann senden — sonst kann die Antwort schneller da sein
	// als der Warteplatz.
	warten := warteAufQuittung(p.IP, cmd, seq)

	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("%s: Zeitueberschreitung beim Senden", p.IP)
	}
	if err := tok.Error(); err != nil {
		return fmt.Errorf("%s: %v", p.IP, err)
	}

	// Auf die Quittung warten. Ohne sie wissen wir nur, dass die Nachricht
	// abgeschickt wurde — nicht, dass sie ausgefuehrt wird.
	var antwort cmdAntwort
	angekommen := false
	select {
	case antwort = <-warten:
		angekommen = true
	case <-time.After(cmdWartezeit):
	}
	if err := deuteQuittung(antwort, angekommen); err != nil {
		log.Printf("PRINT %s -> %s FEHLGESCHLAGEN: %v", p.IP, cmd, err)
		return fmt.Errorf("%s: %v", p.Name, err)
	}
	log.Printf("PRINT %s -> %s bestaetigt\n", p.IP, cmd)

	// Direkt danach den vollen Status anfordern, damit die Oberflaeche nicht
	// bis zum naechsten Turnus auf der alten Anzeige sitzen bleibt.
	go func() {
		time.Sleep(700 * time.Millisecond)
		m.RequestStatus(p)
	}()
	return nil
}

// SendRaw schickt eine fertige Nachricht und wartet wie bei den anderen
// Befehlen auf die Quittung des Druckers.
func (m *MQTTManager) SendRaw(p Printer, command, seq, payload string) error {
	if p.Serial == "" {
		return fmt.Errorf("%s: keine Seriennummer hinterlegt", p.IP)
	}
	m.mu.RLock()
	client, connected := m.clients[p.IP]
	m.mu.RUnlock()
	if !connected || !client.IsConnected() {
		return fmt.Errorf("%s: keine MQTT-Verbindung", p.IP)
	}

	warten := warteAufQuittung(p.IP, command, seq)
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("%s: Zeitueberschreitung beim Senden", p.IP)
	}
	if err := tok.Error(); err != nil {
		return fmt.Errorf("%s: %v", p.IP, err)
	}

	var antwort cmdAntwort
	angekommen := false
	select {
	case antwort = <-warten:
		angekommen = true
	case <-time.After(cmdWartezeit):
	}
	if err := deuteQuittung(antwort, angekommen); err != nil {
		log.Printf("PRINT %s -> %s FEHLGESCHLAGEN: %v", p.IP, command, err)
		return fmt.Errorf("%s: %v", p.Name, err)
	}
	log.Printf("PRINT %s -> %s bestaetigt", p.IP, command)
	return nil
}

// RequestStatus fordert einen vollstaendigen Statusbericht an (pushall).
func (m *MQTTManager) RequestStatus(p Printer) bool {
	if p.Serial == "" {
		return false
	}
	m.mu.RLock()
	client, ok := m.clients[p.IP]
	m.mu.RUnlock()
	if !ok || !client.IsConnected() {
		return false
	}
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false,
		`{"pushing":{"sequence_id":"0","command":"pushall","version":1,"push_target":1}}`)
	return tok.WaitTimeout(3*time.Second) && tok.Error() == nil
}

// PublishRaw sendet eine fertige Nachricht und meldet nur, ob das Verschicken
// geklappt hat — es wartet NICHT auf eine Quittung. Fuer Befehle wie das
// Anstossen eines Firmware-Updates, deren Antwortkanal nicht sicher bekannt ist.
func (m *MQTTManager) PublishRaw(p Printer, payload string) error {
	if p.Serial == "" {
		return fmt.Errorf("%s: keine Seriennummer hinterlegt", p.IP)
	}
	m.mu.RLock()
	client, ok := m.clients[p.IP]
	m.mu.RUnlock()
	if !ok || !client.IsConnected() {
		return fmt.Errorf("%s: keine MQTT-Verbindung", p.IP)
	}
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5*time.Second) || tok.Error() != nil {
		return fmt.Errorf("%s: Senden fehlgeschlagen", p.IP)
	}
	return nil
}

// IsConnected sagt, ob fuer diese IP eine lebende MQTT-Verbindung besteht.
func (m *MQTTManager) IsConnected(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[ip]
	return ok && c.IsConnected()
}

// syncMQTT gleicht die offenen Verbindungen mit der Druckerliste ab: neue
// Drucker werden verbunden, entfernte getrennt. Frueher lief connectAllMQTT nur
// einmal beim Start — ein neu angelegter Drucker blieb deshalb bis zum Neustart
// der App ohne jeden Status.
func syncMQTT() {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	wanted := map[string]bool{}
	for _, p := range printers {
		if p.Serial == "" {
			continue
		}
		wanted[p.IP] = true
		if !mqttMgr.hasClient(p.IP) {
			go mqttMgr.Connect(p)
		}
	}

	for _, ip := range mqttMgr.clientIPs() {
		if !wanted[ip] {
			mqttMgr.Disconnect(ip)
		}
	}
}

func (m *MQTTManager) hasClient(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[ip]
	return ok
}

func (m *MQTTManager) clientIPs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ips := make([]string, 0, len(m.clients))
	for ip := range m.clients {
		ips = append(ips, ip)
	}
	return ips
}

// pushAllLoop fordert regelmaessig den vollen Status an. Ohne das bleibt ein
// Drucker, dessen erste Antwort verloren ging, dauerhaft ohne Daten.
func pushAllLoop() {
	runde := 0
	for {
		time.Sleep(45 * time.Second)
		if netzPausiert() {
			continue
		}
		mu.Lock()
		printers := make([]Printer, len(state.Printers))
		copy(printers, state.Printers)
		mu.Unlock()
		runde++
		for _, p := range printers {
			mqttMgr.RequestStatus(p)
			// Die Baugruppenliste aendert sich fast nie — einmal beim ersten
			// Durchgang und danach etwa stuendlich genuegt vollauf.
			s := mqttMgr.GetStatus(p.IP)
			if s != nil && s.Online && (s.Info == nil || runde%80 == 0) {
				mqttMgr.RequestVersion(p)
			}
		}
	}
}

// ─── KAMMERBELEUCHTUNG ────────────────────────────────────────────────────────
//
// X1E und X2D haben keine eigene Signalleuchte. Damit ein Fehler in einer Halle
// mit 42 Geraeten auffaellt, wird bei diesen Modellen die Kammerbeleuchtung zum
// Blinken gebracht. Der Drucker kann das selbst ("flashing" mit eigenen Zeiten),
// deshalb wird nicht im Sekundentakt gefunkt, sondern nur alle 20 s aufgefrischt.

var defaultBlinkModels = []string{"X1E", "X2D"}

func blinkModels() map[string]bool {
	mu.Lock()
	list := append([]string(nil), state.BlinkModels...)
	mu.Unlock()
	if len(list) == 0 {
		list = defaultBlinkModels
	}
	out := map[string]bool{}
	for _, m := range list {
		if m = strings.ToUpper(strings.TrimSpace(m)); m != "" {
			out[m] = true
		}
	}
	return out
}

// SetChamberLight schaltet die Kammerbeleuchtung. mode ist "on", "off" oder
// "flashing"; die Zeiten gelten nur fuer "flashing".
func (m *MQTTManager) SetChamberLight(p Printer, mode string, onMS, offMS int) error {
	if p.Serial == "" {
		return fmt.Errorf("%s: keine Seriennummer", p.IP)
	}
	m.mu.RLock()
	client, ok := m.clients[p.IP]
	m.mu.RUnlock()
	if !ok || !client.IsConnected() {
		return fmt.Errorf("%s: keine MQTT-Verbindung", p.IP)
	}

	payload := fmt.Sprintf(`{"system":{"sequence_id":"0","command":"ledctrl","led_node":"chamber_light",`+
		`"led_mode":"%s","led_on_time":%d,"led_off_time":%d,"loop_times":0,"interval_time":0}}`,
		mode, onMS, offMS)
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("%s: Zeitueberschreitung", p.IP)
	}
	return tok.Error()
}

// printerHasError sagt, ob ein Drucker gerade einen Fehler meldet.
// openHms zaehlt die Stoerungsmeldungen, die noch nicht als quittiert gelten.
func openHms(s *PrinterStatus) int {
	n := 0
	for _, e := range s.HmsErrors {
		if s.AckedHms == nil || !s.AckedHms[e] {
			n++
		}
	}
	return n
}

func printerHasError(s *PrinterStatus) bool {
	if s == nil || !s.Online {
		return false
	}
	return s.PrintError > 0 || openHms(s) > 0 || strings.EqualFold(s.GcodeState, "FAILED")
}

// printerWantsBlink entscheidet, ob die Kammer blinken soll. Geblinkt wird bei
// einer Stoerung ODER im Pausezustand. Ein fertiger Druck (FINISH) blinkt
// ausdruecklich NIE — auch dann nicht, wenn zum Abschluss noch eine Meldung
// ansteht. So signalisiert das Blinken nur, was wirklich Aufmerksamkeit
// braucht: Fehler und angehaltene Drucke.
func printerWantsBlink(s *PrinterStatus) bool {
	if s == nil || !s.Online {
		return false
	}
	if strings.EqualFold(s.GcodeState, "FINISH") {
		return false
	}
	if strings.EqualFold(s.GcodeState, "PAUSE") || strings.EqualFold(s.GcodeState, "PAUSED") {
		return true
	}
	return printerHasError(s)
}

// errorLightLoop haelt das Blinken aufrecht, solange ein Fehler ansteht, und
// stellt die Beleuchtung genau einmal zurueck, wenn er weg ist.
func errorLightLoop() {
	// Wer beim letzten Lauf geblinkt hat, steht in der Konfiguration. Sonst
	// wuesste die App nach einem Neustart nichts davon und das Licht bliebe
	// fuer immer im Blinkmodus.
	blinking := map[string]bool{}
	lastSent := map[string]time.Time{}
	mu.Lock()
	for _, ip := range state.BlinkingIPs {
		blinking[ip] = true
	}
	mu.Unlock()

	persist := func() {
		list := make([]string, 0, len(blinking))
		for ip := range blinking {
			list = append(list, ip)
		}
		sort.Strings(list)
		mu.Lock()
		same := len(list) == len(state.BlinkingIPs)
		if same {
			for i := range list {
				if list[i] != state.BlinkingIPs[i] {
					same = false
					break
				}
			}
		}
		if !same {
			state.BlinkingIPs = list
		}
		mu.Unlock()
		if !same {
			saveState()
		}
	}

	// Alle fuenf Sekunden nachsehen. Zwanzig waren zu traege: nach dem
	// Fortsetzen eines Drucks blinkte die Kammer noch fast eine halbe Minute
	// weiter, was wie ein neuer Fehler aussah.
	for {
		time.Sleep(5 * time.Second)
		if netzPausiert() {
			continue
		}

		mu.Lock()
		enabled := state.ErrorBlink == nil || *state.ErrorBlink
		printers := make([]Printer, len(state.Printers))
		copy(printers, state.Printers)
		mu.Unlock()

		models := blinkModels()
		for _, p := range printers {
			if p.Serial == "" {
				continue
			}
			wants := enabled && !blinkAusFuer(p.IP) && models[strings.ToUpper(p.Model)] && printerWantsBlink(mqttMgr.GetStatus(p.IP))

			switch {
			case wants:
				// Haeufig pruefen, selten senden. Der 5-Sekunden-Takt aus 1.5.0
				// hat die Befehle an die Drucker vervierfacht, ohne dass es
				// dafuer einen Grund gab: das Auffrischen dient nur dazu, einen
				// Neustart des Druckers zu ueberstehen, und dafuer reicht eine
				// Minute. Neu auftretende Stoerungen werden weiterhin binnen
				// fuenf Sekunden bemerkt.
				if blinking[p.IP] && time.Since(lastSent[p.IP]) < time.Minute {
					continue
				}
				if err := mqttMgr.SetChamberLight(p, "flashing", 1000, 1000); err != nil {
					continue
				}
				lastSent[p.IP] = time.Now()
				blinking[p.IP] = true
			case blinking[p.IP]:
				// Zurueck auf normales Licht. Klappt es nicht, bleibt der Eintrag
				// stehen und es wird beim naechsten Durchgang erneut versucht.
				if err := mqttMgr.SetChamberLight(p, "on", 500, 500); err == nil {
					delete(blinking, p.IP)
					delete(lastSent, p.IP)
					log.Printf("Kammerlicht %s zurueckgesetzt (Stoerung behoben)", p.IP)
				}
			}
		}
		persist()
	}
}

// ResetChamberLights stellt bei allen betroffenen Druckern normales Licht her.
// Wird vom Schalter in den Einstellungen benutzt, wenn die Sonderbeleuchtung
// abgeschaltet wird.
func ResetChamberLights() int {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	wanted := map[string]bool{}
	for _, ip := range state.BlinkingIPs {
		wanted[ip] = true
	}
	mu.Unlock()

	n := 0
	for _, p := range printers {
		if !wanted[p.IP] {
			continue
		}
		if err := mqttMgr.SetChamberLight(p, "on", 500, 500); err == nil {
			n++
		}
	}
	mu.Lock()
	state.BlinkingIPs = nil
	mu.Unlock()
	saveState()
	return n
}

// neueKameraGeneration erkennt die H2-/X2-/P2-Reihe (H2D, H2S, X2D, P2S …). Diese
// Geräte haben eine andere Kamera und unterstützen das alte
// ipcam_resolution_set (720p/1080p) nicht.
func neueKameraGeneration(model string) bool {
	m := strings.ToUpper(strings.TrimSpace(model))
	return strings.HasPrefix(m, "H2") || strings.HasPrefix(m, "X2") || strings.HasPrefix(m, "P2")
}

// reconnectLoop hält offline Drucker im Blick und verbindet sie neu, sobald sie
// wieder erreichbar sind. Nötig, weil paho AutoReconnect NUR nach einer einmal
// geglückten Verbindung greift: War der Drucker beim ersten Versuch aus, würde
// er ohne diesen Takt für immer offline bleiben. Alle 30 s wird für jeden
// Drucker mit Seriennummer geprüft, ob eine lebende MQTT-Verbindung besteht;
// wenn nicht, wird die alte (hängende) getrennt und frisch aufgebaut.
func reconnectLoop() {
	for {
		time.Sleep(30 * time.Second)
		reconnectDurchlauf()
	}
}

// reconnectDurchlauf ist ein einzelner Prüf-Durchlauf (getrennt für Tests).
func reconnectDurchlauf() {
	if netzPausiert() {
		return
	}
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()
	for _, p := range printers {
		if p.Serial == "" {
			continue
		}
		if mqttMgr.IsConnected(p.IP) {
			continue
		}
		mqttMgr.Disconnect(p.IP) // alte, tote Verbindung sauber weg
		mqttMgr.Connect(p)       // frisch versuchen
	}
}
