package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

func TestDiscordWebhookSendsEmbed(t *testing.T) {
	var bodies []discordPayload
	var types []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		types = append(types, r.Header.Get("Content-Type"))
		raw, _ := io.ReadAll(r.Body)
		var payload discordPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Errorf("decode: %v", err)
		}
		bodies = append(bodies, payload)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := NewDiscordWebhook(DiscordSettings{WebhookAPI: srv.URL})
	if d == nil {
		t.Fatal("webhook should be enabled")
	}
	line := "🥳 Alice (1.2k points) is Online! | Playing: Valorant | streak length 2 | expires at 2026-11-25 15:04:05 +0100 CET"
	d.Send(Describe(line), events.EventStreamerOnline)
	d.Send(Describe("tick"), events.EventGainForWatch) // quiet, not selected
	d.Send(Describe("ignored"), "")

	if len(bodies) != 1 {
		t.Fatalf("got %d requests, want 1", len(bodies))
	}
	if types[0] != "application/json" {
		t.Fatalf("content type %q", types[0])
	}
	embed := bodies[0].Embeds[0]
	if embed.Title != "🟢 Alice is online" || embed.Description != "Playing Valorant · Balance 1.2k points · Watch streak 2" || embed.Color != 0x0A7A2F {
		t.Fatalf("unexpected embed %+v", embed)
	}
	if bodies[0].Username != project.Name {
		t.Fatalf("username %q", bodies[0].Username)
	}

	named := NewDiscordWebhook(DiscordSettings{WebhookAPI: srv.URL, Events: []string{"GAIN_FOR_WATCH"}})
	named.Send(Describe("+10 → Alice (1.2k points) - Reason: WATCH"), events.EventGainForWatch)
	if len(bodies) != 2 || bodies[1].Embeds[0].Title != "➕ +10 points on Alice" {
		t.Fatalf("explicit watch event should be delivered, got %+v", bodies)
	}
}

func TestDiscordWebhookDisabledWithoutURL(t *testing.T) {
	if NewDiscordWebhook(DiscordSettings{}) != nil {
		t.Fatal("webhook must be nil without a url")
	}
	var d *DiscordWebhook
	d.Send(Describe("ignored"), events.EventStartup)
}

func TestDiscordEscapesFormattingCharacters(t *testing.T) {
	var got discordPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := NewDiscordWebhook(DiscordSettings{WebhookAPI: srv.URL})
	d.Send(Describe("+10 → d0cc_tv (2.7k points) - Reason: WATCH | Game: Counter_Strike *2*"), events.EventGainForWatchStreak)

	embed := got.Embeds[0]
	if want := "⏫ +10 points on d0cc\\_tv"; embed.Title != want {
		t.Fatalf("title got %q, want %q", embed.Title, want)
	}
	if want := "Watching Counter\\_Strike \\*2\\* · Balance 2.7k points"; embed.Description != want {
		t.Fatalf("description got %q, want %q", embed.Description, want)
	}
}
