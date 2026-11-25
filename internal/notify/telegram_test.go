package notify

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestTelegramNotifierSendsSelectedEvents(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	var forms []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		paths = append(paths, r.Method+" "+r.URL.Path)
		forms = append(forms, r.PostForm)
		mu.Unlock()
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	n := newTelegramNotifier(TelegramSettings{
		ChatID:              " 123456789 ",
		Token:               "42:secret",
		Events:              []string{"streamer_online", "BET_LOSE"},
		DisableNotification: true,
	}, srv.URL)
	if n == nil {
		t.Fatal("notifier should be enabled")
	}

	n.Send(Describe("Alice (1.2k points) is Online! | Playing: Valorant"), events.EventStreamerOnline)
	n.Send(Describe("+10 points"), events.EventGainForWatch) // not selected
	n.Send(Describe("no event"), "")
	n.Send(Message{Headline: strings.Repeat("é", 5000), Details: []string{strings.Repeat("é", 5000)}}, events.EventBetLose)

	if len(forms) != 2 {
		t.Fatalf("got %d requests, want 2", len(forms))
	}
	if paths[0] != "POST /bot42:secret/sendMessage" {
		t.Fatalf("unexpected request %q", paths[0])
	}
	first := forms[0]
	if first.Get("chat_id") != "123456789" || first.Get("text") != "🟢 <b>Alice is online</b>\nPlaying Valorant · Balance 1.2k points" ||
		first.Get("parse_mode") != "HTML" || first.Get("disable_notification") != "true" ||
		first.Get("link_preview_options") != `{"is_disabled":true}` {
		t.Fatalf("unexpected form %v", first)
	}
	long := forms[1].Get("text")
	if !strings.HasPrefix(long, "📉 <b>é") || !strings.Contains(long, "</b>\né") {
		t.Fatalf("long text should keep a bold headline and a details line: %q", long[:40])
	}
	if got := len([]rune(long)); got > telegramTextLimit {
		t.Fatalf("long text should be at most %d characters, got %d", telegramTextLimit, got)
	}
}

func TestTelegramNotifierSkipsQuietEventsUnlessNamed(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got = append(got, r.PostForm.Get("text"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	open := newTelegramNotifier(TelegramSettings{ChatID: "1", Token: "t"}, srv.URL)
	open.Send(Describe("tick"), events.EventGainForWatch)
	open.Send(Describe("Alice (1.2k points) is Online!"), events.EventStreamerOnline)
	if len(got) != 1 || !strings.Contains(got[0], "Alice") {
		t.Fatalf("empty events list should skip watch ticks and send online, got %v", got)
	}

	got = nil
	all := newTelegramNotifier(TelegramSettings{ChatID: "1", Token: "t", Events: []string{"*"}}, srv.URL)
	all.Send(Describe("+10 → Alice (1.2k points) - Reason: WATCH"), events.EventGainForWatch)
	if len(got) != 1 || got[0] != "➕ <b>+10 points on Alice</b>\nWatching · Balance 1.2k points" {
		t.Fatalf("* should opt into watch ticks, got %v", got)
	}
}

func TestTelegramNotifierDisabledWithoutCredentials(t *testing.T) {
	if NewTelegramNotifier(TelegramSettings{ChatID: "1"}) != nil || NewTelegramNotifier(TelegramSettings{Token: "t"}) != nil {
		t.Fatal("notifier must be nil unless chat_id and token are both set")
	}
	var n *TelegramNotifier
	n.Send(Describe("ignored"), events.EventStartup) // nil notifier is a no-op
}

func TestFormatTelegram(t *testing.T) {
	for _, tc := range []struct {
		event   events.Event
		message string
		want    string
	}{
		{events.EventStreamerOnline, "Alice is Online!", "🟢 <b>Alice is online</b>"},
		{events.EventStartup, "", "▶️ <b>Session started</b>"},
		{events.EventChatMention, "nick at #alice wrote: <b>hi</b> & bye", "💬 <b>nick mentioned you in #alice</b>\n\"&lt;b&gt;hi&lt;/b&gt; &amp; bye\""},
		{events.EventBetLose, "EventPrediction: bob - overtime? - Decision: No - Result: LOSE, Gained: -500", "📉 <b>Lost 500 points on bob</b>\n\"overtime?\" · Chose No"},
	} {
		if got := formatTelegram(tc.event, Describe(tc.message)); got != tc.want {
			t.Fatalf("formatTelegram(%s) got %q, want %q", tc.event, got, tc.want)
		}
	}
}
