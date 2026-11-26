package main

import (
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

type filterConditionConfig struct {
	By    string   `json:"by"`
	Where string   `json:"where"`
	Value *float64 `json:"value"`
}

type betConfig struct {
	Strategy        string                 `json:"strategy"`
	Percentage      *int                   `json:"percentage"`
	PercentageGap   *int                   `json:"percentage_gap"`
	MaxPoints       *int                   `json:"max_points"`
	StealthMode     *bool                  `json:"stealth_mode"`
	DelayMode       string                 `json:"delay_mode"`
	Delay           *float64               `json:"delay"`
	MinimumPoints   *int                   `json:"minimum_points"`
	FilterCondition *filterConditionConfig `json:"filter_condition"`
}

func mergeBetSettings(base domain.BetSettings, override betConfig) domain.BetSettings {
	out := base
	if strategy := settingName(override.Strategy); strategy != "" {
		out.Strategy = domain.Strategy(strategy)
	}
	if override.Percentage != nil {
		out.Percentage = override.Percentage
	}
	if override.PercentageGap != nil {
		out.PercentageGap = override.PercentageGap
	}
	if override.MaxPoints != nil {
		out.MaxPoints = override.MaxPoints
	}
	if override.MinimumPoints != nil {
		out.MinimumPoints = override.MinimumPoints
	}
	if override.StealthMode != nil {
		out.StealthMode = override.StealthMode
	}
	if override.FilterCondition != nil {
		out.FilterCondition = mergeFilterCondition(out.FilterCondition, override.FilterCondition)
	}
	if mode := settingName(override.DelayMode); mode != "" {
		out.DelayMode = domain.DelayMode(mode)
	}
	if override.Delay != nil {
		out.Delay = override.Delay
	}
	out.Default()
	return out
}

func mergeFilterCondition(base *domain.FilterCondition, override *filterConditionConfig) *domain.FilterCondition {
	if override == nil {
		return base
	}
	var out domain.FilterCondition
	if base != nil {
		out = *base
	}
	if by := strings.TrimSpace(override.By); by != "" {
		out.By = domain.OutcomeKey(strings.ToUpper(by))
	}
	if where := strings.TrimSpace(override.Where); where != "" {
		out.Where = domain.Condition(strings.ToUpper(where))
	}
	if override.Value != nil {
		out.Value = override.Value
	}
	// If nothing was set, keep nil to avoid activating an empty filter
	if out.By == "" && out.Where == "" && out.Value == nil {
		return base
	}
	return &out
}
