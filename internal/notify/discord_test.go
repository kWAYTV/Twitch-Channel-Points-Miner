package notify

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify/notifytest"
)

var aliceOnline = Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}}

func aliceOnlineOnDiscord() notifytest.DiscordMessage {
	return notifytest.DiscordMessage{
		ContentType: "application/json",
		Username:    "Twitch Channel Points Miner",
		AvatarURL:   "https://raw.githubusercontent.com/kWAYTV/twitch-points-miner/main/docs/assets/gopher.png",
		Embeds: []notifytest.DiscordEmbed{{
			Title:       "🟢 Alice is online",
			Description: "Playing Valorant · Balance 1.2k points",
			Color:       0x0A7A2F,
		}},
	}
}

const discordWarning = "discord: notification not delivered: %s (further failures are not logged)"

func setupDiscord(t *testing.T, eventNames ...string) (*DiscordWebhook, *notifytest.FakeDiscord, *fakeLog) {
	t.Helper()
	discord := notifytest.NewFakeDiscord(t)
	log := &fakeLog{}
	webhook := newDiscordWebhook(DiscordSettings{WebhookAPI: discord.URL, Events: eventNames}, log.Printf)
	return webhook, discord, log
}

func TestNewDiscordWebhook(t *testing.T) {
	t.Run("no webhook URL", func(t *testing.T) {
		if got := NewDiscordWebhook(DiscordSettings{Events: []string{"*"}}); got != nil {
			t.Errorf("NewDiscordWebhook = %+v, want nil", got)
		}
	})

	t.Run("only spaces for a webhook URL", func(t *testing.T) {
		if got := NewDiscordWebhook(DiscordSettings{WebhookAPI: "   "}); got != nil {
			t.Errorf("NewDiscordWebhook = %+v, want nil", got)
		}
	})

	t.Run("a webhook URL", func(t *testing.T) {
		if got := NewDiscordWebhook(DiscordSettings{WebhookAPI: "https://discord.com/api/webhooks/1/secret"}); got == nil {
			t.Errorf("NewDiscordWebhook = nil, want a webhook")
		}
	})
}

func TestDiscordWebhook_Send(t *testing.T) {
	t.Run("an event with no events list", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := discord.Messages(), []notifytest.DiscordMessage{aliceOnlineOnDiscord()}; !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
		if len(log.Lines) != 0 {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})

	t.Run("a webhook URL with spaces around it", func(t *testing.T) {
		discord := notifytest.NewFakeDiscord(t)
		webhook := newDiscordWebhook(DiscordSettings{WebhookAPI: "  " + discord.URL + "\n"}, (&fakeLog{}).Printf)

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := discord.Messages(), []notifytest.DiscordMessage{aliceOnlineOnDiscord()}; !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	t.Run("a quiet event with no events list", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t)

		webhook.Send(aliceOnline, events.EventGainForWatch)

		if got := discord.Messages(); len(got) != 0 {
			t.Errorf("discord received %+v, want nothing", got)
		}
	})

	t.Run("a quiet event named in the events list", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t, "gain_for_watch")

		webhook.Send(Message{"+10 points on Alice", []string{"Watching"}}, events.EventGainForWatch)

		want := []notifytest.DiscordMessage{{
			ContentType: "application/json",
			Username:    "Twitch Channel Points Miner",
			AvatarURL:   "https://raw.githubusercontent.com/kWAYTV/twitch-points-miner/main/docs/assets/gopher.png",
			Embeds:      []notifytest.DiscordEmbed{{Title: "➕ +10 points on Alice", Description: "Watching", Color: 0x0A7A2F}},
		}}
		if got := discord.Messages(); !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	t.Run("an event missing from the events list", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t, "BET_WIN")

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if got := discord.Messages(); len(got) != 0 {
			t.Errorf("discord received %+v, want nothing", got)
		}
	})

	t.Run("no event", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t)

		webhook.Send(aliceOnline, "")

		if got := discord.Messages(); len(got) != 0 {
			t.Errorf("discord received %+v, want nothing", got)
		}
	})

	t.Run("a webhook that is not set up", func(t *testing.T) {
		var webhook *DiscordWebhook

		webhook.Send(aliceOnline, events.EventStreamerOnline)
	})

	t.Run("the same event twice", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t)
		webhook.Send(aliceOnline, events.EventStreamerOnline)

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := discord.Messages(), []notifytest.DiscordMessage{aliceOnlineOnDiscord(), aliceOnlineOnDiscord()}; !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	t.Run("discord refusing", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)
		discord.Refuse(400, `{"message": "Cannot send an empty message", "code": 50006}`+"\n")

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{`discord: notification not delivered: {"message": "Cannot send an empty message", "code": 50006} (further failures are not logged)`}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("discord refusing with an empty answer", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)
		discord.Refuse(500, "")

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"discord: notification not delivered: 500 Internal Server Error (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("discord refusing with an answer over 180 bytes", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)
		discord.Refuse(500, strings.Repeat("x", 181))

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"discord: notification not delivered: " + strings.Repeat("x", 180) + " (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("discord closing the connection", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)
		discord.Drop()

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"discord: notification not delivered: request failed (network error or timeout) (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("a second refusal", func(t *testing.T) {
		webhook, discord, log := setupDiscord(t)
		discord.Refuse(429, "slow down")
		webhook.Send(aliceOnline, events.EventStreamerOnline)
		discord.Refuse(500, "broken")

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"discord: notification not delivered: slow down (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("an event after a refusal", func(t *testing.T) {
		webhook, discord, _ := setupDiscord(t)
		discord.Drop()
		webhook.Send(aliceOnline, events.EventStreamerOnline)
		discord.Accept()

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := discord.Messages(), []notifytest.DiscordMessage{aliceOnlineOnDiscord()}; !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	t.Run("a webhook URL that is no URL", func(t *testing.T) {
		log := &fakeLog{}
		webhook := newDiscordWebhook(DiscordSettings{WebhookAPI: "http://discord webhook"}, log.Printf)

		webhook.Send(aliceOnline, events.EventStreamerOnline)

		if len(log.Lines) != 0 {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})
}

func TestDiscordEmbedFor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		want    discordEmbed
	}{
		{
			"a headline and details",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}},
			discordEmbed{"🟢 Alice is online", "Playing Valorant · Balance 1.2k points", 0x0A7A2F},
		},
		{
			"a headline alone",
			events.EventStreamerOffline,
			Message{Headline: "Alice is offline"},
			discordEmbed{"🔴 Alice is offline", "", 0xC23030},
		},
		{
			"no headline",
			events.EventStartup,
			Message{},
			discordEmbed{"▶️ Session started", "", 0x5C5C66},
		},
		{
			"an event without a glyph",
			"NONSENSE",
			Message{Headline: "Something changed"},
			discordEmbed{"Something changed", "", 0x5C5C66},
		},
		{
			"formatting characters",
			events.EventGainForWatchStreak,
			Message{"+10 points on d0cc_tv", []string{"Watching Counter_Strike *2*", "Balance 2.7k points"}},
			discordEmbed{`⏫ +10 points on d0cc\_tv`, `Watching Counter\_Strike \*2\* · Balance 2.7k points`, 0x0A7A2F},
		},
		{
			"a headline of exactly the limit with its glyph",
			events.EventStreamerOnline,
			Message{Headline: strings.Repeat("é", 254)},
			discordEmbed{"🟢 " + strings.Repeat("é", 254), "", 0x0A7A2F},
		},
		{
			"a headline one rune over the limit with its glyph",
			events.EventStreamerOnline,
			Message{Headline: strings.Repeat("é", 255)},
			discordEmbed{"🟢 " + strings.Repeat("é", 253) + "…", "", 0x0A7A2F},
		},
		{
			"details of exactly the limit",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{strings.Repeat("é", 4096)}},
			discordEmbed{"🟢 Alice is online", strings.Repeat("é", 4096), 0x0A7A2F},
		},
		{
			"details one rune over the limit",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{strings.Repeat("é", 4097)}},
			discordEmbed{"🟢 Alice is online", strings.Repeat("é", 4095) + "…", 0x0A7A2F},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := discordEmbedFor(tc.event, tc.message); got != tc.want {
				t.Errorf("discordEmbedFor(%q, %+v) = %+v, want %+v", tc.event, tc.message, got, tc.want)
			}
		})
	}
}

func TestDiscordFailure(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  string
		snippet string
		want    string
	}{
		{"an answer", "400 Bad Request", `{"message": "Unknown Webhook"}`, `{"message": "Unknown Webhook"}`},
		{"an answer with space around it", "400 Bad Request", "\n unknown webhook \n", "unknown webhook"},
		{"an answer of only space", "400 Bad Request", " \n", "400 Bad Request"},
		{"an empty answer", "500 Internal Server Error", "", "500 Internal Server Error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := discordFailure(tc.status, []byte(tc.snippet)); got != tc.want {
				t.Errorf("discordFailure(%q, %q) = %q, want %q", tc.status, tc.snippet, got, tc.want)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"plain text", "Alice is online", "Alice is online"},
		{"a backslash", `a\b`, `a\\b`},
		{"an asterisk", "a*b", `a\*b`},
		{"an underscore", "d0cc_tv", `d0cc\_tv`},
		{"a tilde", "a~b", `a\~b`},
		{"a backtick", "a`b", "a\\`b"},
		{"a pipe", "a|b", `a\|b`},
		{"a quote mark", "> a", `\> a`},
		{"every formatting character", "\\*_~`|>", "\\\\\\*\\_\\~\\`\\|\\>"},
		{"letters outside ASCII", "Pokémon 👋", "Pokémon 👋"},
		{"an empty string", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeMarkdown(tc.in); got != tc.want {
				t.Errorf("escapeMarkdown(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
