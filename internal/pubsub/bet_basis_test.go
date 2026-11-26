package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func TestChanceNote(t *testing.T) {
	tests := []struct {
		name  string
		basis prediction.Basis
		want  string
	}{
		{"an estimate over the pool's share", prediction.Basis{Pool: 5_000, TotalPool: 10_000, Chance: 0.628}, ", Chance: 63% vs pool 50%"},
		{"an estimate under the pool's share", prediction.Basis{Pool: 100, TotalPool: 150, Chance: 0.4172}, ", Chance: 42% vs pool 67%"},
		{"an estimate of exactly a half", prediction.Basis{Pool: 250, TotalPool: 1_000, Chance: 0.5}, ", Chance: 50% vs pool 25%"},
		{"a strategy that makes no estimate", prediction.Basis{Pool: 5_000, TotalPool: 10_000}, ""},
		{"nothing known", prediction.Basis{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chanceNote(tt.basis); got != tt.want {
				t.Errorf("chanceNote(%+v) = %q, want %q", tt.basis, got, tt.want)
			}
		})
	}
}

func kellyBettor() *domain.Streamer {
	streamer := bettor()
	streamer.Settings.Bet.Strategy = domain.StrategyKelly
	return streamer
}

func TestDecideBet_kelly(t *testing.T) {
	t.Run("nothing underpriced", func(t *testing.T) {
		// Ten of fifteen voters can be trusted for 42%; Yes holds 67% of the pool.
		event := &prediction.Event{
			Streamer: kellyBettor(),
			EventID:  "ev-1",
			Status:   "ACTIVE",
			Outcomes: []prediction.Outcome{
				{ID: "yes", Title: "Yes", TotalUsers: 10, TotalPoints: 100},
				{ID: "no", Title: "No", TotalUsers: 5, TotalPoints: 50},
			},
		}

		_, skip := decideBet(event)

		want := &skippedBet{
			events.EventBetFilters,
			"no outcome is underpriced enough to be worth a bet",
			"no outcome is underpriced enough to be worth a bet (best: Yes, chance 42% vs pool 67%)",
		}
		if !reflect.DeepEqual(skip, want) {
			t.Errorf("skip = %+v, want %+v", skip, want)
		}
	})
}

func TestClient_placePrediction_kellyLine(t *testing.T) {
	// Four hundred of six hundred voters on Yes, the pool split evenly: Yes
	// can be trusted for 63% against the pool's 50%.
	const created = `{"type":"event-created","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":400,"total_points":5000},{"id":"no","title":"No","color":"PINK","total_users":200,"total_points":5000}]}}}`

	t.Run("a bet KELLY places", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{kellyBettor()}})
		h.receive(t, "predictions-channel-v1.123", created)

		h.timers[0].fire()

		want := []string{
			betStartLine,
			":four_leaf_clover: BET_GENERAL: Place 50 points on: Yes (BLUE), Points: 5k, Users: 400 (66.67%), Odds: 2 (50%), Chance: 63% vs pool 50% for Tester",
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a bet KELLY places, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{kellyBettor()}, anonymise: true})
		h.receive(t, "predictions-channel-v1.123", created)

		h.timers[0].fire()

		if got := h.logger.Lines[len(h.logger.Lines)-1]; got != ":four_leaf_clover: BET_GENERAL: Place bet on: Yes (BLUE), Points: 5k, Users: 400 (66.67%), Odds: 2 (50%), Chance: 63% vs pool 50% for Streamer1" {
			t.Errorf("last line = %q", got)
		}
	})

	t.Run("a bet another strategy places", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})

		h.timers[0].fire()

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})
}
