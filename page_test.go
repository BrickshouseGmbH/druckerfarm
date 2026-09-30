package main

import (
	"strings"
	"testing"
)

// The version number must be inserted into the page server-side so it
// already shows on the splash (gorilla loading page).
func TestSeiteMitVersionErsetztPlatzhalter(t *testing.T) {
	html := pageWithVersion()
	if strings.Contains(html, "__APP_VERSION__") {
		t.Error("Platzhalter __APP_VERSION__ blieb in der ausgelieferten Seite stehen")
	}
	if !strings.Contains(html, "Version "+appVersion) {
		t.Errorf("erwartete 'Version %s' in der Seite, nicht gefunden", appVersion)
	}
	// Copyright without the old addendum.
	if strings.Contains(html, "Reinhard @ Brickshouse") {
		t.Error("altes Copyright mit 'Reinhard @' noch vorhanden")
	}
	if !strings.Contains(html, "© Brickshouse GmbH · Druckerfarm.ch") {
		t.Error("neues Copyright nicht gefunden")
	}
	// Translations must be injected — placeholder gone, keys present.
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
