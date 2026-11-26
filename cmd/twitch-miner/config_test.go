package main

import (
	"reflect"
	"regexp"
	"testing"
)

// defaultsDecoded is the config a first run loads.
func defaultsDecoded() config {
	return config{
		Username:         "your-twitch-username",
		AutoUpdate:       true,
		ClaimDrops:       true,
		ClaimMoments:     true,
		MakePredictions:  true,
		FollowRaid:       true,
		Emojis:           true,
		ShowGame:         true,
		IRCMode:          "ONLINE",
		Streamers:        []string{},
		StreamersExclude: []string{},
		GamePriority:     []string{},
		GameExclude:      []string{},
		WatchPriority:    []string{"STREAK", "DROPS", "ORDER"},
		Bet: betConfig{
			Strategy:        "SMART",
			Percentage:      ptr(5),
			PercentageGap:   ptr(20),
			MaxPoints:       ptr(50000),
			StealthMode:     ptr(false),
			DelayMode:       "FROM_END",
			Delay:           ptr(6.0),
			MinimumPoints:   ptr(0),
			FilterCondition: &filterConditionConfig{},
		},
		Discord:           discordConfig{Events: []string{}},
		Dashboard:         dashboardConfig{Enabled: true, Host: "127.0.0.1", Port: 5000},
		Telegram:          telegramConfig{Events: []string{}},
		Matrix:            matrixConfig{Events: []string{}},
		StreamerOverrides: map[string]streamerSettingsConfig{},
	}
}

func TestDecodeConfig(t *testing.T) {
	t.Run("the defaults", func(t *testing.T) {
		got, err := decodeConfig(defaultConfig())
		if err != nil {
			t.Fatalf("decodeConfig() error = %v, want none", err)
		}
		if want := defaultsDecoded(); !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})

	t.Run("no keys", func(t *testing.T) {
		got, err := decodeConfig(map[string]interface{}{})
		if err != nil {
			t.Fatalf("decodeConfig() error = %v, want none", err)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})

	t.Run("every key set", func(t *testing.T) {
		settings := map[string]interface{}{
			"username":            "me",
			"auto_update":         true,
			"debug":               "full",
			"show_watch_queue":    true,
			"show_seconds":        true,
			"show_username":       true,
			"claim_drops":         true,
			"claim_moments":       true,
			"make_predictions":    true,
			"follow_raid":         true,
			"community_goals":     true,
			"emojis":              true,
			"save_logs":           true,
			"show_game":           true,
			"chat_presence":       "ALWAYS",
			"mention_without_at":  true,
			"show_drops_progress": true,
			"streamers":           []interface{}{"one", "two"},
			"streamers_exclude":   []interface{}{"three"},
			"game_priority":       []interface{}{"Chess"},
			"game_exclude":        []interface{}{"Slots"},
			"watch_priority":      []interface{}{"ORDER"},
			"timezone":            "Europe/Madrid",
			"anonymize_logs":      true,
			"discord":             map[string]interface{}{"webhook": "https://discord.com/api/webhooks/1/a", "events": []interface{}{"BET_WIN"}},
			"dashboard":           map[string]interface{}{"enabled": true, "host": "0.0.0.0", "port": 8080},
			"telegram":            map[string]interface{}{"chat_id": -1001234567890, "token": "123:abc", "events": []interface{}{"BET_LOSE"}, "silent": true},
			"matrix":              map[string]interface{}{"homeserver": "https://matrix.example.org", "access_token": "syt_secret", "room_id": "!room:example.org", "events": []interface{}{"BET_REFUND"}},
			"bet": map[string]interface{}{
				"strategy":         "PERCENTAGE",
				"percentage":       10,
				"percentage_gap":   30,
				"max_points":       1000,
				"stealth_mode":     true,
				"delay_mode":       "FROM_START",
				"delay":            2.5,
				"minimum_points":   100,
				"filter_condition": map[string]interface{}{"by": "TOTAL_USERS", "where": "GT", "value": 50},
			},
			"streamer_overrides": map[string]interface{}{
				"one": map[string]interface{}{
					"make_predictions": false,
					"follow_raid":      false,
					"claim_drops":      false,
					"claim_moments":    false,
					"watch_streak":     false,
					"community_goals":  false,
					"chat_presence":    "NEVER",
					"bet":              map[string]interface{}{"max_points": 5},
				},
			},
		}
		got, err := decodeConfig(settings)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v, want none", err)
		}
		want := config{
			Username:          "me",
			AutoUpdate:        true,
			Debug:             debugFull,
			ShowWatchQueue:    true,
			ShowSeconds:       true,
			ShowUsername:      true,
			ClaimDrops:        true,
			ClaimMoments:      true,
			MakePredictions:   true,
			FollowRaid:        true,
			CommunityGoals:    true,
			Emojis:            true,
			SaveLogs:          true,
			ShowGame:          true,
			IRCMode:           "ALWAYS",
			MentionWithoutAt:  true,
			ShowDropsProgress: true,
			Streamers:         []string{"one", "two"},
			StreamersExclude:  []string{"three"},
			GamePriority:      []string{"Chess"},
			GameExclude:       []string{"Slots"},
			WatchPriority:     []string{"ORDER"},
			Timezone:          ptr("Europe/Madrid"),
			AnonymizeLogs:     true,
			Discord:           discordConfig{Webhook: "https://discord.com/api/webhooks/1/a", Events: []string{"BET_WIN"}},
			Dashboard:         dashboardConfig{Enabled: true, Host: "0.0.0.0", Port: 8080},
			Telegram:          telegramConfig{ChatID: "-1001234567890", Token: "123:abc", Events: []string{"BET_LOSE"}, Silent: true},
			Matrix:            matrixConfig{Homeserver: "https://matrix.example.org", AccessToken: "syt_secret", RoomID: "!room:example.org", Events: []string{"BET_REFUND"}},
			Bet: betConfig{
				Strategy:        "PERCENTAGE",
				Percentage:      ptr(10),
				PercentageGap:   ptr(30),
				MaxPoints:       ptr(1000),
				StealthMode:     ptr(true),
				DelayMode:       "FROM_START",
				Delay:           ptr(2.5),
				MinimumPoints:   ptr(100),
				FilterCondition: &filterConditionConfig{By: "TOTAL_USERS", Where: "GT", Value: ptr(50.0)},
			},
			StreamerOverrides: map[string]streamerSettingsConfig{
				"one": {
					MakePredictions: ptr(false),
					FollowRaid:      ptr(false),
					ClaimDrops:      ptr(false),
					ClaimMoments:    ptr(false),
					WatchStreak:     ptr(false),
					CommunityGoals:  ptr(false),
					IRCMode:         ptr("NEVER"),
					Bet:             betConfig{MaxPoints: ptr(5)},
				},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})

	t.Run("a key the program does not know", func(t *testing.T) {
		got, err := decodeConfig(map[string]interface{}{"username": "me", "an_old_key": true})
		if err != nil {
			t.Fatalf("decodeConfig() error = %v, want none", err)
		}
		if want := (config{Username: "me"}); !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})

	t.Run("a number where text belongs", func(t *testing.T) {
		got, err := decodeConfig(map[string]interface{}{"username": 5})
		wantErr := regexp.MustCompile(`^json: cannot unmarshal number into Go struct field .*username of type string$`)
		if err == nil || !wantErr.MatchString(err.Error()) {
			t.Errorf("decodeConfig() error = %v, want one matching %s", err, wantErr)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})

	t.Run("a value JSON cannot hold", func(t *testing.T) {
		got, err := decodeConfig(map[string]interface{}{"username": make(chan int)})
		if err == nil || err.Error() != "json: unsupported type: chan int" {
			t.Errorf("decodeConfig() error = %v, want json: unsupported type: chan int", err)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("decodeConfig() = %+v, want %+v", got, want)
		}
	})
}
