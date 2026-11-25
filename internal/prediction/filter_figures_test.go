package prediction

import (
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func filterEvent(by domain.OutcomeKey, where domain.Condition, value float64, choice int) *Event {
	return &Event{
		Streamer: &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{
			FilterCondition: &domain.FilterCondition{By: by, Where: where, Value: &value},
		}}},
		Outcomes: []Outcome{
			{TotalUsers: 10, TotalPoints: 100, TopPoints: 40, PercentageUsers: 40, Odds: 2.5, OddsPercentage: 40},
			{TotalUsers: 15, TotalPoints: 200, TopPoints: 90, PercentageUsers: 60, Odds: 1.5, OddsPercentage: 66.5},
		},
		Decision: Decision{Choice: choice},
	}
}

func TestFilterComparesTheRightFigure(t *testing.T) {
	for by, want := range map[domain.OutcomeKey]float64{
		domain.OutcomeTotalUsers:      25,
		domain.OutcomeTotalPoints:     300,
		domain.OutcomeDecisionUsers:   15,
		domain.OutcomeDecisionPoints:  200,
		domain.OutcomePercentageUsers: 60,
		domain.OutcomeOdds:            1.5,
		domain.OutcomeOddsPercentage:  66.5,
		domain.OutcomeTopPoints:       90,
	} {
		skip, compared, reason := filterEvent(by, domain.ConditionGTE, 0, 1).ShouldSkipByFilter()
		if skip || compared != want || reason != "" {
			t.Fatalf("%s: got skip=%t compared=%v reason=%q, want compared %v and no skip", by, skip, compared, reason, want)
		}
	}
}

func TestFilterConditions(t *testing.T) {
	for _, tc := range []struct {
		where domain.Condition
		value float64
		skip  bool
	}{
		{domain.ConditionGT, 24, false}, {domain.ConditionGT, 25, true},
		{domain.ConditionGTE, 25, false}, {domain.ConditionGTE, 26, true},
		{domain.ConditionLT, 26, false}, {domain.ConditionLT, 25, true},
		{domain.ConditionLTE, 25, false}, {domain.ConditionLTE, 24, true},
	} {
		skip, compared, reason := filterEvent(domain.OutcomeTotalUsers, tc.where, tc.value, 1).ShouldSkipByFilter()
		if skip != tc.skip || compared != 25 {
			t.Fatalf("25 %s %v: got skip=%t compared=%v, want skip=%t", tc.where, tc.value, skip, compared, tc.skip)
		}
		if tc.skip && !strings.Contains(reason, "not met (current 25)") {
			t.Fatalf("25 %s %v: a skipped bet should say why, got %q", tc.where, tc.value, reason)
		}
	}
}

func TestFilterIsCaseInsensitive(t *testing.T) {
	skip, compared, _ := filterEvent("total_users", "gte", 20, 1).ShouldSkipByFilter()
	if skip || compared != 25 {
		t.Fatalf("lower-case filter names should work, got skip=%t compared=%v", skip, compared)
	}
}

func TestFilterProblemsSkipTheBet(t *testing.T) {
	for _, by := range []domain.OutcomeKey{domain.OutcomeDecisionUsers, domain.OutcomeDecisionPoints, domain.OutcomePercentageUsers,
		domain.OutcomeOdds, domain.OutcomeOddsPercentage, domain.OutcomeTopPoints} {
		for _, choice := range []int{-1, 2} {
			skip, compared, reason := filterEvent(by, domain.ConditionGTE, 0, choice).ShouldSkipByFilter()
			if !skip || compared != 0 || reason != "filter_condition requires a decision outcome" {
				t.Fatalf("%s with choice %d: got skip=%t compared=%v reason=%q", by, choice, skip, compared, reason)
			}
		}
	}

	skip, compared, reason := filterEvent("SHOE_SIZE", domain.ConditionGTE, 0, 1).ShouldSkipByFilter()
	if !skip || compared != 0 || reason != "filter_condition 'by' unsupported: SHOE_SIZE" {
		t.Fatalf("unknown figure: got skip=%t compared=%v reason=%q", skip, compared, reason)
	}
	skip, compared, reason = filterEvent(domain.OutcomeTotalUsers, "ROUGHLY", 0, 1).ShouldSkipByFilter()
	if !skip || compared != 25 || reason != "filter_condition 'where' unsupported: ROUGHLY" {
		t.Fatalf("unknown comparison: got skip=%t compared=%v reason=%q", skip, compared, reason)
	}
}

func TestNoFilterNeverSkips(t *testing.T) {
	for _, event := range []*Event{
		nil,
		{},
		{Streamer: &domain.Streamer{}},
		{Streamer: &domain.Streamer{Settings: domain.StreamerSettings{Bet: domain.BetSettings{FilterCondition: &domain.FilterCondition{By: domain.OutcomeTotalUsers}}}}},
	} {
		if skip, compared, reason := event.ShouldSkipByFilter(); skip || compared != 0 || reason != "" {
			t.Fatalf("an incomplete filter should not skip, got skip=%t compared=%v reason=%q", skip, compared, reason)
		}
	}
}
