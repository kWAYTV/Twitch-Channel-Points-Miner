package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

// pointsLine is the log line for a change of balance, as the console shows it.
func pointsLine(color, change, name, balance, reason string) string {
	return "[INFO] " + color + change + colorReset + " → " + name + " (" + colorCyan + balance + colorReset + " points) - Reason: " + reason
}

func TestMiner_handlePointsUpdate(t *testing.T) {
	t.Run("a first load", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_000}

		h.miner.handlePointsUpdate(alice, 0, "")

		if !alice.PointsInit {
			t.Errorf("PointsInit = false, want true")
		}
		if want := []recordedPoint{{Streamer: "alice", Points: 1_000}}; !reflect.DeepEqual(sink.Points, want) {
			t.Errorf("recorded points = %+v, want %+v", sink.Points, want)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("a first load with no analytics sink", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_000}

		h.miner.handlePointsUpdate(alice, 0, "")

		if !alice.PointsInit {
			t.Errorf("PointsInit = false, want true")
		}
	})

	t.Run("a balance that grew since the last load", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_450, PointsInit: true}

		h.miner.handlePointsUpdate(alice, 1_000, "")

		if want := []recordedPoint{{Streamer: "alice", Points: 1_450, Delta: 450}}; !reflect.DeepEqual(sink.Points, want) {
			t.Errorf("recorded points = %+v, want %+v", sink.Points, want)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("a balance that did not change", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_000, PointsInit: true}

		h.miner.handlePointsUpdate(alice, 1_000, "")

		if sink.Points != nil {
			t.Errorf("recorded points = %+v, want none", sink.Points)
		}
	})
}

func TestMiner_logPointsDelta(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		delta   int
		reason  string
		wantLog []string
	}{
		{"no change", Options{}, 0, "WATCH", nil},
		{"a change with no reason", Options{}, 10, "", nil},
		{"watch points", Options{}, 10, "WATCH", []string{pointsLine(colorGreen, "+10", "Alice", "1.01k", "WATCH")}},
		{
			"watch points with showing games on",
			Options{ShowGame: true},
			10,
			"WATCH",
			[]string{pointsLine(colorGreen, "+10", "Alice", "1.01k", "WATCH | "+colorGameLabel+"Game:"+colorReset+" Poker")},
		},
		{
			"a watch streak bonus with showing games on",
			Options{ShowGame: true},
			450,
			"WATCH_STREAK",
			[]string{pointsLine(colorGreen, "+450", "Alice", "1.01k", "WATCH_STREAK")},
		},
		{"a prediction win", Options{}, 500, "PREDICTION", []string{pointsLine(colorGreen, "+500", "Alice", "1.01k", "PREDICTION")}},
		{"a prediction stake", Options{}, -250, "PREDICTION", []string{pointsLine(colorRed, "-250", "Alice", "1.01k", "PREDICTION")}},
		{
			"a prediction win with emoji on",
			Options{Logger: logging.LoggerSettings{Emoji: true}},
			500,
			"PREDICTION",
			[]string{"[INFO] ➕ " + colorGreen + "+500" + colorReset + " → Alice (" + colorCyan + "1.01k" + colorReset + " points) - Reason: PREDICTION"},
		},
		{
			"a prediction stake with emoji on",
			Options{Logger: logging.LoggerSettings{Emoji: true}},
			-250,
			"PREDICTION",
			[]string{"[INFO] ➖ " + colorRed + "-250" + colorReset + " → Alice (" + colorCyan + "1.01k" + colorReset + " points) - Reason: PREDICTION"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)
			alice := freshStreamer("alice", "1", "Poker")
			alice.ChannelPoints = 1_010

			h.miner.logPointsDelta(alice, tt.delta, tt.reason)

			if got := h.logLines(t); !reflect.DeepEqual(got, tt.wantLog) {
				t.Errorf("log = %q, want %q", got, tt.wantLog)
			}
		})
	}

	t.Run("a stake, recorded for analytics", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 750}

		h.miner.logPointsDelta(alice, -250, "PREDICTION")

		if want := []recordedPoint{{Streamer: "alice", Points: 750, Delta: -250, Reason: "PREDICTION"}}; !reflect.DeepEqual(sink.Points, want) {
			t.Errorf("recorded points = %+v, want %+v", sink.Points, want)
		}
	})

	t.Run("no change, recorded for analytics", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 750}

		h.miner.logPointsDelta(alice, 0, "PREDICTION")

		if sink.Points != nil {
			t.Errorf("recorded points = %+v, want none", sink.Points)
		}
	})
}

func TestMiner_handlePubSubGain(t *testing.T) {
	options := Options{Username: "account"}
	resolvedEntry := streak.WarmStartEntry{
		AccountName:     "account",
		StreamerLogin:   "alice",
		ChannelID:       "1",
		BroadcastID:     "alice-broadcast",
		StreamCreatedAt: testNow.Add(-5 * time.Minute),
		CheckedAt:       testNow,
		IsOnline:        true,
	}

	t.Run("a first watch gain", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 100

		h.miner.handlePubSubGain(alice, 10, "WATCH", 110)

		if alice.ChannelPoints != 110 {
			t.Errorf("ChannelPoints = %d, want 110", alice.ChannelPoints)
		}
		if !alice.PointsInit {
			t.Errorf("PointsInit = false, want true")
		}
		if want := map[string]domain.HistoryEntry{"WATCH": {Count: 1, Amount: 10}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
		wantLog := []string{pointsLine(colorGreen, "+10", "Alice", "110", "WATCH")}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
		if entry, ok := h.cache.Get("alice"); ok {
			t.Errorf("cached entry = %+v, want none", entry)
		}
	})

	t.Run("a second watch gain", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 100
		h.miner.handlePubSubGain(alice, 10, "WATCH", 110)

		h.miner.handlePubSubGain(alice, 10, "WATCH", 120)

		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		if entry, _ := h.cache.Get("alice"); entry != resolvedEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, resolvedEntry)
		}
	})

	t.Run("a watch streak bonus", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 100

		h.miner.handlePubSubGain(alice, 450, "WATCH_STREAK", 550)

		if alice.ChannelPoints != 550 {
			t.Errorf("ChannelPoints = %d, want 550", alice.ChannelPoints)
		}
		if entry, _ := h.cache.Get("alice"); entry != resolvedEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, resolvedEntry)
		}
	})

	t.Run("a prediction stake with a stale balance reported", func(t *testing.T) {
		h := setup(t, options)
		alice := &domain.Streamer{Username: "alice", ChannelID: "1", ChannelPoints: 1_000_000, PointsInit: true}

		h.miner.handlePubSubGain(alice, -250_000, "PREDICTION", 1_000_000)

		if alice.ChannelPoints != 750_000 {
			t.Errorf("ChannelPoints = %d, want 750000", alice.ChannelPoints)
		}
		if want := map[string]domain.HistoryEntry{"PREDICTION": {Count: 1, Amount: -250_000}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
		wantLog := []string{pointsLine(colorRed, "-250000", "Alice", "750k", "PREDICTION")}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a prediction payout after its stake", func(t *testing.T) {
		h := setup(t, options)
		alice := &domain.Streamer{Username: "alice", ChannelID: "1", ChannelPoints: 1_000_000, PointsInit: true}
		h.miner.handlePubSubGain(alice, -250_000, "PREDICTION", 0)

		h.miner.handlePubSubGain(alice, 256_827, "PREDICTION", 0)

		if alice.ChannelPoints != 1_006_827 {
			t.Errorf("ChannelPoints = %d, want 1006827", alice.ChannelPoints)
		}
		if want := map[string]domain.HistoryEntry{"PREDICTION": {Count: 2, Amount: 6_827}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
	})

	t.Run("a balance report with nothing earned", func(t *testing.T) {
		h := setup(t, options)
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelID: "1", ChannelPoints: 1_000, PointsInit: true}

		h.miner.handlePubSubGain(alice, 0, "", 1_200)

		if alice.ChannelPoints != 1_200 {
			t.Errorf("ChannelPoints = %d, want 1200", alice.ChannelPoints)
		}
		if want := []recordedPoint{{Streamer: "alice", Points: 1_200, Delta: 200}}; !reflect.DeepEqual(sink.Points, want) {
			t.Errorf("recorded points = %+v, want %+v", sink.Points, want)
		}
		if alice.History != nil {
			t.Errorf("History = %v, want none", historyOf(alice))
		}
	})

	t.Run("nothing earned and nothing reported", func(t *testing.T) {
		h := setup(t, options)
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelID: "1", ChannelPoints: 1_000, PointsInit: true}

		h.miner.handlePubSubGain(alice, 0, "", 0)

		if alice.ChannelPoints != 1_000 {
			t.Errorf("ChannelPoints = %d, want 1000", alice.ChannelPoints)
		}
		if sink.Points != nil {
			t.Errorf("recorded points = %+v, want none", sink.Points)
		}
	})
}

func TestMiner_updateHistory(t *testing.T) {
	t.Run("a gain with no reason", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.updateHistory(alice, "", 10)

		if alice.History != nil {
			t.Errorf("History = %v, want none", historyOf(alice))
		}
	})

	t.Run("a first gain for a reason", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.updateHistory(alice, "CLAIM", 50)

		if want := map[string]domain.HistoryEntry{"CLAIM": {Count: 1, Amount: 50}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
		if alice.Stream.WatchCount != 0 {
			t.Errorf("WatchCount = %d, want 0", alice.Stream.WatchCount)
		}
		if !alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = false, want true")
		}
	})

	t.Run("a second gain for a reason", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")
		h.miner.updateHistory(alice, "CLAIM", 50)

		h.miner.updateHistory(alice, "CLAIM", 50)

		if want := map[string]domain.HistoryEntry{"CLAIM": {Count: 2, Amount: 100}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
	})

	t.Run("a gain for a streamer with no stream", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", ChannelID: "1"}

		h.miner.updateHistory(alice, "WATCH", 10)

		if want := map[string]domain.HistoryEntry{"WATCH": {Count: 1, Amount: 10}}; !reflect.DeepEqual(historyOf(alice), want) {
			t.Errorf("History = %v, want %v", historyOf(alice), want)
		}
	})

	t.Run("a first watch gain", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.updateHistory(alice, "WATCH", 10)

		if alice.Stream.WatchCount != 1 {
			t.Errorf("WatchCount = %d, want 1", alice.Stream.WatchCount)
		}
		if !alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = false, want true")
		}
		if alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = true, want false")
		}
	})

	t.Run("a second watch gain", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")
		h.miner.updateHistory(alice, "WATCH", 10)

		h.miner.updateHistory(alice, "WATCH", 10)

		if alice.Stream.WatchCount != 2 {
			t.Errorf("WatchCount = %d, want 2", alice.Stream.WatchCount)
		}
		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		if !alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = false, want true")
		}
	})

	t.Run("a second watch gain with the streak already resolved", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")
		alice.Stream.WatchStreakMissing = false
		h.miner.updateHistory(alice, "WATCH", 10)

		h.miner.updateHistory(alice, "WATCH", 10)

		if alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = true, want false")
		}
	})

	t.Run("a watch streak bonus", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.updateHistory(alice, "WATCH_STREAK", 450)

		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		if !alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = false, want true")
		}
		if alice.Stream.WatchCount != 0 {
			t.Errorf("WatchCount = %d, want 0", alice.Stream.WatchCount)
		}
	})
}

func TestMiner_refreshContexts(t *testing.T) {
	t.Run("a balance that grew", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := h.live("alice", "1", "Poker")
		alice.ChannelPoints = 100
		alice.PointsInit = true
		h.twitch.Channels["alice"].Points = 150

		h.miner.refreshContexts([]*domain.Streamer{alice})

		if alice.ChannelPoints != 150 {
			t.Errorf("ChannelPoints = %d, want 150", alice.ChannelPoints)
		}
		if want := []recordedPoint{{Streamer: "alice", Points: 150, Delta: 50}}; !reflect.DeepEqual(sink.Points, want) {
			t.Errorf("recorded points = %+v, want %+v", sink.Points, want)
		}
	})

	t.Run("a streamer whose balance was never loaded", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Points = 150

		h.miner.refreshContexts([]*domain.Streamer{alice})

		if alice.ChannelPoints != 150 {
			t.Errorf("ChannelPoints = %d, want 150", alice.ChannelPoints)
		}
		if !alice.PointsInit {
			t.Errorf("PointsInit = false, want true")
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		alice.ChannelPoints = 100
		h.twitch.Fail["LoadChannelPointsContext"] = errors.New("503")

		h.miner.refreshContexts([]*domain.Streamer{alice})

		if alice.ChannelPoints != 100 {
			t.Errorf("ChannelPoints = %d, want 100", alice.ChannelPoints)
		}
		wantLog := []string{"[INFO] refresh Alice: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})
}

func TestMiner_contextRefresher(t *testing.T) {
	t.Run("stopped before it starts", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Points = 150
		stop := make(chan struct{})
		close(stop)

		h.miner.contextRefresher([]*domain.Streamer{alice}, stop)

		if alice.ChannelPoints != 0 {
			t.Errorf("ChannelPoints = %d, want 0", alice.ChannelPoints)
		}
	})
}
