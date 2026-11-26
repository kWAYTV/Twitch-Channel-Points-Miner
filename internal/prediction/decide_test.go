package prediction

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestEvent_Decide(t *testing.T) {
	t.Run("the most voted outcome", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyMostVoted}}}
		event := &Event{
			Streamer: streamer,
			Outcomes: []Outcome{{ID: "a", TotalUsers: 10}, {ID: "b", TotalUsers: 30}},
		}

		got := event.Decide(1_000)

		want := Decision{Choice: 1, OutcomeID: "b", Amount: 50}
		if got != want {
			t.Errorf("Decide(1000) = %+v, want %+v", got, want)
		}
		wantEvent := &Event{
			Streamer: streamer,
			Outcomes: []Outcome{{ID: "a", TotalUsers: 10}, {ID: "b", TotalUsers: 30}},
			Decision: Decision{Choice: 1, OutcomeID: "b", Amount: 50},
		}
		if !reflect.DeepEqual(event, wantEvent) {
			t.Errorf("event = %+v, want %+v", event, wantEvent)
		}
	})

	t.Run("stealth mode under the chosen outcome's top bet", func(t *testing.T) {
		streamer := &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{
			Strategy:    domain.StrategyMostVoted,
			StealthMode: boolPtr(true),
		}}}
		event := &Event{
			Streamer: streamer,
			Outcomes: []Outcome{{ID: "a", TotalUsers: 10, TopPoints: 20}, {ID: "b", TotalUsers: 30, TopPoints: 40}},
		}

		got := event.Decide(1_000)

		want := Decision{Choice: 1, OutcomeID: "b", Amount: 39}
		if got != want {
			t.Errorf("Decide(1000) = %+v, want %+v", got, want)
		}
		if event.Decision != want {
			t.Errorf("event.Decision = %+v, want %+v", event.Decision, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		event := &Event{Outcomes: []Outcome{{ID: "a", TotalUsers: 10}}}

		got := event.Decide(1_000)

		if got != (Decision{}) {
			t.Errorf("Decide(1000) = %+v, want %+v", got, Decision{})
		}
		if event.Decision != (Decision{}) {
			t.Errorf("event.Decision = %+v, want %+v", event.Decision, Decision{})
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		event := &Event{Streamer: &domain.Streamer{}}

		got := event.Decide(1_000)

		if got != (Decision{}) {
			t.Errorf("Decide(1000) = %+v, want %+v", got, Decision{})
		}
		if event.Decision != (Decision{}) {
			t.Errorf("event.Decision = %+v, want %+v", event.Decision, Decision{})
		}
	})
}

func TestStake(t *testing.T) {
	for _, row := range []struct {
		name      string
		balance   int
		settings  domain.BetSettings
		topPoints int
		want      int
	}{
		{"no settings: five percent", 1_000, domain.BetSettings{}, 0, 50},
		{"a percentage of ten", 1_000, domain.BetSettings{Percentage: intPtr(10)}, 0, 100},
		{"a percentage of zero", 1_000, domain.BetSettings{Percentage: intPtr(0)}, 0, 10},
		{"a percentage above a hundred", 1_000, domain.BetSettings{Percentage: intPtr(150)}, 0, 1_000},

		{"one point under the maximum", 1_000, domain.BetSettings{Percentage: intPtr(10), MaxPoints: intPtr(101)}, 0, 100},
		{"exactly the maximum", 1_000, domain.BetSettings{Percentage: intPtr(10), MaxPoints: intPtr(100)}, 0, 100},
		{"one point over the maximum", 1_000, domain.BetSettings{Percentage: intPtr(10), MaxPoints: intPtr(99)}, 0, 99},

		{"stealth: one point under the top bet", 1_000, domain.BetSettings{StealthMode: boolPtr(true)}, 51, 50},
		{"stealth: exactly the top bet", 1_000, domain.BetSettings{StealthMode: boolPtr(true)}, 50, 49},
		{"stealth: over the top bet", 1_000, domain.BetSettings{StealthMode: boolPtr(true)}, 40, 39},
		{"stealth: nobody has bet", 1_000, domain.BetSettings{StealthMode: boolPtr(true)}, 0, 50},
		{"stealth: a top bet under the minimum bet", 1_000, domain.BetSettings{StealthMode: boolPtr(true)}, 5, 10},
		{"stealth: a top bet of one point", 5, domain.BetSettings{StealthMode: boolPtr(true), Percentage: intPtr(100)}, 1, 1},
		{"stealth off: over the top bet", 1_000, domain.BetSettings{StealthMode: boolPtr(false)}, 40, 50},
		{"stealth unset: over the top bet", 1_000, domain.BetSettings{}, 40, 50},

		{"five percent is exactly the minimum bet", 200, domain.BetSettings{}, 0, 10},
		{"five percent is one under the minimum bet", 180, domain.BetSettings{}, 0, 10},
		{"a balance of exactly the minimum bet", 10, domain.BetSettings{}, 0, 10},
		{"a balance one under the minimum bet", 9, domain.BetSettings{}, 0, 0},
		{"a maximum of exactly the minimum bet", 100, domain.BetSettings{MaxPoints: intPtr(10)}, 0, 10},
		{"a maximum one under the minimum bet", 100, domain.BetSettings{MaxPoints: intPtr(9)}, 0, 9},
		{"a maximum under the minimum bet and over the balance (today: the maximum)", 5, domain.BetSettings{MaxPoints: intPtr(8)}, 0, 8},

		{"a balance of zero", 0, domain.BetSettings{}, 0, 0},
		{"a negative balance", -100, domain.BetSettings{}, 0, -100},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := stake(row.balance, row.settings, row.topPoints); got != row.want {
				t.Errorf("stake(%d, settings, %d) = %d, want %d", row.balance, row.topPoints, got, row.want)
			}
		})
	}

	t.Run("a maximum under the minimum bet and over the balance (bug: should not bet more than the balance)", func(t *testing.T) {
		t.Skip("bug: a maximum under 10 replaces the stake even when the balance is smaller")
		if got := stake(5, domain.BetSettings{MaxPoints: intPtr(8)}, 0); got > 5 {
			t.Errorf("stake(5, settings, 0) = %d, want at most 5", got)
		}
	})
}

func TestSelectOutcome(t *testing.T) {
	four := []Outcome{
		{TotalUsers: 50, PercentageUsers: 50, Odds: 1.2, OddsPercentage: 30, TopPoints: 10},
		{TotalUsers: 20, PercentageUsers: 20, Odds: 9, OddsPercentage: 20, TopPoints: 20},
		{TotalUsers: 20, PercentageUsers: 20, Odds: 2, OddsPercentage: 80, TopPoints: 30},
		{TotalUsers: 10, PercentageUsers: 10, Odds: 3, OddsPercentage: 40, TopPoints: 99},
	}
	eight := []Outcome{{Odds: 8}, {Odds: 7}, {Odds: 6}, {Odds: 5}, {Odds: 4}, {Odds: 3}, {Odds: 2}, {Odds: 1}}

	for _, row := range []struct {
		name     string
		outcomes []Outcome
		settings domain.BetSettings
		want     int
	}{
		{"most voted", four, domain.BetSettings{Strategy: domain.StrategyMostVoted}, 0},
		{"high odds", four, domain.BetSettings{Strategy: domain.StrategyHighOdds}, 1},
		{"percentage", four, domain.BetSettings{Strategy: domain.StrategyPercentage}, 2},
		{"smart money", four, domain.BetSettings{Strategy: domain.StrategySmartMoney}, 3},
		{"an unknown strategy: high odds", four, domain.BetSettings{Strategy: "COIN_FLIP"}, 1},

		{"number 1 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber1}, 0},
		{"number 2 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber2}, 1},
		{"number 3 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber3}, 2},
		{"number 4 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber4}, 3},
		{"number 5 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber5}, 4},
		{"number 6 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber6}, 5},
		{"number 7 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber7}, 6},
		{"number 8 of eight", eight, domain.BetSettings{Strategy: domain.StrategyNumber8}, 7},

		{"number 1 of four", four, domain.BetSettings{Strategy: domain.StrategyNumber1}, 0},
		{"number 2 of one: high odds", eight[:1], domain.BetSettings{Strategy: domain.StrategyNumber2}, 0},
		{"number 3 of two: high odds", eight[:2], domain.BetSettings{Strategy: domain.StrategyNumber3}, 0},
		{"number 4 of three: high odds", eight[:3], domain.BetSettings{Strategy: domain.StrategyNumber4}, 0},
		{"number 5 of four: high odds", eight[:4], domain.BetSettings{Strategy: domain.StrategyNumber5}, 0},
		{"number 6 of five: high odds", eight[:5], domain.BetSettings{Strategy: domain.StrategyNumber6}, 0},
		{"number 7 of six: high odds", eight[:6], domain.BetSettings{Strategy: domain.StrategyNumber7}, 0},
		{"number 8 of seven: high odds", eight[:7], domain.BetSettings{Strategy: domain.StrategyNumber8}, 0},

		{"smart, a clear favourite: most voted", four, domain.BetSettings{Strategy: domain.StrategySmart}, 0},
		{"no strategy: smart", four, domain.BetSettings{}, 0},
		{"smart, a gap wider than the favourite's lead: high odds", four, domain.BetSettings{Strategy: domain.StrategySmart, PercentageGap: intPtr(40)}, 1},
		{
			"smart, a lead one under the gap: high odds",
			[]Outcome{{TotalUsers: 405, PercentageUsers: 40.5, Odds: 2.4}, {TotalUsers: 595, PercentageUsers: 59.5, Odds: 1.6}},
			domain.BetSettings{Strategy: domain.StrategySmart}, 0,
		},
		{
			"smart, a lead of exactly the gap: most voted",
			[]Outcome{{TotalUsers: 40, PercentageUsers: 40, Odds: 2.5}, {TotalUsers: 60, PercentageUsers: 60, Odds: 1.6}},
			domain.BetSettings{Strategy: domain.StrategySmart}, 1,
		},
		{
			"smart, a lead one over the gap: most voted",
			[]Outcome{{TotalUsers: 395, PercentageUsers: 39.5, Odds: 2.5}, {TotalUsers: 605, PercentageUsers: 60.5, Odds: 1.6}},
			domain.BetSettings{Strategy: domain.StrategySmart}, 1,
		},
		{
			"smart, the two leaders are not the first two: high odds",
			[]Outcome{{TotalUsers: 10, PercentageUsers: 10, Odds: 9}, {TotalUsers: 48, PercentageUsers: 48, Odds: 2}, {TotalUsers: 42, PercentageUsers: 42, Odds: 2.3}},
			domain.BetSettings{Strategy: domain.StrategySmart}, 0,
		},
		{"smart, one outcome", eight[:1], domain.BetSettings{Strategy: domain.StrategySmart}, 0},

		{"no outcomes", nil, domain.BetSettings{Strategy: domain.StrategyMostVoted}, -1},
		{"number 1 of none", nil, domain.BetSettings{Strategy: domain.StrategyNumber1}, -1},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := selectOutcome(row.outcomes, row.settings); got != row.want {
				t.Errorf("selectOutcome() = %d, want %d", got, row.want)
			}
		})
	}
}

func TestMaxIndex(t *testing.T) {
	odds := func(o Outcome) float64 { return o.Odds }

	for _, row := range []struct {
		name     string
		outcomes []Outcome
		want     int
	}{
		{"the largest is first", []Outcome{{Odds: 3}, {Odds: 2}, {Odds: 1}}, 0},
		{"the largest is last", []Outcome{{Odds: 1}, {Odds: 2}, {Odds: 3}}, 2},
		{"a tie: the first of them", []Outcome{{Odds: 1}, {Odds: 3}, {Odds: 3}}, 1},
		{"all negative", []Outcome{{Odds: -3}, {Odds: -1}}, 1},
		{"one outcome", []Outcome{{Odds: 1}}, 0},
		{"no outcomes", nil, -1},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := maxIndex(row.outcomes, odds); got != row.want {
				t.Errorf("maxIndex() = %d, want %d", got, row.want)
			}
		})
	}
}
