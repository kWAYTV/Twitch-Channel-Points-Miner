package twitch

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// ClaimMoment redeems a community moment callout.
func (t *Client) ClaimMoment(streamer *domain.Streamer, momentID string) error {
	if momentID == "" {
		return nil
	}
	op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.CommunityMomentCalloutClaim)
	if op.Variables == nil {
		op.Variables = map[string]interface{}{}
	}
	op.Variables["input"] = map[string]interface{}{"momentID": momentID}
	_, err := t.PostGQL(op)
	return err
}
