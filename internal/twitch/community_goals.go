package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// ContributeToCommunityGoals mirrors the site behavior by spending points into active community goals.
func (t *Client) ContributeToCommunityGoals(streamer *domain.Streamer) {
	if streamer == nil || !streamer.Settings.CommunityGoals || len(streamer.CommunityGoals) == 0 {
		return
	}
	if !hasActiveCommunityGoal(streamer.CommunityGoals) {
		return
	}

	op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.UserPointsContribution)
	if op.Variables == nil {
		op.Variables = map[string]interface{}{}
	}
	op.Variables["channelLogin"] = streamer.Username
	resp, err := t.PostGQL(op)
	if err != nil {
		return
	}
	contribs := jsonvalue.At(resp, "data.user.channel.self.communityPoints.goalContributions")
	arr, ok := contribs.([]interface{})
	if !ok {
		return
	}
	for _, raw := range arr {
		goalContribution, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		goalData, _ := goalContribution["goal"].(map[string]interface{})
		goalID, _ := goalData["id"].(string)
		if goalID == "" {
			continue
		}
		goal := streamer.CommunityGoals[goalID]
		if goal == nil {
			continue
		}
		userPoints := int(jsonvalue.Float(goalContribution["userPointsContributedThisStream"]))
		userLeft := goal.PerStreamUserMaximumContribution - userPoints
		amount := minInt(goal.AmountLeft(), userLeft, streamer.ChannelPoints)
		if amount > 0 {
			_ = t.ContributeToCommunityGoal(streamer, goalID, goal.Title, amount)
		}
	}
}

// ContributeToCommunityGoal sends a single contribution transaction.
func (t *Client) ContributeToCommunityGoal(streamer *domain.Streamer, goalID, title string, amount int) error {
	if amount <= 0 || goalID == "" {
		return nil
	}
	op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.ContributeCommunityPointsCommunityGoal)
	if op.Variables == nil {
		op.Variables = map[string]interface{}{}
	}
	op.Variables["input"] = map[string]interface{}{
		"amount":        amount,
		"channelID":     streamer.ChannelID,
		"goalID":        goalID,
		"transactionID": random.Hex(16),
	}
	resp, err := t.PostGQL(op)
	if err != nil {
		return err
	}
	switch refusal := jsonvalue.At(resp, "data.contributeCommunityPointsCommunityGoal.error").(type) {
	case nil:
	case string:
		if refusal != "" {
			return fmt.Errorf("unable to contribute to %s: %s", title, refusal)
		}
	case map[string]interface{}:
		reason := jsonvalue.FirstString(refusal, "code", "message")
		if reason == "" {
			reason = fmt.Sprint(refusal)
		}
		return fmt.Errorf("unable to contribute to %s: %s", title, reason)
	default:
		return fmt.Errorf("unable to contribute to %s: %v", title, refusal)
	}
	streamer.ChannelPoints -= amount
	if streamer.ChannelPoints < 0 {
		streamer.ChannelPoints = 0
	}
	return nil
}

// hasActiveCommunityGoal reports whether any goal is started and still in stock.
func hasActiveCommunityGoal(goals map[string]*domain.CommunityGoal) bool {
	for _, goal := range goals {
		if goal != nil && goal.Status == "STARTED" && goal.IsInStock {
			return true
		}
	}
	return false
}

func minInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}
