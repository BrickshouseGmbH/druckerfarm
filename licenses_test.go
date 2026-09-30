package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The license endpoint writes the embedded file into the data directory
// (and opens it in Explorer on the real system — not testable here).
func TestOpenLicensesSchreibtDatei(t *testing.T) {
	if !strings.Contains(thirdPartyLicenses, "gorilla/websocket") {
		t.Fatal("THIRD_PARTY_LICENSES.md scheint nicht eingebettet")
	}
	// No more explanatory text — only naming + license texts.
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
