package pubsub

import (
	"fmt"
	"math"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func (p *Client) logPredictionResult(event *prediction.Event, result map[string]interface{}) {
	if event == nil || result == nil {
		return
	}
	gained, _, _, resultType, resultString := event.ParseResult(result)
	event.BetConfirmed = true
	if p.onBetSettled != nil {
		p.onBetSettled(event.Settled(resultType, gained))
	}
	decisionLabel := event.DecisionLabel()
	if decisionLabel == "" {
		decisionLabel = fmt.Sprintf("Decision #%d", event.Decision.Choice+1)
	}
	color := resultColor(gained)
	label := event.String()
	if p.anonymizer.Enabled() {
		label = prediction.Label(p.anonymizer.Shown(event.Streamer), event.Title)
		resultString = resultType
	}
	resultEvent := events.EventFromBetResult(resultType)
	if resultEvent != "" {
		p.logger.EmojiEventf(
			":bar_chart:",
			resultEvent,
			"%s - Decision: %s - Result: %s%s%s",
			label,
			decisionLabel,
			color,
			resultString,
			ansi.ColorReset,
		)
	} else {
		p.logger.EmojiPrintf(
			":bar_chart:",
			"%s - Decision: %s - Result: %s%s%s",
			label,
			decisionLabel,
			color,
			resultString,
			ansi.ColorReset,
		)
	}
}

// resultColor is green for a gain, red for a loss and plain for neither.
func resultColor(gained int) string {
	if gained < 0 {
		return ansi.ColorRed
	}
	if gained == 0 {
		return ansi.ColorReset
	}
	return ansi.ColorGreen
}

func (p *Client) resolvePredictionFromChannel(event *prediction.Event, status string, eventMap map[string]interface{}) {
	if event == nil || !event.BetPlaced || event.Decision.Amount == 0 || event.ResultType != "" {
		return
	}
	resultType, pointsWon, settled := settledResult(status, winningOutcomeID(eventMap), event.Decision, event.Outcomes)
	if !settled {
		return
	}

	p.logPredictionResult(event, map[string]interface{}{
		"type":       resultType,
		"points_won": pointsWon,
	})

	p.forget(event.EventID)
}

// settledResult is how a bet came out once its event has a final status: a
// refund when it was cancelled, a win with its payout or a loss when it was
// resolved. settled is false while the event is open, or resolved with no
// winner named.
func settledResult(status, winningID string, decision prediction.Decision, outcomes []prediction.Outcome) (resultType string, pointsWon int, settled bool) {
	switch status {
	case "CANCELED", "CANCELLED":
		return "REFUND", 0, true
	case "RESOLVED":
		if winningID == "" {
			return "", 0, false
		}
		if decision.OutcomeID == winningID {
			return "WIN", payoutForOutcome(decision, outcomes, winningID), true
		}
		return "LOSE", 0, true
	default:
		return "", 0, false
	}
}

func winningOutcomeID(event map[string]interface{}) string {
	if id := jsonvalue.String(event["winning_outcome_id"]); id != "" {
		return id
	}
	if outcomes, ok := event["outcomes"].([]interface{}); ok {
		for _, raw := range outcomes {
			oc, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if win, ok := oc["is_winning_outcome"].(bool); ok && win {
				if id := jsonvalue.String(oc["id"]); id != "" {
					return id
				}
			}
			state := strings.ToUpper(jsonvalue.String(oc["state"]))
			if state == "RESOLVED" || state == "WINNER" || state == "WIN" {
				if id := jsonvalue.String(oc["id"]); id != "" {
					return id
				}
			}
		}
	}
	return ""
}

func payoutForOutcome(decision prediction.Decision, outcomes []prediction.Outcome, winningID string) int {
	if decision.Amount <= 0 || decision.OutcomeID != winningID {
		return 0
	}

	totalPoints := 0
	winPoints := 0
	for _, oc := range outcomes {
		totalPoints += oc.TotalPoints
		if oc.ID == winningID {
			winPoints = oc.TotalPoints
		}
	}

	if totalPoints == 0 || winPoints == 0 {
		return decision.Amount
	}

	payout := int(math.Round(float64(decision.Amount) * (float64(totalPoints) / float64(winPoints))))
	if payout < decision.Amount {
		return decision.Amount
	}
	return payout
}
