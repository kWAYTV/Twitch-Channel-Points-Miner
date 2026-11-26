// Package notifytest holds in-process fakes of the services notifications go to.
package notifytest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// DiscordEmbed is one embed of a webhook message, as Discord's API names it.
type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

// DiscordMessage is one webhook message the fake received.
type DiscordMessage struct {
	ContentType string         `json:"-"`
	Username    string         `json:"username"`
	AvatarURL   string         `json:"avatar_url"`
	Embeds      []DiscordEmbed `json:"embeds"`
}

// FakeDiscord is a Discord webhook endpoint. It answers 204 to a JSON POST, as
// Discord does, and keeps every message it is sent.
type FakeDiscord struct {
	// URL is the webhook URL to configure.
	URL string

	mu       sync.Mutex
	messages []DiscordMessage
	status   int
	answer   string
	drop     bool
}

// NewFakeDiscord starts the fake and stops it when the test ends.
func NewFakeDiscord(t testing.TB) *FakeDiscord {
	t.Helper()
	fake := &FakeDiscord{}
	server := httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(server.Close)
	fake.URL = server.URL + "/api/webhooks/1/secret"
	return fake
}

// Messages returns what the fake has received, oldest first.
func (f *FakeDiscord) Messages() []DiscordMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]DiscordMessage(nil), f.messages...)
}

// Refuse makes every later request fail with this status and answer.
func (f *FakeDiscord) Refuse(status int, answer string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer = status, answer
}

// Accept makes later requests succeed again.
func (f *FakeDiscord) Accept() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer, f.drop = 0, "", false
}

// Drop makes the fake close every later connection without answering.
func (f *FakeDiscord) Drop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drop = true
}

func (f *FakeDiscord) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.drop {
		dropConnection(w)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "405: Method Not Allowed", "code": 0}`, http.StatusMethodNotAllowed)
		return
	}
	raw, _ := io.ReadAll(r.Body)
	message := DiscordMessage{ContentType: r.Header.Get("Content-Type")}
	if err := json.Unmarshal(raw, &message); err != nil {
		http.Error(w, `{"message": "Invalid JSON", "code": 50109}`, http.StatusBadRequest)
		return
	}
	f.messages = append(f.messages, message)
	if f.status != 0 {
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.answer)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func dropConnection(w http.ResponseWriter) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return
	}
	if conn, _, err := hijacker.Hijack(); err == nil {
		_ = conn.Close()
	}
}
