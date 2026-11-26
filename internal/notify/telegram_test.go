package notify

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify/notifytest"
)

func aliceOnlineOnTelegram() notifytest.TelegramMessage {
	return notifytest.TelegramMessage{
		ChatID:              "123456789",
		Text:                "🟢 <b>Alice is online</b>\nPlaying Valorant · Balance 1.2k points",
		ParseMode:           "HTML",
		DisableNotification: "false",
		LinkPreviewOptions:  `{"is_disabled":true}`,
	}
}

func setupTelegram(t *testing.T, eventNames ...string) (*TelegramNotifier, *notifytest.FakeTelegram, *fakeLog) {
	t.Helper()
	telegram := notifytest.NewFakeTelegram(t)
	log := &fakeLog{}
	notifier := newTelegramNotifier(TelegramSettings{ChatID: "123456789", Token: telegram.Token, Events: eventNames}, telegram.URL, log.Printf)
	return notifier, telegram, log
}

func TestNewTelegramNotifier(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings TelegramSettings
	}{
		{"no token", TelegramSettings{ChatID: "123456789"}},
		{"no chat id", TelegramSettings{Token: "42:secret"}},
		{"only spaces for a token", TelegramSettings{ChatID: "123456789", Token: "  "}},
		{"only spaces for a chat id", TelegramSettings{ChatID: "  ", Token: "42:secret"}},
		{"no settings", TelegramSettings{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewTelegramNotifier(tc.settings); got != nil {
				t.Errorf("NewTelegramNotifier(%+v) = %+v, want nil", tc.settings, got)
			}
		})
	}

	t.Run("a chat id and a token", func(t *testing.T) {
		if got := NewTelegramNotifier(TelegramSettings{ChatID: "123456789", Token: "42:secret"}); got == nil {
			t.Errorf("NewTelegramNotifier = nil, want a notifier")
		}
	})
}

func TestTelegramNotifier_Send(t *testing.T) {
	t.Run("an event with no events list", func(t *testing.T) {
		notifier, telegram, log := setupTelegram(t)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := telegram.Messages(), []notifytest.TelegramMessage{aliceOnlineOnTelegram()}; !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
		if len(log.Lines) != 0 {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})

	t.Run("a chat id and a token with spaces around them", func(t *testing.T) {
		telegram := notifytest.NewFakeTelegram(t)
		notifier := newTelegramNotifier(TelegramSettings{ChatID: " 123456789 ", Token: " " + telegram.Token + "\n"}, telegram.URL, (&fakeLog{}).Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := telegram.Messages(), []notifytest.TelegramMessage{aliceOnlineOnTelegram()}; !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
	})

	t.Run("silent notifications turned on", func(t *testing.T) {
		telegram := notifytest.NewFakeTelegram(t)
		notifier := newTelegramNotifier(TelegramSettings{ChatID: "123456789", Token: telegram.Token, DisableNotification: true}, telegram.URL, (&fakeLog{}).Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []notifytest.TelegramMessage{{
			ChatID:              "123456789",
			Text:                "🟢 <b>Alice is online</b>\nPlaying Valorant · Balance 1.2k points",
			ParseMode:           "HTML",
			DisableNotification: "true",
			LinkPreviewOptions:  `{"is_disabled":true}`,
		}}
		if got := telegram.Messages(); !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
	})

	t.Run("a quiet event with no events list", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t)

		notifier.Send(aliceOnline, events.EventGainForWatch)

		if got := telegram.Messages(); len(got) != 0 {
			t.Errorf("telegram received %+v, want nothing", got)
		}
	})

	t.Run("a quiet event with a star in the events list", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t, "*")

		notifier.Send(Message{"+10 points on Alice", []string{"Watching"}}, events.EventGainForWatch)

		want := []notifytest.TelegramMessage{{
			ChatID:              "123456789",
			Text:                "➕ <b>+10 points on Alice</b>\nWatching",
			ParseMode:           "HTML",
			DisableNotification: "false",
			LinkPreviewOptions:  `{"is_disabled":true}`,
		}}
		if got := telegram.Messages(); !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
	})

	t.Run("an event missing from the events list", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t, "streamer_offline", "BET_LOSE")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got := telegram.Messages(); len(got) != 0 {
			t.Errorf("telegram received %+v, want nothing", got)
		}
	})

	t.Run("no event", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t)

		notifier.Send(aliceOnline, "")

		if got := telegram.Messages(); len(got) != 0 {
			t.Errorf("telegram received %+v, want nothing", got)
		}
	})

	t.Run("a notifier that is not set up", func(t *testing.T) {
		var notifier *TelegramNotifier

		notifier.Send(aliceOnline, events.EventStreamerOnline)
	})

	t.Run("the same event twice", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t)
		notifier.Send(aliceOnline, events.EventStreamerOnline)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := telegram.Messages(), []notifytest.TelegramMessage{aliceOnlineOnTelegram(), aliceOnlineOnTelegram()}; !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
	})

	t.Run("another bot's token", func(t *testing.T) {
		telegram := notifytest.NewFakeTelegram(t)
		log := &fakeLog{}
		notifier := newTelegramNotifier(TelegramSettings{ChatID: "123456789", Token: "7:other"}, telegram.URL, log.Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"telegram: notification not delivered: Unauthorized (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("telegram refusing", func(t *testing.T) {
		notifier, telegram, log := setupTelegram(t)
		telegram.Refuse(400, "Bad Request: chat not found")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"telegram: notification not delivered: Bad Request: chat not found (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("telegram refusing with no description", func(t *testing.T) {
		notifier, telegram, log := setupTelegram(t)
		telegram.RefuseWith(502, "<html>Bad Gateway</html>")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"telegram: notification not delivered: 502 Bad Gateway (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("telegram closing the connection", func(t *testing.T) {
		notifier, telegram, log := setupTelegram(t)
		telegram.Drop()

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"telegram: notification not delivered: request failed (network error or timeout) (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("a second refusal", func(t *testing.T) {
		notifier, telegram, log := setupTelegram(t)
		telegram.Refuse(429, "Too Many Requests: retry after 5")
		notifier.Send(aliceOnline, events.EventStreamerOnline)
		telegram.Refuse(400, "Bad Request: chat not found")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"telegram: notification not delivered: Too Many Requests: retry after 5 (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("an event after a refusal", func(t *testing.T) {
		notifier, telegram, _ := setupTelegram(t)
		telegram.Drop()
		notifier.Send(aliceOnline, events.EventStreamerOnline)
		telegram.Accept()

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := telegram.Messages(), []notifytest.TelegramMessage{aliceOnlineOnTelegram()}; !reflect.DeepEqual(got, want) {
			t.Errorf("telegram received %+v, want %+v", got, want)
		}
	})
}

func TestFormatTelegram(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		want    string
	}{
		{
			"a headline and details",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}},
			"🟢 <b>Alice is online</b>\nPlaying Valorant · Balance 1.2k points",
		},
		{"a headline alone", events.EventStreamerOnline, Message{Headline: "Alice is online"}, "🟢 <b>Alice is online</b>"},
		{"no headline", events.EventStartup, Message{}, "▶️ <b>Session started</b>"},
		{"details without a headline", events.EventBetWin, Message{"", []string{"Chose Yes"}}, "📈 <b>Bet won</b>\nChose Yes"},
		{"an event without a glyph", "NONSENSE", Message{Headline: "Something changed"}, "<b>Something changed</b>"},
		{
			"characters HTML reserves",
			events.EventChatMention,
			Message{"nick <3 & you", []string{`"<b>hi</b> & bye"`}},
			"💬 <b>nick &lt;3 &amp; you</b>\n\"&lt;b&gt;hi&lt;/b&gt; &amp; bye\"",
		},
		{
			"a headline of exactly the limit",
			events.EventBetLose,
			Message{Headline: strings.Repeat("é", 256)},
			"📉 <b>" + strings.Repeat("é", 256) + "</b>",
		},
		{
			"a headline one rune over the limit",
			events.EventBetLose,
			Message{Headline: strings.Repeat("é", 257)},
			"📉 <b>" + strings.Repeat("é", 255) + "…</b>",
		},
		{
			"details that fill the message exactly",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4085)}},
			"📉 <b>x</b>\n" + strings.Repeat("é", 4085),
		},
		{
			"details one rune too long for the message",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4086)}},
			"📉 <b>x</b>\n" + strings.Repeat("é", 4084) + "…",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatTelegram(tc.event, tc.message); got != tc.want {
				t.Errorf("formatTelegram(%q, %+v) = %q, want %q", tc.event, tc.message, got, tc.want)
			}
		})
	}
}

func TestTelegramFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status string
		body   string
		want   string
	}{
		{"an answer with a description", "400 Bad Request", `{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`, "Bad Request: chat not found"},
		{"an answer without a description", "400 Bad Request", `{"ok":false,"error_code":400}`, "400 Bad Request"},
		{"an answer with an empty description", "400 Bad Request", `{"ok":false,"description":""}`, "400 Bad Request"},
		{"an answer that is not JSON", "502 Bad Gateway", "<html>Bad Gateway</html>", "502 Bad Gateway"},
		{"an empty answer", "500 Internal Server Error", "", "500 Internal Server Error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := telegramFailure(tc.status, []byte(tc.body)); got != tc.want {
				t.Errorf("telegramFailure(%q, %q) = %q, want %q", tc.status, tc.body, got, tc.want)
			}
		})
	}
}
