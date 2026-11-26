package prediction

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestEvent_Settled(t *testing.T) {
	event := func() *Event {
		return &Event{
			Streamer: &domain.Streamer{
				Username: "alice",
				Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}},
			},
			Title: "Who wins?",
			Outcomes: []Outcome{
				{ID: "yes", Title: "Yes"},
				{ID: "no", Title: "No"},
			},
			Decision: Decision{Choice: 0, OutcomeID: "yes", Amount: 480},
			Basis:    Basis{Voters: 400, TotalVoters: 600, Pool: 5_000, TotalPool: 10_000, Chance: 0.628},
		}
	}

	t.Run("a bet that won", func(t *testing.T) {
		got := event().Settled("WIN", 470)

		want := SettledBet{
			Streamer: "alice", Title: "Who wins?", Strategy: "KELLY", Outcome: "Yes",
			Stake: 480, Voters: 400, TotalVoters: 600, Pool: 5_000, TotalPool: 10_000, Chance: 0.628,
			Result: "WIN", Gained: 470,
		}
		if got != want {
			t.Errorf("Settled = %+v, want %+v", got, want)
		}
	})

	t.Run("a bet that lost", func(t *testing.T) {
		got := event().Settled("LOSE", -480)

		if got.Result != "LOSE" || got.Gained != -480 {
			t.Errorf("Settled = %+v, want a loss of 480", got)
		}
	})

	t.Run("a bet on an outcome that is no longer listed", func(t *testing.T) {
		e := event()
		e.Outcomes = nil

		if got := e.Settled("REFUND", 0); got.Outcome != "yes" {
			t.Errorf("Outcome = %q, want the outcome's id, %q", got.Outcome, "yes")
		}
	})

	t.Run("an event with no streamer", func(t *testing.T) {
		e := event()
		e.Streamer = nil

		got := e.Settled("WIN", 470)

		if got.Streamer != "" || got.Strategy != "" {
			t.Errorf("Settled = %+v, want no streamer and no strategy", got)
		}
	})
}
