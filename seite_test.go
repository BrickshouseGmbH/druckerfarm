package main

import (
	"strings"
	"testing"
)

// Die Versionsnummer muss serverseitig in die Seite eingesetzt werden, damit sie
// schon auf dem Startbild (Gorilla-Ladeseite) steht.
func TestSeiteMitVersionErsetztPlatzhalter(t *testing.T) {
	html := seiteMitVersion()
	if strings.Contains(html, "__APP_VERSION__") {
		t.Error("Platzhalter __APP_VERSION__ blieb in der ausgelieferten Seite stehen")
	}
	if !strings.Contains(html, "Version "+appVersion) {
		t.Errorf("erwartete 'Version %s' in der Seite, nicht gefunden", appVersion)
	}
	// Copyright ohne den alten Zusatz.
	if strings.Contains(html, "Reinhard @ Brickshouse") {
		t.Error("altes Copyright mit 'Reinhard @' noch vorhanden")
	}
	if !strings.Contains(html, "© Brickshouse GmbH · Druckerfarm.ch") {
		t.Error("neues Copyright nicht gefunden")
	}
	// Übersetzungen müssen injiziert sein — Platzhalter weg, Schlüssel da.
	if strings.Contains(html, "__LANGS_JSON__") {
		t.Error("Platzhalter __LANGS_JSON__ blieb stehen — Sprachdateien nicht injiziert")
	}
	if !strings.Contains(html, "pillPaused") || !strings.Contains(html, "Paused") {
		t.Error("englische Übersetzungen (en.json) nicht in der Seite")
	}
	if !strings.Contains(html, "Pausiert") {
		t.Error("deutsche Übersetzungen (de.json) nicht in der Seite")
	}
	if strings.Contains(html, "__DEEN_JSON__") {
		t.Error("Platzhalter __DEEN_JSON__ blieb stehen — Übersetzungstabelle nicht injiziert")
	}
	if !strings.Contains(html, "Saved printers") {
		t.Error("de-en.json (automatische Übersetzung) nicht in der Seite")
	}
}
