package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ─── KI-ASSISTENT ─────────────────────────────────────────────────────────────
//
// Connects to either OpenAI (ChatGPT Platform) or Anthropic (Claude) — with
// dem EIGENEN API-Key des Anwenders. Zwei Funktionen:
//   • Error explainer: error code/message -> plain text + suggested fix.
//   • Camera check: fresh full frame -> assessment "running normally / suspicious
//     auf Fehldruck".
//
// SECURITY: The API key is a secret. It is stored only locally in
// config.json (gitignored), is NEVER logged, NEVER returned to the UI
// and NEVER exposed via the broadcast page/tunnel. It goes
// exclusively over HTTPS to the chosen provider.

var aiClient = &http.Client{Timeout: 90 * time.Second}

// AIConnector is one named AI connection. Created like adding a printer: first
// the connector (name + provider), then its model / key / endpoint. The API key
// is a secret — stored only in config.json, never logged, never sent back to the
// UI (the UI only learns whether a key is set).
type AIConnector struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // openai | anthropic | ollama
	Model    string `json:"model,omitempty"`
	Key      string `json:"key,omitempty"` // secret
	URL      string `json:"url,omitempty"` // endpoint (ollama / custom)
}

// migrateAIConnectors folds the legacy single-provider fields into one connector
// the first time, so an existing setup keeps working and shows up in the new
// list. Idempotent: does nothing once connectors exist.
func migrateAIConnectors() {
	mu.Lock()
	defer mu.Unlock()
	if len(state.AIConnectors) > 0 {
		return
	}
	prov := state.AIProvider
	switch prov {
	case "openai", "anthropic", "ollama":
	default:
		prov = ""
	}
	hasLegacy := prov != "" || strings.TrimSpace(state.OpenAIKey) != "" ||
		strings.TrimSpace(state.AnthropicKey) != "" || strings.TrimSpace(state.AIModel) != ""
	if !hasLegacy {
		return
	}
	if prov == "" {
		prov = "openai"
	}
	c := AIConnector{ID: newConnID(), Name: "Standard", Provider: prov, Model: strings.TrimSpace(state.AIModel), URL: strings.TrimSpace(state.OllamaURL)}
	switch prov {
	case "anthropic":
		c.Key = strings.TrimSpace(state.AnthropicKey)
	case "openai":
		c.Key = strings.TrimSpace(state.OpenAIKey)
	}
	state.AIConnectors = []AIConnector{c}
	state.AIActiveID = c.ID
}

func newConnID() string {
	return fmt.Sprintf("ai_%d", time.Now().UnixNano())
}

// activeConnector returns a copy of the currently selected connector.
func activeConnector() (AIConnector, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range state.AIConnectors {
		if c.ID == state.AIActiveID {
			return c, true
		}
	}
	if len(state.AIConnectors) > 0 {
		return state.AIConnectors[0], true
	}
	return AIConnector{}, false
}

// Standard-Aufforderungen. Im Tool aenderbar (Einstellungen). Platzhalter:
// {lang} = Antwortsprache, {error} = erkannte Fehlermeldung(en).
const defaultPromptExplain = "You are a 3D printer support assistant. Explain the following printer error(s) in simple terms and give concrete step-by-step troubleshooting. Answer in {lang}. Error(s): {error}"
const defaultPromptVision = "You are a 3D printing monitor. Look at this printer camera image and judge whether the print looks normal or shows signs of a failure (spaghetti/stringing, detached part, clog, empty/failed bed). Answer briefly in {lang} with a clear verdict (OK / Suspicious / Failure) and a short reason."

func promptExplain() string {
	mu.Lock()
	p := strings.TrimSpace(state.AIPromptExplain)
	mu.Unlock()
	if p == "" {
		return defaultPromptExplain
	}
	return p
}
func promptVision() string {
	mu.Lock()
	p := strings.TrimSpace(state.AIPromptVision)
	mu.Unlock()
	if p == "" {
		return defaultPromptVision
	}
	return p
}

// aiCfg reads the active connector and fills in sensible defaults.
func aiCfg() (provider, model, key, url string) {
	c, _ := activeConnector()
	provider = c.Provider
	switch provider {
	case "anthropic", "ollama", "openai":
	default:
		provider = "openai"
	}
	model = strings.TrimSpace(c.Model)
	url = strings.TrimSpace(c.URL)
	if url == "" {
		url = "http://localhost:11434"
	}
	switch provider {
	case "anthropic":
		key = strings.TrimSpace(c.Key)
		if model == "" {
			model = "claude-3-5-sonnet-latest"
		}
	case "ollama":
		if model == "" {
			model = "llama3.2-vision"
		}
	default:
		key = strings.TrimSpace(c.Key)
		if model == "" {
			model = "gpt-4o-mini"
		}
	}
	return
}

func languageName() string {
	mu.Lock()
	l := state.Lang
	mu.Unlock()
	switch l {
	case "en":
		return "English"
	case "zh":
		return "Chinese (简体)"
	case "es":
		return "Spanish"
	default:
		return "German"
	}
}

// aiChat sends a prompt (optionally with a JPEG image) to the chosen
// provider and returns the text response.
func aiChat(prompt string, imageJPEG []byte) (string, error) {
	provider, model, key, url := aiCfg()
	if provider == "ollama" {
		return chatOllama(url, model, prompt, imageJPEG)
	}
	if key == "" {
		return "", fmt.Errorf("kein API-Key hinterlegt (%s)", provider)
	}
	if provider == "anthropic" {
		return chatAnthropic(model, key, prompt, imageJPEG)
	}
	return chatOpenAI(model, key, prompt, imageJPEG)
}

// chatOllama talks to a model running locally on THIS PC (Ollama). That way
// nothing leaves the machine for the AI. Vision needs a multimodal
// Modell (z. B. llava, llama3.2-vision).
func chatOllama(url, model, prompt string, img []byte) (string, error) {
	endpoint := strings.TrimRight(url, "/") + "/api/chat"
	msg := map[string]any{"role": "user", "content": prompt}
	if img != nil {
		msg["images"] = []string{base64.StdEncoding.EncodeToString(img)}
	}
	body, _ := json.Marshal(map[string]any{"model": model, "stream": false, "messages": []any{msg}})
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := aiClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama nicht erreichbar (%s): %v", endpoint, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama HTTP %d: %s", resp.StatusCode, kurz(raw))
	}
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Message.Content), nil
}

func chatOpenAI(model, key, prompt string, img []byte) (string, error) {
	content := []any{map[string]any{"type": "text", "text": prompt}}
	if img != nil {
		b64 := base64.StdEncoding.EncodeToString(img)
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]any{"url": "data:image/jpeg;base64," + b64},
		})
	}
	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 600,
		"messages":   []any{map[string]any{"role": "user", "content": content}},
	})
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := aiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, kurz(raw))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("keine Antwort erhalten")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func chatAnthropic(model, key, prompt string, img []byte) (string, error) {
	content := []any{map[string]any{"type": "text", "text": prompt}}
	if img != nil {
		b64 := base64.StdEncoding.EncodeToString(img)
		content = append(content, map[string]any{
			"type":   "image",
			"source": map[string]any{"type": "base64", "media_type": "image/jpeg", "data": b64},
		})
	}
	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 600,
		"messages":   []any{map[string]any{"role": "user", "content": content}},
	})
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := aiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic HTTP %d: %s", resp.StatusCode, kurz(raw))
	}
	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, c := range out.Content {
		sb.WriteString(c.Text)
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("keine Antwort erhalten")
	}
	return strings.TrimSpace(sb.String()), nil
}

func kurz(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}

// ─── HTTP-API ─────────────────────────────────────────────────────────────────

func handleAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Provider      *string `json:"provider"`
			Model         *string `json:"model"`
			OpenAIKey     *string `json:"openai_key"`
			AnthropicKey  *string `json:"anthropic_key"`
			OllamaURL     *string `json:"ollama_url"`
			PromptExplain *string `json:"prompt_explain"`
			PromptVision  *string `json:"prompt_vision"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		if body.Provider != nil && (*body.Provider == "openai" || *body.Provider == "anthropic" || *body.Provider == "ollama") {
			state.AIProvider = *body.Provider
		}
		if body.OllamaURL != nil {
			state.OllamaURL = strings.TrimSpace(*body.OllamaURL)
		}
		if body.PromptExplain != nil {
			state.AIPromptExplain = strings.TrimSpace(*body.PromptExplain)
		}
		if body.PromptVision != nil {
			state.AIPromptVision = strings.TrimSpace(*body.PromptVision)
		}
		if body.Model != nil {
			state.AIModel = strings.TrimSpace(*body.Model)
		}
		// Only set keys when non-empty — an empty field leaves the key unchanged.
		if body.OpenAIKey != nil && strings.TrimSpace(*body.OpenAIKey) != "" {
			state.OpenAIKey = strings.TrimSpace(*body.OpenAIKey)
		}
		if body.AnthropicKey != nil && strings.TrimSpace(*body.AnthropicKey) != "" {
			state.AnthropicKey = strings.TrimSpace(*body.AnthropicKey)
		}
		mu.Unlock()
		saveState()
	}
	mu.Lock()
	provider := state.AIProvider
	if provider == "" {
		provider = "openai"
	}
	ollamaURL := strings.TrimSpace(state.OllamaURL)
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	// IMPORTANT: read the prompt values directly from the state here — do NOT
	// call promptExplain()/promptVision(). They lock mu again, and since we
	// already hold mu, that would be a self-deadlock (Go mutexes are not
	// reentrant). Exactly this froze the whole UI: the handler
	// blocked forever holding mu, so every other API request
	// (printers, status, jobs …) waited endlessly.
	pe := strings.TrimSpace(state.AIPromptExplain)
	if pe == "" {
		pe = defaultPromptExplain
	}
	pv := strings.TrimSpace(state.AIPromptVision)
	if pv == "" {
		pv = defaultPromptVision
	}
	resp := map[string]any{
		"provider":       provider,
		"model":          state.AIModel,
		"has_openai":     strings.TrimSpace(state.OpenAIKey) != "",
		"has_anthropic":  strings.TrimSpace(state.AnthropicKey) != "",
		"ollama_url":     ollamaURL,
		"prompt_explain": pe,
		"prompt_vision":  pv,
	}
	mu.Unlock()
	writeJSON(w, resp)
}

// fehlerTextFuer builds a short description of a printer's current faults
// Druckers (HMS-Codes samt Klartext, print_error).
func fehlerTextFuer(ip string) string {
	s := mqttMgr.GetStatus(ip)
	if s == nil {
		return ""
	}
	var parts []string
	for _, code := range s.HmsErrors {
		if msg, ok := hmsErrorMap[strings.ToUpper(code)]; ok {
			parts = append(parts, code+" ("+msg+")")
		} else {
			parts = append(parts, code)
		}
	}
	if s.PrintError > 0 {
		parts = append(parts, fmt.Sprintf("print_error=%d", s.PrintError))
	}
	return strings.Join(parts, "; ")
}

func handleAIExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP string `json:"ip"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	fehler := fehlerTextFuer(body.IP)
	if strings.TrimSpace(fehler) == "" {
		writeJSON(w, map[string]any{"text": "", "keine_fehler": true})
		return
	}
	prompt := strings.ReplaceAll(promptExplain(), "{lang}", languageName())
	if strings.Contains(prompt, "{error}") {
		prompt = strings.ReplaceAll(prompt, "{error}", fehler)
	} else {
		prompt = prompt + "\n\nError(s): " + fehler
	}
	text, err := aiChat(prompt, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, map[string]any{"text": text, "fehler": fehler})
}

// ─── KI-KONNEKTOREN (mehrere, wie Drucker) ────────────────────────────────────

func validProvider(p string) bool {
	switch p {
	case "openai", "anthropic", "ollama":
		return true
	}
	return false
}

// connectorView is the UI-safe form of a connector: the secret key is replaced
// by a has_key flag so the key never leaves the machine.
func connectorView(c AIConnector) map[string]any {
	return map[string]any{
		"id":       c.ID,
		"name":     c.Name,
		"provider": c.Provider,
		"model":    c.Model,
		"url":      c.URL,
		"has_key":  strings.TrimSpace(c.Key) != "",
	}
}

// handleAIConnectors: GET lists all connectors (+ active id); POST creates one.
func handleAIConnectors(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Name     string `json:"name"`
			Provider string `json:"provider"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			http.Error(w, "Name fehlt", http.StatusBadRequest)
			return
		}
		prov := body.Provider
		if !validProvider(prov) {
			prov = "openai"
		}
		c := AIConnector{ID: newConnID(), Name: name, Provider: prov}
		mu.Lock()
		state.AIConnectors = append(state.AIConnectors, c)
		if state.AIActiveID == "" {
			state.AIActiveID = c.ID
		}
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"id": c.ID})
		return
	}
	mu.Lock()
	views := make([]map[string]any, 0, len(state.AIConnectors))
	for _, c := range state.AIConnectors {
		views = append(views, connectorView(c))
	}
	active := state.AIActiveID
	mu.Unlock()
	writeJSON(w, map[string]any{"connectors": views, "active_id": active})
}

// handleAIConnector: POST updates fields of one connector; DELETE removes it.
func handleAIConnector(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		http.Error(w, "id fehlt", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodDelete {
		mu.Lock()
		out := state.AIConnectors[:0]
		for _, c := range state.AIConnectors {
			if c.ID != id {
				out = append(out, c)
			}
		}
		state.AIConnectors = out
		if state.AIActiveID == id {
			state.AIActiveID = ""
			if len(state.AIConnectors) > 0 {
				state.AIActiveID = state.AIConnectors[0].ID
			}
		}
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "POST/DELETE erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name     *string `json:"name"`
		Provider *string `json:"provider"`
		Model    *string `json:"model"`
		Key      *string `json:"key"`
		URL      *string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	found := false
	for i := range state.AIConnectors {
		if state.AIConnectors[i].ID != id {
			continue
		}
		found = true
		if body.Name != nil && strings.TrimSpace(*body.Name) != "" {
			state.AIConnectors[i].Name = strings.TrimSpace(*body.Name)
		}
		if body.Provider != nil && validProvider(*body.Provider) {
			state.AIConnectors[i].Provider = *body.Provider
		}
		if body.Model != nil {
			state.AIConnectors[i].Model = strings.TrimSpace(*body.Model)
		}
		if body.URL != nil {
			state.AIConnectors[i].URL = strings.TrimSpace(*body.URL)
		}
		// Only overwrite the key when a non-empty value is sent — an empty field
		// leaves the stored secret untouched.
		if body.Key != nil && strings.TrimSpace(*body.Key) != "" {
			state.AIConnectors[i].Key = strings.TrimSpace(*body.Key)
		}
		break
	}
	mu.Unlock()
	if !found {
		http.Error(w, "unbekannter Konnektor", http.StatusNotFound)
		return
	}
	saveState()
	writeJSON(w, map[string]any{"ok": true})
}

// handleAIActive selects the active connector.
func handleAIActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		var body struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		id = strings.TrimSpace(body.ID)
	}
	mu.Lock()
	ok := false
	for _, c := range state.AIConnectors {
		if c.ID == id {
			ok = true
			break
		}
	}
	if ok {
		state.AIActiveID = id
	}
	mu.Unlock()
	if !ok {
		http.Error(w, "unbekannter Konnektor", http.StatusNotFound)
		return
	}
	saveState()
	writeJSON(w, map[string]any{"ok": true, "active_id": id})
}

func handleAIVision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erwartet", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IP string `json:"ip"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	stream := streamNameFor(body.IP)
	if stream == "" {
		http.Error(w, "kein Stream fuer diesen Drucker", http.StatusBadRequest)
		return
	}
	frame, err := fetchFrame(stream)
	if err != nil || len(frame) == 0 {
		http.Error(w, "kein Kamerabild verfuegbar (go2rtc/Drucker?)", http.StatusServiceUnavailable)
		return
	}
	prompt := strings.ReplaceAll(promptVision(), "{lang}", languageName())
	text, err := aiChat(prompt, frame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, map[string]any{"text": text})
}
