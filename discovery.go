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
// The printers respond to SSDP M-SEARCH — but on port 2021, not on
// dem ueblichen 1900. Die Antwort kommt als Unicast zurueck ("HTTP/1.1 200 OK")
// and carry the device data in custom headers: DevModel, DevName,
// DevConnect, DevBind, USN (serial number) and DevVersion.
//
// Two things matter here:
//   - UDP gets lost. With ~36 devices on the network a single attempt is
//     not enough, so it is sent multiple times.
//   - Der Suchtyp bleibt "ssdp:all". Damit antwortet jedes SSDP-Geraet; gefiltert
//     via the headers only printers send.

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
	Known   bool   `json:"known"` // already in the printer list

	// Moved: the same serial number is known, but at a different address.
	// Exactly this case paralysed half the farm without anyone
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

// parseSSDPResponse parses a response. The second return value is false
// when it is recognisably not a printer — on the network
// routers, TVs and other kinds of printers also answer SSDP.
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

	// Depending on firmware the fields are "DevModel" or "DevModel.suffix" —
	// so the header prefix is searched for, not compared exactly.
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

	// Without model and serial it is not a device we can do anything with
	// — this drops foreign SSDP participants.
	if d.Model == "" || d.Serial == "" {
		return DiscoveredPrinter{}, false
	}
	return d, true
}

// searchTargets returns the addresses to search: the SSDP multicast
// address plus the broadcast address of each active IPv4 network. The latter helps
// on networks where multicast is not forwarded between switches.
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

// discoverPrinters sends multiple times and collects the responses until the
// time window elapses.
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

	// Send multiple times: a single packet gets lost with many devices on the net
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

// Only one search at a time — otherwise a frantic click on the button sends
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

	// Plus everything the continuous listener heard in the last few minutes.
	// Many devices do not answer M-SEARCH but announce themselves —
	// without this part the search stayed empty.
	nachSerie := map[string]DiscoveredPrinter{}
	for _, d := range list {
		schluessel := d.Serial
		if schluessel == "" {
			schluessel = d.IP
		}
		nachSerie[schluessel] = d
	}
	ausMithoeren := 0
	for _, d := range heardPrinters(10 * time.Minute) {
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

	// Mark known ones — and work out the case that got the H series
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
