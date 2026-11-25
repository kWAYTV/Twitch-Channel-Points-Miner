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

// placePrediction bets on a tracked prediction, or stops tracking it when the
// bet is skipped or fails.
func (p *Client) placePrediction(eventID string) {
	event := p.trackedPrediction(eventID)
	if event == nil {
		return
	}
	decision, skipped := p.decideBet(event)
	if skipped != nil {
		p.logSkippedBet(event.Streamer, *skipped)
		p.stopTracking(event, "SKIPPED")
		return
	}
	if err := p.account.MakePrediction(event); err != nil {
		p.logger.ErrorEventf(events.EventBetFailed, "prediction %s: %v", p.streamerName(event.Streamer), err)
		p.stopTracking(event, "ERROR")
		return
	}
	p.deductStake(event.Streamer, decision.Amount)
	event.BetPlaced = true
	// Twitch does not always confirm a bet, and the result must be logged either way.
	event.BetConfirmed = true
	p.logPlacedBet(event, decision)
}

func (p *Client) trackedPrediction(eventID string) *prediction.Event {
	p.predMu.Lock()
	defer p.predMu.Unlock()
	event, ok := p.predictions[eventID]
	if !ok || event == nil || event.Streamer == nil {
		return nil
	}
	return event
}

func (p *Client) stopTracking(event *prediction.Event, result string) {
	event.ResultType = result
	p.predMu.Lock()
	delete(p.predictions, event.EventID)
	p.predMu.Unlock()
}

// decideBet chooses the outcome and the stake, or returns why no bet is placed.
func (p *Client) decideBet(event *prediction.Event) (prediction.Decision, *skippedBet) {
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

// logSkippedBet logs the skip. The two "general" reasons always use the display
// name; the filter reasons show the real name and figures unless logs are anonymised.
func (p *Client) logSkippedBet(streamer *domain.Streamer, skipped skippedBet) {
	if p.anonymizeLogs() || skipped.event == events.EventBetGeneral {
		p.logger.Eventf(skipped.event, "Skip bet for %s: %s", p.streamerName(streamer), skipped.public)
		return
	}
	p.logger.Eventf(skipped.event, "Skip bet for %s: %s", streamer.Username, skipped.detail)
}

// deductStake takes the stake off the balance right away when the streamer's
// settings ask for it. The payout or refund arrives later as a points event.
func (p *Client) deductStake(streamer *domain.Streamer, stake int) {
	deduct := true
	if setting := streamer.Settings.Bet.DeductStakeOnPlace; setting != nil {
		deduct = *setting
	}
	if !deduct || stake <= 0 {
		return
	}
	if p.onGain != nil {
		p.onGain(streamer, -stake, "PREDICTION", 0)
		return
	}
	streamer.ChannelPoints -= stake
	if streamer.ChannelPoints < 0 {
		streamer.ChannelPoints = 0
	}
}

func (p *Client) logPlacedBet(event *prediction.Event, decision prediction.Decision) {
	outcome := event.DecisionOutcomeString()
	if outcome == "" {
		outcome = decision.OutcomeID
	}
	if p.anonymizeLogs() {
		p.logger.EmojiEventf(":four_leaf_clover:", events.EventBetGeneral, "Place bet on: %s for %s", outcome, p.streamerName(event.Streamer))
		return
	}
	p.logger.EmojiEventf(":four_leaf_clover:", events.EventBetGeneral, "Place %s points on: %s for %s", prediction.FormatNumber(decision.Amount), outcome, event.Streamer.Username)
}
