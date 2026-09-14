package main

import (
	"strings"
	"testing"
	"time"
)

// Die Quittung ist der Kern der Reparatur: ein Befehl gilt erst als erledigt,
// wenn der Drucker ihn bestaetigt. Vorher galt "abgeschickt" als "erledigt" —
// deshalb meldete die Oberflaeche Erfolg, waehrend nichts geschah.
func TestQuittungErfolg(t *testing.T) {
	ip := "10.1.1.1"
	warten := warteAufQuittung(ip, "pause", "9001")
	pruefeQuittung(ip, []byte(`{"print":{"command":"pause","sequence_id":"9001","result":"success","reason":""}}`))
	select {
	case a := <-warten:
		if err := deuteQuittung(a, true); err != nil {
			t.Fatalf("Erfolg wurde als Fehler gedeutet: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("keine Quittung durchgereicht")
	}
}

func TestQuittungAblehnung(t *testing.T) {
	ip := "10.1.1.2"
	warten := warteAufQuittung(ip, "stop", "9002")
	pruefeQuittung(ip, []byte(`{"print":{"command":"stop","sequence_id":"9002","result":"FAIL","reason":"device is busy"}}`))
	a := <-warten
	err := deuteQuittung(a, true)
	if err == nil {
		t.Fatal("Ablehnung wurde als Erfolg gedeutet")
	}
	if !strings.Contains(err.Error(), "device is busy") {
		t.Fatalf("Grund fehlt in der Meldung: %v", err)
	}
}

// Der wichtigste Fall: der Drucker schweigt. Genau das passiert, solange er mit
// der Herstellercloud verbunden ist — er nimmt Statusabfragen an, aber keine
// Steuerbefehle, und antwortet auf sie gar nicht.
func TestQuittungSchweigenWirdErklaert(t *testing.T) {
	err := deuteQuittung(cmdAntwort{}, false)
	if err == nil {
		t.Fatal("Schweigen muss ein Fehler sein")
	}
	for _, wort := range []string{"keine Rückmeldung", "LAN"} {
		if !strings.Contains(err.Error(), wort) {
			t.Fatalf("Meldung hilft nicht weiter (%q fehlt): %v", wort, err)
		}
	}
}

// Eine Antwort auf einen anderen Befehl darf den Warter nicht wecken.
func TestQuittungFremderBefehlWecktNicht(t *testing.T) {
	ip := "10.1.1.3"
	warten := warteAufQuittung(ip, "pause", "9003")
	pruefeQuittung(ip, []byte(`{"print":{"command":"resume","sequence_id":"9003","result":"success"}}`))
	pruefeQuittung(ip, []byte(`{"print":{"command":"pause","sequence_id":"9099","result":"success"}}`))
	select {
	case <-warten:
		t.Fatal("falsche Antwort hat den Warter geweckt")
	case <-time.After(200 * time.Millisecond):
	}
	// Die richtige Antwort kommt durch
	pruefeQuittung(ip, []byte(`{"print":{"command":"pause","sequence_id":"9003","result":"success"}}`))
	select {
	case <-warten:
	case <-time.After(time.Second):
		t.Fatal("richtige Antwort kam nicht durch")
	}
}

// Statusmeldungen ohne "result" duerfen nichts ausloesen — davon kommen
// hunderte pro Minute.
func TestQuittungIgnoriertStatusmeldungen(t *testing.T) {
	ip := "10.1.1.4"
	warten := warteAufQuittung(ip, "pause", "9004")
	pruefeQuittung(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":42}}`))
	pruefeQuittung(ip, []byte(`{"print":{"command":"push_status","sequence_id":"1"}}`))
	select {
	case <-warten:
		t.Fatal("Statusmeldung wurde als Quittung gewertet")
	case <-time.After(200 * time.Millisecond):
	}
}

// Die Firmware ab 01.08.05 lehnt Fremdsteuerung mit "mqtt message verify
// failed" ab. Der Anwender soll nicht die Originalmeldung sehen, sondern den
// Hinweis auf den Developer Mode.
func TestQuittungVerifyFehlerNenntDeveloperMode(t *testing.T) {
	err := deuteQuittung(cmdAntwort{Result: "failed", Reason: "mqtt message verify failed"}, true)
	if err == nil {
		t.Fatal("verify-Fehler muss ein Fehler sein")
	}
	if !strings.Contains(err.Error(), "Developer Mode") {
		t.Fatalf("Hinweis auf Developer Mode fehlt: %v", err)
	}
}
