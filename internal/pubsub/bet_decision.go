package pubsub

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// skippedBet says why a bet is not placed. public is safe to log when names and
// balances are anonymised; detail is logged otherwise.
type skippedBet struct {
	event  events.Event
	public string
	detail string
}

// decideBet chooses the outcome and the stake, or returns why no bet is placed.
// The choice is kept on the event.
func decideBet(event *prediction.Event) (prediction.Decision, *skippedBet) {
	streamer := event.Streamer
	bet := streamer.Settings.Bet
	if event.Status != "ACTIVE" {
		reason := "event status is " + event.Status
		return prediction.Decision{}, &skippedBet{events.EventBetGeneral, reason, reason}
	}
	if belowMinimumPoints(streamer) {
		return prediction.Decision{}, &skippedBet{
			events.EventBetFilters,
			"balance below minimum_points",
			fmt.Sprintf("balance %d <= minimum_points %d", streamer.ChannelPoints, *bet.MinimumPoints),
		}
	}
	decision := event.Decide(streamer.ChannelPoints)
	if decision.OutcomeID == "" {
		return decision, &skippedBet{events.EventBetGeneral, "no outcome selected", "no outcome selected"}
	}
	if skip, compared, reason := event.ShouldSkipByFilter(); skip {
		return decision, filterSkip(reason, compared)
	}
	if decision.Amount < prediction.MinimumStake {
		return decision, &skippedBet{events.EventBetFilters, "below Twitch minimum", stakeTooSmall(streamer, decision.Amount)}
	}
	return decision, nil
}

// belowMinimumPoints says whether the balance is at or under the streamer's
// minimum_points, where no bet is placed.
func belowMinimumPoints(streamer *domain.Streamer) bool {
	minimum := streamer.Settings.Bet.MinimumPoints
	return minimum != nil && streamer.ChannelPoints <= *minimum
}

func filterSkip(reason string, compared float64) *skippedBet {
	if reason != "" {
		return &skippedBet{events.EventBetFilters, reason, reason}
	}
	return &skippedBet{
		events.EventBetFilters,
		"filter_condition not satisfied",
		fmt.Sprintf("filter_condition not satisfied (current %s)", prediction.FormatFloat(compared)),
	}
}

// stakeTooSmall says which limit kept the stake under Twitch's minimum.
func stakeTooSmall(streamer *domain.Streamer, stake int) string {
	if streamer.ChannelPoints < prediction.MinimumStake {
		return fmt.Sprintf("balance %d below Twitch minimum %d", streamer.ChannelPoints, prediction.MinimumStake)
	}
	if streamer.Settings.Bet.Strategy == domain.StrategyKelly && stake == 0 {
		return "no outcome is underpriced: none has a larger share of the viewers than of the pool"
	}
	if limit := streamer.Settings.Bet.MaxPoints; limit != nil && *limit < prediction.MinimumStake {
		return fmt.Sprintf("max_points %d below Twitch minimum %d", *limit, prediction.MinimumStake)
	}
	return fmt.Sprintf("calculated stake %d below Twitch minimum", stake)
}
