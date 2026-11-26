package pubsub

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// twitchMinimumStake is the smallest bet Twitch accepts.
const twitchMinimumStake = 10

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
	if bet.MinimumPoints != nil && streamer.ChannelPoints <= *bet.MinimumPoints {
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
	if decision.Amount < twitchMinimumStake {
		return decision, &skippedBet{events.EventBetFilters, "below Twitch minimum", stakeTooSmall(streamer, decision.Amount)}
	}
	return decision, nil
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
	if streamer.ChannelPoints < twitchMinimumStake {
		return fmt.Sprintf("balance %d below Twitch minimum %d", streamer.ChannelPoints, twitchMinimumStake)
	}
	if limit := streamer.Settings.Bet.MaxPoints; limit != nil && *limit < twitchMinimumStake {
		return fmt.Sprintf("max_points %d below Twitch minimum %d", *limit, twitchMinimumStake)
	}
	return fmt.Sprintf("calculated stake %d below Twitch minimum", stake)
}
