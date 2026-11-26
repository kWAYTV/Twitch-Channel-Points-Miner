package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// MakePrediction places a bet for the provided event.
func (t *Client) MakePrediction(event *prediction.Event) error {
	if event == nil || event.Streamer == nil {
		return fmt.Errorf("nil prediction event")
	}
	if event.Decision.Amount < prediction.MinimumStake || event.Decision.OutcomeID == "" || event.EventID == "" {
		return fmt.Errorf("invalid prediction decision")
	}
	op := twitchapi.GQLOperations.MakePrediction.With(map[string]interface{}{
		"input": map[string]interface{}{
			"eventID":       event.EventID,
			"outcomeID":     event.Decision.OutcomeID,
			"points":        event.Decision.Amount,
			"transactionID": random.Hex(16),
		},
	})
	_, err := t.PostGQL(op)
	return err
}
