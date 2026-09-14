package main

import "testing"

// Der Python-Traceback landete frueher unveraendert in der Oberflaeche. Diese
// Faelle stammen aus echten Ausgaben des Hilfsskripts.
func TestFriendlyFTPError(t *testing.T) {
	cases := []struct {
		name, raw, want string
	}{
		{"Zeitueberschreitung",
			"Traceback (most recent call last):\n  File \"x.py\", line 40, in <module>\n    ftp.connect(host, 990)\nTimeoutError: [WinError 10060] Ein Verbindungsversuch ist fehlgeschlagen\n",
			"Zeitüberschreitung — Drucker antwortet nicht auf Port 990"},
		{"abgelehnt",
			"ConnectionRefusedError: [WinError 10061] Es konnte keine Verbindung hergestellt werden",
			"Verbindung abgelehnt — FTP am Drucker aus oder falscher Port"},
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

// Unbekanntes bleibt erhalten, aber nur die letzte Zeile und gekuerzt.
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

// Waehrend eines laufenden Drucks darf eine alte HMS-Meldung das Licht nicht
// weiter blinken lassen. Genau das war der Grund, warum die Kammer nach dem
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

// Der Ablauf, den der Anwender gemeldet hat: Stoerung, quittiert am Geraet,
// Druck laeuft weiter — und der X1 schickt die alte Meldung trotzdem weiter mit.
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

	// 1. Stoerung waehrend der Pause -> es muss blinken
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"PAUSE","hms":[{"ecode":"0300_0100"}]}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("Stoerung in der Pause muss blinken")
	}

	// 2. Anwender quittiert am Geraet und setzt fort
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":41}}`))
	if printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("nach dem Fortsetzen darf nicht mehr geblinkt werden")
	}

	// 3. Der X1 schickt seine vollstaendige Statusmeldung inklusive alter Liste
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","mc_percent":42,"hms":[{"ecode":"0300_0100"}]}}`))
	if printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("die wiederholte alte Meldung darf das Blinken nicht neu ausloesen")
	}

	// 4. Eine wirklich neue Stoerung waehrend des Drucks muss durchkommen
	mqttMgr.handleMessage(ip, []byte(`{"print":{"gcode_state":"RUNNING","hms":[{"ecode":"0300_0100"},{"ecode":"0C00_0300"}]}}`))
	if !printerHasError(mqttMgr.GetStatus(ip)) {
		t.Fatal("neue Stoerung waehrend des Drucks muss blinken")
	}
}
