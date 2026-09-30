package main

import (
	"net/http"
	"testing"
)

func TestPasswordHashVerify(t *testing.T) {
	h := hashPassword("geheim123")
	if !verifyPassword("geheim123", h) {
		t.Fatal("correct password should verify")
	}
	if verifyPassword("falsch", h) {
		t.Fatal("wrong password must not verify")
	}
	if verifyPassword("geheim123", "garbage") {
		t.Fatal("garbage hash must not verify")
	}
	// Two hashes of the same password differ (random salt).
	if h == hashPassword("geheim123") {
		t.Fatal("hashes should be salted/unique")
	}
}

func TestRequiredRole(t *testing.T) {
	cases := []struct {
		method, path, want string
	}{
		{http.MethodGet, "/api/status", "viewer"},
		{http.MethodGet, "/api/printers", "viewer"},
		{http.MethodPost, "/api/print/command", "operator"},
		{http.MethodPost, "/api/sync/start", "operator"},
		{http.MethodPost, "/api/plan", "operator"},
		{http.MethodPost, "/api/settings", "admin"},
		{http.MethodPost, "/api/printers", "admin"},
		{http.MethodPost, "/api/ai/connector", "admin"},
		{http.MethodGet, "/api/users", "admin"},
		{http.MethodDelete, "/api/token", "admin"},
	}
	for _, c := range cases {
		if got := requiredRole(c.method, c.path); got != c.want {
			t.Errorf("%s %s: got %q want %q", c.method, c.path, got, c.want)
		}
	}
}

func TestOpenPaths(t *testing.T) {
	for _, p := range []string{"/api/login", "/api/needsetup", "/api/setup-admin", "/logo.svg"} {
		if !openPath(p) {
			t.Errorf("%s should be open", p)
		}
	}
	for _, p := range []string{"/api/status", "/api/printers", "/"} {
		if openPath(p) {
			t.Errorf("%s should NOT be open", p)
		}
	}
}

func TestIsLocalRequest(t *testing.T) {
	mk := func(remote string, hdr map[string]string) *http.Request {
		r, _ := http.NewRequest("GET", "/api/status", nil)
		r.RemoteAddr = remote
		for k, v := range hdr {
			r.Header.Set(k, v)
		}
		return r
	}
	if !isLocalRequest(mk("127.0.0.1:5051", nil)) {
		t.Error("loopback should be local")
	}
	if !isLocalRequest(mk("[::1]:5051", nil)) {
		t.Error("ipv6 loopback should be local")
	}
	if isLocalRequest(mk("192.168.1.20:5051", nil)) {
		t.Error("LAN IP must NOT be local")
	}
	// Proxied in (tunnel) from localhost is remote.
	if isLocalRequest(mk("127.0.0.1:5051", map[string]string{"X-Forwarded-For": "8.8.8.8"})) {
		t.Error("forwarded request must NOT be local")
	}
	if isLocalRequest(mk("127.0.0.1:5051", map[string]string{"Cf-Connecting-Ip": "8.8.8.8"})) {
		t.Error("cloudflare-forwarded request must NOT be local")
	}
}

func TestTokenIdentify(t *testing.T) {
	mu.Lock()
	saved := state.APITokens
	clear := authRandToken()
	state.APITokens = []APIToken{{ID: "t1", Name: "test", Role: "operator", Hash: tokenHash(clear)}}
	mu.Unlock()
	defer func() { mu.Lock(); state.APITokens = saved; mu.Unlock() }()

	r, _ := http.NewRequest("GET", "/api/status", nil)
	r.Header.Set("Authorization", "Bearer "+clear)
	role, name := identify(r)
	if role != "operator" {
		t.Fatalf("token role: got %q want operator", role)
	}
	if name != "token:test" {
		t.Fatalf("token name: got %q", name)
	}
	// Wrong token → no role.
	r2, _ := http.NewRequest("GET", "/api/status", nil)
	r2.Header.Set("X-API-Token", "wrong")
	if role2, _ := identify(r2); role2 != "" {
		t.Fatalf("wrong token must not authenticate, got %q", role2)
	}
}
