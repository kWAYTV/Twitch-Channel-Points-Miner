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
