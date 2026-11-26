package events

import (
	"testing"
)

func TestNormalizeEventName(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want Event
	}{
		{"an event name as written", "BET_WIN", EventBetWin},
		{"a lower case name", "bet_win", EventBetWin},
		{"a name with spaces around it", "  Streamer_Online\t", EventStreamerOnline},
		{"a name that is no event", "nonsense", "NONSENSE"},
		{"a star", "*", "*"},
		{"a name outside ASCII", "évènement", "ÉVÈNEMENT"},
		{"an empty name", "", ""},
		{"only spaces", "   ", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeEventName(tc.raw); got != tc.want {
				t.Errorf("NormalizeEventName(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestEventFromGainReason(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		want   Event
	}{
		{"a watch tick", "WATCH", EventGainForWatch},
		{"a watch streak", "WATCH_STREAK", EventGainForWatchStreak},
		{"a bonus claim", "CLAIM", EventGainForClaim},
		{"a raid", "RAID", EventGainForRaid},
		{"a lower case reason with spaces around it", " watch ", EventGainForWatch},
		{"a reason with no event", "PREDICTION", ""},
		{"an empty reason", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := EventFromGainReason(tc.reason); got != tc.want {
				t.Errorf("EventFromGainReason(%q) = %q, want %q", tc.reason, got, tc.want)
			}
		})
	}
}

func TestEventFromBetResult(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result string
		want   Event
	}{
		{"a win", "WIN", EventBetWin},
		{"a loss", "LOSE", EventBetLose},
		{"a refund", "REFUND", EventBetRefund},
		{"a lower case result with spaces around it", " win ", EventBetWin},
		{"a result with no event", "CANCELED", ""},
		{"an empty result", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := EventFromBetResult(tc.result); got != tc.want {
				t.Errorf("EventFromBetResult(%q) = %q, want %q", tc.result, got, tc.want)
			}
		})
	}
}

func TestEventGlyph(t *testing.T) {
	type glyph struct {
		emoji string
		tone  Tone
		ok    bool
	}
	for _, tc := range []struct {
		name  string
		event Event
		want  glyph
	}{
		{"startup", EventStartup, glyph{"▶️", TonePositive, true}},
		{"shutdown", EventShutdown, glyph{"⏹️", ToneNeutral, true}},
		{"streamer online", EventStreamerOnline, glyph{"🟢", TonePositive, true}},
		{"streamer offline", EventStreamerOffline, glyph{"🔴", ToneNegative, true}},
		{"gain for watch", EventGainForWatch, glyph{"➕", TonePositive, true}},
		{"gain for watch streak", EventGainForWatchStreak, glyph{"⏫", TonePositive, true}},
		{"gain for claim", EventGainForClaim, glyph{"✅", TonePositive, true}},
		{"bonus claim", EventBonusClaim, glyph{"✅", TonePositive, true}},
		{"gain for raid", EventGainForRaid, glyph{"📡", TonePositive, true}},
		{"join raid", EventJoinRaid, glyph{"↪️", TonePositive, true}},
		{"bet win", EventBetWin, glyph{"📈", TonePositive, true}},
		{"bet lose", EventBetLose, glyph{"📉", ToneNegative, true}},
		{"bet refund", EventBetRefund, glyph{"↩️", ToneNeutral, true}},
		{"bet start", EventBetStart, glyph{"⏳", ToneNeutral, true}},
		{"bet general", EventBetGeneral, glyph{"🎯", ToneNeutral, true}},
		{"bet filters", EventBetFilters, glyph{"⏭️", ToneNeutral, true}},
		{"bet failed", EventBetFailed, glyph{"⛔", ToneNegative, true}},
		{"drop claim", EventDropClaim, glyph{"📦", TonePositive, true}},
		{"drop status", EventDropStatus, glyph{"📋", ToneNeutral, true}},
		{"moment claim", EventMomentClaim, glyph{"⚡", TonePositive, true}},
		{"chat mention", EventChatMention, glyph{"💬", ToneNeutral, true}},
		{"an unknown event", "NONSENSE", glyph{"", "", false}},
		{"no event", "", glyph{"", "", false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			emoji, tone, ok := EventGlyph(tc.event)
			if got := (glyph{emoji, tone, ok}); got != tc.want {
				t.Errorf("EventGlyph(%q) = %+v, want %+v", tc.event, got, tc.want)
			}
		})
	}
}
