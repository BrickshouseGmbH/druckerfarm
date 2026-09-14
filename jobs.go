package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── SUCH- UND LÖSCHVORGÄNGE ──────────────────────────────────────────────────
//
// Beides läuft über viele Drucker und dauert. Statt einmal am Ende zu antworten,
// wird ein Vorgang gestartet, dessen Zwischenstand die Oberfläche abfragt — so
// erscheinen Treffer, sobald der jeweilige Drucker geantwortet hat, und ein
// laufender Vorgang lässt sich abbrechen.

const (
	jobParallel   = 6                // gleichzeitige FTP-Sitzungen
	jobPerPrinter = 40 * time.Second // Zeitfenster je Drucker
	jobKeepFor    = 10 * time.Minute // wie lange ein beendeter Vorgang abrufbar bleibt
)

type sdSearchHit struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	File  string `json:"file"`
	Size  int64  `json:"size"`
	Model string `json:"model"`
}

type jobFailure struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	File  string `json:"file,omitempty"`
	Error string `json:"error"`
}

type job struct {
	mu       sync.Mutex
	id       string
	kind     string // search | delete
	query    string
	total    int
	done     int
	hits     []sdSearchHit
	failures []jobFailure
	deleted  int
	running  bool
	stopped  bool
	cancel   chan struct{}
	finished time.Time
}

var (
	jobsMu sync.Mutex
	jobs   = map[string]*job{}
	jobSeq int
)

func newJob(kind, query string, total int) *job {
	jobsMu.Lock()
	defer jobsMu.Unlock()

	// Abgelaufene Vorgänge aufräumen, damit die Sammlung nicht wächst
	for id, j := range jobs {
		j.mu.Lock()
		expired := !j.running && !j.finished.IsZero() && time.Since(j.finished) > jobKeepFor
		j.mu.Unlock()
		if expired {
			delete(jobs, id)
		}
	}

	jobSeq++
	j := &job{
		id:      fmt.Sprintf("%s-%d-%d", kind, time.Now().Unix(), jobSeq),
		kind:    kind,
		query:   query,
		total:   total,
		running: true,
		cancel:  make(chan struct{}),
	}
	jobs[j.id] = j
	return j
}

func getJob(id string) *job {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	return jobs[id]
}

func (j *job) stop() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.stopped {
		return
	}
	j.stopped = true
	close(j.cancel)
}

func (j *job) cancelled() bool {
	select {
	case <-j.cancel:
		return true
	default:
		return false
	}
}

func (j *job) addHits(h []sdSearchHit) {
	j.mu.Lock()
	j.hits = append(j.hits, h...)
	j.done++
	j.mu.Unlock()
}

func (j *job) addFailure(f jobFailure) {
	j.mu.Lock()
	j.failures = append(j.failures, f)
	j.done++
	j.mu.Unlock()
}

func (j *job) finish() {
	j.mu.Lock()
	j.running = false
	j.finished = time.Now()
	j.mu.Unlock()
}

// snapshot liefert den Stand ab dem angegebenen Treffer — die Oberfläche holt
// so nur das Neue und hängt es an, statt die Liste jedes Mal neu zu bauen.
func (j *job) snapshot(sinceHits, sinceFails int) map[string]any {
	j.mu.Lock()
	defer j.mu.Unlock()

	newHits := []sdSearchHit{}
	if sinceHits < len(j.hits) {
		newHits = append(newHits, j.hits[sinceHits:]...)
	}
	newFails := []jobFailure{}
	if sinceFails < len(j.failures) {
		newFails = append(newFails, j.failures[sinceFails:]...)
	}
	return map[string]any{
		"id":           j.id,
		"kind":         j.kind,
		"query":        j.query,
		"total":        j.total,
		"done":         j.done,
		"running":      j.running,
		"stopped":      j.stopped,
		"deleted":      j.deleted,
		"hits":         newHits,
		"failed":       newFails,
		"total_hits":   len(j.hits),
		"total_failed": len(j.failures),
		"complete":     !j.running,
	}
}

// ─── SUCHE ────────────────────────────────────────────────────────────────────

func runSearch(j *job, targets []Printer, q string) {
	defer j.finish()

	sem := make(chan struct{}, jobParallel)
	var wg sync.WaitGroup
	for _, p := range targets {
		if j.cancelled() {
			break
		}
		wg.Add(1)
		go func(pr Printer) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-j.cancel:
				return
			}
			defer func() { <-sem }()
			if j.cancelled() {
				return
			}

			type res struct {
				files []SDFileInfo
				err   error
			}
			ch := make(chan res, 1)
			go func() {
				files, err := listSD(pr)
				ch <- res{files, err}
			}()

			select {
			case r := <-ch:
				if r.err != nil {
					j.addFailure(jobFailure{IP: pr.IP, Name: pr.Name, Error: r.err.Error()})
					return
				}
				var hits []sdSearchHit
				for _, f := range r.files {
					if strings.Contains(strings.ToLower(f.Name), q) {
						hits = append(hits, sdSearchHit{IP: pr.IP, Name: pr.Name, File: f.Name, Size: f.Size, Model: pr.Model})
					}
				}
				j.addHits(hits)
			case <-time.After(jobPerPrinter):
				j.addFailure(jobFailure{IP: pr.IP, Name: pr.Name,
					Error: fmt.Sprintf("keine Antwort binnen %.0f s", jobPerPrinter.Seconds())})
			case <-j.cancel:
			}
		}(p)
	}
	wg.Wait()
}

func handleSearchStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Query string   `json:"q"`
		IPs   []string `json:"ips"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	q := strings.ToLower(strings.TrimSpace(body.Query))
	if len([]rune(q)) < 3 {
		http.Error(w, "Suchbegriff braucht mindestens 3 Zeichen", http.StatusBadRequest)
		return
	}

	wanted := map[string]bool{}
	for _, ip := range body.IPs {
		if ip = strings.TrimSpace(ip); ip != "" {
			wanted[ip] = true
		}
	}
	mu.Lock()
	var targets []Printer
	for _, p := range state.Printers {
		if len(wanted) == 0 || wanted[p.IP] {
			targets = append(targets, p)
		}
	}
	mu.Unlock()
	if len(targets) == 0 {
		http.Error(w, "kein Drucker im Filter", http.StatusBadRequest)
		return
	}

	j := newJob("search", q, len(targets))
	go runSearch(j, targets, q)
	writeJSON(w, map[string]any{"id": j.id, "total": len(targets)})
}

// ─── LÖSCHEN ──────────────────────────────────────────────────────────────────

type deleteItem struct {
	IP   string `json:"ip"`
	File string `json:"file"`
}

func runDelete(j *job, byPrinter map[string][]string, names map[string]string) {
	defer j.finish()

	sem := make(chan struct{}, jobParallel)
	var wg sync.WaitGroup
	for ip, files := range byPrinter {
		if j.cancelled() {
			break
		}
		wg.Add(1)
		go func(ip string, files []string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-j.cancel:
				return
			}
			defer func() { <-sem }()

			// Alle Dateien eines Druckers in einer FTP-Sitzung
			res, err := deleteSDFiles(ip, files)
			if err != nil {
				for _, f := range files {
					j.addFailure(jobFailure{IP: ip, Name: names[ip], File: f, Error: err.Error()})
				}
				return
			}
			for _, r := range res {
				if r.OK {
					j.mu.Lock()
					j.deleted++
					j.done++
					j.mu.Unlock()
					continue
				}
				j.addFailure(jobFailure{IP: ip, Name: names[ip], File: r.Path, Error: r.Error})
			}
		}(ip, files)
	}
	wg.Wait()
	log.Printf("Löschvorgang %s beendet: %d entfernt, %d Fehler", j.id, j.deleted, len(j.failures))
}

func handleDeleteStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Items []deleteItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body.Items) == 0 {
		http.Error(w, "nichts ausgewählt", http.StatusBadRequest)
		return
	}

	mu.Lock()
	known := map[string]string{}
	for _, p := range state.Printers {
		known[p.IP] = p.Name
	}
	mu.Unlock()

	byPrinter := map[string][]string{}
	for _, it := range body.Items {
		if it.IP == "" || it.File == "" {
			continue
		}
		if _, ok := known[it.IP]; !ok {
			continue // unbekannter Drucker — nicht blind löschen
		}
		byPrinter[it.IP] = append(byPrinter[it.IP], it.File)
	}
	n := 0
	for _, f := range byPrinter {
		n += len(f)
	}
	if n == 0 {
		http.Error(w, "keine gültigen Einträge", http.StatusBadRequest)
		return
	}

	j := newJob("delete", "", n)
	go runDelete(j, byPrinter, known)
	writeJSON(w, map[string]any{"id": j.id, "total": n, "printers": len(byPrinter)})
}

// ─── Status und Abbruch ───────────────────────────────────────────────────────

// listJobs fasst alle laufenden und kuerzlich beendeten Vorgaenge zusammen —
// die Registry-Jobs (Suche/Loeschen) plus den Datei-Sync als synthetischen
// Eintrag. Damit kann die Oberflaeche ein gemeinsames "Auftraege"-Panel zeigen.
func listJobs() []map[string]any {
	out := []map[string]any{}

	// Datei-Sync laeuft ueber ein eigenes Zustandsobjekt, nicht ueber die
	// Registry — hier als Auftrag mit aufgenommen, solange er aktiv ist.
	syncMu.Lock()
	if syncProgress.Status == SyncRunning || syncProgress.Status == SyncPaused {
		out = append(out, map[string]any{
			"id": "sync", "kind": "sync", "query": syncProgress.Current,
			"total": syncProgress.Total, "done": syncProgress.Done,
			"running": true, "stopped": syncProgress.Status == SyncPaused,
			"failed": len(syncProgress.Errors), "deleted": 0, "finished": int64(0),
		})
	}
	syncMu.Unlock()

	jobsMu.Lock()
	list := make([]*job, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, j)
	}
	jobsMu.Unlock()

	sort.Slice(list, func(a, b int) bool {
		list[a].mu.Lock()
		ra, fa := list[a].running, list[a].finished
		list[a].mu.Unlock()
		list[b].mu.Lock()
		rb, fb := list[b].running, list[b].finished
		list[b].mu.Unlock()
		if ra != rb {
			return ra // laufende zuerst
		}
		return fa.After(fb) // dann die zuletzt beendeten
	})

	for _, j := range list {
		j.mu.Lock()
		var fin int64
		if !j.finished.IsZero() {
			fin = j.finished.Unix()
		}
		out = append(out, map[string]any{
			"id": j.id, "kind": j.kind, "query": j.query,
			"total": j.total, "done": j.done,
			"running": j.running, "stopped": j.stopped,
			"failed": len(j.failures), "deleted": j.deleted, "finished": fin,
		})
		j.mu.Unlock()
	}
	return out
}

func handleJobsList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"jobs": listJobs()})
}

func handleJobStatus(w http.ResponseWriter, r *http.Request) {
	j := getJob(r.URL.Query().Get("id"))
	if j == nil {
		http.Error(w, "unbekannter Vorgang", http.StatusNotFound)
		return
	}
	atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }
	writeJSON(w, j.snapshot(atoi(r.URL.Query().Get("hits")), atoi(r.URL.Query().Get("failed"))))
}

func handleJobStop(w http.ResponseWriter, r *http.Request) {
	j := getJob(r.URL.Query().Get("id"))
	if j == nil {
		http.Error(w, "unbekannter Vorgang", http.StatusNotFound)
		return
	}
	j.stop()
	writeJSON(w, map[string]any{"stopped": true})
}
