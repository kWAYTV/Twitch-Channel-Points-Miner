package main

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/miner"
)

func TestMinerOptionsFrom(t *testing.T) {
	t.Run("every setting set", func(t *testing.T) {
		cfg := config{
			Username:                   "me",
			Password:                   "secret",
			ClaimDropsStartup:          true,
			DisableSSLCertVerification: true,
			ClaimDrops:                 true,
			ClaimMoments:               true,
			BettingMakePredictions:     true,
			FollowRaid:                 true,
			CommunityGoals:             true,
			IRCMode:                    "ALWAYS",
			StreamerOverrides:          map[string]streamerSettingsConfig{"One": {ClaimDrops: ptr(false)}},
			WatchPriority:              []string{"ORDER"},
			StreamersExclude:           []string{"three"},
			GamePriority:               []string{"Chess"},
			GameExclude:                []string{"Slots"},
			DisableAtInNickname:        true,
			ShowGame:                   true,
			WatchQueueLogging:          true,
			WatchStreakWarmStartCache:  true,
			ShowDropsProgress:          true,
		}
		loggerSettings := logging.LoggerSettings{Save: true, Emoji: true}

		got := minerOptionsFrom(cfg, loggerSettings)

		want := miner.Options{
			Username:                   "me",
			Password:                   "secret",
			ClaimDropsStartup:          true,
			DisableSSLCertVerification: true,
			Logger:                     logging.LoggerSettings{Save: true, Emoji: true},
			StreamerSettings: domain.StreamerSettings{
				MakePredictions: true,
				FollowRaid:      true,
				ClaimDrops:      true,
				ClaimMoments:    true,
				WatchStreak:     true,
				CommunityGoals:  true,
				Bet:             defaultBet(),
				IRCMode:         "ALWAYS",
			},
			StreamerOverrides: map[string]domain.StreamerSettings{
				"one": {
					MakePredictions: true,
					FollowRaid:      true,
					ClaimDrops:      false,
					ClaimMoments:    true,
					WatchStreak:     true,
					CommunityGoals:  true,
					Bet:             defaultBet(),
					IRCMode:         "ALWAYS",
				},
			},
			WatchPriority:             []string{"ORDER"},
			StreamersExclude:          []string{"three"},
			GamePriority:              []string{"Chess"},
			GameExclude:               []string{"Slots"},
			DisableAtInNickname:       true,
			ShowGame:                  true,
			WatchQueueLogging:         true,
			WatchStreakWarmStartCache: true,
			ShowDropsProgress:         true,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("minerOptionsFrom() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing set", func(t *testing.T) {
		got := minerOptionsFrom(config{}, logging.LoggerSettings{})

		want := miner.Options{
			StreamerSettings:  domain.StreamerSettings{WatchStreak: true, Bet: defaultBet(), IRCMode: "ONLINE"},
			StreamerOverrides: map[string]domain.StreamerSettings{},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("minerOptionsFrom() = %+v, want %+v", got, want)
		}
	})

	// Each row sets one switch or list and expects only its counterpart set.
	base := func() miner.Options {
		return miner.Options{
			StreamerSettings:  domain.StreamerSettings{WatchStreak: true, Bet: defaultBet(), IRCMode: "ONLINE"},
			StreamerOverrides: map[string]domain.StreamerSettings{},
		}
	}
	oneSetting := []struct {
		name   string
		cfg    config
		change func(want *miner.Options)
	}{
		{name: "only the username", cfg: config{Username: "me"}, change: func(want *miner.Options) { want.Username = "me" }},
		{name: "only the password", cfg: config{Password: "secret"}, change: func(want *miner.Options) { want.Password = "secret" }},
		{name: "only drops at startup", cfg: config{ClaimDropsStartup: true}, change: func(want *miner.Options) { want.ClaimDropsStartup = true }},
		{name: "only certificate checks off", cfg: config{DisableSSLCertVerification: true}, change: func(want *miner.Options) { want.DisableSSLCertVerification = true }},
		{name: "only the watch priority", cfg: config{WatchPriority: []string{"ORDER"}}, change: func(want *miner.Options) { want.WatchPriority = []string{"ORDER"} }},
		{name: "only excluded streamers", cfg: config{StreamersExclude: []string{"three"}}, change: func(want *miner.Options) { want.StreamersExclude = []string{"three"} }},
		{name: "only the game priority", cfg: config{GamePriority: []string{"Chess"}}, change: func(want *miner.Options) { want.GamePriority = []string{"Chess"} }},
		{name: "only excluded games", cfg: config{GameExclude: []string{"Slots"}}, change: func(want *miner.Options) { want.GameExclude = []string{"Slots"} }},
		{name: "only no @ in nicknames", cfg: config{DisableAtInNickname: true}, change: func(want *miner.Options) { want.DisableAtInNickname = true }},
		{name: "only showing the game", cfg: config{ShowGame: true}, change: func(want *miner.Options) { want.ShowGame = true }},
		{name: "only watch queue logging", cfg: config{WatchQueueLogging: true}, change: func(want *miner.Options) { want.WatchQueueLogging = true }},
		{name: "only the warm start cache", cfg: config{WatchStreakWarmStartCache: true}, change: func(want *miner.Options) { want.WatchStreakWarmStartCache = true }},
		{name: "only drops progress", cfg: config{ShowDropsProgress: true}, change: func(want *miner.Options) { want.ShowDropsProgress = true }},
	}
	for _, tt := range oneSetting {
		t.Run(tt.name, func(t *testing.T) {
			want := base()
			tt.change(&want)
			if got := minerOptionsFrom(tt.cfg, logging.LoggerSettings{}); !reflect.DeepEqual(got, want) {
				t.Errorf("minerOptionsFrom() = %+v, want %+v", got, want)
			}
		})
	}
}
