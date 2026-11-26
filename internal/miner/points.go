package miner

import (
	"fmt"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func (m *Miner) handlePointsUpdate(streamer *domain.Streamer, previous int, reason string) {
	if !streamer.PointsInit {
		streamer.PointsInit = true
		m.analyticsRecordPoint(streamer, 0, "")
		return
	}
	delta := streamer.ChannelPoints - previous
	m.logPointsDelta(streamer, delta, reason)
}

func (m *Miner) logPointsDelta(streamer *domain.Streamer, delta int, reason string) {
	if delta == 0 {
		return
	}
	m.analyticsRecordPoint(streamer, delta, reason)
	who := m.nameWithPoints(streamer)
	valueColor, sign, amount := signedChange(delta)
	if reason == "" {
		return
	}
	reasonDisplay := reason
	if reason == "WATCH" {
		if ctx := m.watchContext(streamer); ctx != "" {
			reasonDisplay = fmt.Sprintf("%s %s", reason, ctx)
		}
	}
	line := fmt.Sprintf("%s%s%d%s → %s - Reason: %s", valueColor, sign, amount, colorReset, who, reasonDisplay)
	if event := events.EventFromGainReason(reason); event != "" {
		m.logger.EmojiEventf(":rocket:", event, "%s", line)
		return
	}
	mark := "➕"
	if sign == "-" {
		mark = "➖"
	}
	m.logger.EmojiPrintf(mark, "%s", line)
}

func (m *Miner) handlePubSubGain(streamer *domain.Streamer, earned int, reason string, balance int) {
	prev := streamer.ChannelPoints
	prevWatchStreakMissing := true
	if streamer != nil && streamer.Stream != nil {
		prevWatchStreakMissing = streamer.Stream.WatchStreakMissing
	}

	streamer.ChannelPoints = balanceAfterGain(prev, earned, balance)
	if !streamer.PointsInit {
		streamer.PointsInit = true
	}
	delta := earned
	if delta == 0 {
		delta = streamer.ChannelPoints - prev
	}
	m.logPointsDelta(streamer, delta, reason)
	m.updateHistory(streamer, reason, earned)
	if streamer != nil && streamer.Stream != nil && prevWatchStreakMissing != streamer.Stream.WatchStreakMissing {
		m.syncWarmStartCacheFromStreamer(streamer)
	}
}

func (m *Miner) updateHistory(streamer *domain.Streamer, reason string, amount int) {
	if reason == "" {
		return
	}
	streamer.HistoryMu.Lock()
	if streamer.History == nil {
		streamer.History = make(map[string]*domain.HistoryEntry)
	}
	entry, ok := streamer.History[reason]
	if !ok {
		entry = &domain.HistoryEntry{}
		streamer.History[reason] = entry
	}
	entry.Count++
	entry.Amount += amount
	streamer.HistoryMu.Unlock()
	if streamer.Stream == nil {
		return
	}
	if reason == "WATCH" {
		streamer.Stream.WatchCount++
		if streamer.Stream.WatchStreakMissing && streamer.Stream.WatchCount >= 2 {
			streak.MarkCompleted(streamer)
		}
		m.syncActiveStreakWatch(streamer, m.now())
		return
	}
	if reason == "WATCH_STREAK" {
		streak.MarkCompleted(streamer)
	}
	m.syncActiveStreakWatch(streamer, m.now())
}

func (m *Miner) contextRefresher(streamers []*domain.Streamer, stop <-chan struct{}) {
	ticks := m.human.Every(20*time.Minute, periodSpread, stop)
	for {
		select {
		case <-ticks:
			m.refreshContexts(streamers)
		case <-stop:
			return
		}
	}
}

// refreshContexts reloads every streamer's balance, a few at a time.
func (m *Miner) refreshContexts(streamers []*domain.Streamer) {
	forEachIndex(len(streamers), streamerConcurrency, func(i int) {
		s := streamers[i]
		prev := s.ChannelPoints
		if _, err := m.loadChannelPointsContextWithIdentityRepair(s); err != nil {
			m.logger.Printf("refresh %s: %v", m.styledStreamerName(s), err)
		} else {
			m.handlePointsUpdate(s, prev, "")
		}
	})
}
