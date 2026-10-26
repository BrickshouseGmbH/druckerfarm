package main

import (
	"strings"
	"testing"
	"time"
)

// The acknowledgement is the core of the fix: a command is done only
// when the printer confirms it. Previously "sent" counted as "done" —
// so the UI reported success while nothing happened.
func TestQuittungErfolg(t *testing.T) {
	ip := "10.1.1.1"
	warten := waitForAck(ip, "pause", "9001")
	checkAck(ip, []byte(`{"print":{"command":"pause","sequence_id":"9001","result":"success","reason":""}}`))
	select {
	case a := <-warten:
		if err := interpretAck(a, true); err != nil {
			t.Fatalf("Erfolg wurde als Fehler gedeutet: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("keine Quittung durchgereicht")
	}
}

func TestQuittungAblehnung(t *testing.T) {
	ip := "10.1.1.2"
	warten := waitForAck(ip, "stop", "9002")
	checkAck(ip, []byte(`{"print":{"command":"stop","sequence_id":"9002","result":"FAIL","reason":"device is busy"}}`))
	a := <-warten
	err := interpretAck(a, true)
	if err == nil {
		t.Fatal("Ablehnung wurde als Erfolg gedeutet")
	}
	if !strings.Contains(err.Error(), "device is busy") {
		t.Fatalf("Grund fehlt in der Meldung: %v", err)
	}
}

// The most important case: the printer is silent. Exactly that happens while it
// is connected to the vendor cloud — it accepts status queries but no
// control commands, and does not answer them at all.
func TestQuittungSchweigenWirdErklaert(t *testing.T) {
	err := interpretAck(cmdResponse{}, false)
	if err == nil {
		t.Fatal("Schweigen muss ein Fehler sein")
	}
	for _, wort := range []string{"keine Rückmeldung", "LAN"} {
		if !strings.Contains(err.Error(), wort) {
			t.Fatalf("Meldung hilft nicht weiter (%q fehlt): %v", wort, err)
		}
	}
}

// A reply to a different command must not wake the waiter.
func TestQuittungFremderBefehlWecktNicht(t *testing.T) {
	ip := "10.1.1.3"
	warten := waitForAck(ip, "pause", "9003")
	checkAck(ip, []byte(`{"print":{"command":"resume","sequence_id":"9003","result":"success"}}`))
	checkAck(ip, []byte(`{"print":{"command":"pause","sequence_id":"9099","result":"success"}}`))
	select {
	case <-warten:
		t.Fatal("falsche Antwort hat den Warter geweckt")
	case <-time.After(200 * time.Millisecond):
	}
	// The correct reply comes through
	checkAck(ip, []byte(`{"print":{"command":"pause","sequence_id":"9003","result":"success"}}`))
	select {
	case <-warten:
	case <-time.After(time.Second):
		t.Fatal("richtige Antwort kam nicht durch")
	}
}

// Status messages without "result" must trigger nothing — many of
// hunderte pro Minute.
func TestQuittungIgnoriertStatusmeldungen(t *testing.T) {
	ip := "10.1.1.4"
	warten := waitForAck(ip, "pause", "9004")
	checkAck(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":42}}`))
	checkAck(ip, []byte(`{"print":{"command":"push_status","sequence_id":"1"}}`))
	select {
	case <-warten:
		t.Fatal("Statusmeldung wurde als Quittung gewertet")
	case <-time.After(200 * time.Millisecond):
	}
}

// Firmware from 01.08.05 rejects external control with "mqtt message verify
// failed". The user should not see the original message but the
// hint about Developer Mode.
func TestQuittungVerifyFehlerNenntDeveloperMode(t *testing.T) {
	err := interpretAck(cmdResponse{Result: "failed", Reason: "mqtt message verify failed"}, true)
	if err == nil {
		t.Fatal("verify-Fehler muss ein Fehler sein")
	}
	if !strings.Contains(err.Error(), "Developer Mode") {
		t.Fatalf("Hinweis auf Developer Mode fehlt: %v", err)
	}
}
