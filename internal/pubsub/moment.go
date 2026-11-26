package pubsub

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func (p *Client) processMomentMessage(topic string, payload map[string]interface{}) error {
	channelID := strings.TrimPrefix(topic, "community-moments-channel-v1.")
	streamer := p.streamerMap[channelID]
	if streamer == nil || !streamer.Settings.ClaimMoments {
		return nil
	}
	if strings.ToLower(fmt.Sprint(payload["type"])) != "active" {
		return nil
	}
	data, _ := payload["data"].(map[string]interface{})
	momentID, _ := data["moment_id"].(string)
	if momentID == "" {
		return nil
	}
	if err := p.account.ClaimMoment(streamer, momentID); err != nil {
		p.logger.Errorf("claim moment %s: %v", p.streamerName(streamer), err)
		return nil
	}
	p.logger.EmojiEventf(":video_camera:", events.EventMomentClaim, "%s Claimed Moment", p.streamerName(streamer))
	return nil
}
