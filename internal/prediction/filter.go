package prediction

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const noDecisionOutcome = "filter_condition requires a decision outcome"

// chosenOutcomeFigures are the figures a filter can read from the outcome the
// bet would be placed on.
var chosenOutcomeFigures = map[domain.OutcomeKey]func(Outcome) float64{
	domain.OutcomeDecisionUsers:   func(o Outcome) float64 { return float64(o.TotalUsers) },
	domain.OutcomeDecisionPoints:  func(o Outcome) float64 { return float64(o.TotalPoints) },
	domain.OutcomePercentageUsers: func(o Outcome) float64 { return o.PercentageUsers },
	domain.OutcomeOdds:            func(o Outcome) float64 { return o.Odds },
	domain.OutcomeOddsPercentage:  func(o Outcome) float64 { return o.OddsPercentage },
	domain.OutcomeTopPoints:       func(o Outcome) float64 { return float64(o.TopPoints) },
}

// ShouldSkipByFilter applies the streamer's optional filter_condition. It
// reports whether the bet should be skipped, the figure that was compared, and
// the reason for skipping.
func (p *Event) ShouldSkipByFilter() (skip bool, compared float64, reason string) {
	if p == nil || p.Streamer == nil {
		return false, 0, ""
	}
	filter := p.Streamer.Settings.Bet.FilterCondition
	if filter == nil || filter.Value == nil || filter.By == "" || filter.Where == "" {
		return false, 0, ""
	}
	by := domain.OutcomeKey(strings.ToUpper(string(filter.By)))
	where := domain.Condition(strings.ToUpper(string(filter.Where)))
	threshold := *filter.Value

	compared, problem := p.filterFigure(by)
	if problem != "" {
		return true, 0, problem
	}
	met, known := conditionMet(where, compared, threshold)
	if !known {
		return true, compared, fmt.Sprintf("filter_condition 'where' unsupported: %s", where)
	}
	if met {
		return false, compared, ""
	}
	return true, compared, fmt.Sprintf("filter_condition %s %s %s not met (current %s)", by, where, FormatFloat(threshold), FormatFloat(compared))
}

// filterFigure returns the figure a filter compares, or the problem that
// prevents reading it.
func (p *Event) filterFigure(by domain.OutcomeKey) (figure float64, problem string) {
	switch by {
	case domain.OutcomeTotalUsers:
		for _, outcome := range p.Outcomes {
			figure += float64(outcome.TotalUsers)
		}
		return figure, ""
	case domain.OutcomeTotalPoints:
		for _, outcome := range p.Outcomes {
			figure += float64(outcome.TotalPoints)
		}
		return figure, ""
	}
	read, known := chosenOutcomeFigures[by]
	if !known {
		return 0, fmt.Sprintf("filter_condition 'by' unsupported: %s", by)
	}
	if p.Decision.Choice < 0 || p.Decision.Choice >= len(p.Outcomes) {
		return 0, noDecisionOutcome
	}
	return read(p.Outcomes[p.Decision.Choice]), ""
}

// conditionMet compares a figure with the threshold. known is false for a
// comparison the filter does not support.
func conditionMet(where domain.Condition, figure, threshold float64) (met, known bool) {
	switch where {
	case domain.ConditionGT:
		return figure > threshold, true
	case domain.ConditionGTE:
		return figure >= threshold, true
	case domain.ConditionLT:
		return figure < threshold, true
	case domain.ConditionLTE:
		return figure <= threshold, true
	default:
		return false, false
	}
}
