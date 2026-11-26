package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

// describe prints settings with the values their pointers hold.
func describe(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(raw)
}

func TestStreamerSettings_Default(t *testing.T) {
	t.Run("nothing set", func(t *testing.T) {
		settings := StreamerSettings{}

		settings.Default()

		want := StreamerSettings{
			Bet: BetSettings{
				Strategy:      StrategySmart,
				Percentage:    intPtr(5),
				PercentageGap: intPtr(20),
				MaxPoints:     intPtr(50000),
				MinimumPoints: intPtr(0),
				StealthMode:   boolPtr(false),
				Delay:         floatPtr(6),
				DelayMode:     DelayModeFromEnd,
			},
			IRCMode: IRCModeOnline,
		}
		if !reflect.DeepEqual(settings, want) {
			t.Errorf("Default() left %s, want %s", describe(settings), describe(want))
		}
	})

	t.Run("everything set", func(t *testing.T) {
		settings := StreamerSettings{
			MakePredictions: true,
			FollowRaid:      true,
			ClaimDrops:      true,
			ClaimMoments:    true,
			WatchStreak:     true,
			CommunityGoals:  true,
			Bet: BetSettings{
				Strategy:      StrategyHighOdds,
				Percentage:    intPtr(10),
				PercentageGap: intPtr(35),
				MaxPoints:     intPtr(900),
				MinimumPoints: intPtr(2000),
				StealthMode:   boolPtr(true),
				Delay:         floatPtr(0.5),
				DelayMode:     DelayModePercentage,
			},
			IRCMode: IRCModeNever,
		}

		settings.Default()

		want := StreamerSettings{
			MakePredictions: true,
			FollowRaid:      true,
			ClaimDrops:      true,
			ClaimMoments:    true,
			WatchStreak:     true,
			CommunityGoals:  true,
			Bet: BetSettings{
				Strategy:      StrategyHighOdds,
				Percentage:    intPtr(10),
				PercentageGap: intPtr(35),
				MaxPoints:     intPtr(900),
				MinimumPoints: intPtr(2000),
				StealthMode:   boolPtr(true),
				Delay:         floatPtr(0.5),
				DelayMode:     DelayModePercentage,
			},
			IRCMode: IRCModeNever,
		}
		if !reflect.DeepEqual(settings, want) {
			t.Errorf("Default() left %s, want %s", describe(settings), describe(want))
		}
	})
}
