package pubsub

import (
	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func (p *Client) processRaidMessage(topic string, payload map[string]interface{}) error {
	channelID := topicChannelID(topic, raidTopic)
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
	// Twitch repeats the message while the raid counts down. The raid is
	// remembered only once it is joined, so a join that failed is tried again.
	p.react("raid:"+raidID, func() {
		if err := p.account.JoinRaid(streamer, raidID); err != nil {
			p.logger.Errorf("join raid %s->%s: %v", p.anonymizer.Shown(streamer), p.anonymizer.Name(target), err)
			return
		}
		streamer.LastRaidID = raidID
		if target == "" {
			target = "raid target"
		}
		p.logger.EmojiEventf(":performing_arts:", events.EventJoinRaid, "Joining raid from %s to %s", p.anonymizer.Shown(streamer), p.anonymizer.Name(target))
	})
	return nil
}
