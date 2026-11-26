package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

func TestStreamerWatchBackedOff(t *testing.T) {
	until := testNow.Add(5 * time.Minute)
	tests := []struct {
		name     string
		streamer *domain.Streamer
		now      time.Time
		want     bool
	}{
		{"no streamer", nil, testNow, false},
		{"no backoff", &domain.Streamer{}, testNow, false},
		{"a second before the backoff ends", &domain.Streamer{WatchBackoffUntil: until}, until.Add(-time.Second), true},
		{"the moment the backoff ends", &domain.Streamer{WatchBackoffUntil: until}, until, false},
		{"a second after the backoff ends", &domain.Streamer{WatchBackoffUntil: until}, until.Add(time.Second), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streamerWatchBackedOff(tt.streamer, tt.now); got != tt.want {
				t.Errorf("streamerWatchBackedOff() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRecordMinuteWatchFailure(t *testing.T) {
	tests := []struct {
		name           string
		failuresBefore int
		wantFailures   int
		wantUntil      time.Time
	}{
		{"a first failure", 0, 1, time.Time{}},
		{"a second failure", 1, 2, time.Time{}},
		{"a third failure", 2, 3, testNow.Add(5 * time.Minute)},
		{"a fourth failure", 3, 4, testNow.Add(10 * time.Minute)},
		{"an eighth failure", 7, 8, testNow.Add(30 * time.Minute)},
		{"a ninth failure", 8, 9, testNow.Add(30 * time.Minute)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamer := &domain.Streamer{WatchFailureCount: tt.failuresBefore}

			recordMinuteWatchFailure(streamer, testNow)

			if streamer.WatchFailureCount != tt.wantFailures {
				t.Errorf("WatchFailureCount = %d, want %d", streamer.WatchFailureCount, tt.wantFailures)
			}
			if !streamer.WatchBackoffUntil.Equal(tt.wantUntil) {
				t.Errorf("WatchBackoffUntil = %s, want %s", streamer.WatchBackoffUntil, tt.wantUntil)
			}
		})
	}

	t.Run("no streamer", func(t *testing.T) {
		recordMinuteWatchFailure(nil, testNow)
	})

	t.Run("a third failure at an unset time", func(t *testing.T) {
		streamer := &domain.Streamer{WatchFailureCount: 2}
		before := time.Now()

		recordMinuteWatchFailure(streamer, time.Time{})

		after := time.Now()
		if got := streamer.WatchBackoffUntil; got.Before(before.Add(5*time.Minute)) || got.After(after.Add(5*time.Minute)) {
			t.Errorf("WatchBackoffUntil = %s, want five minutes after the wall clock (%s to %s)", got, before, after)
		}
	})
}

func TestRecordMinuteWatchSuccess(t *testing.T) {
	t.Run("a streamer backed off", func(t *testing.T) {
		streamer := &domain.Streamer{WatchFailureCount: 3, WatchBackoffUntil: testNow.Add(5 * time.Minute)}

		recordMinuteWatchSuccess(streamer)

		if streamer.WatchFailureCount != 0 {
			t.Errorf("WatchFailureCount = %d, want 0", streamer.WatchFailureCount)
		}
		if !streamer.WatchBackoffUntil.IsZero() {
			t.Errorf("WatchBackoffUntil = %s, want unset", streamer.WatchBackoffUntil)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		recordMinuteWatchSuccess(nil)
	})
}

func TestWatchInterval(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  time.Duration
	}{
		{"a negative count", -1, 20 * time.Second},
		{"no streamers", 0, 20 * time.Second},
		{"one streamer", 1, 20 * time.Second},
		{"two streamers", 2, 10 * time.Second},
		{"three streamers", 3, 6666666666 * time.Nanosecond},
		{"four streamers", 4, 5 * time.Second},
		{"five streamers", 5, 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := watchInterval(tt.count); got != tt.want {
				t.Errorf("watchInterval(%d) = %s, want %s", tt.count, got, tt.want)
			}
		})
	}
}

func TestSleepWithStop(t *testing.T) {
	stopped := make(chan struct{})
	close(stopped)
	tests := []struct {
		name     string
		duration time.Duration
		stop     chan struct{}
		want     bool
	}{
		{"no time to wait", 0, make(chan struct{}), false},
		{"a negative time to wait", -time.Second, make(chan struct{}), false},
		{"stopped before the wait", time.Hour, stopped, true},
		{"a wait that runs out", 300 * time.Microsecond, make(chan struct{}), false},
		{"a wait too short to split", 2 * time.Nanosecond, make(chan struct{}), false},
		{"a wait that does not split evenly", 7 * time.Nanosecond, make(chan struct{}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sleepWithStop(tt.duration, tt.stop); got != tt.want {
				t.Errorf("sleepWithStop(%s) = %t, want %t", tt.duration, got, tt.want)
			}
		})
	}

	t.Run("a wait that runs out lasts its whole time", func(t *testing.T) {
		started := time.Now()

		sleepWithStop(2*time.Millisecond, make(chan struct{}))

		if elapsed := time.Since(started); elapsed < 2*time.Millisecond {
			t.Errorf("sleepWithStop(2ms) returned after %s, want at least 2ms", elapsed)
		}
	})
}

func TestMiner_watchMinute(t *testing.T) {
	options := Options{Username: "account", WatchStreakWarmStartCache: true}

	t.Run("a minute sent", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")

		got := h.miner.watchMinute(alice)

		if !got {
			t.Errorf("watchMinute() = false, want true")
		}
		if want := []string{"alice"}; !reflect.DeepEqual(h.twitch.MinutesWatched, want) {
			t.Errorf("minutes watched = %q, want %q", h.twitch.MinutesWatched, want)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("a minute sent after two failures", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.WatchFailureCount = 2

		h.miner.watchMinute(alice)

		if alice.WatchFailureCount != 0 {
			t.Errorf("WatchFailureCount = %d, want 0", alice.WatchFailureCount)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Fail["SendMinuteWatched"] = errors.New("503")

		got := h.miner.watchMinute(alice)

		if !got {
			t.Errorf("watchMinute() = false, want true")
		}
		if alice.WatchFailureCount != 1 {
			t.Errorf("WatchFailureCount = %d, want 1", alice.WatchFailureCount)
		}
		if !alice.WatchBackoffUntil.IsZero() {
			t.Errorf("WatchBackoffUntil = %s, want unset", alice.WatchBackoffUntil)
		}
		wantLog := []string{"[ERROR] minute watch Alice: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch failing a third time", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.WatchFailureCount = 2
		h.twitch.Fail["SendMinuteWatched"] = errors.New("503")

		h.miner.watchMinute(alice)

		if want := testNow.Add(5 * time.Minute); !alice.WatchBackoffUntil.Equal(want) {
			t.Errorf("WatchBackoffUntil = %s, want %s", alice.WatchBackoffUntil, want)
		}
		wantLog := []string{
			"[INFO] minute watch Alice: backing off until 2026-05-17T12:05:00Z after repeated failures",
			"[ERROR] minute watch Alice: 503",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch failing a fourth time", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.WatchFailureCount = 3
		h.twitch.Fail["SendMinuteWatched"] = errors.New("503")

		h.miner.watchMinute(alice)

		if want := testNow.Add(10 * time.Minute); !alice.WatchBackoffUntil.Equal(want) {
			t.Errorf("WatchBackoffUntil = %s, want %s", alice.WatchBackoffUntil, want)
		}
		wantLog := []string{"[ERROR] minute watch Alice: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("an offline answer and the stream no longer live", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Online = false

		h.miner.watchMinute(alice)

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
		if alice.WatchFailureCount != 0 {
			t.Errorf("WatchFailureCount = %d, want 0", alice.WatchFailureCount)
		}
		wantLog := []string{offlineLine, "[ERROR] minute watch Alice: streamer offline"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("an offline answer while the stream is still live", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Fail["SendMinuteWatched"] = twitch.ErrStreamerOffline

		h.miner.watchMinute(alice)

		if !alice.IsOnline {
			t.Errorf("IsOnline = false, want true")
		}
		if alice.WatchFailureCount != 0 {
			t.Errorf("WatchFailureCount = %d, want 0", alice.WatchFailureCount)
		}
		wantLog := []string{
			"[INFO] minute watch Alice: transient offline response, keeping online",
			"[ERROR] minute watch Alice: streamer offline",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("an offline answer and the live check failing", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Fail["SendMinuteWatched"] = twitch.ErrStreamerOffline
		h.twitch.Fail["IsStreamLive"] = errors.New("503")

		h.miner.watchMinute(alice)

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
		wantLog := []string{
			"[INFO] live check Alice: 503",
			offlineLine,
			"[ERROR] minute watch Alice: streamer offline",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch reporting the streak complete", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].StreakComplete = true

		h.miner.watchMinute(alice)

		entry, _ := h.cache.Get("alice")
		wantEntry := streak.WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "alice",
			ChannelID:       "1",
			BroadcastID:     "alice-broadcast",
			StreamCreatedAt: testNow.Add(-5 * time.Minute),
			CheckedAt:       testNow,
			IsOnline:        true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})

	t.Run("a minute sent with nothing new about the streak", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")

		h.miner.watchMinute(alice)

		if entry, ok := h.cache.Get("alice"); ok {
			t.Errorf("cached entry = %+v, want none", entry)
		}
	})
}

func TestMiner_minuteWatcher(t *testing.T) {
	t.Run("stopped before it starts", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		stop := make(chan struct{})
		close(stop)

		h.miner.minuteWatcher([]*domain.Streamer{alice}, stop)

		if h.twitch.MinutesWatched != nil {
			t.Errorf("minutes watched = %q, want none", h.twitch.MinutesWatched)
		}
	})
}
