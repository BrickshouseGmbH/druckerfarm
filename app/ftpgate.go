package main

import (
	"strings"
	"sync"
	"time"
)

// ─── FTP CONNECTION SERIALISATION + 421 BACKOFF ───────────────────────────────
//
// The printer's FTP server (vsftpd) limits how many connections a single client
// IP may hold at once (max_per_ip). When the tool opens several FTP sessions to
// the SAME printer at the same time — the FTP/storage check, the media browser
// and a sync all at once — it runs into that limit and the printer answers every
// further attempt with the plaintext line:
//
//	421 There are too many connections from your internet address.
//
// That 421 arrives BEFORE the TLS handshake, which is why implicit-TLS clients
// then report confusing "wrong version number" / "unexpected TLS packet" errors.
// It is not a login, TLS or firmware problem — just too many parallel sessions.
//
// Two measures fix it: serialise FTP per printer IP (only one session at a time
// to a given printer), and, on a 421, wait briefly and retry instead of failing.

var (
	ftpLocksMu sync.Mutex
	ftpLocks   = map[string]*sync.Mutex{}
)

// ftpGlobalMax caps how many FTP sessions may run across the WHOLE farm at the
// same time. Set to 1: the printers' vsftpd is so quick to hit "421 too many
// connections" (and so slow to free stale sessions) that talking to more than
// one printer at a time is what caused the farm-wide 421 storms. One session at
// a time is slower but means the tool never opens a second FTP connection while
// another is still in flight — no restarts needed. Per-printer serialisation
// (lockFTP) still applies on top. Uploads reuse ONE session for all of a
// printer's files (see uploadmany), so this ceiling does not slow a batch.
const ftpGlobalMax = 1

var ftpSem = make(chan struct{}, ftpGlobalMax)

// acquireFTPSlot blocks until a farm-wide FTP slot is free and returns the
// release func. Always call AFTER lockFTP(ip) so a slot is held by at most one
// in-flight operation per printer.
func acquireFTPSlot() func() {
	ftpSem <- struct{}{}
	return func() { <-ftpSem }
}

// lockFTP serialises FTP access per printer IP and returns the unlock func.
// After acquiring the per-printer lock it also waits out a short COOL-DOWN since
// that printer's previous FTP session ended (D-FTP-C): the printer's vsftpd is
// slow to release a just-closed session, so hammering it with the next connect
// immediately is what tips it into a lingering "421 too many connections". The
// unlock records the release time so the next caller's cool-down is measured
// from here.
func lockFTP(ip string) func() {
	if ip == "" {
		return func() {}
	}
	ftpLocksMu.Lock()
	m, ok := ftpLocks[ip]
	if !ok {
		m = &sync.Mutex{}
		ftpLocks[ip] = m
	}
	ftpLocksMu.Unlock()
	m.Lock()
	waitFTPCooldown(ip)
	return func() {
		markFTPUsed(ip)
		m.Unlock()
	}
}

// ─── PER-PRINTER COOL-DOWN (D-FTP-C) ──────────────────────────────────────────
const ftpCooldown = 1500 * time.Millisecond

var (
	ftpLastMu  sync.Mutex
	ftpLastUse = map[string]time.Time{}
)

func markFTPUsed(ip string) {
	ftpLastMu.Lock()
	ftpLastUse[ip] = time.Now()
	ftpLastMu.Unlock()
}

// waitFTPCooldown sleeps until at least ftpCooldown has passed since this
// printer's last FTP session ended.
func waitFTPCooldown(ip string) {
	ftpLastMu.Lock()
	last, ok := ftpLastUse[ip]
	ftpLastMu.Unlock()
	if !ok {
		return
	}
	if d := ftpCooldown - time.Since(last); d > 0 {
		time.Sleep(d)
	}
}

// ─── PER-PRINTER CIRCUIT BREAKER (D-FTP-D) ────────────────────────────────────
//
// If a printer answers 421 repeatedly, retrying only pours fuel on the fire.
// After ftpTripThreshold consecutive 421s we stop talking FTP to that printer
// for ftpBlockDuration and fail fast with a clear message — no printer restart
// needed, it recovers on its own once vsftpd frees its stale sessions.
const (
	ftpTripThreshold = 3
	ftpBlockDuration = 3 * time.Minute
)

var (
	ftpCBMu    sync.Mutex
	ftp421Runs = map[string]int{}
	ftpBlocked = map[string]time.Time{}
)

// ftpBlockedUntil reports whether FTP to ip is currently circuit-broken.
func ftpBlockedUntil(ip string) (time.Time, bool) {
	ftpCBMu.Lock()
	defer ftpCBMu.Unlock()
	until, ok := ftpBlocked[ip]
	if ok && time.Now().Before(until) {
		return until, true
	}
	if ok {
		delete(ftpBlocked, ip) // expired
	}
	return time.Time{}, false
}

// ftpNote421 records a 421 for ip and trips the breaker past the threshold.
func ftpNote421(ip string) {
	if ip == "" {
		return
	}
	ftpCBMu.Lock()
	ftp421Runs[ip]++
	if ftp421Runs[ip] >= ftpTripThreshold {
		ftpBlocked[ip] = time.Now().Add(ftpBlockDuration)
	}
	ftpCBMu.Unlock()
}

// ftpNoteOK clears the 421 streak for ip after any successful FTP exchange.
func ftpNoteOK(ip string) {
	if ip == "" {
		return
	}
	ftpCBMu.Lock()
	delete(ftp421Runs, ip)
	delete(ftpBlocked, ip)
	ftpCBMu.Unlock()
}

// argIP pulls the value after "--ip" out of the python helper's argument list,
// so the runners know which printer they are talking to.
func argIP(args []string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--ip" {
			return args[i+1]
		}
	}
	return ""
}

// is421 reports whether a message is the vsftpd per-IP connection-limit reply.
func is421(s string) bool {
	low := strings.ToLower(s)
	return strings.Contains(low, "421") &&
		(strings.Contains(low, "too many") || strings.Contains(low, "connection"))
}

// firstArg returns the python helper's action (args[0]) or "".
func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
