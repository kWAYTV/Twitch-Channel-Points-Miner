package miner

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func (m *Miner) streakRecovery(streamers []*domain.Streamer, stop <-chan struct{}) {
	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()
	for {
		m.recoverStreaks(streamers)

		select {
		case <-ticker.C:
			continue
		case <-stop:
			return
		}
	}
}

// recoverStreaks asks Twitch to save every streamer's expiring watch streak, and logs the tally.
func (m *Miner) recoverStreaks(streamers []*domain.Streamer) {
	expiringstreaks := 0
	savedstreaks := 0
	for _, s := range streamers {
		if saved, err := m.twitch.RecoverStreak(s); err == nil {
			if saved {
				expiringstreaks += 1
				savedstreaks += 1
			}
		} else {
			expiringstreaks += 1
			m.logger.EmojiPrintf(":ambulance:", "streak recovery %s: %v", m.styledStreamerName(s), err)
		}
	}
	m.logger.EmojiPrintf(":ambulance:", "streak recovery: %d streaks expiring, %d saved", expiringstreaks, savedstreaks)
}

func (m *Miner) resolveTimedOutStreak(streamer *domain.Streamer, now time.Time) bool {
	if streamer == nil || streamer.Stream == nil {
		return false
	}
	if !streamer.Settings.WatchStreak || !streamer.Stream.WatchStreakMissing {
		return false
	}
	if streamer.Stream.MinuteWatched < m.streakPriorityLimit(now) {
		return false
	}
	streamer.Stream.WatchStreakMissing = false
	return true
}

func streakCooldownBlocksCurrentStream(streamer *domain.Streamer, now time.Time) bool {
	if streamer == nil || streamer.OfflineAt.IsZero() {
		return false
	}
	if now.Sub(streamer.OfflineAt) > 30*time.Minute {
		return false
	}
	if streamer.Stream == nil {
		return true
	}
	if streak.StreamStartSurvivesOffline(streamer.Stream.StreamUpAt, streamer.OfflineAt) {
		return false
	}
	if streak.StreamStartSurvivesOffline(streamer.Stream.CreatedAt, streamer.OfflineAt) {
		return false
	}
	return true
}

func (m *Miner) shouldPrioritizeStreak(streamer *domain.Streamer, now time.Time) bool {
	if streamer == nil || streamer.Stream == nil {
		return false
	}
	if !streamer.Settings.WatchStreak || streak.CarryoverActive(streamer, now) || !streamer.Stream.WatchStreakMissing {
		return false
	}
	if streakCooldownBlocksCurrentStream(streamer, now) {
		return false
	}
	// Keep streak priority long enough for Twitch to issue the streak check (typically ~15 minutes).
	return streamer.Stream.MinuteWatched < m.streakPriorityLimit(now)
}

func (m *Miner) shouldKeepActiveStreak(streamer *domain.Streamer, now time.Time) bool {
	if !m.shouldPrioritizeStreak(streamer, now) {
		return false
	}
	return streamer.Stream.MinuteWatched > 0
}

func (m *Miner) syncActiveStreakWatch(streamer *domain.Streamer, now time.Time) {
	if m == nil {
		return
	}
	key := streamerKey(streamer)
	if key == "" {
		return
	}
	if !m.shouldKeepActiveStreak(streamer, now) {
		m.activeStreakMu.Lock()
		if m.activeStreakWatches != nil {
			delete(m.activeStreakWatches, key)
		}
		m.activeStreakMu.Unlock()
		return
	}
	m.activeStreakMu.Lock()
	if m.activeStreakWatches == nil {
		m.activeStreakWatches = make(map[string]streak.ActiveWatch)
	}
	m.activeStreakWatches[key] = streak.ActiveWatch{
		BroadcastID: streamer.Stream.BroadcastID,
		CreatedAt:   streamer.Stream.CreatedAt,
	}
	m.activeStreakMu.Unlock()
}

// streakPriorityLimit is how many watched minutes a pending streak keeps its priority for.
func (m *Miner) streakPriorityLimit(now time.Time) float64 {
	if m == nil {
		return streak.PriorityMinutesBase
	}
	return streakPriorityMinutes(m.startedAt, now)
}

// streakPriorityMinutes adjusts streak priority duration:
// - default 7 minutes
// - extended to 20 minutes after 10 hours runtime to avoid churn late in long sessions.
func streakPriorityMinutes(startedAt, now time.Time) float64 {
	if startedAt.IsZero() {
		return streak.PriorityMinutesBase
	}
	if now.Sub(startedAt) > 10*time.Hour {
		return streak.PriorityMinutesExtended
	}
	return streak.PriorityMinutesBase
}
