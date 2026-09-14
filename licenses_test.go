package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Der Lizenz-Endpunkt schreibt die eingebettete Datei ins Datenverzeichnis
// (und öffnet sie am echten System im Explorer — das ist im Test nicht prüfbar).
func TestOpenLicensesSchreibtDatei(t *testing.T) {
	if !strings.Contains(thirdPartyLicenses, "gorilla/websocket") {
		t.Fatal("THIRD_PARTY_LICENSES.md scheint nicht eingebettet")
	}
	// Keine Erklärungstexte mehr — nur Nennung + Lizenztexte.
	if strings.Contains(thirdPartyLicenses, "Diese Prüfung") || strings.Contains(thirdPartyLicenses, "Einschätzung") {
		t.Error("Erklärungstexte sollten aus der Lizenzdatei entfernt sein")
	}

	tmp := t.TempDir()
	alt := appDir
	appDir = tmp
	defer func() { appDir = alt }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/open-licenses", nil)
	handleOpenLicenses(rec, req)

	pfad := filepath.Join(tmp, "THIRD_PARTY_LICENSES.md")
	data, err := os.ReadFile(pfad)
	if err != nil {
		t.Fatalf("Datei nicht geschrieben: %v", err)
	}
	if !strings.Contains(string(data), "gorilla/websocket") {
		t.Error("geschriebene Datei hat den erwarteten Inhalt nicht")
	}
}
