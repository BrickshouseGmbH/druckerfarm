package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// ─── ERREICHBARKEIT ───────────────────────────────────────────────────────────
//
// Wenn kein Bild kommt, gibt es drei mögliche Gründe, und sie sehen in der
// Oberfläche alle gleich aus: das Gerät ist aus, die Kamera ist aus, oder das
// Programm hat einen Fehler. Diese Prüfung trennt die drei, indem sie die drei
// Ports einzeln anklopft, auf die es ankommt:
//
//	322  RTSP  — die Kamera. Genau der Port, an dem go2rtc scheitert.
//	990  FTPS  — der Dateizugriff.
//	8883 MQTT  — Status und Steuerung.
//
// Damit ist ohne Ratespiel ablesbar: antwortet gar nichts, ist das Gerät aus
// oder nicht im selben Netz. Antworten 990 und 8883, aber 322 nicht, dann läuft
// der Drucker und nur die Kamera ist abgeschaltet — das ist ein Schalter am
// Gerät, kein Programmfehler.

type portResult struct {
	Port  int    `json:"port"`
	Was   string `json:"was"`
	Offen bool   `json:"offen"`
	Grund string `json:"grund,omitempty"`
	MS    int64  `json:"ms"`
}

type reachResult struct {
	IP     string       `json:"ip"`
	Name   string       `json:"name"`
	Modell string       `json:"modell"`
	Ports  []portResult `json:"ports"`
	Urteil string       `json:"urteil"`
}

var reachPorts = []struct {
	port int
	was  string
}{
	{322, "Kamera"},
	{990, "Dateien"},
	{8883, "Status"},
}

// probePort klopft an und übersetzt das Ergebnis. Die Unterscheidung zwischen
// "keine Antwort" und "aktiv abgelehnt" ist die wichtigste Information
// überhaupt: abgelehnt heißt, das Gerät lebt und nur der Dienst fehlt.
func probePort(ip string, port int, timeout time.Duration) portResult {
	res := portResult{Port: port}
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, itoa(port)), timeout)
	res.MS = time.Since(start).Milliseconds()
	if err == nil {
		conn.Close()
		res.Offen = true
		return res
	}
	low := strings.ToLower(err.Error())
	switch {
	case strings.Contains(low, "refused"):
		res.Grund = "abgelehnt — Gerät antwortet, Dienst läuft nicht"
	case strings.Contains(low, "timeout"), strings.Contains(low, "deadline"):
		res.Grund = "keine Antwort"
	case strings.Contains(low, "no route"), strings.Contains(low, "unreachable"):
		res.Grund = "nicht erreichbar — anderes Netz?"
	default:
		res.Grund = err.Error()
	}
	return res
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// urteil fasst zusammen, was die drei Ports gemeinsam bedeuten.
func urteil(ports []portResult) string {
	offen := map[int]bool{}
	abgelehnt := map[int]bool{}
	for _, p := range ports {
		offen[p.Port] = p.Offen
		abgelehnt[p.Port] = strings.HasPrefix(p.Grund, "abgelehnt")
	}
	switch {
	case offen[322]:
		return "Kamera erreichbar — Bild sollte kommen"
	case !offen[322] && !offen[990] && !offen[8883] && !abgelehnt[322]:
		return "Gerät aus oder nicht im Netz"
	case abgelehnt[322] && (offen[990] || offen[8883]):
		return "Drucker läuft, Kamera ist abgeschaltet — LAN Mode Liveview am Gerät einschalten"
	case !offen[322] && (offen[990] || offen[8883]):
		return "Drucker läuft, Kamera antwortet nicht — LAN Mode Liveview prüfen, danach Drucker neu starten"
	default:
		return "Teilweise erreichbar"
	}
}

func checkPrinter(p Printer, timeout time.Duration) reachResult {
	r := reachResult{IP: p.IP, Name: p.Name, Modell: p.Model}
	var wg sync.WaitGroup
	out := make([]portResult, len(reachPorts))
	for i, rp := range reachPorts {
		wg.Add(1)
		go func(i int, port int, was string) {
			defer wg.Done()
			pr := probePort(p.IP, port, timeout)
			pr.Was = was
			out[i] = pr
		}(i, rp.port, rp.was)
	}
	wg.Wait()
	r.Ports = out
	r.Urteil = urteil(out)
	return r
}

// handleReach prüft entweder einen Drucker (?ip=…) oder alle auf einmal.
// Parallel, aber gedeckelt — 42 Geräte × 3 Ports gleichzeitig wäre unhöflich.
func handleReach(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.Unlock()

	if ip := r.URL.Query().Get("ip"); ip != "" {
		for _, p := range printers {
			if p.IP == ip {
				writeJSON(w, checkPrinter(p, 3*time.Second))
				return
			}
		}
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}

	sem := make(chan struct{}, 8)
	res := make([]reachResult, len(printers))
	var wg sync.WaitGroup
	for i, p := range printers {
		wg.Add(1)
		go func(i int, p Printer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			res[i] = checkPrinter(p, 3*time.Second)
		}(i, p)
	}
	wg.Wait()
	sort.Slice(res, func(a, b int) bool { return res[a].Name < res[b].Name })

	zus := map[string]int{}
	for _, x := range res {
		zus[x.Urteil]++
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"drucker":         res,
		"zusammenfassung": zus,
	})
}

// handleReconnect stößt für EINEN Drucker eine frische MQTT-Verbindung an:
// alte Verbindung trennen, dann neu aufbauen. Nützlich direkt nach dem
// Bearbeiten oder beim manuellen Anpingen — so wird „gefunden, aber offline"
// nicht durch eine hängengebliebene Verbindung verursacht. Antwortet mit einem
// Hinweis, wenn die Seriennummer fehlt (dann ist MQTT-Status gar nicht möglich).
func handleReconnect(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		ip = r.FormValue("ip")
	}
	w.Header().Set("Content-Type", "application/json")

	mu.Lock()
	var gefunden *Printer
	for i := range state.Printers {
		if state.Printers[i].IP == ip {
			p := state.Printers[i]
			gefunden = &p
			break
		}
	}
	mu.Unlock()

	if gefunden == nil {
		http.Error(w, "unbekannter Drucker", http.StatusNotFound)
		return
	}

	hatSerial := strings.TrimSpace(gefunden.Serial) != ""
	if hatSerial {
		go func(p Printer) {
			mqttMgr.Disconnect(p.IP)
			syncMQTT()
		}(*gefunden)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ip":         ip,
		"hat_serial": hatSerial,
		"reconnect":  hatSerial,
	})
}
