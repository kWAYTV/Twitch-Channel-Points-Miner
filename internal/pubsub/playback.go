package pubsub

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func (p *Client) processPlaybackMessage(topic string, payload map[string]interface{}) error {
	channelID := topicChannelID(topic, playbackTopic)
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
	msgType := payloadType(payload)
	switch msgType {
	case "stream-up":
		if streamer.Stream == nil {
			streamer.Stream = domain.NewStream()
		} else {
			streamer.Stream.StartNewBroadcast()
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
			p.logger.Errorf("live check %s: %v", p.anonymizer.Shown(streamer), err)
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
