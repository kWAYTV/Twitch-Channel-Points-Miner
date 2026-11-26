package notify

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestHeadlineOf(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		want    string
	}{
		{"a message with a headline", events.EventStreamerOnline, Message{Headline: "Alice is online"}, "Alice is online"},
		{"a message without one", events.EventStreamerOnline, Message{}, "Online"},
		{"only details", events.EventDropClaim, Message{Details: []string{"Summer"}}, "Drop claimed"},
		{"no headline and no event", "", Message{}, "Miner"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := headlineOf(tc.event, tc.message); got != tc.want {
				t.Errorf("headlineOf(%q, %+v) = %q, want %q", tc.event, tc.message, got, tc.want)
			}
		})
	}
}

func TestEventTitle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		event events.Event
		want  string
	}{
		{"startup", events.EventStartup, "Session started"},
		{"shutdown", events.EventShutdown, "Session ended"},
		{"streamer online", events.EventStreamerOnline, "Online"},
		{"streamer offline", events.EventStreamerOffline, "Offline"},
		{"gain for raid", events.EventGainForRaid, "Raid points"},
		{"gain for claim", events.EventGainForClaim, "Bonus claimed"},
		{"bonus claim", events.EventBonusClaim, "Bonus claimed"},
		{"gain for watch", events.EventGainForWatch, "Watch"},
		{"gain for watch streak", events.EventGainForWatchStreak, "Watch streak"},
		{"bet win", events.EventBetWin, "Bet won"},
		{"bet lose", events.EventBetLose, "Bet lost"},
		{"bet refund", events.EventBetRefund, "Bet refunded"},
		{"bet filters", events.EventBetFilters, "Bet skipped"},
		{"bet general", events.EventBetGeneral, "Bet"},
		{"bet failed", events.EventBetFailed, "Bet failed"},
		{"bet start", events.EventBetStart, "Bet starting"},
		{"moment claim", events.EventMomentClaim, "Moment claimed"},
		{"join raid", events.EventJoinRaid, "Raid joined"},
		{"drop claim", events.EventDropClaim, "Drop claimed"},
		{"drop status", events.EventDropStatus, "Drop progress"},
		{"chat mention", events.EventChatMention, "Chat mention"},
		{"an unknown event", "NONSENSE", "NONSENSE"},
		{"no event", "", "Miner"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := eventTitle(tc.event); got != tc.want {
				t.Errorf("eventTitle(%q) = %q, want %q", tc.event, got, tc.want)
			}
		})
	}
}

func TestEventColor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		event events.Event
		want  int
	}{
		{"streamer online", events.EventStreamerOnline, 0x0A7A2F},
		{"gain for watch", events.EventGainForWatch, 0x0A7A2F},
		{"gain for watch streak", events.EventGainForWatchStreak, 0x0A7A2F},
		{"bet win", events.EventBetWin, 0x0A7A2F},
		{"gain for claim", events.EventGainForClaim, 0x0A7A2F},
		{"bonus claim", events.EventBonusClaim, 0x0A7A2F},
		{"streamer offline", events.EventStreamerOffline, 0xC23030},
		{"bet lose", events.EventBetLose, 0xC23030},
		{"bet failed", events.EventBetFailed, 0xC23030},
		{"bet refund", events.EventBetRefund, 0xEDA100},
		{"drop claim", events.EventDropClaim, 0xE87BA4},
		{"drop status", events.EventDropStatus, 0xE87BA4},
		{"moment claim", events.EventMomentClaim, 0xE87BA4},
		{"join raid", events.EventJoinRaid, 0x2A78D6},
		{"gain for raid", events.EventGainForRaid, 0x2A78D6},
		{"chat mention", events.EventChatMention, 0x6250D6},
		{"bet start", events.EventBetStart, 0x6250D6},
		{"bet general", events.EventBetGeneral, 0x6250D6},
		{"startup", events.EventStartup, 0x5C5C66},
		{"shutdown", events.EventShutdown, 0x5C5C66},
		{"bet filters", events.EventBetFilters, 0x5C5C66},
		{"an unknown event", "NONSENSE", 0x5C5C66},
		{"no event", "", 0x5C5C66},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := eventColor(tc.event); got != tc.want {
				t.Errorf("eventColor(%q) = %#06X, want %#06X", tc.event, got, tc.want)
			}
		})
	}
}

func TestEventGlyphPrefix(t *testing.T) {
	for _, tc := range []struct {
		name  string
		event events.Event
		want  string
	}{
		{"an event with a glyph", events.EventStreamerOnline, "🟢 "},
		{"an unknown event", "NONSENSE", ""},
		{"no event", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := eventGlyphPrefix(tc.event); got != tc.want {
				t.Errorf("eventGlyphPrefix(%q) = %q, want %q", tc.event, got, tc.want)
			}
		})
	}
}
