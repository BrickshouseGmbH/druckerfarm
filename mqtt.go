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
	GcodeFile   string `json:"gcode_file"` // fallback source for the print file name

	NozzleTemp   float64 `json:"nozzle_temp"`
	NozzleTarget float64 `json:"nozzle_target"`
	BedTemp      float64 `json:"bed_temp"`
	BedTarget    float64 `json:"bed_target"`
	ChamberTemp  float64 `json:"chamber_temp"`

	FanSpeed   int      `json:"fan_speed"`
	PrintError int      `json:"print_error"`
	HmsErrors  []string `json:"hms_errors"`

	// AI/camera monitoring (xcam): the printer's built-in error detection
	// (Erstschicht-Inspektion, Spaghetti-/Fremdkoerper-Erkennung). Rohfelder
	// are included so nothing is lost; AiMonitoring is the
	// derived yes/no answer. The raw LIDAR point cloud is NOT exposed by the
	// printer over the local interface.
	Xcam         map[string]interface{} `json:"xcam,omitempty"`
	AiMonitoring bool                   `json:"ai_monitoring,omitempty"`

	// Timelapse is the timelapse state reported by the printer from
	// ipcam.timelapse ("enable"/"disable"). Empty when the printer has
	// not (yet) reported the field. TimelapseKnown distinguishes "off" from
	// "unbekannt".
	Timelapse      string `json:"timelapse,omitempty"`
	TimelapseKnown bool   `json:"timelapse_known,omitempty"`

	// AckedHms are fault messages that were already present when the print
	// started. The X1 sends the full HMS list with every message,
	// even when the print has long continued and the
	// message was merely not acknowledged on the device. Without this memory
	// HmsErrors refilled right after clearing and the chamber light
	// blinkte endlos weiter.
	AckedHms map[string]bool `json:"-"`

	// AMS and filament
	AMS      []AMSUnit `json:"ams"`
	TrayNow  string    `json:"tray_now"`
	ExtSpool *AMSTray  `json:"ext_spool,omitempty"`

	// Device info from info.get_version — firmware and AMS levels
	Info *DeviceInfo `json:"info,omitempty"`

	// Open firmware updates as the printer reports them (upgrade_state /
	// new_ver_list im pushall). Rein lesend — dieses Programm loest nichts aus.
	Upgrade []UpgradeMessage `json:"upgrade,omitempty"`

	// Debug
	MsgCount  int    `json:"msg_count"`
	LastTopic string `json:"last_topic,omitempty"`
}

// AMSTray is a single tray. Color comes as RRGGBBAA from the printer.
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

// mqttDebug enables logging of complete MQTT payloads. By default
// off: with 36 printers this would otherwise write megabytes per minute to the log.
var mqttDebug = os.Getenv("DRUCKERFARM_MQTT_DEBUG") != ""

// ─── MQTT MANAGER ─────────────────────────────────────────────────────────────

type MQTTManager struct {
	mu          sync.RWMutex
	clients     map[string]mqtt.Client
	statuses    map[string]*PrinterStatus
	lastPayload map[string]string // raw payload for debug
	lastUpgrade map[string]string // raw upgrade_state per IP, for diagnostics
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

	// MQTT runs over TLS on port 8883
	opts.AddBroker(fmt.Sprintf("ssl://%s:8883", p.IP))

	// The printer requires a specific client-ID format
	// STABLE client ID (no timestamp): the printer allows only ONE local
	// MQTT connection. With a changing ID a new connection cannot take over a
	// stuck old one — the slot stays occupied by a "zombie"
	// and the printer stays offline. With a fixed ID the broker kicks the
	// old session automatically on reconnect (MQTT standard), and the
	// reconnect works reliably.
	opts.SetClientID("druckerfarm-" + p.Serial)
	opts.SetUsername("bblp")
	opts.SetPassword(p.Code)

	// The printer uses a self-signed certificate
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

	// If the printer replies to a control command, someone is waiting for it.
	checkAck(ip, payload)

	// The module list comes only on request and in its own envelope.
	if info, ok := parseVersionReport(payload); ok {
		m.mu.Lock()
		if st, vorhanden := m.statuses[ip]; vorhanden {
			st.Info = info
		}
		m.mu.Unlock()
		log.Printf("MQTT %s: Firmware %s, %d AMS-Baugruppe(n)", ip, info.Firmware, len(info.AMS))
		return
	}

	// The printer wraps the status in {"print": {...}}
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

	// The printer sends partial updates: a missing key means "unchanged",
	// a present one with value 0 means really 0. Previously everything here said
	// "> 0" — a printer at 0 % thus never got a progress set and
	// kept the value of the previous job instead.
	has := func(key string) bool { _, ok := raw2[key]; return ok }

	if has("gcode_state") {
		prev := strings.ToUpper(s.GcodeState)
		next := strings.ToUpper(getString("gcode_state"))
		s.GcodeState = getString("gcode_state")
		// When the printer switches from pause or error back to running, the
		// fault counts as acknowledged. Without this old HMS messages hang forever in
		// memory — the printer does not send the field with every
		// message, and then the light keeps blinking endlessly.
		if next == "RUNNING" && prev != "RUNNING" && prev != "" {
			if len(s.HmsErrors) > 0 || s.PrintError > 0 {
				log.Printf("MQTT %s: Druck laeuft wieder — alte Stoerungsmeldungen verworfen", ip)
			}
			// Whatever is still pending now counts as acknowledged. If it comes right back via
			// pushall, it no longer triggers blinking. Something
			// truly new during the print does, though.
			s.AckedHms = map[string]bool{}
			for _, e := range s.HmsErrors {
				s.AckedHms[e] = true
			}
			s.HmsErrors = nil
			s.PrintError = 0
		}
		if next != "RUNNING" && prev == "RUNNING" {
			// Print over: the memory is cleared, otherwise an
			// old message would stay muted forever.
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
	// Some models/firmwares report the name only in gcode_file (full path).
	if has("gcode_file") {
		s.GcodeFile = getString("gcode_file")
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

	// AMS — only replace when the message actually contains AMS data
	if units, trayNow, ok := parseAMS(raw2); ok {
		s.AMS = units
		s.TrayNow = trayNow
	}
	if vt, ok := raw2["vt_tray"]; ok {
		if tray, ok2 := parseTray(vt); ok2 {
			s.ExtSpool = &tray
		}
	}

	// Firmware level: the printer stores in upgrade_state whether and which
	// module has an update pending. If the block is present, it is
	// authoritative — even an empty list then means "nothing pending".
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

	// xcam: AI/camera monitoring (if the printer reports it).
	if xcamRaw, ok := raw2["xcam"].(map[string]interface{}); ok {
		s.Xcam = xcamRaw
		truthy := func(v interface{}) bool {
			switch x := v.(type) {
			case bool:
				return x
			case string:
				u := strings.ToLower(strings.TrimSpace(x))
				return u == "on" || u == "enable" || u == "enabled" || u == "1" || u == "true"
			case float64:
				return x != 0
			}
			return false
		}
		s.AiMonitoring = truthy(xcamRaw["printing_monitor"]) || truthy(xcamRaw["spaghetti_detector"]) ||
			truthy(xcamRaw["first_layer_inspector"]) || truthy(xcamRaw["buildplate_marker_detector"])
	}

	// ipcam.timelapse: timelapse on/off as the printer reports it itself.
	if ipcamRaw, ok := raw2["ipcam"].(map[string]interface{}); ok {
		if tl, ok := ipcamRaw["timelapse"].(string); ok {
			u := strings.ToLower(strings.TrimSpace(tl))
			if u != "" {
				s.Timelapse = u
				s.TimelapseKnown = true
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
		// The new generation (H2/X2/P2) does not know the old 720p/1080p switching
		// via ipcam_resolution_set. If sent anyway, the
		// printer may turn off its camera/liveview (X2D "goes off again by
		// itself"). So the command is NOT sent for these models — the
		// livestream keeps running independently.
		if newCameraGeneration(p.Model) {
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

// SetTimelapse turns a printer's timelapse recording on or off.
// Bambu accepts this via the same "camera" command as switching
// the resolution (ipcam_timelapse, control enable/disable). Returns true when
// the command was sent.
func (m *MQTTManager) SetTimelapse(p Printer, on bool) error {
	if p.Serial == "" {
		return fmt.Errorf("%s: keine Seriennummer", p.IP)
	}
	ctrl := "disable"
	if on {
		ctrl = "enable"
	}
	payload := fmt.Sprintf(`{"camera":{"sequence_id":"0","command":"ipcam_timelapse","control":"%s"}}`, ctrl)
	m.mu.RLock()
	client, ok := m.clients[p.IP]
	m.mu.RUnlock()
	if !ok || !client.IsConnected() {
		return fmt.Errorf("%s: nicht verbunden", p.IP)
	}
	topic := fmt.Sprintf("device/%s/request", p.Serial)
	tok := client.Publish(topic, 1, false, payload)
	tok.Wait()
	if tok.Error() != nil {
		return tok.Error()
	}
	log.Printf("TIMELAPSE %s -> %s\n", p.IP, ctrl)
	return nil
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

// parseTray reads a single tray. Empty trays are returned with an empty type
// so the UI can distinguish "empty" from "not present".
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
	// factory spools always fill it, third-party filament often with only one entry.
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
	// "00000000" is the color of an empty tray — treat as unknown
	if t.Color == "00000000" {
		t.Color = ""
	}
	return t, true
}

// parseAMS extracts the AMS units from a status message. The third
// return value says whether the message contained any AMS data at all — the
// printer only includes them occasionally.
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

// The "param" field is mandatory per protocol, even though it always
// stays empty. Without the field the printer accepts the command and does
// nothing — exactly the reported behaviour: no error, no effect.
// The "param" field is mandatory per protocol, even though it always
// stays empty. The sequence_id is assigned per command so the printer's reply
// can be matched unambiguously.
var printCommands = map[string]string{
	"pause":  `{"print":{"sequence_id":"%s","command":"pause","param":""}}`,
	"resume": `{"print":{"sequence_id":"%s","command":"resume","param":""}}`,
	"stop":   `{"print":{"sequence_id":"%s","command":"stop","param":""}}`,
}

// SendPrintCommand sends pause/resume/stop to a printer.
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

	seq := nextSeq()
	payload := fmt.Sprintf(vorlage, seq)

	// Listen first, then send — otherwise the reply may arrive faster
	// than the waiting slot.
	warten := waitForAck(p.IP, cmd, seq)

	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("%s: Zeitueberschreitung beim Senden", p.IP)
	}
	if err := tok.Error(); err != nil {
		return fmt.Errorf("%s: %v", p.IP, err)
	}

	// Wait for the acknowledgement. Without it we only know the message
	// was sent — not that it is executed.
	var antwort cmdResponse
	angekommen := false
	select {
	case antwort = <-warten:
		angekommen = true
	case <-time.After(cmdTimeout):
	}
	if err := interpretAck(antwort, angekommen); err != nil {
		log.Printf("PRINT %s -> %s FEHLGESCHLAGEN: %v", p.IP, cmd, err)
		return fmt.Errorf("%s: %v", p.Name, err)
	}
	log.Printf("PRINT %s -> %s bestaetigt\n", p.IP, cmd)

	// Request full status right after, so the UI does not
	// sit on the old display until the next cycle.
	go func() {
		time.Sleep(700 * time.Millisecond)
		m.RequestStatus(p)
	}()
	return nil
}

// SendRaw sends a ready message and waits, as with the other
// commands, for the printer's acknowledgement.
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

	warten := waitForAck(p.IP, command, seq)
	tok := client.Publish(fmt.Sprintf("device/%s/request", p.Serial), 1, false, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("%s: Zeitueberschreitung beim Senden", p.IP)
	}
	if err := tok.Error(); err != nil {
		return fmt.Errorf("%s: %v", p.IP, err)
	}

	var antwort cmdResponse
	angekommen := false
	select {
	case antwort = <-warten:
		angekommen = true
	case <-time.After(cmdTimeout):
	}
	if err := interpretAck(antwort, angekommen); err != nil {
		log.Printf("PRINT %s -> %s FEHLGESCHLAGEN: %v", p.IP, command, err)
		return fmt.Errorf("%s: %v", p.Name, err)
	}
	log.Printf("PRINT %s -> %s bestaetigt", p.IP, command)
	return nil
}

// RequestStatus requests a full status report (pushall).
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

// PublishRaw sends a ready message and only reports whether sending
// succeeded — it does NOT wait for an acknowledgement. For commands like
// triggering a firmware update, whose reply channel is not reliably known.
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

// IsConnected reports whether a live MQTT connection exists for this IP.
func (m *MQTTManager) IsConnected(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[ip]
	return ok && c.IsConnected()
}

// syncMQTT reconciles the open connections with the printer list: new
// printers are connected, removed ones disconnected. Previously connectAllMQTT ran only
// once at start — a newly added printer therefore stayed without any status
// until the app restarted.
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

// pushAllLoop requests the full status regularly. Without it a
// printer whose first reply was lost stays without data permanently.
func pushAllLoop() {
	runde := 0
	for {
		time.Sleep(45 * time.Second)
		if netPaused() {
			continue
		}
		mu.Lock()
		printers := make([]Printer, len(state.Printers))
		copy(printers, state.Printers)
		mu.Unlock()
		runde++
		for _, p := range printers {
			mqttMgr.RequestStatus(p)
			// The module list almost never changes — once on the first
			// pass and then about hourly is plenty.
			s := mqttMgr.GetStatus(p.IP)
			if s != nil && s.Online && (s.Info == nil || runde%80 == 0) {
				mqttMgr.RequestVersion(p)
			}
		}
	}
}

// ─── KAMMERBELEUCHTUNG ────────────────────────────────────────────────────────
//
// X1E and X2D have no signal light of their own. So an error in a hall
// with 42 devices stands out, the chamber light of these models is made to
// blink. The printer can do this itself ("flashing" with its own timings),
// so it is not sent every second but only refreshed every 20 s.

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

// SetChamberLight switches the chamber light. mode is "on", "off" or
// "flashing"; the timings apply only to "flashing".
// SetXcam toggles the built-in AI/camera monitoring (printing_monitor) on
// or off and sets the abort sensitivity. Best-effort — depending on model
// and firmware it varies; the printer ignores unknown fields.
func (m *MQTTManager) SetXcam(p Printer, monitoring bool, sensitivity string) error {
	ctrl := "false"
	if monitoring {
		ctrl = "true"
	}
	sens := sensitivity
	if sens == "" {
		sens = "medium"
	}
	payload := fmt.Sprintf(`{"xcam":{"sequence_id":"0","command":"xcam_control_set","module_name":"printing_monitor","control":%s,"print_halt":true,"halt_print_sensitivity":"%s"}}`, ctrl, sens)
	return m.PublishRaw(p, payload)
}

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

// printerHasError reports whether a printer is currently reporting an error.
// openHms counts the fault messages not yet considered acknowledged.
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

// printerWantsBlink decides whether the chamber should blink. Blinking happens on
// einer Stoerung ODER im Pausezustand. Ein fertiger Druck (FINISH) blinkt
// explicitly NEVER — not even when a message is still pending at the end.
// So the blinking signals only what truly needs attention:
// errors and paused prints.
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

// errorLightLoop keeps the blinking going while an error is pending, and
// resets the light exactly once when it is gone.
func errorLightLoop() {
	// Which ones blinked last run is in the configuration. Otherwise
	// the app would not know after a restart and the light would stay
	// in blink mode forever.
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

	// Check every five seconds. Twenty was too sluggish: after
	// resuming a print the chamber kept blinking for nearly half a minute
	// more, which looked like a new error.
	for {
		time.Sleep(5 * time.Second)
		if netPaused() {
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
				// Check often, send rarely. The 5-second cadence from 1.5.0
				// quadrupled the commands to the printers without any
				// reason: the refresh only serves to survive a
				// printer restart, and a minute is enough for that.
				// Newly occurring faults are still noticed within
				// five seconds.
				if blinking[p.IP] && time.Since(lastSent[p.IP]) < time.Minute {
					continue
				}
				if err := mqttMgr.SetChamberLight(p, "flashing", 1000, 1000); err != nil {
					continue
				}
				lastSent[p.IP] = time.Now()
				blinking[p.IP] = true
			case blinking[p.IP]:
				// Back to normal light. If it fails, the entry stays
				// and it is retried on the next pass.
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
// Used by the switch in the settings when the special lighting
// is turned off.
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

// newCameraGeneration detects the H2/X2/P2 series (H2D, H2S, X2D, P2S …). These
// devices have a different camera and do not support the old
// ipcam_resolution_set (720p/1080p).
func newCameraGeneration(model string) bool {
	m := strings.ToUpper(strings.TrimSpace(model))
	return strings.HasPrefix(m, "H2") || strings.HasPrefix(m, "X2") || strings.HasPrefix(m, "P2")
}

// reconnectLoop keeps an eye on offline printers and reconnects them once they
// are reachable again. Needed because paho AutoReconnect only kicks in after a
// once-successful connection: if the printer was off on the first try, it would
// stay offline forever without this cadence. Every 30 s each
// printer with a serial is checked for a live MQTT connection;
// if not, the old (stuck) one is dropped and freshly established.
func reconnectLoop() {
	for {
		time.Sleep(30 * time.Second)
		reconnectDurchlauf()
	}
}

// reconnectDurchlauf is a single check pass (separated for tests).
func reconnectDurchlauf() {
	if netPaused() {
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
