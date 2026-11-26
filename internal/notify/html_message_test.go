package notify

import (
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestHTMLMessage(t *testing.T) {
	for _, tc := range []struct {
		name      string
		event     events.Event
		message   Message
		lineBreak string
		limit     int
		want      string
	}{
		{
			"a newline for a line break",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}},
			"\n",
			4096,
			"🟢 <b>Alice is online</b>\nPlaying Valorant · Balance 1.2k points",
		},
		{
			"a tag for a line break",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}},
			"<br>",
			4096,
			"🟢 <b>Alice is online</b><br>Playing Valorant · Balance 1.2k points",
		},
		{"a headline alone", events.EventStreamerOnline, Message{Headline: "Alice is online"}, "<br>", 4096, "🟢 <b>Alice is online</b>"},
		{"no headline", events.EventStartup, Message{}, "<br>", 4096, "▶️ <b>Session started</b>"},
		{
			"characters HTML reserves",
			events.EventChatMention,
			Message{"nick <3 & you", []string{`"<b>hi</b> & bye"`}},
			"<br>",
			4096,
			"💬 <b>nick &lt;3 &amp; you</b><br>\"&lt;b&gt;hi&lt;/b&gt; &amp; bye\"",
		},
		{
			"a headline one rune over its limit",
			events.EventBetLose,
			Message{Headline: strings.Repeat("é", 257)},
			"<br>",
			4096,
			"📉 <b>" + strings.Repeat("é", 255) + "…</b>",
		},
		{
			"details that fill the limit exactly",
			events.EventBetLose,
			Message{"x", []string{"abcdef"}},
			"<br>",
			20,
			"📉 <b>x</b><br>abcdef",
		},
		{
			"details one rune over the limit",
			events.EventBetLose,
			Message{"x", []string{"abcdefg"}},
			"<br>",
			20,
			"📉 <b>x</b><br>abcde…",
		},
		{
			"a limit with room for one rune of details",
			events.EventBetLose,
			Message{"x", []string{"abcdefg"}},
			"<br>",
			15,
			"📉 <b>x</b><br>…",
		},
		{
			"a limit with no room for details",
			events.EventBetLose,
			Message{"x", []string{"abcdefg"}},
			"<br>",
			14,
			"📉 <b>x</b>",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := htmlMessage(tc.event, tc.message, tc.lineBreak, tc.limit); got != tc.want {
				t.Errorf("htmlMessage(%q, %+v, %q, %d) = %q, want %q", tc.event, tc.message, tc.lineBreak, tc.limit, got, tc.want)
			}
		})
	}
}

func TestHeadlineText(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		want    string
	}{
		{"a headline", events.EventStreamerOnline, Message{Headline: "Alice is online"}, "Alice is online"},
		{"no headline", events.EventBetWin, Message{}, "Bet won"},
		{"a headline of exactly the limit", events.EventBetLose, Message{Headline: strings.Repeat("é", 256)}, strings.Repeat("é", 256)},
		{"a headline one rune over the limit", events.EventBetLose, Message{Headline: strings.Repeat("é", 257)}, strings.Repeat("é", 255) + "…"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := headlineText(tc.event, tc.message); got != tc.want {
				t.Errorf("headlineText(%q, %+v) = %q, want %q", tc.event, tc.message, got, tc.want)
			}
		})
	}
}

func TestDetailLine(t *testing.T) {
	for _, tc := range []struct {
		name    string
		message Message
		want    string
	}{
		{"two details", Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}}, "Playing Valorant · Balance 1.2k points"},
		{"one detail", Message{"Alice is online", []string{"Playing Valorant"}}, "Playing Valorant"},
		{"no details", Message{Headline: "Alice is online"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := detailLine(tc.message); got != tc.want {
				t.Errorf("detailLine(%+v) = %q, want %q", tc.message, got, tc.want)
			}
		})
	}
}

func TestHTMLEscape(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"plain text", "Alice is online", "Alice is online"},
		{"an ampersand", "a & b", "a &amp; b"},
		{"a less-than sign", "a < b", "a &lt; b"},
		{"a greater-than sign", "a > b", "a &gt; b"},
		{"a tag", "<b>hi</b>", "&lt;b&gt;hi&lt;/b&gt;"},
		{"text that is already escaped", "&amp;", "&amp;amp;"},
		{"quotes", `"hi" 'there'`, `"hi" 'there'`},
		{"an empty string", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := htmlEscape(tc.in); got != tc.want {
				t.Errorf("htmlEscape(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestLimitEscaped(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"one rune under the limit", "abcd", 5, "abcd"},
		{"exactly the limit", "abcde", 5, "abcde"},
		{"one rune over the limit", "abcdef", 5, "abcd…"},
		{"letters outside ASCII over the limit", "éèêëàâ", 5, "éèêë…"},
		{"a space where the cut falls", "abc def", 5, "abc…"},
		{"a cut inside an entity", "ab&amp;cd", 5, "ab…"},
		{"a cut right after an entity", "a&lt;bcdef", 6, "a&lt;…"},
		{"a cut inside the second entity", "&lt;a&amp;b", 7, "&lt;a…"},
		{"a limit of one", "abc", 1, "…"},
		{"a limit of zero", "abc", 0, ""},
		{"a negative limit", "abc", -1, ""},
		{"an empty string and a limit of zero", "", 0, ""},
		{"an empty string", "", 5, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := limitEscaped(tc.s, tc.n); got != tc.want {
				t.Errorf("limitEscaped(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
			}
		})
	}
}
