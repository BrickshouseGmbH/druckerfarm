package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestXiaomiProxyLeitetWeiter(t *testing.T) {
	var gesehenBody, gesehenPfad, gesehenQuery string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gesehenPfad = r.URL.Path
		gesehenQuery = r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gesehenBody = string(b)
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer fake.Close()
	u, _ := url.Parse(fake.URL)
	p, _ := strconv.Atoi(u.Port())
	old := go2rtcPort
	go2rtcPort = p
	defer func() { go2rtcPort = old }()

	body := `{"username":"me@example.com","password":"geheim","server":"de"}`
	req := httptest.NewRequest(http.MethodPost, "/api/xiaomi-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleXiaomiProxy(rec, req)

	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("Antwort nicht durchgereicht: %d %s", rec.Code, rec.Body.String())
	}
	if gesehenPfad != "/api/xiaomi" {
		t.Fatalf("falscher Zielpfad: %q", gesehenPfad)
	}
	if !strings.Contains(gesehenBody, "geheim") {
		t.Fatalf("Body nicht weitergeleitet: %q", gesehenBody)
	}
	// Zugangsdaten dürfen NICHT in der Query landen
	if strings.Contains(gesehenQuery, "geheim") || strings.Contains(gesehenQuery, "password") {
		t.Fatalf("Passwort in der URL/Query gelandet: %q", gesehenQuery)
	}
}
