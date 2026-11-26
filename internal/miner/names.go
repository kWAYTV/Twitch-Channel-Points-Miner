package miner

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/points"
)

func (m *Miner) styledStreamerName(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	name := m.anonymizer.Shown(streamer)
	if streamer.HasActiveMultipliers() {
		return fmt.Sprintf("%s%s%s", colorDropsAccent, name, colorReset)
	}
	return name
}

func (m *Miner) formattedStreamerPoints(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	balance := m.anonymizer.PseudoChannelPoints(streamer)
	if balance < 0 {
		balance = -balance
	}
	return points.Format(balance)
}

// nameWithPoints is the streamer's name followed by the balance held there.
func (m *Miner) nameWithPoints(streamer *domain.Streamer) string {
	return fmt.Sprintf("%s (%s%s%s points)", m.styledStreamerName(streamer), colorCyan, m.formattedStreamerPoints(streamer), colorReset)
}
