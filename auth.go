package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── ANMELDUNG, BENUTZER, ROLLEN, API-TOKENS ──────────────────────────────────
//
// The tool is a hub: it holds the printers' access codes and is the only client
// that talks to them. Everyone else — people in the browser, other programs —
// goes through here, behind a login. Three roles:
//
//	viewer   — read only (status, cameras, lists)
//	operator — read + control prints (pause/resume/stop, send file, sync)
//	admin    — everything incl. printers, access codes, settings, users, tokens
//
// Passwords are stored only as a salted PBKDF2/HMAC-SHA256 hash (no external
// dependency). API tokens are stored only as a SHA-256 hash; the clear token is
// shown once at creation. Nothing here is ever logged.

type User struct {
	Name     string `json:"name"`
	Role     string `json:"role"` // viewer | operator | admin
	PassHash string `json:"pass_hash"`
}

type APIToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Hash    string `json:"hash"` // sha256(token) hex
	Created int64  `json:"created"`
}

var roleRank = map[string]int{"viewer": 1, "operator": 2, "admin": 3}

func validRole(r string) bool { _, ok := roleRank[r]; return ok }

// ─── Passwort-Hashing (PBKDF2/HMAC-SHA256, nur Standardbibliothek) ────────────

const pbkdf2Iter = 120000

func pbkdf2Key(password, salt []byte, iter, keyLen int) []byte {
	out := []byte{}
	block := 1
	for len(out) < keyLen {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		var b [4]byte
		b[0] = byte(block >> 24)
		b[1] = byte(block >> 16)
		b[2] = byte(block >> 8)
		b[3] = byte(block)
		mac.Write(b[:])
		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iter; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
		block++
	}
	return out[:keyLen]
}

func hashPassword(pw string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	dk := pbkdf2Key([]byte(pw), salt, pbkdf2Iter, 32)
	return fmt.Sprintf("pbkdf2$%d$%s$%s", pbkdf2Iter,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk))
}

func verifyPassword(pw, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got := pbkdf2Key([]byte(pw), salt, iter, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func tokenHash(tok string) string {
	h := sha256.Sum256([]byte(tok))
	return fmt.Sprintf("%x", h[:])
}

func authRandToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// ─── Sitzungen (im Speicher; nach Neustart neu anmelden) ──────────────────────

type session struct {
	user string
	role string
	exp  time.Time
}

var (
	sessMu   sync.Mutex
	sessions = map[string]session{}
)

const sessionTTL = 7 * 24 * time.Hour
const sessCookie = "df_sess"

func newSession(user, role string) string {
	id := authRandToken()
	sessMu.Lock()
	sessions[id] = session{user: user, role: role, exp: time.Now().Add(sessionTTL)}
	sessMu.Unlock()
	return id
}

func lookupSession(id string) (session, bool) {
	sessMu.Lock()
	defer sessMu.Unlock()
	s, ok := sessions[id]
	if !ok {
		return session{}, false
	}
	if time.Now().After(s.exp) {
		delete(sessions, id)
		return session{}, false
	}
	return s, true
}

func dropSession(id string) {
	sessMu.Lock()
	delete(sessions, id)
	sessMu.Unlock()
}

// identify returns the caller's role and name, from an API token (header) or the
// session cookie. Empty role means not authenticated.
func identify(r *http.Request) (role, name string) {
	// API token: "Authorization: Bearer <t>" or "X-API-Token: <t>".
	tok := ""
	if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
		tok = strings.TrimSpace(a[7:])
	}
	if tok == "" {
		tok = strings.TrimSpace(r.Header.Get("X-API-Token"))
	}
	if tok != "" {
		h := tokenHash(tok)
		mu.RLock()
		for _, t := range state.APITokens {
			if subtle.ConstantTimeCompare([]byte(t.Hash), []byte(h)) == 1 {
				role = t.Role
				name = "token:" + t.Name
				break
			}
		}
		mu.RUnlock()
		if role != "" {
			return role, name
		}
	}
	if c, err := r.Cookie(sessCookie); err == nil && c.Value != "" {
		if s, ok := lookupSession(c.Value); ok {
			return s.role, s.user
		}
	}
	return "", ""
}

// ─── Pfad-Berechtigungen ──────────────────────────────────────────────────────

func hasPrefixAny(s string, ps ...string) bool {
	for _, p := range ps {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// openPath: reachable without login (needed to render/serve the login page).
func openPath(p string) bool {
	switch p {
	case "/api/login", "/api/logout", "/api/me", "/api/needsetup", "/api/setup-admin",
		"/logo.svg", "/splash.jpg", "/favicon.ico":
		return true
	}
	return false
}

// requiredRole says which role a path+method needs (for authenticated callers).
func requiredRole(method, path string) string {
	// Account/token management is admin-only, even for reading.
	if hasPrefixAny(path, "/api/users", "/api/user", "/api/tokens", "/api/token") {
		return "admin"
	}
	if method == http.MethodGet || method == http.MethodHead {
		return "viewer" // any logged-in user may read
	}
	// Writes: a fixed set is admin, the rest is operator.
	adminWrite := hasPrefixAny(path,
		"/api/settings", "/api/printers", "/api/import", "/api/export", "/api/yaml",
		"/api/ai/", "/api/cameras", "/api/camera-off", "/api/tunnel/", "/api/youtube/",
		"/api/update/", "/api/components", "/api/discover", "/api/blink",
		"/api/go2rtc/restart", "/api/go2rtc/killall", "/api/open-folder")
	if adminWrite {
		return "admin"
	}
	return "operator"
}

// authMiddleware enforces login for everything and role for each path.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if openPath(p) {
			next.ServeHTTP(w, r)
			return
		}
		// Local access (the person at the server machine) runs without login.
		// Only connections from elsewhere need credentials. Requests proxied in
		// (e.g. via the Cloudflare tunnel) count as remote even though they
		// arrive from localhost.
		if isLocalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		role, _ := identify(r)
		if role == "" {
			// Not logged in.
			if strings.HasPrefix(p, "/api/") {
				http.Error(w, `{"error":"nicht angemeldet","auth":true}`, http.StatusUnauthorized)
				return
			}
			serveLoginPage(w, r)
			return
		}
		need := requiredRole(r.Method, p)
		if roleRank[role] < roleRank[need] {
			if strings.HasPrefix(p, "/api/") {
				http.Error(w, `{"error":"keine Berechtigung fuer diese Aktion"}`, http.StatusForbidden)
				return
			}
			// Pages are fine; the UI hides what the role may not do.
		}
		next.ServeHTTP(w, r)
	})
}

// isLocalRequest reports whether the request comes from the machine itself
// (loopback) and is NOT proxied in. A forwarding header means it came through a
// proxy/tunnel from outside, so it counts as remote even from 127.0.0.1.
func isLocalRequest(r *http.Request) bool {
	if r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != "" ||
		r.Header.Get("Cf-Connecting-Ip") != "" || r.Header.Get("Forwarded") != "" {
		return false
	}
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

// ─── HTTP-Handler ─────────────────────────────────────────────────────────────

func usersEmpty() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(state.Users) == 0
}

func handleNeedSetup(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"needsetup": usersEmpty()})
}

// handleSetupAdmin creates the very first admin — only while no users exist.
func handleSetupAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	// Only the person at the server machine may create accounts — a remote
	// visitor must not be able to claim admin over an unconfigured tool.
	if !isLocalRequest(r) {
		http.Error(w, `{"error":"nur am Server-Rechner moeglich"}`, http.StatusForbidden)
		return
	}
	if !usersEmpty() {
		http.Error(w, `{"error":"bereits eingerichtet"}`, http.StatusForbidden)
		return
	}
	var b struct{ Name, Password string }
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" || len(b.Password) < 6 {
		http.Error(w, `{"error":"Name noetig, Passwort min. 6 Zeichen"}`, http.StatusBadRequest)
		return
	}
	mu.Lock()
	state.Users = []User{{Name: b.Name, Role: "admin", PassHash: hashPassword(b.Password)}}
	mu.Unlock()
	saveState()
	id := newSession(b.Name, "admin")
	setSessionCookie(w, id)
	writeJSON(w, map[string]any{"ok": true, "name": b.Name, "role": "admin"})
}

func setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name: sessCookie, Value: id, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
		MaxAge: int(sessionTTL / time.Second),
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var b struct{ Name, Password string }
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	mu.RLock()
	var found *User
	for i := range state.Users {
		if strings.EqualFold(state.Users[i].Name, b.Name) {
			u := state.Users[i]
			found = &u
			break
		}
	}
	mu.RUnlock()
	if found == nil || !verifyPassword(b.Password, found.PassHash) {
		http.Error(w, `{"error":"Benutzer oder Passwort falsch"}`, http.StatusUnauthorized)
		return
	}
	id := newSession(found.Name, found.Role)
	setSessionCookie(w, id)
	writeJSON(w, map[string]any{"ok": true, "name": found.Name, "role": found.Role})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessCookie); err == nil {
		dropSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessCookie, Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, map[string]any{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	// Local access (the person at the server machine) has full rights without
	// logging in.
	if isLocalRequest(r) {
		writeJSON(w, map[string]any{"name": "lokal", "role": "admin", "local": true})
		return
	}
	role, name := identify(r)
	if role == "" {
		writeJSON(w, map[string]any{"anon": true, "needsetup": usersEmpty()})
		return
	}
	writeJSON(w, map[string]any{"name": name, "role": role})
}

// handleUsers: GET lists users (admin), POST creates one (admin).
func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var b struct{ Name, Password, Role string }
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		b.Name = strings.TrimSpace(b.Name)
		if !validRole(b.Role) {
			b.Role = "viewer"
		}
		if b.Name == "" || len(b.Password) < 6 {
			http.Error(w, `{"error":"Name noetig, Passwort min. 6 Zeichen"}`, http.StatusBadRequest)
			return
		}
		mu.Lock()
		for _, u := range state.Users {
			if strings.EqualFold(u.Name, b.Name) {
				mu.Unlock()
				http.Error(w, `{"error":"Benutzer existiert schon"}`, http.StatusConflict)
				return
			}
		}
		state.Users = append(state.Users, User{Name: b.Name, Role: b.Role, PassHash: hashPassword(b.Password)})
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true})
		return
	}
	mu.RLock()
	list := make([]map[string]any, 0, len(state.Users))
	for _, u := range state.Users {
		list = append(list, map[string]any{"name": u.Name, "role": u.Role})
	}
	mu.RUnlock()
	writeJSON(w, map[string]any{"users": list})
}

// handleUser: POST updates role/password, DELETE removes. ?name=
func handleUser(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		http.Error(w, "name fehlt", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodDelete {
		mu.Lock()
		// Never remove the last admin.
		admins := 0
		for _, u := range state.Users {
			if u.Role == "admin" {
				admins++
			}
		}
		out := state.Users[:0]
		removedAdmin := false
		for _, u := range state.Users {
			if strings.EqualFold(u.Name, name) {
				if u.Role == "admin" {
					removedAdmin = true
				}
				continue
			}
			out = append(out, u)
		}
		if removedAdmin && admins <= 1 {
			mu.Unlock()
			http.Error(w, `{"error":"letzter Admin kann nicht geloescht werden"}`, http.StatusBadRequest)
			return
		}
		state.Users = out
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "POST/DELETE erwartet", http.StatusMethodNotAllowed)
		return
	}
	var b struct{ Role, Password string }
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	found := false
	for i := range state.Users {
		if !strings.EqualFold(state.Users[i].Name, name) {
			continue
		}
		found = true
		if validRole(b.Role) {
			// Do not demote the last admin.
			if state.Users[i].Role == "admin" && b.Role != "admin" {
				admins := 0
				for _, u := range state.Users {
					if u.Role == "admin" {
						admins++
					}
				}
				if admins <= 1 {
					mu.Unlock()
					http.Error(w, `{"error":"letzter Admin kann nicht herabgestuft werden"}`, http.StatusBadRequest)
					return
				}
			}
			state.Users[i].Role = b.Role
		}
		if strings.TrimSpace(b.Password) != "" {
			if len(b.Password) < 6 {
				mu.Unlock()
				http.Error(w, `{"error":"Passwort min. 6 Zeichen"}`, http.StatusBadRequest)
				return
			}
			state.Users[i].PassHash = hashPassword(b.Password)
		}
		break
	}
	mu.Unlock()
	if !found {
		http.Error(w, `{"error":"unbekannter Benutzer"}`, http.StatusNotFound)
		return
	}
	saveState()
	writeJSON(w, map[string]any{"ok": true})
}

// handleTokens: GET lists tokens (no secret), POST creates one (returns the
// clear token exactly once).
func handleTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var b struct{ Name, Role string }
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		b.Name = strings.TrimSpace(b.Name)
		if b.Name == "" {
			http.Error(w, `{"error":"Name fehlt"}`, http.StatusBadRequest)
			return
		}
		if !validRole(b.Role) {
			b.Role = "viewer"
		}
		clear := authRandToken()
		t := APIToken{ID: fmt.Sprintf("tok_%d", time.Now().UnixNano()), Name: b.Name, Role: b.Role, Hash: tokenHash(clear), Created: time.Now().Unix()}
		mu.Lock()
		state.APITokens = append(state.APITokens, t)
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true, "token": clear, "id": t.ID})
		return
	}
	mu.RLock()
	list := make([]map[string]any, 0, len(state.APITokens))
	for _, t := range state.APITokens {
		list = append(list, map[string]any{"id": t.ID, "name": t.Name, "role": t.Role, "created": t.Created})
	}
	mu.RUnlock()
	writeJSON(w, map[string]any{"tokens": list})
}

// handleToken: DELETE removes a token by ?id=.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE erwartet", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	mu.Lock()
	out := state.APITokens[:0]
	for _, t := range state.APITokens {
		if t.ID != id {
			out = append(out, t)
		}
	}
	state.APITokens = out
	mu.Unlock()
	saveState()
	writeJSON(w, map[string]any{"ok": true})
}
