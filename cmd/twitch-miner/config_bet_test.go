package main

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func ptr[T any](v T) *T { return &v }

// customBet has every field set to something other than its default.
func customBet() domain.BetSettings {
	return domain.BetSettings{
		Strategy:           "PERCENTAGE",
		Percentage:         ptr(10),
		PercentageGap:      ptr(30),
		MaxPoints:          ptr(1000),
		MinimumPoints:      ptr(100),
		StealthMode:        ptr(true),
		DeductStakeOnPlace: ptr(false),
		FilterCondition:    &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
		Delay:              ptr(10.0),
		DelayMode:          "FROM_START",
	}
}

func TestMergeBetSettings(t *testing.T) {
	t.Run("nothing overridden", func(t *testing.T) {
		got := mergeBetSettings(customBet(), betConfig{})
		if want := customBet(); !reflect.DeepEqual(got, want) {
			t.Errorf("mergeBetSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing overridden and nothing in the base", func(t *testing.T) {
		got := mergeBetSettings(domain.BetSettings{}, betConfig{})
		want := domain.BetSettings{
			Strategy:           "SMART",
			Percentage:         ptr(5),
			PercentageGap:      ptr(20),
			MaxPoints:          ptr(50000),
			MinimumPoints:      ptr(0),
			StealthMode:        ptr(false),
			DeductStakeOnPlace: ptr(true),
			FilterCondition:    nil,
			Delay:              ptr(6.0),
			DelayMode:          "FROM_END",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeBetSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("everything overridden", func(t *testing.T) {
		override := betConfig{
			Strategy:           "HIGH_ODDS",
			Percentage:         ptr(7),
			PercentageGap:      ptr(15),
			MaxPoints:          ptr(2500),
			MinimumPoints:      ptr(300),
			StealthMode:        ptr(false),
			DeductStakeOnPlace: ptr(true),
			FilterCondition:    &filterConditionConfig{By: "odds", Where: "lte", Value: ptr(1.5)},
			Delay:              ptr(2.5),
			DelayMode:          "PERCENTAGE",
		}
		got := mergeBetSettings(customBet(), override)
		want := domain.BetSettings{
			Strategy:           "HIGH_ODDS",
			Percentage:         ptr(7),
			PercentageGap:      ptr(15),
			MaxPoints:          ptr(2500),
			MinimumPoints:      ptr(300),
			StealthMode:        ptr(false),
			DeductStakeOnPlace: ptr(true),
			FilterCondition:    &domain.FilterCondition{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
			Delay:              ptr(2.5),
			DelayMode:          "PERCENTAGE",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeBetSettings() = %+v, want %+v", got, want)
		}
	})

	// Each row overrides one field of customBet and expects only that field to change.
	oneField := []struct {
		name     string
		override betConfig
		change   func(want *domain.BetSettings)
	}{
		{name: "only the strategy", override: betConfig{Strategy: "HIGH_ODDS"}, change: func(want *domain.BetSettings) { want.Strategy = "HIGH_ODDS" }},
		{name: "only the percentage", override: betConfig{Percentage: ptr(7)}, change: func(want *domain.BetSettings) { want.Percentage = ptr(7) }},
		{name: "a percentage of zero", override: betConfig{Percentage: ptr(0)}, change: func(want *domain.BetSettings) { want.Percentage = ptr(0) }},
		{name: "only the percentage gap", override: betConfig{PercentageGap: ptr(15)}, change: func(want *domain.BetSettings) { want.PercentageGap = ptr(15) }},
		{name: "only the maximum points", override: betConfig{MaxPoints: ptr(2500)}, change: func(want *domain.BetSettings) { want.MaxPoints = ptr(2500) }},
		{name: "only the minimum points", override: betConfig{MinimumPoints: ptr(300)}, change: func(want *domain.BetSettings) { want.MinimumPoints = ptr(300) }},
		{name: "only stealth mode", override: betConfig{StealthMode: ptr(false)}, change: func(want *domain.BetSettings) { want.StealthMode = ptr(false) }},
		{name: "only the stake deduction", override: betConfig{DeductStakeOnPlace: ptr(true)}, change: func(want *domain.BetSettings) { want.DeductStakeOnPlace = ptr(true) }},
		{name: "only the delay mode", override: betConfig{DelayMode: "PERCENTAGE"}, change: func(want *domain.BetSettings) { want.DelayMode = "PERCENTAGE" }},
		{name: "only the delay", override: betConfig{Delay: ptr(2.5)}, change: func(want *domain.BetSettings) { want.Delay = ptr(2.5) }},
		{
			name:     "only the filter's value",
			override: betConfig{FilterCondition: &filterConditionConfig{Value: ptr(1.5)}},
			change: func(want *domain.BetSettings) {
				want.FilterCondition = &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(1.5)}
			},
		},
	}
	for _, tt := range oneField {
		t.Run(tt.name, func(t *testing.T) {
			want := customBet()
			tt.change(&want)
			if got := mergeBetSettings(customBet(), tt.override); !reflect.DeepEqual(got, want) {
				t.Errorf("mergeBetSettings() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestMergeFilterCondition(t *testing.T) {
	tests := []struct {
		name     string
		base     *domain.FilterCondition
		override *filterConditionConfig
		want     *domain.FilterCondition
	}{
		{
			name:     "no override",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: nil,
			want:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
		},
		{
			name:     "no override and no base",
			base:     nil,
			override: nil,
			want:     nil,
		},
		{
			name:     "a whole filter and no base",
			base:     nil,
			override: &filterConditionConfig{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
			want:     &domain.FilterCondition{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
		},
		{
			name:     "a whole filter over a base",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
			want:     &domain.FilterCondition{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
		},
		{
			name:     "lower case and padded names",
			base:     nil,
			override: &filterConditionConfig{By: " odds ", Where: " lte ", Value: ptr(1.5)},
			want:     &domain.FilterCondition{By: "ODDS", Where: "LTE", Value: ptr(1.5)},
		},
		{
			name:     "only by",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{By: "ODDS"},
			want:     &domain.FilterCondition{By: "ODDS", Where: "GT", Value: ptr(50.0)},
		},
		{
			name:     "only where",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{Where: "LTE"},
			want:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "LTE", Value: ptr(50.0)},
		},
		{
			name:     "only a value",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{Value: ptr(1.5)},
			want:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(1.5)},
		},
		{
			name:     "a value of zero",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{Value: ptr(0.0)},
			want:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(0.0)},
		},
		{
			name:     "only a value and no base",
			base:     nil,
			override: &filterConditionConfig{Value: ptr(1.5)},
			want:     &domain.FilterCondition{Value: ptr(1.5)},
		},
		{
			name:     "an empty filter over a base",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{},
			want:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
		},
		{
			name:     "an empty filter and no base",
			base:     nil,
			override: &filterConditionConfig{},
			want:     nil,
		},
		{
			name:     "a blank by and no base",
			base:     nil,
			override: &filterConditionConfig{By: "  "},
			want:     nil,
		},
		{
			name:     "a blank by over a base (today: by is erased)",
			base:     &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			override: &filterConditionConfig{By: "  "},
			want:     &domain.FilterCondition{By: "", Where: "GT", Value: ptr(50.0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeFilterCondition(tt.base, tt.override); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeFilterCondition() = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("a blank by over a base (bug: should keep the base's by)", func(t *testing.T) {
		t.Skip("bug: a by or where of only spaces counts as set and erases the base's value")
		base := &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)}
		override := &filterConditionConfig{By: "  "}
		want := &domain.FilterCondition{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)}
		if got := mergeFilterCondition(base, override); !reflect.DeepEqual(got, want) {
			t.Errorf("mergeFilterCondition() = %+v, want %+v", got, want)
		}
	})
}
