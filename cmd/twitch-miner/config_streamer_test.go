package main

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// defaultBet is what an empty bet becomes once its defaults are filled in.
func defaultBet() domain.BetSettings {
	return domain.BetSettings{
		Strategy:           "SMART",
		Percentage:         ptr(5),
		PercentageGap:      ptr(20),
		MaxPoints:          ptr(50000),
		MinimumPoints:      ptr(0),
		StealthMode:        ptr(false),
		DeductStakeOnPlace: ptr(true),
		Delay:              ptr(6.0),
		DelayMode:          "FROM_END",
	}
}

// customStreamer has every switch on and a chat presence other than the default.
func customStreamer() domain.StreamerSettings {
	return domain.StreamerSettings{
		MakePredictions: true,
		FollowRaid:      true,
		ClaimDrops:      true,
		ClaimMoments:    true,
		WatchStreak:     true,
		CommunityGoals:  true,
		Bet:             customBet(),
		IRCMode:         "NEVER",
	}
}

func TestMergeStreamerSettings(t *testing.T) {
	t.Run("nothing overridden", func(t *testing.T) {
		got := mergeStreamerSettings(customStreamer(), streamerSettingsConfig{})
		if want := customStreamer(); !reflect.DeepEqual(got, want) {
			t.Errorf("mergeStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing overridden and nothing in the base", func(t *testing.T) {
		got := mergeStreamerSettings(domain.StreamerSettings{}, streamerSettingsConfig{})
		want := domain.StreamerSettings{Bet: defaultBet(), IRCMode: "ONLINE"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("everything overridden", func(t *testing.T) {
		override := streamerSettingsConfig{
			MakePredictions: ptr(false),
			FollowRaid:      ptr(false),
			ClaimDrops:      ptr(false),
			ClaimMoments:    ptr(false),
			WatchStreak:     ptr(false),
			CommunityGoals:  ptr(false),
			Bet:             betConfig{MaxPoints: ptr(2500)},
			IRCMode:         ptr("always"),
		}
		got := mergeStreamerSettings(customStreamer(), override)
		wantBet := customBet()
		wantBet.MaxPoints = ptr(2500)
		want := domain.StreamerSettings{Bet: wantBet, IRCMode: "ALWAYS"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	// Each row overrides one field of customStreamer and expects only that field to change.
	oneField := []struct {
		name     string
		override streamerSettingsConfig
		change   func(want *domain.StreamerSettings)
	}{
		{name: "only predictions", override: streamerSettingsConfig{MakePredictions: ptr(false)}, change: func(want *domain.StreamerSettings) { want.MakePredictions = false }},
		{name: "only raids", override: streamerSettingsConfig{FollowRaid: ptr(false)}, change: func(want *domain.StreamerSettings) { want.FollowRaid = false }},
		{name: "only drops", override: streamerSettingsConfig{ClaimDrops: ptr(false)}, change: func(want *domain.StreamerSettings) { want.ClaimDrops = false }},
		{name: "only moments", override: streamerSettingsConfig{ClaimMoments: ptr(false)}, change: func(want *domain.StreamerSettings) { want.ClaimMoments = false }},
		{name: "only the watch streak", override: streamerSettingsConfig{WatchStreak: ptr(false)}, change: func(want *domain.StreamerSettings) { want.WatchStreak = false }},
		{name: "only community goals", override: streamerSettingsConfig{CommunityGoals: ptr(false)}, change: func(want *domain.StreamerSettings) { want.CommunityGoals = false }},
		{name: "only a bet setting", override: streamerSettingsConfig{Bet: betConfig{MaxPoints: ptr(2500)}}, change: func(want *domain.StreamerSettings) { want.Bet.MaxPoints = ptr(2500) }},
		{name: "only the chat presence", override: streamerSettingsConfig{IRCMode: ptr("online")}, change: func(want *domain.StreamerSettings) { want.IRCMode = "ONLINE" }},
		{name: "a chat presence that is not one", override: streamerSettingsConfig{IRCMode: ptr("maybe")}, change: func(want *domain.StreamerSettings) {}},
	}
	for _, tt := range oneField {
		t.Run(tt.name, func(t *testing.T) {
			want := customStreamer()
			tt.change(&want)
			if got := mergeStreamerSettings(customStreamer(), tt.override); !reflect.DeepEqual(got, want) {
				t.Errorf("mergeStreamerSettings() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestParseChatPresence(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		fallback domain.IRCMode
		want     domain.IRCMode
	}{
		{name: "ALWAYS", mode: "ALWAYS", fallback: "ONLINE", want: "ALWAYS"},
		{name: "NEVER", mode: "NEVER", fallback: "ONLINE", want: "NEVER"},
		{name: "OFFLINE", mode: "OFFLINE", fallback: "ONLINE", want: "OFFLINE"},
		{name: "ONLINE", mode: "ONLINE", fallback: "NEVER", want: "ONLINE"},
		{name: "lower case", mode: "always", fallback: "ONLINE", want: "ALWAYS"},
		{name: "padded", mode: "  never ", fallback: "ONLINE", want: "NEVER"},
		{name: "empty", mode: "", fallback: "NEVER", want: "NEVER"},
		{name: "another word", mode: "nope", fallback: "NEVER", want: "NEVER"},
		{name: "another word and no fallback", mode: "nope", fallback: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseChatPresence(tt.mode, tt.fallback); got != tt.want {
				t.Errorf("parseChatPresence(%q, %q) = %q, want %q", tt.mode, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestBuildBaseStreamerSettings(t *testing.T) {
	t.Run("an empty config", func(t *testing.T) {
		got := buildBaseStreamerSettings(config{})
		want := domain.StreamerSettings{WatchStreak: true, Bet: defaultBet(), IRCMode: "ONLINE"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("buildBaseStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("every setting set", func(t *testing.T) {
		cfg := config{
			BettingMakePredictions: true,
			FollowRaid:             true,
			ClaimDrops:             true,
			ClaimMoments:           true,
			CommunityGoals:         true,
			IRCMode:                "never",
			Bet: betConfig{
				Strategy:           "PERCENTAGE",
				Percentage:         ptr(10),
				PercentageGap:      ptr(30),
				MaxPoints:          ptr(1000),
				MinimumPoints:      ptr(100),
				StealthMode:        ptr(true),
				DeductStakeOnPlace: ptr(false),
				FilterCondition:    &filterConditionConfig{By: "total_users", Where: "gt", Value: ptr(50.0)},
				Delay:              ptr(10.0),
				DelayMode:          "FROM_START",
			},
		}
		got := buildBaseStreamerSettings(cfg)
		want := domain.StreamerSettings{
			MakePredictions: true,
			FollowRaid:      true,
			ClaimDrops:      true,
			ClaimMoments:    true,
			WatchStreak:     true,
			CommunityGoals:  true,
			IRCMode:         "NEVER",
			Bet: domain.BetSettings{
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
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("buildBaseStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("a filter with nothing in it", func(t *testing.T) {
		cfg := config{Bet: betConfig{FilterCondition: &filterConditionConfig{}}}
		got := buildBaseStreamerSettings(cfg)
		want := domain.StreamerSettings{WatchStreak: true, Bet: defaultBet(), IRCMode: "ONLINE"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("buildBaseStreamerSettings() = %+v, want %+v", got, want)
		}
	})

	t.Run("a chat presence that is not one", func(t *testing.T) {
		got := buildBaseStreamerSettings(config{IRCMode: "maybe"})
		want := domain.StreamerSettings{WatchStreak: true, Bet: defaultBet(), IRCMode: "ONLINE"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("buildBaseStreamerSettings() = %+v, want %+v", got, want)
		}
	})
}

func TestBuildOverrideSettings(t *testing.T) {
	withoutMoments := customStreamer()
	withoutMoments.ClaimMoments = false
	withoutDrops := customStreamer()
	withoutDrops.ClaimDrops = false

	tests := []struct {
		name      string
		overrides map[string]streamerSettingsConfig
		want      map[string]domain.StreamerSettings
	}{
		{
			name:      "no overrides",
			overrides: nil,
			want:      map[string]domain.StreamerSettings{},
		},
		{
			name:      "one streamer",
			overrides: map[string]streamerSettingsConfig{"somestreamer": {ClaimMoments: ptr(false)}},
			want:      map[string]domain.StreamerSettings{"somestreamer": withoutMoments},
		},
		{
			name:      "a name in mixed case with spaces around it",
			overrides: map[string]streamerSettingsConfig{" SomeStreamer ": {ClaimMoments: ptr(false)}},
			want:      map[string]domain.StreamerSettings{"somestreamer": withoutMoments},
		},
		{
			name:      "a streamer with nothing overridden",
			overrides: map[string]streamerSettingsConfig{"somestreamer": {}},
			want:      map[string]domain.StreamerSettings{"somestreamer": customStreamer()},
		},
		{
			name:      "a blank name",
			overrides: map[string]streamerSettingsConfig{"  ": {ClaimMoments: ptr(false)}},
			want:      map[string]domain.StreamerSettings{},
		},
		{
			name: "two streamers",
			overrides: map[string]streamerSettingsConfig{
				"first":  {ClaimMoments: ptr(false)},
				"second": {ClaimDrops: ptr(false)},
			},
			want: map[string]domain.StreamerSettings{"first": withoutMoments, "second": withoutDrops},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildOverrideSettings(customStreamer(), tt.overrides); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildOverrideSettings() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
