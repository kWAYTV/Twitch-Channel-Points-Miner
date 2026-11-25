package miner

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func (m *Miner) loadWarmStartCache() {
	if m == nil || !m.WatchStreakWarmStartCache {
		return
	}
	m.warmStartCache = streak.LoadWarmStartCache(m.warmStartCachePath, m.Username)
}

func (m *Miner) saveWarmStartCacheSnapshot(streamers []*domain.Streamer) {
	if m == nil || m.warmStartCache == nil {
		return
	}
	now := time.Now()
	for _, streamer := range streamers {
		m.warmStartCache.UpdateFromStreamer(streamer, now)
	}
	if err := m.warmStartCache.SaveIfDirty(); err != nil && m.logger != nil {
		m.logger.Debugf("warm-start cache save failed: %v", err)
	}
}

func (m *Miner) syncWarmStartCacheFromStreamer(streamer *domain.Streamer) {
	if m == nil || m.warmStartCache == nil || streamer == nil {
		return
	}
	m.warmStartCache.UpdateFromStreamer(streamer, time.Now())
	if err := m.warmStartCache.SaveIfDirty(); err != nil && m.logger != nil {
		m.logger.Debugf("warm-start cache save failed for %s: %v", m.rawStreamerName(streamer.Username), err)
	}
}

func (m *Miner) applyWarmStartCache(streamer *domain.Streamer) bool {
	if m == nil || m.warmStartCache == nil || streamer == nil || streamer.Stream == nil || !streamer.IsOnline {
		return false
	}
	entry, ok := m.warmStartCache.ResolvedEntryFor(streamer, time.Now())
	if !ok {
		return false
	}
	streamer.Stream.WatchStreakMissing = entry.WatchStreakMissing
	return true
}

// warmStartCachePathFor is where an account's watch-streak state is kept between runs.
func warmStartCachePathFor(username string) string {
	safeUsername := logging.SanitizeFilename(strings.ToLower(strings.TrimSpace(username)))
	if safeUsername == "" {
		return ""
	}
	return filepath.Join("log", fmt.Sprintf("watch_streak_cache.%s.json", safeUsername))
}
