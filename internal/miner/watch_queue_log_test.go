package miner

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestMiner_logWatchQueueSlots(t *testing.T) {
	logging := Options{WatchQueueLogging: true}

	t.Run("nobody to watch", func(t *testing.T) {
		h := setup(t, logging)

		h.miner.pickStreamersToWatch(nil)

		want := []string{"[INFO] WATCH queue (≈20s between streamers):"}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("one streamer with no pending streak", func(t *testing.T) {
		h := setup(t, logging)
		alice := freshStreamer("alice", "1", "Poker")
		alice.Stream.WatchStreakMissing = false

		h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		want := []string{
			"[INFO] WATCH queue (≈20s between streamers):",
			"SLOT 1: Alice (reason=ORDER, streak=false, priorityGame=false, rank=1, pos=0)",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a pending streak on a priority game, then a plain pick", func(t *testing.T) {
		withGames := logging
		withGames.GamePriority = []string{"chess", "go"}
		h := setup(t, withGames)
		alice := freshStreamer("alice", "1", "Poker")
		alice.Stream.WatchStreakMissing = false
		carol := streakStreamer("carol", "3", "Go", 3)

		h.miner.pickStreamersToWatch([]*domain.Streamer{alice, carol})

		want := []string{
			"[INFO] WATCH queue (≈10s between streamers):",
			"SLOT 1: Carol (reason=STREAK, streak=true, priorityGame=true, rank=1, pos=1, streakRemaining=04m 00s)",
			"SLOT 2: Alice (reason=ORDER, streak=false, priorityGame=false, rank=3, pos=0)",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streak watched past its priority time, on a streamer with watch streaks off", func(t *testing.T) {
		h := setup(t, logging)
		alice := freshStreamer("alice", "1", "Poker")
		alice.Stream.MinuteWatched = 9

		h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		want := []string{
			"[INFO] WATCH queue (≈20s between streamers):",
			"SLOT 1: Alice (reason=ORDER, streak=false, priorityGame=false, rank=1, pos=0, streakRemaining=00s)",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("queue logging off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})
}
