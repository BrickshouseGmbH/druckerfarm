package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ─── QUITTUNG FUER DRUCKBEFEHLE ───────────────────────────────────────────────
//
// A command used to count as successful as soon as the MQTT broker accepted
// it. That only means the message went out. Whether the printer executes it
// was a different matter — and that was exactly the reported bug:
// the UI said "paused", the printer did nothing, no message.
//
// The protocol provides a reply: the printer sends the same command
// back with the same sequence_id, plus "result" and "reason". This is
// jetzt gewartet.
//
// Important to know: while a printer is connected to the vendor cloud,
// it accepts status queries over the local connection but no
// control commands. It then does not reply at all. Without this wait
// this case was indistinguishable from a success.

type cmdWaiter struct {
	command string
	seq     string
	antwort chan cmdResponse
}

type cmdResponse struct {
	Result string
	Reason string
}

var (
	cmdMu     sync.Mutex
	cmdWarten = map[string][]*cmdWaiter{} // ip -> offene Warter
	cmdSeq    atomic.Int64
)

func nextSeq() string {
	return fmt.Sprintf("%d", 9000+cmdSeq.Add(1)%1000)
}

func waitForAck(ip, command, seq string) chan cmdResponse {
	w := &cmdWaiter{command: command, seq: seq, antwort: make(chan cmdResponse, 1)}
	cmdMu.Lock()
	cmdWarten[ip] = append(cmdWarten[ip], w)
	cmdMu.Unlock()
	return w.antwort
}

func releaseWaiter(ip string, w *cmdWaiter) {
	cmdMu.Lock()
	liste := cmdWarten[ip]
	for i, x := range liste {
		if x == w {
			cmdWarten[ip] = append(liste[:i], liste[i+1:]...)
			break
		}
	}
	if len(cmdWarten[ip]) == 0 {
		delete(cmdWarten, ip)
	}
	cmdMu.Unlock()
}

// checkAck is called for every incoming message and wakes a
// waiting command when the reply belongs to it.
func checkAck(ip string, payload []byte) {
	cmdMu.Lock()
	offen := len(cmdWarten[ip])
	cmdMu.Unlock()
	if offen == 0 {
		return
	}

	var w struct {
		Print struct {
			Command    string `json:"command"`
			SequenceID string `json:"sequence_id"`
			Result     string `json:"result"`
			Reason     string `json:"reason"`
		} `json:"print"`
	}
	if err := json.Unmarshal(payload, &w); err != nil {
		return
	}
	if w.Print.Command == "" || w.Print.Result == "" {
		return
	}

	cmdMu.Lock()
	liste := cmdWarten[ip]
	rest := liste[:0]
	var treffer []*cmdWaiter
	for _, x := range liste {
		if x.command == w.Print.Command && (x.seq == w.Print.SequenceID || w.Print.SequenceID == "") {
			treffer = append(treffer, x)
		} else {
			rest = append(rest, x)
		}
	}
	cmdWarten[ip] = rest
	if len(cmdWarten[ip]) == 0 {
		delete(cmdWarten, ip)
	}
	cmdMu.Unlock()

	for _, x := range treffer {
		select {
		case x.antwort <- cmdResponse{Result: w.Print.Result, Reason: w.Print.Reason}:
		default:
		}
	}
}

// cmdTimeout is kept deliberately short: the printer normally replies
// in under a second. Anything silent longer is not executing the command
// aus.
const cmdTimeout = 4 * time.Second

// interpretAck turns the reply into a sentence one can act on
func interpretAck(a cmdResponse, ok bool) error {
	if !ok {
		return fmt.Errorf("keine Rückmeldung vom Drucker — steht er auf LAN-Modus? " +
			"Mit der Herstellercloud verbunden nimmt er nur Statusabfragen an, keine Steuerbefehle")
	}
	if strings.EqualFold(a.Result, "success") {
		return nil
	}
	if a.Reason != "" {
		// Firmware from 01.08.05 on rejects external control while no
		// Developer Mode is active ("mqtt message verify failed"). Instead of the
		// cryptic original message, a hint that helps.
		if strings.Contains(strings.ToLower(a.Reason), "verify") {
			return fmt.Errorf("Drucker lehnt die Steuerung ab — am Gerät den Developer Mode " +
				"aktivieren (Einstellungen → WLAN → LAN Mode Only → Developer Mode). Ohne ihn prüft " +
				"die Firmware die Herkunft der Befehle und weist sie ab")
		}
		return fmt.Errorf("Drucker lehnt ab: %s (%s)", a.Reason, a.Result)
	}
	return fmt.Errorf("Drucker lehnt ab: %s", a.Result)
}
