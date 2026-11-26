package prediction

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

type verdict struct {
	skip     bool
	compared float64
	reason   string
}

func shouldSkip(event *Event) verdict {
	skip, compared, reason := event.ShouldSkipByFilter()
	return verdict{skip, compared, reason}
}

func floatPtr(v float64) *float64 { return &v }

// twoOutcomes total 25 users and 300 points.
func twoOutcomes() []Outcome {
	return []Outcome{
		{TotalUsers: 10, TotalPoints: 100, TopPoints: 40, PercentageUsers: 40, Odds: 2.5, OddsPercentage: 40},
		{TotalUsers: 15, TotalPoints: 200, TopPoints: 90, PercentageUsers: 60, Odds: 1.5, OddsPercentage: 66.5},
	}
}

func filtered(filter *domain.FilterCondition, choice int) *Event {
	return &Event{
		Streamer: &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{FilterCondition: filter}}},
		Outcomes: twoOutcomes(),
		Decision: Decision{Choice: choice},
	}
}

func TestEvent_ShouldSkipByFilter(t *testing.T) {
	for _, row := range []struct {
		name   string
		filter *domain.FilterCondition
		choice int
		want   verdict
	}{
		{
			"a condition that is met",
			&domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGTE, Value: floatPtr(20)}, 1,
			verdict{false, 25, ""},
		},
		{
			"a condition that is not met",
			&domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGTE, Value: floatPtr(30)}, 1,
			verdict{true, 25, "filter_condition TOTAL_USERS GTE 30 not met (current 25)"},
		},
		{
			"a condition on the chosen outcome that is met",
			&domain.FilterCondition{By: domain.OutcomeDecisionUsers, Where: domain.ConditionLT, Value: floatPtr(12)}, 0,
			verdict{false, 10, ""},
		},
		{
			"a condition on the chosen outcome that is not met",
			&domain.FilterCondition{By: domain.OutcomeOdds, Where: domain.ConditionLT, Value: floatPtr(1.25)}, 1,
			verdict{true, 1.5, "filter_condition ODDS LT 1.25 not met (current 1.5)"},
		},
		{
			"lower-case names",
			&domain.FilterCondition{By: "total_users", Where: "gte", Value: floatPtr(30)}, 1,
			verdict{true, 25, "filter_condition TOTAL_USERS GTE 30 not met (current 25)"},
		},
		{
			"a threshold of zero",
			&domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGT, Value: floatPtr(0)}, 1,
			verdict{false, 25, ""},
		},
		{
			"no chosen outcome to read",
			&domain.FilterCondition{By: domain.OutcomeDecisionUsers, Where: domain.ConditionGTE, Value: floatPtr(0)}, 2,
			verdict{true, 0, "filter_condition requires a decision outcome"},
		},
		{
			"an unknown figure",
			&domain.FilterCondition{By: "SHOE_SIZE", Where: domain.ConditionGTE, Value: floatPtr(0)}, 1,
			verdict{true, 0, "filter_condition 'by' unsupported: SHOE_SIZE"},
		},
		{
			"an unknown comparison",
			&domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: "ROUGHLY", Value: floatPtr(0)}, 1,
			verdict{true, 25, "filter_condition 'where' unsupported: ROUGHLY"},
		},
		{"no filter", nil, 1, verdict{}},
		{"a filter with no value", &domain.FilterCondition{By: domain.OutcomeTotalUsers, Where: domain.ConditionGTE}, 1, verdict{}},
		{"a filter with no figure", &domain.FilterCondition{Where: domain.ConditionGTE, Value: floatPtr(30)}, 1, verdict{}},
		{"a filter with no comparison", &domain.FilterCondition{By: domain.OutcomeTotalUsers, Value: floatPtr(30)}, 1, verdict{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := shouldSkip(filtered(row.filter, row.choice)); got != row.want {
				t.Errorf("ShouldSkipByFilter() = %+v, want %+v", got, row.want)
			}
		})
	}

	t.Run("no event", func(t *testing.T) {
		var event *Event
		if got := shouldSkip(event); got != (verdict{}) {
			t.Errorf("ShouldSkipByFilter() = %+v, want %+v", got, verdict{})
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		event := &Event{Outcomes: twoOutcomes(), Decision: Decision{Choice: 1}}
		if got := shouldSkip(event); got != (verdict{}) {
			t.Errorf("ShouldSkipByFilter() = %+v, want %+v", got, verdict{})
		}
	})
}

type figure struct {
	value   float64
	problem string
}

func TestEvent_filterFigure(t *testing.T) {
	const noOutcome = "filter_condition requires a decision outcome"

	for _, row := range []struct {
		name   string
		by     domain.OutcomeKey
		choice int
		want   figure
	}{
		{"total users", domain.OutcomeTotalUsers, 1, figure{25, ""}},
		{"total points", domain.OutcomeTotalPoints, 1, figure{300, ""}},
		{"users on the chosen outcome", domain.OutcomeDecisionUsers, 1, figure{15, ""}},
		{"points on the chosen outcome", domain.OutcomeDecisionPoints, 1, figure{200, ""}},
		{"share of users on the chosen outcome", domain.OutcomePercentageUsers, 1, figure{60, ""}},
		{"odds of the chosen outcome", domain.OutcomeOdds, 1, figure{1.5, ""}},
		{"odds percentage of the chosen outcome", domain.OutcomeOddsPercentage, 1, figure{66.5, ""}},
		{"top bet on the chosen outcome", domain.OutcomeTopPoints, 1, figure{90, ""}},

		{"the first outcome chosen", domain.OutcomeDecisionUsers, 0, figure{10, ""}},
		{"a choice below zero", domain.OutcomeDecisionUsers, -1, figure{0, noOutcome}},
		{"a choice past the last outcome", domain.OutcomeDecisionUsers, 2, figure{0, noOutcome}},
		{"points with no chosen outcome", domain.OutcomeDecisionPoints, 2, figure{0, noOutcome}},
		{"share of users with no chosen outcome", domain.OutcomePercentageUsers, 2, figure{0, noOutcome}},
		{"odds with no chosen outcome", domain.OutcomeOdds, 2, figure{0, noOutcome}},
		{"odds percentage with no chosen outcome", domain.OutcomeOddsPercentage, 2, figure{0, noOutcome}},
		{"top bet with no chosen outcome", domain.OutcomeTopPoints, 2, figure{0, noOutcome}},
		{"total users with no chosen outcome", domain.OutcomeTotalUsers, 2, figure{25, ""}},
		{"total points with no chosen outcome", domain.OutcomeTotalPoints, 2, figure{300, ""}},

		{"an unknown figure", "SHOE_SIZE", 1, figure{0, "filter_condition 'by' unsupported: SHOE_SIZE"}},
		{"a lower-case figure", "total_users", 1, figure{0, "filter_condition 'by' unsupported: total_users"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			event := &Event{Outcomes: twoOutcomes(), Decision: Decision{Choice: row.choice}}
			value, problem := event.filterFigure(row.by)
			if got := (figure{value, problem}); got != row.want {
				t.Errorf("filterFigure(%s) = %+v, want %+v", row.by, got, row.want)
			}
		})
	}

	t.Run("total users of no outcomes", func(t *testing.T) {
		event := &Event{}
		value, problem := event.filterFigure(domain.OutcomeTotalUsers)
		if got := (figure{value, problem}); got != (figure{}) {
			t.Errorf("filterFigure(TOTAL_USERS) = %+v, want %+v", got, figure{})
		}
	})

	t.Run("total points of no outcomes", func(t *testing.T) {
		event := &Event{}
		value, problem := event.filterFigure(domain.OutcomeTotalPoints)
		if got := (figure{value, problem}); got != (figure{}) {
			t.Errorf("filterFigure(TOTAL_POINTS) = %+v, want %+v", got, figure{})
		}
	})
}

type comparison struct{ met, known bool }

func TestConditionMet(t *testing.T) {
	for _, row := range []struct {
		name      string
		where     domain.Condition
		figure    float64
		threshold float64
		want      comparison
	}{
		{"GT, one under", domain.ConditionGT, 24, 25, comparison{false, true}},
		{"GT, exactly on", domain.ConditionGT, 25, 25, comparison{false, true}},
		{"GT, one over", domain.ConditionGT, 26, 25, comparison{true, true}},
		{"GTE, one under", domain.ConditionGTE, 24, 25, comparison{false, true}},
		{"GTE, exactly on", domain.ConditionGTE, 25, 25, comparison{true, true}},
		{"GTE, one over", domain.ConditionGTE, 26, 25, comparison{true, true}},
		{"LT, one under", domain.ConditionLT, 24, 25, comparison{true, true}},
		{"LT, exactly on", domain.ConditionLT, 25, 25, comparison{false, true}},
		{"LT, one over", domain.ConditionLT, 26, 25, comparison{false, true}},
		{"LTE, one under", domain.ConditionLTE, 24, 25, comparison{true, true}},
		{"LTE, exactly on", domain.ConditionLTE, 25, 25, comparison{true, true}},
		{"LTE, one over", domain.ConditionLTE, 26, 25, comparison{false, true}},
		{"a negative threshold", domain.ConditionGT, 0, -1, comparison{true, true}},
		{"an unknown comparison", "ROUGHLY", 25, 25, comparison{false, false}},
		{"a lower-case comparison", "gt", 26, 25, comparison{false, false}},
		{"no comparison", "", 26, 25, comparison{false, false}},
	} {
		t.Run(row.name, func(t *testing.T) {
			met, known := conditionMet(row.where, row.figure, row.threshold)
			if got := (comparison{met, known}); got != row.want {
				t.Errorf("conditionMet(%q, %v, %v) = %+v, want %+v", row.where, row.figure, row.threshold, got, row.want)
			}
		})
	}
}
