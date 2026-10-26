package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ─── FARM-WIDE FTP / STORAGE CHECK ────────────────────────────────────────────
//
// FTP works on some printers and not on others, and the cause differs per
// device: LAN/developer mode off (port 990 closed or refused), a stale access
// code (login rejected), or no SD/USB card inserted (connected but empty). This
// endpoint checks every printer at once and reports, per printer, exactly which
// of these it is — so the offenders can be worked through instead of guessed at.

type ftpCheckResult struct {
	IP       string `json:"ip"`
	Name     string `json:"name"`
	Model    string `json:"model"`
	Port     string `json:"port"`     // open | refused | timeout | unreachable
	Login    string `json:"login"`    // ok | bad_code | no_code | fail | -
	Storage  bool   `json:"storage"`  // at least one folder/file found on the card
	Files    int    `json:"files"`    // entries found at the root
	Reason   string `json:"reason"`   // short human-readable cause
	Code     string `json:"code"`     // stable key for the UI translation (ftp*), if any
	Switched bool   `json:"switched"` // stored IP was switched to a sibling that answers on 990
}

// checkOneFTP runs the port + login + storage check for a single printer.
// altIPs are the other addresses SSDP heard for this printer's serial (used to
// recover a dual-NIC printer whose FTP lives on the other interface).
func checkOneFTP(p Printer, altIPs []string) ftpCheckResult {
	r := ftpCheckResult{IP: p.IP, Name: p.Name, Model: p.Model, Login: "-"}

	// 1) Knock on port 990 and, if it answers in the clear, read what it says —
	// this is where a "421 too many connections" is caught explicitly.
	fp := probeFTP990(p.IP)
	// Dual-NIC recovery: the stored IP is dead on 990, but the printer may
	// answer FTP on a sibling address it also announces. Switch to it.
	if fp.State != "open" && len(altIPs) > 0 {
		if newIP, ok := switchToWorkingFTPIP(p, altIPs); ok {
			p.IP = newIP
			r.IP = newIP
			r.Switched = true
			fp = probeFTP990(p.IP)
		}
	}
	switch {
	case fp.Busy421:
		// Reachable, but the printer is refusing new FTP sessions because this
		// IP already has too many open. Not a LAN/developer-mode problem.
		r.Port = "open"
		r.Login = "-"
		r.Code = "ftpBusy"
		r.Reason = "Drucker meldet zu viele FTP-Verbindungen (421) — andere FTP-Clients schließen oder Drucker kurz neu starten"
		return r
	case fp.State == "open":
		r.Port = "open"
	case fp.State == "refused":
		r.Port = "refused"
		r.Code = "ftpRRefused"
		r.Reason = "Port 990 abgelehnt — FTP/LAN-Modus am Drucker aus"
		return r
	case fp.State == "timeout":
		r.Port = "timeout"
		r.Code = "ftpRTimeout"
		r.Reason = "keine Antwort auf Port 990 — LAN-/Entwicklermodus aus oder Cloud-Only"
		return r
	default:
		r.Port = "unreachable"
		r.Code = "ftpRUnreach"
		r.Reason = "nicht erreichbar — anderes Netz/VLAN oder Gerät aus"
		return r
	}

	// 2) Port is open — try to log in and list the root.
	if strings.TrimSpace(p.Code) == "" {
		r.Login = "no_code"
		r.Code = "ftpRNoCode"
		r.Reason = "kein Zugangscode hinterlegt"
		return r
	}

	done := make(chan struct{})
	var out string
	var err error
	go func() {
		out, err = runPythonFTP([]string{"scan", "--ip", p.IP, "--code", p.Code, "--path", "/", "--depth", "1"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(12 * time.Second):
		r.Login = "fail"
		r.Reason = "Zeitüberschreitung bei der Anmeldung"
		r.Code = "ftpTimeout"
		return r
	}

	if err != nil {
		fe := friendlyFTPError(err.Error())
		r.Code = ftpErrKey(fe)
		if r.Code == "ftpBadCode" {
			r.Login = "bad_code"
			r.Reason = "Zugangscode wird nicht angenommen"
		} else {
			r.Login = "fail"
			r.Reason = fe
		}
		return r
	}

	// 3) Login worked. How many entries at the root → is a card present?
	r.Login = "ok"
	var items []struct {
		Path string `json:"path"`
		Dir  bool   `json:"dir"`
	}
	json.Unmarshal([]byte(strings.TrimSpace(out)), &items)
	r.Files = len(items)
	r.Storage = len(items) > 0
	if r.Storage {
		r.Code = "ftpROk"
		r.Reason = "OK"
	} else {
		r.Code = "ftpREmpty"
		r.Reason = "verbunden, aber leer — keine SD-/USB-Karte eingehängt?"
	}
	return r
}

// handleFtpCheck checks every configured printer and returns the per-printer
// result. Runs with limited concurrency so 50 printers do not open 50 FTP
// sessions at once.
func handleFtpCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// First heal stale IPs via SSDP so a printer that moved over DHCP is not
	// reported as "timeout on 990" just because we still hold its old address.
	ipFixed := refreshStoredIPs()

	// Every address SSDP heard per serial — used below to recover a dual-NIC
	// printer whose FTP server sits on a different interface than the one we
	// stored (probed only when its stored IP is dead on 990).
	altBySerial := heardIPsBySerial()

	mu.RLock()
	printers := make([]Printer, len(state.Printers))
	copy(printers, state.Printers)
	mu.RUnlock()

	results := make([]ftpCheckResult, len(printers))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for i, p := range printers {
		wg.Add(1)
		go func(idx int, pr Printer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = checkOneFTP(pr, altBySerial[pr.Serial])
		}(i, p)
	}
	wg.Wait()
	json.NewEncoder(w).Encode(map[string]any{"results": results, "ip_fixed": ipFixed})
}
