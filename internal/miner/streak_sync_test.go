package miner

import (
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestMiner_syncResolvedStreakState(t *testing.T) {
	startedAt := testNow.Add(-5 * time.Minute)
	earlier := testNow.Add(-3 * time.Hour)

	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.syncResolvedStreakState(nil, "a", earlier, true)
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", CompletedWatchStreak: true}

		h.miner.syncResolvedStreakState(alice, "a", earlier, true)

		if alice.ResolvedStreakCarryover {
			t.Errorf("ResolvedStreakCarryover = true, want false")
		}
	})

	tests := []struct {
		name            string
		completedBefore bool
		carryoverUntil  time.Time
		prevBroadcastID string
		prevCreatedAt   time.Time
		prevMissing     bool
		broadcastID     string
		missing         bool
		wantCompleted   bool
		wantMissing     bool
		wantCarryover   time.Time
	}{
		{
			name:            "a streak resolved since",
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "a", missing: false,
			wantCompleted: true, wantMissing: false,
		},
		{
			name:            "a streak still pending",
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "a", missing: true,
			wantCompleted: false, wantMissing: true,
		},
		{
			name:            "a streak that was already resolved",
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: false,
			broadcastID: "a", missing: false,
			wantCompleted: false, wantMissing: false,
		},
		{
			name:            "a new broadcast after a completed streak",
			completedBefore: true,
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "b", missing: true,
			wantCompleted: true, wantMissing: false, wantCarryover: testNow.Add(30 * time.Minute),
		},
		{
			name:            "a new start time after a completed streak",
			completedBefore: true,
			prevBroadcastID: "", prevCreatedAt: earlier, prevMissing: true,
			broadcastID: "", missing: true,
			wantCompleted: true, wantMissing: false, wantCarryover: testNow.Add(30 * time.Minute),
		},
		{
			name:            "a new broadcast with no completed streak",
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "b", missing: true,
			wantCompleted: false, wantMissing: true,
		},
		{
			name:            "the same broadcast after a completed streak",
			completedBefore: true,
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "a", missing: true,
			wantCompleted: true, wantMissing: true,
		},
		{
			name:            "an unknown earlier broadcast after a completed streak",
			completedBefore: true,
			prevBroadcastID: "", prevCreatedAt: time.Time{}, prevMissing: true,
			broadcastID: "b", missing: true,
			wantCompleted: true, wantMissing: true,
		},
		{
			name:            "a carryover already running",
			completedBefore: true, carryoverUntil: testNow.Add(10 * time.Minute),
			prevBroadcastID: "a", prevCreatedAt: startedAt, prevMissing: true,
			broadcastID: "a", missing: true,
			wantCompleted: true, wantMissing: false, wantCarryover: testNow.Add(10 * time.Minute),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, Options{})
			alice := &domain.Streamer{
				Username:                     "alice",
				CompletedWatchStreak:         tt.completedBefore,
				ResolvedStreakCarryover:      !tt.carryoverUntil.IsZero(),
				ResolvedStreakCarryoverUntil: tt.carryoverUntil,
				Stream: &domain.Stream{
					BroadcastID:        tt.broadcastID,
					CreatedAt:          startedAt,
					WatchStreakMissing: tt.missing,
				},
			}

			h.miner.syncResolvedStreakState(alice, tt.prevBroadcastID, tt.prevCreatedAt, tt.prevMissing)

			if alice.CompletedWatchStreak != tt.wantCompleted {
				t.Errorf("CompletedWatchStreak = %t, want %t", alice.CompletedWatchStreak, tt.wantCompleted)
			}
			if alice.Stream.WatchStreakMissing != tt.wantMissing {
				t.Errorf("WatchStreakMissing = %t, want %t", alice.Stream.WatchStreakMissing, tt.wantMissing)
			}
			if !alice.ResolvedStreakCarryoverUntil.Equal(tt.wantCarryover) {
				t.Errorf("ResolvedStreakCarryoverUntil = %s, want %s", alice.ResolvedStreakCarryoverUntil, tt.wantCarryover)
			}
		})
	}
}

func TestMarkStream(t *testing.T) {
	startedAt := testNow.Add(-5 * time.Minute)

	t.Run("a streamer on a stream", func(t *testing.T) {
		alice := &domain.Streamer{Username: "alice", Stream: &domain.Stream{BroadcastID: "a", CreatedAt: startedAt}}

		want := streamMark{broadcastID: "a", createdAt: startedAt, watchStreakMissing: false}
		if got := markStream(alice); got != want {
			t.Errorf("markStream() = %+v, want %+v", got, want)
		}
	})

	t.Run("a stream with a streak to collect", func(t *testing.T) {
		alice := &domain.Streamer{Username: "alice", Stream: &domain.Stream{BroadcastID: "a", CreatedAt: startedAt, WatchStreakMissing: true}}

		want := streamMark{broadcastID: "a", createdAt: startedAt, watchStreakMissing: true}
		if got := markStream(alice); got != want {
			t.Errorf("markStream() = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		alice := &domain.Streamer{Username: "alice"}

		want := streamMark{watchStreakMissing: true}
		if got := markStream(alice); got != want {
			t.Errorf("markStream() = %+v, want %+v", got, want)
		}
	})
}

func TestMiner_syncStreakSince(t *testing.T) {
	startedAt := testNow.Add(-5 * time.Minute)
	cached := Options{Username: "account", WatchStreakWarmStartCache: true}
	tests := []struct {
		name       string
		before     streamMark
		missing    bool
		wantCached bool
	}{
		{"nothing changed", streamMark{broadcastID: "a", createdAt: startedAt, watchStreakMissing: true}, true, false},
		{"another broadcast", streamMark{broadcastID: "b", createdAt: startedAt, watchStreakMissing: true}, true, true},
		{"another start time", streamMark{broadcastID: "a", createdAt: startedAt.Add(-time.Hour), watchStreakMissing: true}, true, true},
		{"the same start time in another zone", streamMark{broadcastID: "a", createdAt: startedAt.In(time.FixedZone("CEST", 2*60*60)), watchStreakMissing: true}, true, false},
		{"a streak collected since", streamMark{broadcastID: "a", createdAt: startedAt, watchStreakMissing: true}, false, true},
		{"no stream before", streamMark{watchStreakMissing: true}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, cached)
			alice := &domain.Streamer{
				Username: "alice",
				Stream:   &domain.Stream{BroadcastID: "a", CreatedAt: startedAt, WatchStreakMissing: tt.missing},
			}

			h.miner.syncStreakSince(alice, tt.before)

			if _, ok := h.cache.Get("alice"); ok != tt.wantCached {
				t.Errorf("cached = %t, want %t", ok, tt.wantCached)
			}
		})
	}

	t.Run("a streak collected since is marked complete", func(t *testing.T) {
		h := setup(t, cached)
		alice := &domain.Streamer{
			Username: "alice",
			Stream:   &domain.Stream{BroadcastID: "a", CreatedAt: startedAt},
		}

		h.miner.syncStreakSince(alice, streamMark{broadcastID: "a", createdAt: startedAt, watchStreakMissing: true})

		if !alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = false, want true")
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		h := setup(t, cached)
		alice := &domain.Streamer{Username: "alice"}

		h.miner.syncStreakSince(alice, streamMark{watchStreakMissing: true})

		if entry, ok := h.cache.Get("alice"); ok {
			t.Errorf("cached entry = %+v, want none", entry)
		}
	})
}
