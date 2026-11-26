package miner

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func (m *Miner) resolveGameName(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	if streamer.Stream != nil {
		if name := streamer.Stream.GameName(); name != "" {
			return name
		}
	}
	if !m.showGameInfo || m.twitch == nil || !streamer.IsOnline {
		return ""
	}
	before := markStream(streamer)
	if err := m.updateStreamWithIdentityRepair(streamer); err != nil {
		if m.logger != nil && m.logger.DebugEnabled() {
			m.logger.Debugf("update stream %s for game: %v", m.styledStreamerName(streamer), err)
		}
		return ""
	}
	m.syncStreakSince(streamer, before)
	if streamer.Stream != nil {
		return streamer.Stream.GameName()
	}
	return ""
}

func (m *Miner) gameInfo(streamer *domain.Streamer) string {
	if m == nil || !m.showGameInfo {
		return ""
	}
	return m.resolveGameName(streamer)
}

func (m *Miner) watchContext(streamer *domain.Streamer) string {
	name := m.gameInfo(streamer)
	if name != "" {
		label := fmt.Sprintf("| %sGame:%s", colorGameLabel, colorReset)
		return fmt.Sprintf("%s %s", label, name)
	}
	return ""
}

func (m *Miner) handleGameChange(streamer *domain.Streamer, previous, current string) {
	if m == nil || m.logger == nil || !m.showGameInfo {
		return
	}
	current = strings.TrimSpace(current)
	previous = strings.TrimSpace(previous)
	// Skip initial load while bringing a streamer online to avoid duplicate "now playing" spam.
	if current == "" || previous == "" || strings.EqualFold(previous, current) {
		return
	}
	m.logger.EmojiPrintf(":video_game:", "%s now playing: %s!", m.styledStreamerName(streamer), current)
}
