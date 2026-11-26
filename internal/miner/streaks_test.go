package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestStreakPriorityMinutes(t *testing.T) {
	tests := []struct {
		name      string
		startedAt time.Time
		want      float64
	}{
		{"a session that has not started", time.Time{}, 7},
		{"a session just started", testNow, 7},
		{"a second under ten hours in", testNow.Add(-10*time.Hour + time.Second), 7},
		{"exactly ten hours in", testNow.Add(-10 * time.Hour), 7},
		{"a second over ten hours in", testNow.Add(-10*time.Hour - time.Second), 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streakPriorityMinutes(tt.startedAt, testNow); got != tt.want {
				t.Errorf("streakPriorityMinutes(%s, %s) = %v, want %v", tt.startedAt, testNow, got, tt.want)
			}
		})
	}
}

func TestMiner_streakPriorityLimit(t *testing.T) {
	t.Run("a miner that has not started a session", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.streakPriorityLimit(h.now); got != 7 {
			t.Errorf("streakPriorityLimit() = %v, want 7", got)
		}
	})

	t.Run("no miner", func(t *testing.T) {
		var m *Miner

		if got := m.streakPriorityLimit(testNow); got != 7 {
			t.Errorf("streakPriorityLimit() = %v, want 7", got)
		}
	})
}

func TestStreakCooldownBlocksCurrentStream(t *testing.T) {
	offlineAt := testNow.Add(-10 * time.Minute)
	tests := []struct {
		name     string
		streamer *domain.Streamer
		want     bool
	}{
		{"no streamer", nil, false},
		{"never seen offline", &domain.Streamer{Stream: &domain.Stream{}}, false},
		{"offline a second over 30 minutes ago", &domain.Streamer{OfflineAt: testNow.Add(-30*time.Minute - time.Second)}, false},
		{"offline exactly 30 minutes ago", &domain.Streamer{OfflineAt: testNow.Add(-30 * time.Minute)}, true},
		{"offline 10 minutes ago and no stream", &domain.Streamer{OfflineAt: offlineAt}, true},
		{"a stream that went up after the offline", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{StreamUpAt: offlineAt.Add(time.Minute)}}, false},
		{"a stream that went up at the offline", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{StreamUpAt: offlineAt}}, false},
		{"a stream that went up exactly 2 minutes before the offline", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{StreamUpAt: offlineAt.Add(-2 * time.Minute)}}, false},
		{"a stream that went up a second over 2 minutes before the offline", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{StreamUpAt: offlineAt.Add(-2*time.Minute - time.Second)}}, true},
		{"a stream created after the offline, up time unknown", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{CreatedAt: offlineAt.Add(time.Minute)}}, false},
		{"a stream up and created well before the offline", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{StreamUpAt: offlineAt.Add(-time.Hour), CreatedAt: offlineAt.Add(-time.Hour)}}, true},
		{"a stream with no times", &domain.Streamer{OfflineAt: offlineAt, Stream: &domain.Stream{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streakCooldownBlocksCurrentStream(tt.streamer, testNow); got != tt.want {
				t.Errorf("streakCooldownBlocksCurrentStream() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestMiner_resolveTimedOutStreak(t *testing.T) {
	watching := domain.StreamerSettings{WatchStreak: true}
	tests := []struct {
		name        string
		streamer    *domain.Streamer
		want        bool
		wantMissing bool
	}{
		{"watch streaks off", &domain.Streamer{Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 7}}, false, true},
		{"a streak already resolved", &domain.Streamer{Settings: watching, Stream: &domain.Stream{MinuteWatched: 7}}, false, false},
		{"watched just under 7 minutes", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 6.99}}, false, true},
		{"watched exactly 7 minutes", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 7}}, true, false},
		{"watched over 7 minutes", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 7.01}}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, Options{})

			got := h.miner.resolveTimedOutStreak(tt.streamer, h.now)

			if got != tt.want {
				t.Errorf("resolveTimedOutStreak() = %t, want %t", got, tt.want)
			}
			if tt.streamer.Stream.WatchStreakMissing != tt.wantMissing {
				t.Errorf("WatchStreakMissing = %t, want %t", tt.streamer.Stream.WatchStreakMissing, tt.wantMissing)
			}
		})
	}

	t.Run("watched exactly 7 minutes, the streak never confirmed", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 7}}

		h.miner.resolveTimedOutStreak(alice, h.now)

		if alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = true, want false")
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.resolveTimedOutStreak(nil, h.now); got {
			t.Errorf("resolveTimedOutStreak() = true, want false")
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.resolveTimedOutStreak(&domain.Streamer{Settings: watching}, h.now); got {
			t.Errorf("resolveTimedOutStreak() = true, want false")
		}
	})
}

func TestMiner_shouldPrioritizeStreak(t *testing.T) {
	watching := domain.StreamerSettings{WatchStreak: true}
	tests := []struct {
		name     string
		streamer *domain.Streamer
		want     bool
	}{
		{"no streamer", nil, false},
		{"a streamer with no stream", &domain.Streamer{Settings: watching}, false},
		{"watch streaks off", &domain.Streamer{Stream: &domain.Stream{WatchStreakMissing: true}}, false},
		{"a pending streak", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true}}, true},
		{"a resolved streak", &domain.Streamer{Settings: watching, Stream: &domain.Stream{}}, false},
		{
			"a carryover still running",
			&domain.Streamer{
				Settings:                     watching,
				ResolvedStreakCarryover:      true,
				ResolvedStreakCarryoverUntil: testNow.Add(time.Minute),
				Stream:                       &domain.Stream{WatchStreakMissing: true},
			},
			false,
		},
		{
			"a carryover that ran out",
			&domain.Streamer{
				Settings:                     watching,
				ResolvedStreakCarryover:      true,
				ResolvedStreakCarryoverUntil: testNow.Add(-time.Second),
				Stream:                       &domain.Stream{WatchStreakMissing: true},
			},
			true,
		},
		{
			"offline 10 minutes ago on the same stream",
			&domain.Streamer{
				Settings:  watching,
				OfflineAt: testNow.Add(-10 * time.Minute),
				Stream:    &domain.Stream{WatchStreakMissing: true, StreamUpAt: testNow.Add(-time.Hour)},
			},
			false,
		},
		{
			"offline 10 minutes ago and restarted since",
			&domain.Streamer{
				Settings:  watching,
				OfflineAt: testNow.Add(-10 * time.Minute),
				Stream:    &domain.Stream{WatchStreakMissing: true, StreamUpAt: testNow.Add(-9 * time.Minute)},
			},
			true,
		},
		{"watched just under 7 minutes", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 6.99}}, true},
		{"watched exactly 7 minutes", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 7}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, Options{})

			if got := h.miner.shouldPrioritizeStreak(tt.streamer, h.now); got != tt.want {
				t.Errorf("shouldPrioritizeStreak() = %t, want %t", got, tt.want)
			}
		})
	}

	t.Run("a carryover that ran out is forgotten", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{
			Settings:                     domain.StreamerSettings{WatchStreak: true},
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: testNow.Add(-time.Second),
			Stream:                       &domain.Stream{WatchStreakMissing: true},
		}

		h.miner.shouldPrioritizeStreak(alice, h.now)

		if alice.ResolvedStreakCarryover || !alice.ResolvedStreakCarryoverUntil.IsZero() {
			t.Errorf("carryover = %t until %s, want none", alice.ResolvedStreakCarryover, alice.ResolvedStreakCarryoverUntil)
		}
	})
}

func TestMiner_shouldKeepActiveStreak(t *testing.T) {
	watching := domain.StreamerSettings{WatchStreak: true}
	tests := []struct {
		name     string
		streamer *domain.Streamer
		want     bool
	}{
		{"no streamer", nil, false},
		{"a pending streak not watched yet", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true}}, false},
		{"a pending streak watched for a minute", &domain.Streamer{Settings: watching, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 1}}, true},
		{"a resolved streak watched for a minute", &domain.Streamer{Settings: watching, Stream: &domain.Stream{MinuteWatched: 1}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, Options{})

			if got := h.miner.shouldKeepActiveStreak(tt.streamer, h.now); got != tt.want {
				t.Errorf("shouldKeepActiveStreak() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestMiner_syncActiveStreakWatch(t *testing.T) {
	options := Options{WatchPriority: []string{"ORDER"}}
	streakStreamer := func() *domain.Streamer {
		alice := freshStreamer("alice", "1", "Poker")
		alice.Settings.WatchStreak = true
		return alice
	}

	t.Run("a pending streak watched for a minute", func(t *testing.T) {
		h := setup(t, options)
		bob := freshStreamer("bob", "2", "Chess")
		alice := streakStreamer()
		alice.Stream.MinuteWatched = 1

		h.miner.syncActiveStreakWatch(alice, h.now)

		got := logins(h.miner.pickStreamersToWatch([]*domain.Streamer{bob, alice}))
		if want := []string{"alice", "bob"}; !reflect.DeepEqual(got, want) {
			t.Errorf("picked %q, want %q", got, want)
		}
	})

	t.Run("a pending streak not watched yet", func(t *testing.T) {
		h := setup(t, options)
		bob := freshStreamer("bob", "2", "Chess")
		alice := streakStreamer()

		h.miner.syncActiveStreakWatch(alice, h.now)

		got := logins(h.miner.pickStreamersToWatch([]*domain.Streamer{bob, alice}))
		if want := []string{"bob", "alice"}; !reflect.DeepEqual(got, want) {
			t.Errorf("picked %q, want %q", got, want)
		}
	})

	t.Run("a streamer with no channel id or login", func(t *testing.T) {
		h := setup(t, options)
		nobody := &domain.Streamer{Settings: domain.StreamerSettings{WatchStreak: true}, Stream: &domain.Stream{WatchStreakMissing: true, MinuteWatched: 1}}

		h.miner.syncActiveStreakWatch(nobody, h.now)

		got := logins(h.miner.pickStreamersToWatch([]*domain.Streamer{nobody}))
		if want := []string{}; !reflect.DeepEqual(got, want) {
			t.Errorf("picked %q, want %q", got, want)
		}
	})
}

func TestMiner_recoverStreaks(t *testing.T) {
	t.Run("no streamers", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.recoverStreaks(nil)

		wantLog := []string{"[INFO] streak recovery: 0 streaks expiring, 0 saved"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("an expiring streak saved", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].StreakSaved = true

		h.miner.recoverStreaks([]*domain.Streamer{alice})

		if want := []string{"alice"}; !reflect.DeepEqual(h.twitch.Recoveries, want) {
			t.Errorf("recoveries = %q, want %q", h.twitch.Recoveries, want)
		}
		wantLog := []string{"[INFO] streak recovery: 1 streaks expiring, 1 saved"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("no streak expiring", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")

		h.miner.recoverStreaks([]*domain.Streamer{alice})

		wantLog := []string{"[INFO] streak recovery: 0 streaks expiring, 0 saved"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Fail["RecoverStreak"] = errors.New("503")

		h.miner.recoverStreaks([]*domain.Streamer{alice})

		wantLog := []string{
			"[INFO] streak recovery Alice: 503",
			"[INFO] streak recovery: 1 streaks expiring, 0 saved",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})
}

func TestMiner_streakRecovery(t *testing.T) {
	t.Run("stopped before it starts", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		stop := make(chan struct{})
		close(stop)

		h.miner.streakRecovery([]*domain.Streamer{alice}, stop)

		if want := []string{"alice"}; !reflect.DeepEqual(h.twitch.Recoveries, want) {
			t.Errorf("recoveries = %q, want %q", h.twitch.Recoveries, want)
		}
	})
}
