package notifytest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

// TelegramMessage is one sendMessage call the fake received.
type TelegramMessage struct {
	ChatID              string
	Text                string
	ParseMode           string
	DisableNotification string
	LinkPreviewOptions  string
}

// FakeTelegram is the Telegram Bot API's sendMessage method for one bot. A call
// with another bot's token gets Telegram's 401, and is not kept.
type FakeTelegram struct {
	// URL is the API base to configure, in place of https://api.telegram.org.
	URL string
	// Token is the only bot token the fake accepts.
	Token string

	mu       sync.Mutex
	messages []TelegramMessage
	status   int
	answer   string
	drop     bool
}

// NewFakeTelegram starts the fake and stops it when the test ends.
func NewFakeTelegram(t testing.TB) *FakeTelegram {
	t.Helper()
	fake := &FakeTelegram{Token: "42:secret"}
	server := httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(server.Close)
	fake.URL = server.URL
	return fake
}

// Messages returns what the fake has received, oldest first.
func (f *FakeTelegram) Messages() []TelegramMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]TelegramMessage(nil), f.messages...)
}

// Refuse makes every later call fail the way Telegram reports an error.
func (f *FakeTelegram) Refuse(status int, description string) {
	f.RefuseWith(status, `{"ok":false,"error_code":`+strconv.Itoa(status)+`,"description":`+strconv.Quote(description)+`}`)
}

// RefuseWith makes every later call fail with this status and raw answer.
func (f *FakeTelegram) RefuseWith(status int, answer string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer = status, answer
}

// Accept makes later calls succeed again.
func (f *FakeTelegram) Accept() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer, f.drop = 0, "", false
}

// Drop makes the fake close every later connection without answering.
func (f *FakeTelegram) Drop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drop = true
}

func (f *FakeTelegram) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.drop {
		dropConnection(w)
		return
	}
	if r.URL.Path != "/bot"+f.Token+"/sendMessage" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"ok":false,"error_code":401,"description":"Unauthorized"}`)
		return
	}
	if r.Method != http.MethodPost || r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"ok":false,"error_code":400,"description":"Bad Request"}`)
		return
	}
	f.messages = append(f.messages, TelegramMessage{
		ChatID:              r.PostForm.Get("chat_id"),
		Text:                r.PostForm.Get("text"),
		ParseMode:           r.PostForm.Get("parse_mode"),
		DisableNotification: r.PostForm.Get("disable_notification"),
		LinkPreviewOptions:  r.PostForm.Get("link_preview_options"),
	})
	if f.status != 0 {
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.answer)
		return
	}
	_, _ = io.WriteString(w, `{"ok":true,"result":{}}`)
}
