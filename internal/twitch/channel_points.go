package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// LoadChannelPointsContext fetches points and claims any bonuses.
func (t *Client) LoadChannelPointsContext(streamer *domain.Streamer) (int, error) {
	op := twitchapi.GQLOperations.ChannelPointsContext.With(map[string]interface{}{"channelLogin": streamer.Username})
	var resp gqlChannelPointsContextResponse
	if err := t.PostGQLDecode(op, &resp); err != nil {
		return 0, err
	}
	name := t.anonymizer.Shown(streamer)
	channel := resp.Data.Community.Channel
	if channel == nil {
		return 0, fmt.Errorf("channel missing for %s: %w", name, ErrChannelNotFound)
	}
	pointsData := channel.Self.CommunityPoints
	balance := pointsData.Balance
	streamer.ChannelPoints = balance
	if len(pointsData.ActiveMultipliers) > 0 {
		streamer.ActiveMultipliers = pointsData.ActiveMultipliers
	} else {
		streamer.ActiveMultipliers = nil
	}
	if streamer.Settings.CommunityGoals {
		streamer.CommunityGoals = parseCommunityGoalsJSON(channel.CommunityPointsSettings.Goals)
		t.ContributeToCommunityGoals(streamer)
	}
	if pointsData.AvailableClaim != nil {
		claimID := pointsData.AvailableClaim.ID
		if claimID == "" {
			t.debugf("availableClaim present but missing id for %s", name)
			return balance, nil
		}
		if t.logger != nil {
			if t.anonymizer.Enabled() {
				t.logger.EmojiPrintf(":gift:", "Pending bonus detected for %s", name)
			} else {
				t.logger.EmojiPrintf(":gift:", "Pending bonus detected for %s (claim %s, channel %s)", name, claimID, streamer.ChannelID)
			}
		}
		if err := t.ClaimBonus(streamer, claimID); err != nil {
			if t.logger != nil {
				t.logger.Errorf("claim bonus on context load for %s failed: %v", name, err)
			}
		} else if t.logger != nil {
			if t.anonymizer.Enabled() {
				t.logger.Printf("Claim bonus success for %s", name)
			} else {
				t.logger.Printf("Claim bonus success for %s (claim %s)", name, claimID)
			}
		}
	}
	return balance, nil
}
