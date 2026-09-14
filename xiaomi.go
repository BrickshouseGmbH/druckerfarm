package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// ─── XIAOMI / MI-HOME-LOGIN (im Tool, über go2rtc) ────────────────────────────
//
// Xiaomi-Kameras (z. B. CW300) brauchen einen Mi-Home-Login: go2rtc holt damit
// bei jedem Verbindungsaufbau den Sitzungsschlüssel aus der Xiaomi-Cloud. Der
// Login erzeugt einen passToken, der nach ~3 Tagen abläuft (bekannte go2rtc-
// Eigenheit). Ohne gültigen Token: „401 Unauthorized" → kein Bild.
//
// Statt Xiaomis Anmeldeprotokoll (Krypto, Captcha, 2FA) selbst nachzubauen,
// nutzt das Tool die BEREITS VORHANDENE Anmeldung von go2rtc: Diese Funktion
// leitet die Login-Anfrage unverändert an go2rtcs eigenes `/api/xiaomi` weiter
// (auf 127.0.0.1). So findet der Login IM TOOL statt — der Anwender öffnet
// go2rtc nicht selbst —, aber die eigentliche Anmeldung macht go2rtc.
//
// WICHTIG (Datenschutz): Zugangsdaten laufen ausschließlich an das lokale
// go2rtc. Sie werden hier NICHT gespeichert und NICHT geloggt.

func handleXiaomiProxy(w http.ResponseWriter, r *http.Request) {
	ziel := fmt.Sprintf("http://127.0.0.1:%d/api/xiaomi", go2rtcPort)
	if r.URL.RawQuery != "" {
		ziel += "?" + r.URL.RawQuery
	}
	var body io.Reader
	if r.Body != nil {
		body = r.Body
	}
	req, err := http.NewRequest(r.Method, ziel, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Kein Detail-Logging (könnte den Endpunkt/Parameter enthalten).
		http.Error(w, "go2rtc nicht erreichbar", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
