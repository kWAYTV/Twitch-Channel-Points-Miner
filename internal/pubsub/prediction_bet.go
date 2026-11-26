package pubsub

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// placePrediction bets on a tracked prediction, or stops tracking it when the
// bet is skipped or fails.
func (p *Client) placePrediction(eventID string) {
	event := p.trackedPrediction(eventID)
	if event == nil {
		return
	}
	decision, skipped := decideBet(event)
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
