package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── DRUCKERSUCHE IM NETZ ─────────────────────────────────────────────────────
//
// Die Drucker antworten auf SSDP-M-SEARCH — allerdings auf Port 2021, nicht auf
// dem ueblichen 1900. Die Antwort kommt als Unicast zurueck ("HTTP/1.1 200 OK")
// und traegt die Geraetedaten in eigenen Kopfzeilen: DevModel, DevName,
// DevConnect, DevBind, USN (Seriennummer) und DevVersion.
//
// Zwei Dinge sind dabei wichtig:
//   - UDP geht verloren. Bei ~36 Geraeten im Netz reicht ein einzelner Versuch
//     nicht, deshalb wird mehrfach gesendet.
//   - Der Suchtyp bleibt "ssdp:all". Damit antwortet jedes SSDP-Geraet; gefiltert
//     wird ueber die Kopfzeilen, die nur Drucker mitschicken.

const (
	ssdpPort             = 2021
	ssdpMulticast        = "239.255.255.250"
	discoverSearchRounds = 3
)

type DiscoveredPrinter struct {
	IP      string `json:"ip"`
	Model   string `json:"model"`
	Name    string `json:"name"`
	Serial  string `json:"serial"`
	Connect string `json:"connect"` // lan / cloud
	Bind    string `json:"bind"`    // free / occupied
	Version string `json:"version"`
	Known   bool   `json:"known"` // steht schon in der Druckerliste

	// Umgezogen: dieselbe Seriennummer ist bekannt, aber unter anderer Adresse.
	// Genau dieser Fall hat die halbe Farm lahmgelegt, ohne dass es jemand
	// sehen konnte.
	Umgezogen bool   `json:"umgezogen,omitempty"`
	AlteIP    string `json:"alte_ip,omitempty"`
}

func buildMSearch(host string) []byte {
	return []byte("M-SEARCH * HTTP/1.1\r\n" +
		"HOST: " + host + "\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 1\r\n" +
		"ST: ssdp:all\r\n\r\n")
}

// parseSSDPResponse liest eine Antwort. Der zweite Rueckgabewert ist false,
// wenn es sich erkennbar nicht um einen Drucker handelt — im Netz antworten
// auch Router, Fernseher und Drucker anderer Bauart auf SSDP.
func parseSSDPResponse(data []byte, srcIP string) (DiscoveredPrinter, bool) {
	text := string(data)
	if !strings.HasPrefix(strings.ToUpper(text), "HTTP/1.1 200") &&
		!strings.HasPrefix(strings.ToUpper(text), "NOTIFY") {
		return DiscoveredPrinter{}, false
	}

	headers := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := sc.Text()
		i := strings.IndexByte(line, ':')
		if i <= 0 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(line[:i]))
		headers[key] = strings.TrimSpace(line[i+1:])
	}

	// Je nach Firmware heissen die Felder "DevModel" oder "DevModel.suffix" —
	// deshalb wird nach dem Anfang der Kopfzeile gesucht, nicht exakt verglichen.
	get := func(prefix string) string {
		if v, ok := headers[prefix]; ok {
			return v
		}
		for k, v := range headers {
			if strings.HasPrefix(k, prefix+".") {
				return v
			}
		}
		return ""
	}

	d := DiscoveredPrinter{
		IP:      srcIP,
		Model:   modellName(get("DEVMODEL")),
		Name:    get("DEVNAME"),
		Connect: strings.ToLower(get("DEVCONNECT")),
		Bind:    strings.ToLower(get("DEVBIND")),
		Version: get("DEVVERSION"),
		Serial:  strings.TrimPrefix(get("USN"), "uuid:"),
	}
	if loc := get("LOCATION"); d.IP == "" && loc != "" {
		if u := strings.TrimPrefix(strings.TrimPrefix(loc, "http://"), "https://"); u != "" {
			d.IP = strings.SplitN(strings.SplitN(u, "/", 2)[0], ":", 2)[0]
		}
	}

	// Ohne Modell und Seriennummer ist es kein Geraet, mit dem wir etwas anfangen
	// koennen — damit fallen fremde SSDP-Teilnehmer heraus.
	if d.Model == "" || d.Serial == "" {
		return DiscoveredPrinter{}, false
	}
	return d, true
}

// searchTargets liefert die Adressen, an die gesucht wird: die SSDP-Multicast-
// Adresse plus die Broadcast-Adresse jedes aktiven IPv4-Netzes. Letzteres hilft
// in Netzen, in denen Multicast zwischen Switches nicht durchgereicht wird.
func searchTargets() []string {
	targets := []string{fmt.Sprintf("%s:%d", ssdpMulticast, ssdpPort)}
	ifaces, err := net.Interfaces()
	if err != nil {
		return targets
	}
	seen := map[string]bool{}
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			bc := broadcastAddr(ipnet)
			if bc == nil || seen[bc.String()] {
				continue
			}
			seen[bc.String()] = true
			targets = append(targets, fmt.Sprintf("%s:%d", bc, ssdpPort))
		}
	}
	return targets
}

func broadcastAddr(n *net.IPNet) net.IP {
	ip := n.IP.To4()
	mask := net.IP(n.Mask).To4()
	if ip == nil || mask == nil {
		return nil
	}
	bc := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		bc[i] = ip[i] | ^mask[i]
	}
	return bc
}

// discoverPrinters sendet mehrfach und sammelt die Antworten bis zum Ablauf des
// Zeitfensters ein.
func discoverPrinters(targets []string, rounds int, window time.Duration) ([]DiscoveredPrinter, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("UDP-Socket: %v", err)
	}
	defer conn.Close()

	found := map[string]DiscoveredPrinter{}
	var mu sync.Mutex
	done := make(chan struct{})
	deadline := time.Now().Add(window)

	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return
			}
			if remaining > 200*time.Millisecond {
				remaining = 200 * time.Millisecond
			}
			conn.SetReadDeadline(time.Now().Add(remaining))
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return
			}
			d, ok := parseSSDPResponse(buf[:n], src.IP.String())
			if !ok {
				continue
			}
			mu.Lock()
			key := d.Serial
			if key == "" {
				key = d.IP
			}
			found[key] = d
			mu.Unlock()
		}
	}()

	// Mehrfach senden: ein einzelnes Paket geht bei vielen Geraeten im Netz
	// verlaesslich irgendwo verloren.
	for r := 0; r < rounds; r++ {
		for _, t := range targets {
			addr, err := net.ResolveUDPAddr("udp4", t)
			if err != nil {
				continue
			}
			if _, err := conn.WriteToUDP(buildMSearch(t), addr); err != nil {
				log.Printf("Suche: senden an %s fehlgeschlagen: %v", t, err)
			}
		}
		if r < rounds-1 {
			time.Sleep(400 * time.Millisecond)
		}
	}

	<-done

	mu.Lock()
	defer mu.Unlock()
	out := make([]DiscoveredPrinter, 0, len(found))
	for _, d := range found {
		out = append(out, d)
	}
	return out, nil
}

// Nur eine Suche gleichzeitig — sonst schickt ein hektischer Klick auf den Knopf
// mehrfach Suchpakete an alle Geraete im Netz.
var discoverMu sync.Mutex

func handleDiscover(w http.ResponseWriter, r *http.Request) {
	window := 4 * time.Second
	if v := r.URL.Query().Get("timeout"); v != "" {
		if secs, err := strconv.ParseFloat(v, 64); err == nil && secs > 0 && secs <= 20 {
			window = time.Duration(secs * float64(time.Second))
		}
	}

	if !discoverMu.TryLock() {
		http.Error(w, "Suche läuft bereits", http.StatusConflict)
		return
	}
	defer discoverMu.Unlock()

	list, err := discoverPrinters(searchTargets(), discoverSearchRounds, window)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Dazu alles, was der Dauerlauscher in den letzten Minuten gehoert hat.
	// Viele Geraete antworten nicht auf M-SEARCH, melden sich aber von selbst —
	// ohne diesen Teil blieb die Suche leer.
	nachSerie := map[string]DiscoveredPrinter{}
	for _, d := range list {
		schluessel := d.Serial
		if schluessel == "" {
			schluessel = d.IP
		}
		nachSerie[schluessel] = d
	}
	ausMithoeren := 0
	for _, d := range gehoerteDrucker(10 * time.Minute) {
		schluessel := d.Serial
		if schluessel == "" {
			schluessel = d.IP
		}
		if _, schon := nachSerie[schluessel]; !schon {
			nachSerie[schluessel] = d
			ausMithoeren++
		}
	}
	list = list[:0]
	for _, d := range nachSerie {
		list = append(list, d)
	}
	sort.Slice(list, func(a, b int) bool {
		if list[a].Name != list[b].Name {
			return list[a].Name < list[b].Name
		}
		return list[a].IP < list[b].IP
	})

	// Bekannte markieren — und dabei den Fall herausarbeiten, der uns die H-Reihe
	// gekostet hat: dieselbe Seriennummer unter neuer Adresse.
	mu.Lock()
	knownIP := map[string]bool{}
	serieZuIP := map[string]string{}
	for _, p := range state.Printers {
		knownIP[p.IP] = true
		if p.Serial != "" {
			serieZuIP[p.Serial] = p.IP
		}
	}
	mu.Unlock()

	umgezogen := 0
	for i := range list {
		alteIP, kennenWir := serieZuIP[list[i].Serial]
		list[i].Known = knownIP[list[i].IP] || (kennenWir && alteIP == list[i].IP)
		if kennenWir && alteIP != list[i].IP {
			list[i].Umgezogen = true
			list[i].AlteIP = alteIP
			list[i].Known = false
			umgezogen++
		}
	}

	log.Printf("Netzwerksuche: %d Geraet(e) (%d nur ueber Mithoeren, %d umgezogen)",
		len(list), ausMithoeren, umgezogen)
	writeJSON(w, map[string]any{
		"printers": list, "seconds": window.Seconds(),
		"aus_mithoeren": ausMithoeren, "umgezogen": umgezogen,
		"empfang": lauschBericht(),
	})
}
