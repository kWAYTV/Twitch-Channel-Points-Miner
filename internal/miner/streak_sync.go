package miner

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func (m *Miner) syncResolvedStreakState(streamer *domain.Streamer, prevBroadcastID string, prevCreatedAt time.Time, prevWatchStreakMissing bool) {
	if streamer == nil || streamer.Stream == nil {
		return
	}
	if prevWatchStreakMissing && !streamer.Stream.WatchStreakMissing {
		streamer.CompletedWatchStreak = true
	}
	segmentRolled := (prevBroadcastID != "" && streamer.Stream.BroadcastID != "" && prevBroadcastID != streamer.Stream.BroadcastID) ||
		(!prevCreatedAt.IsZero() && !streamer.Stream.CreatedAt.IsZero() && !prevCreatedAt.Equal(streamer.Stream.CreatedAt))
	now := m.now()
	if segmentRolled && streamer.CompletedWatchStreak {
		streak.ActivateCarryover(streamer, now)
	}
	streak.ApplyCarryover(streamer, now)
}

// streamMark is what is remembered of a streamer's stream before asking Twitch
// about it again, to tell afterwards what changed.
type streamMark struct {
	broadcastID        string
	createdAt          time.Time
	watchStreakMissing bool
}

// markStream marks the streamer's stream as it is now. A streamer with no
// stream yet has a streak still to collect.
func markStream(streamer *domain.Streamer) streamMark {
	if streamer.Stream == nil {
		return streamMark{watchStreakMissing: true}
	}
	return streamMark{
		broadcastID:        streamer.Stream.BroadcastID,
		createdAt:          streamer.Stream.CreatedAt,
		watchStreakMissing: streamer.Stream.WatchStreakMissing,
	}
}

// syncStreakSince brings the streak state up to date with a stream refreshed
// since the mark, and saves it to the warm-start cache when the stream changed.
func (m *Miner) syncStreakSince(streamer *domain.Streamer, before streamMark) {
	m.syncResolvedStreakState(streamer, before.broadcastID, before.createdAt, before.watchStreakMissing)
	if streamer.Stream != nil && !before.same(markStream(streamer)) {
		m.syncWarmStartCacheFromStreamer(streamer)
	}
}

func (a streamMark) same(b streamMark) bool {
	return a.broadcastID == b.broadcastID && a.createdAt.Equal(b.createdAt) && a.watchStreakMissing == b.watchStreakMissing
}
