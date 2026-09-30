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

// lockFTP serialises FTP access per printer IP and returns the unlock func.
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
	return m.Unlock
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

// ftpRetryDelays are the waits between attempts when a 421 is hit — enough for
// the printer to time out a stale session and free a slot.
var ftpRetryDelays = []time.Duration{2 * time.Second, 4 * time.Second}
