package main

import "testing"

// The Python traceback used to land unchanged in the UI. These
// Faelle stammen aus echten Ausgaben des Hilfsskripts.
func TestFriendlyFTPError(t *testing.T) {
	cases := []struct {
		name, raw, want string
	}{
		{"Zeitueberschreitung",
			"Traceback (most recent call last):\n  File \"x.py\", line 40, in <module>\n    ftp.connect(host, 990)\nTimeoutError: [WinError 10060] Ein Verbindungsversuch ist fehlgeschlagen\n",
			"Zeitüberschreitung auf Port 990 — am Drucker den LAN-/Entwicklermodus einschalten (Einstellungen › Allgemein) und sicherstellen, dass er nicht nur im Cloud-Modus läuft. Einen anderen FTP-Port gibt es beim Drucker nicht."},
		{"abgelehnt",
			"ConnectionRefusedError: [WinError 10061] Es konnte keine Verbindung hergestellt werden",
			"Verbindung auf Port 990 abgelehnt — FTP/LAN-Modus am Drucker ist aus. Einen anderen Port bietet der Drucker nicht."},
		{"kein Weg",
			"OSError: [Errno 113] No route to host",
			"Drucker nicht erreichbar — im Netz nicht auffindbar"},
		{"Zugangscode",
			"ftplib.error_perm: 530 Login incorrect.",
			"Zugangscode wird nicht angenommen"},
		{"leer",
			"   \n  ",
			"Drucker antwortet nicht"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := friendlyFTPError(c.raw); got != c.want {
				t.Fatalf("friendlyFTPError()\n  bekommen: %q\n  erwartet: %q", got, c.want)
			}
		})
	}
}

// Unknown text is kept, but only the last line and truncated.
func TestFriendlyFTPErrorUnbekannt(t *testing.T) {
	raw := "Traceback (most recent call last):\n  File \"x.py\", line 1\nValueError: irgendwas Eigenartiges"
	got := friendlyFTPError(raw)
	if got != "ValueError: irgendwas Eigenartiges" {
		t.Fatalf("bekommen %q", got)
	}
	long := "Fehler: " + string(make([]byte, 400))
	if len(friendlyFTPError(long)) > 200 {
		t.Fatalf("nicht gekuerzt: %d Zeichen", len(friendlyFTPError(long)))
	}
}

// During a running print an old HMS message must not keep the light
// blinking. That was exactly why the chamber, after
// Fortsetzen weiter blinkte.
func TestPrinterHasErrorWaehrendDruck(t *testing.T) {
	cases := []struct {
		name string
		s    *PrinterStatus
		want bool
	}{
		{"laeuft, Meldung war beim Anlaufen schon da",
			&PrinterStatus{Online: true, GcodeState: "RUNNING",
				HmsErrors: []string{"HMS_0300_0100"},
				AckedHms:  map[string]bool{"HMS_0300_0100": true}}, false},
		{"laeuft, Meldung ist neu dazugekommen",
			&PrinterStatus{Online: true, GcodeState: "RUNNING",
				HmsErrors: []string{"HMS_0500_0200"},
				AckedHms:  map[string]bool{"HMS_0300_0100": true}}, true},
		{"laeuft mit echtem Druckfehler",
			&PrinterStatus{Online: true, GcodeState: "RUNNING", PrintError: 105}, true},
		{"pausiert mit HMS-Meldung",
			&PrinterStatus{Online: true, GcodeState: "PAUSE", HmsErrors: []string{"HMS_0300_0100"}}, true},
		{"abgebrochen",
			&PrinterStatus{Online: true, GcodeState: "FAILED"}, true},
		{"offline",
			&PrinterStatus{Online: false, GcodeState: "FAILED"}, false},
		{"sauber fertig",
			&PrinterStatus{Online: true, GcodeState: "FINISH"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := printerHasError(c.s); got != c.want {
				t.Fatalf("printerHasError() = %v, erwartet %v", got, c.want)
			}
		})
	}
}

// The sequence the user reported: fault, acknowledged on the device,
// print continues — and the X1 keeps sending the old message anyway.
func TestBlinkenEndetNachFortsetzen(t *testing.T) {
	ip := "10.9.9.77"
	mqttMgr.mu.Lock()
	mqttMgr.statuses[ip] = &PrinterStatus{Online: true, GcodeState: "PAUSE"}
	mqttMgr.mu.Unlock()
	defer func() {
		mqttMgr.mu.Lock()
		delete(mqttMgr.statuses, ip)
		mqttMgr.mu.Unlock()
	}()

	// 1. Fault during the pause -> it must blink
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"PAUSE","hms":[{"ecode":"0300_0100"}]}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("Stoerung in der Pause muss blinken")
	}

	// 2. User acknowledges on the device and resumes
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":41}}`))
	if printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("nach dem Fortsetzen darf nicht mehr geblinkt werden")
	}

	// 3. Der X1 schickt seine vollstaendige Statusmeldung inklusive alter Liste
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":42,"hms":[{"ecode":"0300_0100"}]}}`))
	if printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("die wiederholte alte Meldung darf das Blinken nicht neu ausloesen")
	}

	// 4. A truly new fault during the print must get through
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","hms":[{"ecode":"0300_0100"},{"ecode":"0C00_0300"}]}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("neue Stoerung waehrend des Drucks muss blinken")
	}
}
