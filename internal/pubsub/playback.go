package pubsub

import (
	"fmt"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func (p *Client) processPlaybackMessage(topic string, payload map[string]interface{}) error {
	channelID := ""
	if strings.HasPrefix(topic, "video-playback-by-id.") {
		channelID = strings.TrimPrefix(topic, "video-playback-by-id.")
	}
	if channelID == "" {
		return nil
	}
	streamer, ok := p.streamerMap[channelID]
	if !ok {
		return nil
	}
	if p.onPresence == nil {
		return nil
	}
	msgType := strings.ToLower(fmt.Sprint(payload["type"]))
	switch msgType {
	case "stream-up":
		if streamer.Stream == nil {
			streamer.Stream = domain.NewStream()
		} else {
			streamer.Stream.WatchStreakMissing = true
			streamer.Stream.ResetWatchProgress()
		}
		streamer.Stream.StreamUpAt = p.env.now()
		p.onPresence(streamer, true, msgType)
	case "viewcount":
		if streamer.Stream != nil && !streamer.Stream.StreamUpElapsed() {
			return nil
		}
		if justWentOffline(streamer.OfflineAt, p.env.now()) {
			return nil
		}
		live, err := p.account.IsStreamLive(streamer.ChannelID)
		if err != nil {
			p.logger.Errorf("live check %s: %v", p.streamerName(streamer), err)
			return nil
		}
		if live {
			p.onPresence(streamer, true, msgType)
		}
	case "stream-down":
		p.onPresence(streamer, false, msgType)
	}
	return nil
}

// justWentOffline says whether a streamer went offline less than a minute before now.
// A zero offlineAt means the streamer has not gone offline.
func justWentOffline(offlineAt, now time.Time) bool {
	return !offlineAt.IsZero() && now.Sub(offlineAt) < time.Minute
}
