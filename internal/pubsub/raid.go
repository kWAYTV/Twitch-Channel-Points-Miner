package pubsub

import (
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func (p *Client) processRaidMessage(topic string, payload map[string]interface{}) error {
	channelID := strings.TrimPrefix(topic, "raid.")
	streamer := p.streamerMap[channelID]
	if streamer == nil || !streamer.Settings.FollowRaid {
		return nil
	}
	raidData, _ := payload["raid"].(map[string]interface{})
	raidID, _ := raidData["id"].(string)
	target, _ := raidData["target_login"].(string)
	if raidID == "" {
		return nil
	}
	if streamer.LastRaidID == raidID {
		return nil
	}
	streamer.LastRaidID = raidID
	if err := p.account.JoinRaid(streamer, raidID); err != nil {
		p.logger.Errorf("join raid %s->%s: %v", p.streamerName(streamer), p.name(target), err)
		return nil
	}
	if target == "" {
		target = "raid target"
	}
	p.logger.EmojiEventf(":performing_arts:", events.EventJoinRaid, "Joining raid from %s to %s", p.streamerName(streamer), p.name(target))
	return nil
}
