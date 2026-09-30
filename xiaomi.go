package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// ─── XIAOMI / MI-HOME LOGIN (in the tool, via go2rtc) ─────────────────────────
//
// Xiaomi cameras (e.g. CW300) need a Mi-Home login: go2rtc uses it to fetch
// the session key from the Xiaomi cloud on every connect. The
// login creates a passToken that expires after ~3 days (known go2rtc
// quirk). Without a valid token: "401 Unauthorized" → no image.
//
// Statt Xiaomis Anmeldeprotokoll (Krypto, Captcha, 2FA) selbst nachzubauen,
// the tool uses go2rtc's ALREADY EXISTING login: this function
// forwards the login request unchanged to go2rtc's own `/api/xiaomi`
// (on 127.0.0.1). So the login happens IN THE TOOL — the user does not open
// go2rtc themselves — but go2rtc does the actual login.
//
// IMPORTANT (privacy): credentials go exclusively to the local
// go2rtc. They are NOT stored and NOT logged here.

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
		// No detailed logging (could contain the endpoint/parameters).
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
