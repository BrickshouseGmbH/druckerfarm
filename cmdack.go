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
// Bisher galt ein Befehl als erfolgreich, sobald er beim MQTT-Broker angenommen
// war. Das heisst nur: die Nachricht ist raus. Ob der Drucker sie ausfuehrt,
// stand auf einem anderen Blatt — und genau das war der gemeldete Fehler:
// die Oberflaeche sagte "pausiert", der Drucker tat nichts, keine Meldung.
//
// Das Protokoll sieht eine Antwort vor: der Drucker schickt denselben Befehl
// mit derselben sequence_id zurueck, dazu "result" und "reason". Darauf wird
// jetzt gewartet.
//
// Wichtig zu wissen: solange ein Drucker mit der Herstellercloud verbunden ist,
// nimmt er ueber die lokale Verbindung zwar Statusabfragen an, aber keine
// Steuerbefehle. Er meldet sich dann gar nicht zurueck. Ohne diese Wartezeit
// war dieser Fall nicht von einem Erfolg zu unterscheiden.

type cmdWarter struct {
	command string
	seq     string
	antwort chan cmdAntwort
}

type cmdAntwort struct {
	Result string
	Reason string
}

var (
	cmdMu     sync.Mutex
	cmdWarten = map[string][]*cmdWarter{} // ip -> offene Warter
	cmdSeq    atomic.Int64
)

func naechsteSeq() string {
	return fmt.Sprintf("%d", 9000+cmdSeq.Add(1)%1000)
}

func warteAufQuittung(ip, command, seq string) chan cmdAntwort {
	w := &cmdWarter{command: command, seq: seq, antwort: make(chan cmdAntwort, 1)}
	cmdMu.Lock()
	cmdWarten[ip] = append(cmdWarten[ip], w)
	cmdMu.Unlock()
	return w.antwort
}

func loeseWarter(ip string, w *cmdWarter) {
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

// pruefeQuittung wird fuer jede eingehende Nachricht aufgerufen und weckt einen
// wartenden Befehl, wenn die Antwort zu ihm gehoert.
func pruefeQuittung(ip string, payload []byte) {
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
	var treffer []*cmdWarter
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
		case x.antwort <- cmdAntwort{Result: w.Print.Result, Reason: w.Print.Reason}:
		default:
		}
	}
}

// cmdWartezeit ist bewusst kurz gehalten: der Drucker antwortet normalerweise
// in weniger als einer Sekunde. Wer laenger schweigt, fuehrt den Befehl nicht
// aus.
const cmdWartezeit = 4 * time.Second

// deuteQuittung macht aus der Antwort einen Satz, mit dem man etwas anfangen
// kann.
func deuteQuittung(a cmdAntwort, ok bool) error {
	if !ok {
		return fmt.Errorf("keine Rückmeldung vom Drucker — steht er auf LAN-Modus? " +
			"Mit der Herstellercloud verbunden nimmt er nur Statusabfragen an, keine Steuerbefehle")
	}
	if strings.EqualFold(a.Result, "success") {
		return nil
	}
	if a.Reason != "" {
		// Die Firmware ab 01.08.05 lehnt Fremdsteuerung ab, solange kein
		// Developer Mode aktiv ist ("mqtt message verify failed"). Statt der
		// kryptischen Originalmeldung ein Hinweis, der weiterhilft.
		if strings.Contains(strings.ToLower(a.Reason), "verify") {
			return fmt.Errorf("Drucker lehnt die Steuerung ab — am Gerät den Developer Mode " +
				"aktivieren (Einstellungen → WLAN → LAN Mode Only → Developer Mode). Ohne ihn prüft " +
				"die Firmware die Herkunft der Befehle und weist sie ab")
		}
		return fmt.Errorf("Drucker lehnt ab: %s (%s)", a.Reason, a.Result)
	}
	return fmt.Errorf("Drucker lehnt ab: %s", a.Result)
}
