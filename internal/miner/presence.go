package miner

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func (m *Miner) updatePresence(streamer *domain.Streamer) {
	online, err := m.checkStreamerOnlineWithIdentityRepair(streamer)
	if err != nil {
		m.logger.Printf("online check %s: %v", m.styledStreamerName(streamer), err)
		return
	}
	if online && !streamer.PresenceKnown {
		m.applyWarmStartCache(streamer)
	}
	m.setPresence(streamer, online, "poll")
}

func (m *Miner) logOnline(streamer *domain.Streamer) {
	who := m.nameWithPoints(streamer)
	gameSuffix := ""
	if suffix := m.gameInfo(streamer); suffix != "" {
		gameSuffix = fmt.Sprintf(" | %s %s", fmt.Sprintf("%sPlaying:%s", colorGameLabel, colorReset), suffix)
	}
	extrainfoSuffix := ""
	if m.twitch != nil && streamer != nil && streamer.ChannelID != "" {
		extrainfoSuffix = m.twitch.LogExtraInfo(streamer.ChannelID, true)
	}
	m.logger.EmojiEventf(":partying_face:", events.EventStreamerOnline, "%s is %sOnline%s!%s%s", who, colorGreen, colorReset, gameSuffix, extrainfoSuffix)
}

func (m *Miner) logOffline(streamer *domain.Streamer) {
	who := m.nameWithPoints(streamer)
	extrainfoSuffix := ""
	if m.twitch != nil && streamer != nil && streamer.ChannelID != "" {
		extrainfoSuffix = m.twitch.LogExtraInfo(streamer.ChannelID, false)
	}
	m.logger.EmojiEventf(":sleeping:", events.EventStreamerOffline, "%s is %sOffline%s!%s", who, colorRed, colorReset, extrainfoSuffix)
}

// announcesPresence is false while the streamer list is first loaded: the load
// summary reports everyone in one line instead.
func (m *Miner) announcesPresence() bool {
	return !m.startingUp.Load()
}

func (m *Miner) announcePresence(streamer *domain.Streamer, online bool) {
	if !m.announcesPresence() {
		return
	}
	if online {
		m.logOnline(streamer)
	} else {
		m.logOffline(streamer)
	}
}

func (m *Miner) handlePubSubPresence(streamer *domain.Streamer, online bool, reason string) {
	m.setPresence(streamer, online, fmt.Sprintf("pubsub:%s", reason))
}

func (m *Miner) setPresence(streamer *domain.Streamer, online bool, reason string) {
	prevKnown := streamer.PresenceKnown
	prevOnline := streamer.IsOnline
	now := m.now()
	streamer.PresenceKnown = true
	if online != prevOnline || !prevKnown {
		if online {
			streamer.OnlineAt = now
		} else {
			streamer.OfflineAt = now
		}
	}
	if prevKnown && prevOnline != online {
		if online {
			streak.RestoreCarryover(streamer, now)
		} else {
			streak.RememberCarryover(streamer, now)
		}
	}
	streamer.IsOnline = online
	m.syncActiveStreakWatch(streamer, now)
	if m.chat != nil {
		m.chat.Update(streamer, online)
	}
	if online && m.showGameInfo {
		m.resolveGameName(streamer)
	}
	if !prevKnown || prevOnline != online {
		m.announcePresence(streamer, online)
	}
	m.syncWarmStartCacheFromStreamer(streamer)
}
