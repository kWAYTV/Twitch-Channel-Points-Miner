package privacy

import (
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func streamerKey(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	if id := strings.TrimSpace(streamer.ChannelID); id != "" {
		return "id:" + id
	}
	if name := normalizeKey(streamer.Username); name != "" {
		return "name:" + name
	}
	return ""
}

func normalizeKey(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
