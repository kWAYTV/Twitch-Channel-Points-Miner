package domain

import (
	"reflect"
	"testing"
)

func intPtr(v int) *int           { return &v }
func boolPtr(v bool) *bool        { return &v }
func floatPtr(v float64) *float64 { return &v }

func TestBetSettings_Default(t *testing.T) {
	t.Run("nothing set", func(t *testing.T) {
		settings := BetSettings{}

		settings.Default()

		want := BetSettings{
			Strategy:      StrategyKelly,
			Percentage:    intPtr(5),
			PercentageGap: intPtr(20),
			MaxPoints:     intPtr(50000),
			MinimumPoints: intPtr(0),
			StealthMode:   boolPtr(false),
			Delay:         floatPtr(0.75),
			DelayMode:     DelayModePercentage,
		}
		if !reflect.DeepEqual(settings, want) {
			t.Errorf("Default() left %s, want %s", describe(settings), describe(want))
		}
	})

	t.Run("everything set", func(t *testing.T) {
		settings := BetSettings{
			Strategy:        StrategyHighOdds,
			Percentage:      intPtr(10),
			PercentageGap:   intPtr(35),
			MaxPoints:       intPtr(900),
			MinimumPoints:   intPtr(2000),
			StealthMode:     boolPtr(true),
			FilterCondition: &FilterCondition{By: OutcomeTotalUsers, Where: ConditionGTE, Value: floatPtr(100)},
			Delay:           floatPtr(0.5),
			DelayMode:       DelayModePercentage,
		}

		settings.Default()

		want := BetSettings{
			Strategy:        StrategyHighOdds,
			Percentage:      intPtr(10),
			PercentageGap:   intPtr(35),
			MaxPoints:       intPtr(900),
			MinimumPoints:   intPtr(2000),
			StealthMode:     boolPtr(true),
			FilterCondition: &FilterCondition{By: OutcomeTotalUsers, Where: ConditionGTE, Value: floatPtr(100)},
			Delay:           floatPtr(0.5),
			DelayMode:       DelayModePercentage,
		}
		if !reflect.DeepEqual(settings, want) {
			t.Errorf("Default() left %s, want %s", describe(settings), describe(want))
		}
	})

	t.Run("everything set to zero", func(t *testing.T) {
		settings := BetSettings{
			Strategy:      StrategyNumber1,
			Percentage:    intPtr(0),
			PercentageGap: intPtr(0),
			MaxPoints:     intPtr(0),
			MinimumPoints: intPtr(0),
			StealthMode:   boolPtr(false),
			Delay:         floatPtr(0),
			DelayMode:     DelayModeFromStart,
		}

		settings.Default()

		want := BetSettings{
			Strategy:      StrategyNumber1,
			Percentage:    intPtr(0),
			PercentageGap: intPtr(0),
			MaxPoints:     intPtr(0),
			MinimumPoints: intPtr(0),
			StealthMode:   boolPtr(false),
			Delay:         floatPtr(0),
			DelayMode:     DelayModeFromStart,
		}
		if !reflect.DeepEqual(settings, want) {
			t.Errorf("Default() left %s, want %s", describe(settings), describe(want))
		}
	})
}
