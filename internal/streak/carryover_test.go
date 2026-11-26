package streak

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestCarryoverExpiry(t *testing.T) {
	for _, row := range []struct {
		name     string
		streamer *domain.Streamer
		want     time.Time
	}{
		{"went offline twenty seconds ago", &domain.Streamer{OfflineAt: now.Add(-20 * time.Second)}, now.Add(-20 * time.Second).Add(30 * time.Minute)},
		{"went offline this moment", &domain.Streamer{OfflineAt: now}, now.Add(30 * time.Minute)},
		{"went offline one second under thirty minutes ago", &domain.Streamer{OfflineAt: now.Add(-30*time.Minute + time.Second)}, now.Add(time.Second)},
		{"went offline exactly thirty minutes ago", &domain.Streamer{OfflineAt: now.Add(-30 * time.Minute)}, now},
		{"went offline one second over thirty minutes ago", &domain.Streamer{OfflineAt: now.Add(-30*time.Minute - time.Second)}, now.Add(30 * time.Minute)},
		{"goes offline one second from now", &domain.Streamer{OfflineAt: now.Add(time.Second)}, now.Add(30 * time.Minute)},
		{"never went offline", &domain.Streamer{}, now.Add(30 * time.Minute)},
		{"no streamer", nil, now.Add(30 * time.Minute)},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := CarryoverExpiry(row.streamer, now); !got.Equal(row.want) {
				t.Errorf("CarryoverExpiry() = %s, want %s", got, row.want)
			}
		})
	}
}

func TestCarryoverActive(t *testing.T) {
	t.Run("a carryover that ends later", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(time.Second)}

		got := CarryoverActive(streamer, now)

		if got != true {
			t.Errorf("CarryoverActive() = %t, want true", got)
		}
		want := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(time.Second)}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that ends this moment", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now}

		got := CarryoverActive(streamer, now)

		if got != true {
			t.Errorf("CarryoverActive() = %t, want true", got)
		}
		want := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that ended a second ago is forgotten", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(-time.Second)}

		got := CarryoverActive(streamer, now)

		if got != false {
			t.Errorf("CarryoverActive() = %t, want false", got)
		}
		want := &domain.Streamer{}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("an end time with no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryoverUntil: now.Add(time.Second)}

		got := CarryoverActive(streamer, now)

		if got != false {
			t.Errorf("CarryoverActive() = %t, want false", got)
		}
		want := &domain.Streamer{ResolvedStreakCarryoverUntil: now.Add(time.Second)}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover with no end time", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true}

		got := CarryoverActive(streamer, now)

		if got != false {
			t.Errorf("CarryoverActive() = %t, want false", got)
		}
		want := &domain.Streamer{ResolvedStreakCarryover: true}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		if got := CarryoverActive(nil, now); got != false {
			t.Errorf("CarryoverActive(nil) = %t, want false", got)
		}
	})
}

func TestActivateCarryover(t *testing.T) {
	t.Run("a streamer who went offline twenty seconds ago", func(t *testing.T) {
		streamer := &domain.Streamer{OfflineAt: now.Add(-20 * time.Second)}

		ActivateCarryover(streamer, now)

		want := &domain.Streamer{
			OfflineAt:                    now.Add(-20 * time.Second),
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(-20 * time.Second).Add(30 * time.Minute),
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a streamer who never went offline", func(t *testing.T) {
		streamer := &domain.Streamer{}

		ActivateCarryover(streamer, now)

		want := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(30 * time.Minute)}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		ActivateCarryover(nil, now)
	})
}

func TestRememberCarryover(t *testing.T) {
	t.Run("a completed streak", func(t *testing.T) {
		streamer := &domain.Streamer{CompletedWatchStreak: true, OfflineAt: now}

		RememberCarryover(streamer, now)

		want := &domain.Streamer{
			CompletedWatchStreak:         true,
			OfflineAt:                    now,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(30 * time.Minute),
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover still running is extended from the latest offline time", func(t *testing.T) {
		streamer := &domain.Streamer{
			OfflineAt:                    now,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
		}

		RememberCarryover(streamer, now)

		want := &domain.Streamer{
			OfflineAt:                    now,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(30 * time.Minute),
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that has ended", func(t *testing.T) {
		streamer := &domain.Streamer{
			OfflineAt:                    now,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(-time.Second),
		}

		RememberCarryover(streamer, now)

		want := &domain.Streamer{OfflineAt: now}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("an end time with no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{OfflineAt: now, ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute)}

		RememberCarryover(streamer, now)

		want := &domain.Streamer{OfflineAt: now}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streak and no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{OfflineAt: now}

		RememberCarryover(streamer, now)

		want := &domain.Streamer{OfflineAt: now}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		RememberCarryover(nil, now)
	})
}

func TestRestoreCarryover(t *testing.T) {
	t.Run("a carryover still running, on a stream", func(t *testing.T) {
		streamer := &domain.Streamer{
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true},
		}

		RestoreCarryover(streamer, now)

		want := &domain.Streamer{
			CompletedWatchStreak:         true,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1"},
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover still running, with no stream", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute)}

		RestoreCarryover(streamer, now)

		want := &domain.Streamer{
			CompletedWatchStreak:         true,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that has ended", func(t *testing.T) {
		streamer := &domain.Streamer{
			CompletedWatchStreak:         true,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(-time.Second),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true},
		}

		RestoreCarryover(streamer, now)

		want := &domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true}}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a completed streak with no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{CompletedWatchStreak: true, ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute)}

		RestoreCarryover(streamer, now)

		want := &domain.Streamer{}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		RestoreCarryover(nil, now)
	})
}

func TestApplyCarryover(t *testing.T) {
	t.Run("a carryover still running", func(t *testing.T) {
		streamer := &domain.Streamer{
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true},
		}

		ApplyCarryover(streamer, now)

		want := &domain.Streamer{
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(10 * time.Minute),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1"},
		}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that has ended", func(t *testing.T) {
		streamer := &domain.Streamer{
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: now.Add(-time.Second),
			Stream:                       &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true},
		}

		ApplyCarryover(streamer, now)

		want := &domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true}}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true}}

		ApplyCarryover(streamer, now)

		want := &domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true}}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a carryover that has ended, with no stream, is kept", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(-time.Second)}

		ApplyCarryover(streamer, now)

		want := &domain.Streamer{ResolvedStreakCarryover: true, ResolvedStreakCarryoverUntil: now.Add(-time.Second)}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		ApplyCarryover(nil, now)
	})
}

func TestMarkCompleted(t *testing.T) {
	t.Run("a streamer on a stream", func(t *testing.T) {
		streamer := &domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true}}

		MarkCompleted(streamer)

		want := &domain.Streamer{CompletedWatchStreak: true, Stream: &domain.Stream{BroadcastID: "broadcast-1"}}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		streamer := &domain.Streamer{}

		MarkCompleted(streamer)

		want := &domain.Streamer{CompletedWatchStreak: true}
		if !reflect.DeepEqual(streamer, want) {
			t.Errorf("streamer = %+v, want %+v", streamer, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		MarkCompleted(nil)
	})
}

func TestClearCarryover(t *testing.T) {
	until := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)

	t.Run("a carryover", func(t *testing.T) {
		streamer := &domain.Streamer{
			CompletedWatchStreak:         true,
			ResolvedStreakCarryover:      true,
			ResolvedStreakCarryoverUntil: until,
		}

		clearCarryover(streamer)

		if streamer.ResolvedStreakCarryover || !streamer.ResolvedStreakCarryoverUntil.IsZero() {
			t.Errorf("carryover = %t until %s, want none", streamer.ResolvedStreakCarryover, streamer.ResolvedStreakCarryoverUntil)
		}
		if !streamer.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = false, want it left as it was")
		}
	})

	t.Run("an end time with no carryover", func(t *testing.T) {
		streamer := &domain.Streamer{ResolvedStreakCarryoverUntil: until}

		clearCarryover(streamer)

		if !streamer.ResolvedStreakCarryoverUntil.IsZero() {
			t.Errorf("ResolvedStreakCarryoverUntil = %s, want none", streamer.ResolvedStreakCarryoverUntil)
		}
	})
}
