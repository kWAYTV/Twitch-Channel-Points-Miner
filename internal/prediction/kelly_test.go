package prediction

import (
	"math"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// near reports whether a stake found by search is within a point of want.
func near(got, want float64) bool {
	return math.Abs(got-want) <= 1
}

func TestKellyStake(t *testing.T) {
	tests := []struct {
		name        string
		balance     float64
		p           float64
		outcomePool float64
		totalPool   float64
		want        float64
	}{
		// With a pool far larger than the balance the stake hardly moves the
		// odds, and the classic formula applies: (p*odds - 1) / (odds - 1) of
		// the balance. Here odds are 2 and p is 0.6, so a fifth of the balance.
		{"an even pool far larger than the balance, a chance of 0.6", 1_000, 0.6, 50_000_000, 100_000_000, 200},
		{"the same with a chance of 0.75", 1_000, 0.75, 50_000_000, 100_000_000, 500},
		// Odds of 4 and a chance of 0.4: (0.4*4 - 1) / 3 = a fifth.
		{"a longshot at odds of 4 with a chance of 0.4", 1_000, 0.4, 25_000_000, 100_000_000, 200},
		{"a chance equal to the share of the pool", 1_000, 0.5, 50_000_000, 100_000_000, 0},
		{"a chance under the share of the pool", 1_000, 0.4, 50_000_000, 100_000_000, 0},
		{"no chance", 1_000, 0, 50_000, 100_000, 0},
		{"no balance", 0, 0.6, 50_000, 100_000, 0},
		{"nothing bet on the other outcomes", 1_000, 0.9, 100_000, 100_000, 0},
		{"a pool that does not add up", 1_000, 0.9, 200_000, 100_000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kellyStake(tt.balance, tt.p, tt.outcomePool, tt.totalPool); !near(got, tt.want) {
				t.Errorf("kellyStake(%v, %v, %v, %v) = %.2f, want %.0f", tt.balance, tt.p, tt.outcomePool, tt.totalPool, got, tt.want)
			}
		})
	}

	t.Run("a small pool, where the stake eats its own odds", func(t *testing.T) {
		// The same edge as the first example, in a pool the size of the
		// balance: the more is bet, the less each point pays, so the stake
		// has to be smaller than a fifth of the balance.
		got := kellyStake(1_000, 0.6, 500, 1_000)
		if got <= 0 || got >= 200 {
			t.Errorf("kellyStake = %.2f, want more than 0 and less than the 200 of a large pool", got)
		}
	})

	t.Run("a certain win is not bet in full", func(t *testing.T) {
		// Even at a chance of 1 the stake only wins what the others bet, and
		// betting the whole balance is never the peak of the growth.
		got := kellyStake(1_000, 1, 500, 1_000)
		if got <= 0 || got >= 1_000 {
			t.Errorf("kellyStake = %.2f, want more than 0 and less than the balance", got)
		}
	})
}

func TestKellyChoice(t *testing.T) {
	t.Run("most viewers on the side with less of the pool", func(t *testing.T) {
		// Yes has 60% of the viewers and half of a very large pool.
		outcomes := []Outcome{
			{ID: "yes", TotalUsers: 60, TotalPoints: 50_000_000},
			{ID: "no", TotalUsers: 40, TotalPoints: 50_000_000},
		}

		index, stake := kellyChoice(outcomes, 1_000)

		if index != 0 || !near(stake, 200) {
			t.Errorf("kellyChoice = %d, %.2f, want 0, 200", index, stake)
		}
	})

	t.Run("the second outcome underpriced", func(t *testing.T) {
		outcomes := []Outcome{
			{ID: "yes", TotalUsers: 40, TotalPoints: 50_000_000},
			{ID: "no", TotalUsers: 60, TotalPoints: 50_000_000},
		}

		index, stake := kellyChoice(outcomes, 1_000)

		if index != 1 || !near(stake, 200) {
			t.Errorf("kellyChoice = %d, %.2f, want 1, 200", index, stake)
		}
	})

	t.Run("viewers and pool split the same way", func(t *testing.T) {
		outcomes := []Outcome{
			{ID: "yes", TotalUsers: 70, TotalPoints: 70_000},
			{ID: "no", TotalUsers: 30, TotalPoints: 30_000},
		}

		index, stake := kellyChoice(outcomes, 1_000)

		if index != -1 || stake != 0 {
			t.Errorf("kellyChoice = %d, %.2f, want -1, 0: nothing is underpriced", index, stake)
		}
	})

	t.Run("three outcomes, the middle one underpriced", func(t *testing.T) {
		outcomes := []Outcome{
			{ID: "a", TotalUsers: 30, TotalPoints: 40_000_000},
			{ID: "b", TotalUsers: 50, TotalPoints: 30_000_000},
			{ID: "c", TotalUsers: 20, TotalPoints: 30_000_000},
		}

		index, _ := kellyChoice(outcomes, 1_000)

		if index != 1 {
			t.Errorf("kellyChoice picked %d, want 1", index)
		}
	})

	t.Run("nobody has voted", func(t *testing.T) {
		outcomes := []Outcome{{ID: "yes"}, {ID: "no"}}

		if index, stake := kellyChoice(outcomes, 1_000); index != -1 || stake != 0 {
			t.Errorf("kellyChoice = %d, %.2f, want -1, 0", index, stake)
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		if index, stake := kellyChoice(nil, 1_000); index != -1 || stake != 0 {
			t.Errorf("kellyChoice = %d, %.2f, want -1, 0", index, stake)
		}
	})
}

func TestKellyAmount(t *testing.T) {
	tests := []struct {
		name     string
		full     float64
		balance  int
		settings domain.BetSettings
		want     int
	}{
		{"a quarter of the full stake", 200, 10_000, domain.BetSettings{}, 50},
		{"a quarter that rounds", 190, 10_000, domain.BetSettings{}, 48},
		{"over percentage of the balance", 8_000, 10_000, domain.BetSettings{Percentage: intPtr(5)}, 500},
		{"over max_points", 8_000, 10_000, domain.BetSettings{Percentage: intPtr(50), MaxPoints: intPtr(300)}, 300},
		{"under Twitch's minimum, the limit allowing it", 20, 10_000, domain.BetSettings{}, 10},
		{"under Twitch's minimum, the limit under it too", 20, 100, domain.BetSettings{Percentage: intPtr(5)}, 5},
		{"nothing worth betting", 0, 10_000, domain.BetSettings{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kellyAmount(tt.full, tt.balance, tt.settings); got != tt.want {
				t.Errorf("kellyAmount(%v, %d, settings) = %d, want %d", tt.full, tt.balance, got, tt.want)
			}
		})
	}
}

func TestStakeLimit(t *testing.T) {
	tests := []struct {
		name     string
		balance  int
		settings domain.BetSettings
		want     int
	}{
		{"the default 5 percent", 10_000, domain.BetSettings{}, 500},
		{"a percentage", 10_000, domain.BetSettings{Percentage: intPtr(20)}, 2_000},
		{"max_points under the percentage", 10_000, domain.BetSettings{Percentage: intPtr(20), MaxPoints: intPtr(1_500)}, 1_500},
		{"max_points over the percentage", 10_000, domain.BetSettings{Percentage: intPtr(20), MaxPoints: intPtr(5_000)}, 2_000},
		{"a percentage over a hundred", 1_000, domain.BetSettings{Percentage: intPtr(150)}, 1_000},
		{"no balance", 0, domain.BetSettings{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stakeLimit(tt.balance, tt.settings); got != tt.want {
				t.Errorf("stakeLimit(%d, settings) = %d, want %d", tt.balance, got, tt.want)
			}
		})
	}
}

func TestKellyDecision(t *testing.T) {
	underpriced := []Outcome{
		{ID: "yes", TotalUsers: 60, TotalPoints: 50_000_000, TopPoints: 40},
		{ID: "no", TotalUsers: 40, TotalPoints: 50_000_000},
	}
	fair := []Outcome{
		{ID: "yes", TotalUsers: 30, TotalPoints: 30_000},
		{ID: "no", TotalUsers: 70, TotalPoints: 70_000},
	}

	t.Run("an underpriced outcome", func(t *testing.T) {
		// Full Kelly is a fifth of 10,000; a quarter of that is 500, which is
		// also the 5 percent limit.
		choice, amount := kellyDecision(underpriced, 10_000, domain.BetSettings{})

		if choice != 0 || amount != 500 {
			t.Errorf("kellyDecision = %d, %d, want 0, 500", choice, amount)
		}
	})

	t.Run("an underpriced outcome, max_points lower", func(t *testing.T) {
		choice, amount := kellyDecision(underpriced, 10_000, domain.BetSettings{MaxPoints: intPtr(120)})

		if choice != 0 || amount != 120 {
			t.Errorf("kellyDecision = %d, %d, want 0, 120", choice, amount)
		}
	})

	t.Run("an underpriced outcome, stealth mode on", func(t *testing.T) {
		choice, amount := kellyDecision(underpriced, 10_000, domain.BetSettings{StealthMode: boolPtr(true)})

		if choice != 0 || amount != 39 {
			t.Errorf("kellyDecision = %d, %d, want 0, 39: one under the top bet of 40", choice, amount)
		}
	})

	t.Run("nothing underpriced", func(t *testing.T) {
		choice, amount := kellyDecision(fair, 10_000, domain.BetSettings{})

		if choice != 1 || amount != 0 {
			t.Errorf("kellyDecision = %d, %d, want 1, 0: the most voted outcome, with no stake", choice, amount)
		}
	})
}

func TestEvent_Decide_kelly(t *testing.T) {
	outcomes := []Outcome{
		{ID: "yes", TotalUsers: 60, TotalPoints: 50_000_000},
		{ID: "no", TotalUsers: 40, TotalPoints: 50_000_000},
	}

	t.Run("the KELLY strategy", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}}}
		event := &Event{Streamer: streamer, Outcomes: outcomes}

		got := event.Decide(10_000)

		if want := (Decision{Choice: 0, OutcomeID: "yes", Amount: 500}); got != want {
			t.Errorf("Decide = %+v, want %+v", got, want)
		}
	})

	t.Run("no strategy set: the default, KELLY", func(t *testing.T) {
		event := &Event{Streamer: &domain.Streamer{}, Outcomes: outcomes}

		got := event.Decide(10_000)

		if want := (Decision{Choice: 0, OutcomeID: "yes", Amount: 500}); got != want {
			t.Errorf("Decide = %+v, want %+v", got, want)
		}
	})

	t.Run("the KELLY strategy with nothing underpriced", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}}}
		event := &Event{Streamer: streamer, Outcomes: []Outcome{
			{ID: "yes", TotalUsers: 30, TotalPoints: 30_000},
			{ID: "no", TotalUsers: 70, TotalPoints: 70_000},
		}}

		got := event.Decide(10_000)

		if want := (Decision{Choice: 1, OutcomeID: "no", Amount: 0}); got != want {
			t.Errorf("Decide = %+v, want %+v", got, want)
		}
	})
}
