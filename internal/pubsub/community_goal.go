package pubsub

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
)

func (p *Client) processCommunityPointChannel(topic string, payload map[string]interface{}) error {
	channelID := strings.TrimPrefix(topic, "community-points-channel-v1.")
	streamer := p.streamerMap[channelID]
	if streamer == nil || !streamer.Settings.CommunityGoals {
		return nil
	}
	msgType := strings.ToLower(fmt.Sprint(payload["type"]))
	if streamer.CommunityGoals == nil {
		streamer.CommunityGoals = make(map[string]*domain.CommunityGoal)
	}
	switch msgType {
	case "community-goal-created", "community-goal-updated":
		data, _ := payload["data"].(map[string]interface{})
		goalData, _ := data["community_goal"].(map[string]interface{})
		if goal := domain.NewCommunityGoalFromPubSub(goalData); goal != nil && goal.ID != "" {
			streamer.CommunityGoals[goal.ID] = goal
		}
		p.account.ContributeToCommunityGoals(streamer)
	case "community-goal-deleted":
		data, _ := payload["data"].(map[string]interface{})
		goalData, _ := data["community_goal"].(map[string]interface{})
		id := jsonvalue.String(goalData["id"])
		if id != "" {
			delete(streamer.CommunityGoals, id)
		}
	}
	return nil
}
