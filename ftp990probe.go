package main

import (
	"net"
	"strings"
	"time"
)

// ─── ACTIVE PORT-990 PROBE (incl. 421 "too many connections") ─────────────────
//
// A plain TCP knock on port 990 can only say open / refused / timeout. It cannot
// tell the difference between a healthy FTP server and one that is turning us
// away because this IP already has too many open FTP sessions (vsftpd's
// max_per_ip limit). That matters, because a printer at its limit answers every
// new client with a plaintext "421 There are too many connections from your
// internet address" and hangs up — which, to a TLS client, looks like a broken
// handshake and is easy to misread as "developer/LAN mode off".
//
// probeFTP990 catches it directly. Bambu's port 990 is *implicit* FTPS: a
// healthy server stays silent after the TCP connect and waits for the client's
// TLS ClientHello, so a short read returns nothing — that is the good case. A
// printer over its per-IP limit instead speaks in the clear first, so any bytes
// we read here are the 421 (or another plaintext status) and tell us exactly
// what is wrong before we even attempt TLS.

type ftp990Probe struct {
	State   string // open | refused | timeout | unreachable
	Busy421 bool   // printer answered 421 "too many connections" in the clear
	Banner  string // plaintext greeting, if the server sent one
}

func probeFTP990(ip string) ftp990Probe {
	res := ftp990Probe{}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "990"), 5*time.Second)
	if err != nil {
		low := strings.ToLower(err.Error())
		switch {
		case strings.Contains(low, "refused"):
			res.State = "refused"
		case strings.Contains(low, "timeout"), strings.Contains(low, "deadline"):
			res.State = "timeout"
		case strings.Contains(low, "no route"), strings.Contains(low, "unreachable"):
			res.State = "unreachable"
		default:
			res.State = "timeout"
		}
		return res
	}
	defer conn.Close()
	res.State = "open"

	// Read a short window. Silence (0 bytes / timeout) is the healthy implicit
	// FTPS case. Plaintext bytes mean the server talked before TLS — check for
	// the 421 too-many-connections greeting.
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	if n > 0 {
		res.Banner = strings.TrimSpace(string(buf[:n]))
		if is421(res.Banner) {
			res.Busy421 = true
		}
	}
	return res
}
