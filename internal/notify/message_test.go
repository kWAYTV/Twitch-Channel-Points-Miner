package notify

import (
	"reflect"
	"strings"
	"testing"
)

func TestDescribe(t *testing.T) {
	online := "🥳 Alice (1.2k points) is Online! | Playing: Valorant | streak length 3 | expires at 2026-11-25 15:04:05.123 +0100 CET | started at 2026-11-25 12:00:00 +0100 CET | streak achievement timestamp 2026-11-25 13:00:00 +0100 CET"
	place := "🍀 Place 1,234 points on: Yes (BLUE), Points: 50,000, Users: 120 (45.20%), Odds: 1.80 (55.56%) for bob"
	won := "📊 EventPrediction: bob - will alice win? - Decision: A: Yes (BLUE) - Result: WIN, Gained: +500"
	drop := "DROP_PROGRESS [####----------------] 40% | Game / Camp / chan / 12m of 30m / remaining 18m / active"

	cases := []struct {
		in       string
		headline string
		details  []string
	}{
		{online, "Alice is online", []string{"Playing Valorant", "Balance 1.2k points", "Watch streak 3"}},
		{"😴 Alice (1.2k points) is Offline! | streak length 0 | expires at 2026-11-25 15:04:05 +0100 CET", "Alice is offline", []string{"Balance 1.2k points"}},
		{"😴 Alice (1.2k points) is Offline! | streak length -1 | started at sometime", "Alice is offline", []string{"Balance 1.2k points"}},
		{"Alice is Online!", "Alice is online", nil},
		// Points are the account's points on a channel, never the streamer's.
		{"🚀 +50 → Alice (12.4k points) - Reason: WATCH | Game: Valorant", "+50 points on Alice", []string{"Watching Valorant", "Balance 12.4k points"}},
		{"+10 → Alice (12.4k points) - Reason: WATCH", "+10 points on Alice", []string{"Watching", "Balance 12.4k points"}},
		{"+450 → Alice (1.5k points) - Reason: WATCH_STREAK | streak length 4", "+450 points on Alice", []string{"Watch streak 4", "Balance 1.5k points"}},
		{"+450 → Alice (1.5k points) - Reason: WATCH_STREAK", "+450 points on Alice", []string{"Watch streak", "Balance 1.5k points"}},
		{"+50 → Alice (1.5k points) - Reason: CLAIM", "+50 points on Alice", []string{"Bonus claim", "Balance 1.5k points"}},
		{"-500 → Alice (1k points) - Reason: PREDICTION", "−500 points on Alice", []string{"Prediction", "Balance 1k points"}},
		{place, "Bet placed on bob", []string{"1,234 points on Yes"}},
		{"Place bet on: Yes (BLUE), Points: 10, Users: 2 (50.00%), Odds: 2 (50%) for Stream", "Bet placed on Stream", []string{"On Yes"}},
		{won, "Won 500 points on bob", []string{`"will alice win?"`, "Chose Yes"}},
		{"EventPrediction: bob - overtime? - Decision: No - Result: LOSE, Gained: -500", "Lost 500 points on bob", []string{`"overtime?"`, "Chose No"}},
		{"EventPrediction: bob - overtime? - Decision: No - Result: REFUND, Refunded: +500", "Refunded 500 points on bob", []string{`"overtime?"`, "Chose No"}},
		{"Place bet after 4m0s for Alice", "Bet on Alice in 4m0s", nil},
		{"Skip bet for Alice: balance below minimum_points", "Bet skipped on Alice", []string{"The balance is below the minimum"}},
		{"prediction Alice: timeout", "Failed prediction on Alice", []string{"timeout"}},
		{"Joining raid from Alice to Bob", "Joined the raid from Alice to Bob", nil},
		{"Alice Claimed Moment", "Claimed a moment on Alice", nil},
		{"Claim Cool Hat (Summer Drops) 60/60 (100%)", "Claimed Cool Hat", []string{"Summer Drops campaign", "Progress 60/60 (100%)"}},
		{drop, "Drop 40% complete", []string{"Game / Camp / chan", "12m watched of 30m"}},
		{"nick at #alice wrote: hello <b>there</b>", "nick mentioned you in #alice", []string{`"hello <b>there</b>"`}},
		{"Start session: 'abc'", "Session started", nil},
		{"Ending session: 'abc'", "Session ended", nil},
		{"19 streamers loaded in 3.3 seconds: 3 online (Arrowcs, Fl0m, Bob), 16 offline", "19 streamers loaded", []string{"3 online: Arrowcs, Fl0m, Bob", "16 offline", "Took 3.3 seconds"}},
		{"1 streamer loaded in 0.4 seconds: 0 online, 1 offline", "1 streamer loaded", []string{"0 online", "1 offline", "Took 0.4 seconds"}},
		{"Something nobody planned for | Game: Valorant", "Something nobody planned for", []string{"Playing Valorant"}},
	}
	for _, tc := range cases {
		got := Describe(tc.in)
		if got.Headline != tc.headline || !reflect.DeepEqual(got.Details, tc.details) {
			t.Fatalf("Describe(%q)\n got %q %q\nwant %q %q", tc.in, got.Headline, got.Details, tc.headline, tc.details)
		}
		sentence := got.Sentence()
		if again := Sentence(sentence); again != sentence {
			t.Fatalf("sentence is not stable: %q -> %q", sentence, again)
		}
	}
}

func TestMessageSentence(t *testing.T) {
	m := Message{Headline: "+50 points on Alice", Details: []string{"Watching Valorant", "Balance 12.4k points"}}
	if got, want := m.Sentence(), "+50 points on Alice. Watching Valorant. Balance 12.4k points."; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := (Message{}).Sentence(); got != "" {
		t.Fatalf("an empty message should be an empty sentence, got %q", got)
	}
	if got, want := Sentence("Alice is Online!"), "Alice is online."; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	long := "nick at #alice wrote: " + strings.Repeat("é", 800)
	if got := Sentence(long); len([]rune(got)) > sentenceLimit {
		t.Fatalf("a sentence should be capped at %d runes, got %d", sentenceLimit, len([]rune(got)))
	}
}
