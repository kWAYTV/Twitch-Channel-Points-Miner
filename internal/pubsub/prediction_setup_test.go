package pubsub

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// whoWinsCreated opens the prediction ev-1 at noon for two minutes: ten
// users with 100 points on Yes, five with 50 on No.
const whoWinsCreated = `{"type":"event-created","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`

const (
	betStartLine  = ":alarm_clock: BET_START: Place bet after 2m0s for tester"
	betPlacedLine = ":four_leaf_clover: BET_GENERAL: Place 50 points on: Yes (BLUE), Points: 100, Users: 10 (66.67%), Odds: 1.5 (66.67%) for tester"
	betWonLine    = ":bar_chart: BET_WIN: EventPrediction: tester - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorGreen + "WIN, Gained: +25" + ansi.ColorReset
)

// bettor follows channel 123 with 1,000 points and bets 5% of them on the
// outcome most users chose.
func bettor() *domain.Streamer {
	return &domain.Streamer{
		Username:      "tester",
		ChannelID:     "123",
		ChannelPoints: 1000,
		Settings: domain.StreamerSettings{
			MakePredictions: true,
			Bet:             domain.BetSettings{Strategy: domain.StrategyMostVoted},
		},
	}
}

// betDue is a client that has seen whoWinsCreated and not yet bet on it.
func betDue(t *testing.T, w world) *harness {
	t.Helper()
	h := setup(w)
	if err := h.receive(t, "predictions-channel-v1.123", whoWinsCreated); err != nil {
		t.Fatalf("opening the prediction: %v", err)
	}
	if len(h.timers) != 1 {
		t.Fatalf("%d bets due after opening the prediction, want 1", len(h.timers))
	}
	return h
}

// betPlaced is a client that has bet 50 points on Yes in whoWinsCreated.
func betPlaced(t *testing.T, w world) *harness {
	t.Helper()
	h := betDue(t, w)
	h.timers[0].fire()
	if len(h.account.Bets) != 1 {
		t.Fatalf("%d bets after the bet came due, want 1", len(h.account.Bets))
	}
	return h
}
