package prediction

import (
	"math"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestBasis_PoolShare(t *testing.T) {
	tests := []struct {
		name  string
		basis Basis
		want  float64
	}{
		{"half the pool", Basis{Pool: 500, TotalPool: 1_000}, 0.5},
		{"a quarter of the pool", Basis{Pool: 250, TotalPool: 1_000}, 0.25},
		{"the whole pool", Basis{Pool: 1_000, TotalPool: 1_000}, 1},
		{"none of the pool", Basis{Pool: 0, TotalPool: 1_000}, 0},
		{"an empty pool", Basis{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.basis.PoolShare(); got != tt.want {
				t.Errorf("PoolShare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_Decide_basis(t *testing.T) {
	outcomes := func() []Outcome {
		return []Outcome{
			{ID: "yes", TotalUsers: 40, TotalPoints: 5_000},
			{ID: "no", TotalUsers: 20, TotalPoints: 5_000},
		}
	}

	t.Run("a strategy that makes no estimate", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyMostVoted}}}
		event := &Event{Streamer: streamer, Outcomes: outcomes()}

		event.Decide(10_000)

		if want := (Basis{Voters: 40, TotalVoters: 60, Pool: 5_000, TotalPool: 10_000}); event.Basis != want {
			t.Errorf("Basis = %+v, want %+v", event.Basis, want)
		}
	})

	t.Run("KELLY, with its estimate", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}}}
		event := &Event{Streamer: streamer, Outcomes: outcomes()}

		event.Decide(10_000)

		// Forty of sixty voters can be trusted for 54.06%.
		if math.Abs(event.Basis.Chance-0.5405662987) > 1e-9 {
			t.Errorf("Chance = %.10f, want 0.5405662987", event.Basis.Chance)
		}
		event.Basis.Chance = 0
		if want := (Basis{Voters: 40, TotalVoters: 60, Pool: 5_000, TotalPool: 10_000}); event.Basis != want {
			t.Errorf("Basis without its chance = %+v, want %+v", event.Basis, want)
		}
	})

	t.Run("KELLY with nothing underpriced: the most voted outcome", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}}}
		event := &Event{Streamer: streamer, Outcomes: []Outcome{
			{ID: "yes", TotalUsers: 4, TotalPoints: 5_000},
			{ID: "no", TotalUsers: 2, TotalPoints: 5_000},
		}}

		event.Decide(10_000)

		// Four of six voters can be trusted for 30%.
		if math.Abs(event.Basis.Chance-0.2999883213) > 1e-9 {
			t.Errorf("Chance = %.10f, want 0.2999883213", event.Basis.Chance)
		}
		event.Basis.Chance = 0
		if want := (Basis{Voters: 4, TotalVoters: 6, Pool: 5_000, TotalPool: 10_000}); event.Basis != want {
			t.Errorf("Basis without its chance = %+v, want %+v", event.Basis, want)
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		event := &Event{Streamer: &domain.Streamer{}}

		event.Decide(10_000)

		if event.Basis != (Basis{}) {
			t.Errorf("Basis = %+v, want none", event.Basis)
		}
	})
}
