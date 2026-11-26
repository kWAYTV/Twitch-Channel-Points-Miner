package logging

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestEmojize(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
		want string
	}{
		{"a known shortcode", ":rocket:", "🚀"},
		{"an unknown shortcode", ":unknown:", ":unknown:"},
		{"a known name without its colons", "rocket", "rocket"},
		{"an empty code", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := emojize(tc.code); got != tc.want {
				t.Errorf("emojize(%q) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestEmojiFor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		code  string
		event events.Event
		want  string
	}{
		{"a known event with a shortcode", ":partying_face:", events.EventStreamerOnline, "🟢"},
		{"a known event without a shortcode", "", events.EventBetLose, "📉"},
		{"no event with a shortcode", ":rocket:", "", "🚀"},
		{"no event with an unknown shortcode", ":unknown:", "", ":unknown:"},
		{"no event and no shortcode", "", "", ""},
		{"an unknown event with a shortcode", ":rocket:", "NONSENSE", "🚀"},
		{"an unknown event without a shortcode", "", "NONSENSE", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := emojiFor(tc.code, tc.event); got != tc.want {
				t.Errorf("emojiFor(%q, %q) = %q, want %q", tc.code, tc.event, got, tc.want)
			}
		})
	}
}
