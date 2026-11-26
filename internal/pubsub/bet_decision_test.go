package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func pointer[T any](value T) *T {
	return &value
}

// crowdedFilter only lets a bet through when over a million users have voted.
func crowdedFilter() *domain.FilterCondition {
	return &domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGT, Value: pointer(1_000_000.0)}
}

func TestDecideBet(t *testing.T) {
	mostVoted := domain.StrategyMostVoted

	for _, row := range []struct {
		name     string
		status   string
		points   int
		bet      domain.BetSettings
		want     prediction.Decision
		wantSkip *skippedBet
	}{
		{
			"an open event and 1,000 points",
			"ACTIVE", 1000, domain.BetSettings{Strategy: mostVoted},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 50},
			nil,
		},
		{
			"a locked event",
			"LOCKED", 1000, domain.BetSettings{Strategy: mostVoted},
			prediction.Decision{},
			&skippedBet{events.EventBetGeneral, "event status is LOCKED", "event status is LOCKED"},
		},
		{
			"a balance one under minimum_points",
			"ACTIVE", 4999, domain.BetSettings{Strategy: mostVoted, MinimumPoints: pointer(5000)},
			prediction.Decision{},
			&skippedBet{events.EventBetFilters, "balance below minimum_points", "balance 4999 <= minimum_points 5000"},
		},
		{
			"a balance equal to minimum_points",
			"ACTIVE", 5000, domain.BetSettings{Strategy: mostVoted, MinimumPoints: pointer(5000)},
			prediction.Decision{},
			&skippedBet{events.EventBetFilters, "balance below minimum_points", "balance 5000 <= minimum_points 5000"},
		},
		{
			"a balance one over minimum_points",
			"ACTIVE", 5001, domain.BetSettings{Strategy: mostVoted, MinimumPoints: pointer(5000)},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 250},
			nil,
		},
		{
			"a filter the event doesn't meet",
			"ACTIVE", 1000, domain.BetSettings{Strategy: mostVoted, FilterCondition: crowdedFilter()},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 50},
			&skippedBet{
				events.EventBetFilters,
				"filter_condition TOTAL_USERS GT 1000000 not met (current 15)",
				"filter_condition TOTAL_USERS GT 1000000 not met (current 15)",
			},
		},
		{
			"a filter the event meets",
			"ACTIVE", 1000, domain.BetSettings{
				Strategy:        mostVoted,
				FilterCondition: &domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGT, Value: pointer(14.0)},
			},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 50},
			nil,
		},
		{
			"a stake of 5 raised to Twitch's minimum",
			"ACTIVE", 100, domain.BetSettings{Strategy: mostVoted},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 10},
			nil,
		},
		{
			"a balance of 10",
			"ACTIVE", 10, domain.BetSettings{Strategy: mostVoted},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 10},
			nil,
		},
		{
			"a balance of 9",
			"ACTIVE", 9, domain.BetSettings{Strategy: mostVoted},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 0},
			&skippedBet{events.EventBetFilters, "below Twitch minimum", "balance 9 below Twitch minimum 10"},
		},
		{
			"max_points of 10",
			"ACTIVE", 1000, domain.BetSettings{Strategy: mostVoted, MaxPoints: pointer(10)},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 10},
			nil,
		},
		{
			"max_points of 9",
			"ACTIVE", 1000, domain.BetSettings{Strategy: mostVoted, MaxPoints: pointer(9)},
			prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 9},
			&skippedBet{events.EventBetFilters, "below Twitch minimum", "max_points 9 below Twitch minimum 10"},
		},
	} {
		t.Run(row.name, func(t *testing.T) {
			event := &prediction.Event{
				Streamer: &domain.Streamer{Username: "tester", ChannelPoints: row.points, Settings: domain.StreamerSettings{Bet: row.bet}},
				EventID:  "ev-1",
				Status:   row.status,
				Outcomes: []prediction.Outcome{
					{ID: "yes", Title: "Yes", TotalUsers: 10, TotalPoints: 100},
					{ID: "no", Title: "No", TotalUsers: 5, TotalPoints: 50},
				},
			}

			got, skip := decideBet(event)

			if got != row.want || !reflect.DeepEqual(skip, row.wantSkip) {
				t.Errorf("decideBet = %+v, %+v, want %+v, %+v", got, skip, row.want, row.wantSkip)
			}
			if event.Decision != row.want {
				t.Errorf("decision kept on the event = %+v, want %+v", event.Decision, row.want)
			}
		})
	}

	t.Run("an event with no outcomes", func(t *testing.T) {
		event := &prediction.Event{
			Streamer: &domain.Streamer{Username: "tester", ChannelPoints: 1000, Settings: domain.StreamerSettings{Bet: domain.BetSettings{Strategy: mostVoted}}},
			EventID:  "ev-1",
			Status:   "ACTIVE",
		}

		got, skip := decideBet(event)

		want, wantSkip := prediction.Decision{}, &skippedBet{events.EventBetGeneral, "no outcome selected", "no outcome selected"}
		if got != want || !reflect.DeepEqual(skip, wantSkip) {
			t.Errorf("decideBet = %+v, %+v, want %+v, %+v", got, skip, want, wantSkip)
		}
	})
}

func TestFilterSkip(t *testing.T) {
	for _, row := range []struct {
		name     string
		reason   string
		compared float64
		want     skippedBet
	}{
		{
			"a filter that gives its reason",
			"filter_condition ODDS LT 2 not met (current 2.5)", 2.5,
			skippedBet{events.EventBetFilters, "filter_condition ODDS LT 2 not met (current 2.5)", "filter_condition ODDS LT 2 not met (current 2.5)"},
		},
		{
			"a filter that gives no reason",
			"", 2.5,
			skippedBet{events.EventBetFilters, "filter_condition not satisfied", "filter_condition not satisfied (current 2.5)"},
		},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := *filterSkip(row.reason, row.compared); got != row.want {
				t.Errorf("filterSkip(%q, %v) = %+v, want %+v", row.reason, row.compared, got, row.want)
			}
		})
	}
}

func TestStakeTooSmall(t *testing.T) {
	for _, row := range []struct {
		name      string
		points    int
		maxPoints *int
		stake     int
		want      string
	}{
		{"a balance of 9", 9, nil, 0, "balance 9 below Twitch minimum 10"},
		{"an empty balance", 0, nil, 0, "balance 0 below Twitch minimum 10"},
		{"a balance of 9 and max_points of 5", 9, pointer(5), 0, "balance 9 below Twitch minimum 10"},
		{"a balance of 10 and max_points of 9", 10, pointer(9), 9, "max_points 9 below Twitch minimum 10"},
		{"a balance of 10 and max_points of 10", 10, pointer(10), 3, "calculated stake 3 below Twitch minimum"},
		{"a balance of 10 and no max_points", 10, nil, 3, "calculated stake 3 below Twitch minimum"},
	} {
		t.Run(row.name, func(t *testing.T) {
			streamer := &domain.Streamer{
				ChannelPoints: row.points,
				Settings:      domain.StreamerSettings{Bet: domain.BetSettings{MaxPoints: row.maxPoints}},
			}
			if got := stakeTooSmall(streamer, row.stake); got != row.want {
				t.Errorf("stakeTooSmall(%d points, stake %d) = %q, want %q", row.points, row.stake, got, row.want)
			}
		})
	}
}

func TestBelowMinimumPoints(t *testing.T) {
	for _, row := range []struct {
		name    string
		points  int
		minimum *int
		want    bool
	}{
		{"no minimum set", 0, nil, false},
		{"a balance under the minimum", 499, pointer(500), true},
		{"a balance at the minimum", 500, pointer(500), true},
		{"a balance one over the minimum", 501, pointer(500), false},
	} {
		t.Run(row.name, func(t *testing.T) {
			streamer := &domain.Streamer{ChannelPoints: row.points}
			streamer.Settings.Bet.MinimumPoints = row.minimum
			if got := belowMinimumPoints(streamer); got != row.want {
				t.Errorf("belowMinimumPoints(balance %d) = %t, want %t", row.points, got, row.want)
			}
		})
	}
}

func TestStakeTooSmall_kelly(t *testing.T) {
	kelly := domain.StreamerSettings{Bet: domain.BetSettings{Strategy: domain.StrategyKelly}}

	t.Run("a stake held under the minimum by the limits", func(t *testing.T) {
		streamer := &domain.Streamer{ChannelPoints: 100, Settings: kelly}

		want := "calculated stake 5 below Twitch minimum"
		if got := stakeTooSmall(streamer, 5); got != want {
			t.Errorf("stakeTooSmall = %q, want %q", got, want)
		}
	})

	t.Run("a balance under the minimum", func(t *testing.T) {
		streamer := &domain.Streamer{ChannelPoints: 9, Settings: kelly}

		want := "balance 9 below Twitch minimum 10"
		if got := stakeTooSmall(streamer, 0); got != want {
			t.Errorf("stakeTooSmall = %q, want %q", got, want)
		}
	})
}

func TestNothingToWin(t *testing.T) {
	tests := []struct {
		name     string
		outcomes []prediction.Outcome
		choice   int
		want     bool
	}{
		{"every point on the chosen outcome", []prediction.Outcome{{TotalPoints: 2_220}, {TotalPoints: 0}}, 0, true},
		{"a few points on the other outcome", []prediction.Outcome{{TotalPoints: 2_220}, {TotalPoints: 10}}, 0, false},
		{"every point on the other outcome", []prediction.Outcome{{TotalPoints: 0}, {TotalPoints: 2_220}}, 0, false},
		{"three outcomes, every point on the chosen one", []prediction.Outcome{{TotalPoints: 0}, {TotalPoints: 500}, {TotalPoints: 0}}, 1, true},
		{"an empty pool", []prediction.Outcome{{TotalPoints: 0}, {TotalPoints: 0}}, 0, false},
		{"a choice outside the outcomes", []prediction.Outcome{{TotalPoints: 500}}, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &prediction.Event{Outcomes: tt.outcomes}
			if got := nothingToWin(event, prediction.Decision{Choice: tt.choice}); got != tt.want {
				t.Errorf("nothingToWin() = %t, want %t", got, tt.want)
			}
		})
	}
}
