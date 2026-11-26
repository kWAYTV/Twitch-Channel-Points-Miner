package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

// olderConfig is a complete config.json as the version before the settings
// were simplified wrote it, with a value other than the default under every
// key that was since renamed, merged or removed.
const olderConfig = `{
  "username": "me",
  "password": "secret",
  "auto_update": false,

  "streamers": ["one", "two"],
  "streamers_exclude": [],
  "streamer_overrides": {
    "one": {
      "make_predictions": false,
      "bet": {
        "max_points": 500,
        "deduct_stake_on_place": false
      }
    }
  },
  "watch_priority": ["ORDER"],
  "game_priority": [],
  "game_exclude": [],

  "claim_drops": true,
  "claim_drops_startup": false,
  "claim_moments": true,
  "follow_raid": true,
  "community_goals": false,
  "chat_presence": "ONLINE",
  "disable_at_in_nickname": true,
  "watch_streak_warm_start_cache": false,

  "betting(make_predictions)": false,
  "bet": {
    "strategy": "SMART",
    "percentage": 9,
    "percentage_gap": 20,
    "max_points": 50000,
    "minimum_points": 0,
    "stealth_mode": false,
    "deduct_stake_on_place": false,
    "delay_mode": "FROM_END",
    "delay": 6,
    "filter_condition": {
      "by": null,
      "where": null,
      "value": null
    }
  },

  "discord": {
    "webhook_api": "https://discord.com/api/webhooks/1/a",
    "events": ["BET_WIN"]
  },
  "telegram": {
    "token": "123:abc",
    "chat_id": 42,
    "events": [],
    "disable_notification": true
  },

  "analytics": {
    "enabled": false,
    "host": "0.0.0.0",
    "port": 8080
  },

  "emojis": true,
  "show_seconds": false,
  "show_game": true,
  "show_drops_progress": false,
  "watch_queue_logging": true,
  "save_logs": false,
  "timezone": null,
  "privacy": {
    "anonymize_logs": true
  },
  "debug": true,
  "debug_deep": true,
  "disable_ssl_cert_verification": true,

  "smart_logging": true
}
`

// olderConfigConverted is olderConfig as the loader writes it back.
const olderConfigConverted = `{
  "username": "me",
  "auto_update": false,

  "streamers": ["one", "two"],
  "streamers_exclude": [],
  "streamer_overrides": {
    "one": {
      "make_predictions": false,
      "bet": {
        "max_points": 500
      }
    }
  },
  "watch_priority": ["ORDER"],
  "game_priority": [],
  "game_exclude": [],

  "claim_drops": true,
  "claim_moments": true,
  "follow_raid": true,
  "community_goals": false,
  "chat_presence": "ONLINE",
  "mention_without_at": true,

  "make_predictions": false,
  "bet": {
    "strategy": "SMART",
    "percentage": 9,
    "percentage_gap": 20,
    "max_points": 50000,
    "minimum_points": 0,
    "stealth_mode": false,
    "delay_mode": "FROM_END",
    "delay": 6,
    "filter_condition": {
      "by": null,
      "where": null,
      "value": null
    }
  },

  "discord": {
    "webhook": "https://discord.com/api/webhooks/1/a",
    "events": ["BET_WIN"]
  },
  "telegram": {
    "token": "123:abc",
    "chat_id": 42,
    "events": [],
    "silent": true
  },

  "dashboard": {
    "enabled": false,
    "host": "0.0.0.0",
    "port": 8080
  },

  "emojis": true,
  "show_seconds": false,
  "show_game": true,
  "show_drops_progress": false,
  "show_watch_queue": true,
  "save_logs": false,
  "timezone": null,
  "anonymize_logs": true,
  "debug": "full",

  "smart_logging": true
}
`

func TestLoadOrCreateConfig_olderVersion(t *testing.T) {
	t.Run("a complete file from an older version", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, olderConfig)

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		want := config{
			Username:         "me",
			Debug:            debugFull,
			ShowWatchQueue:   true,
			ClaimDrops:       true,
			ClaimMoments:     true,
			FollowRaid:       true,
			Emojis:           true,
			ShowGame:         true,
			IRCMode:          "ONLINE",
			MentionWithoutAt: true,
			Streamers:        []string{"one", "two"},
			StreamersExclude: []string{},
			GamePriority:     []string{},
			GameExclude:      []string{},
			WatchPriority:    []string{"ORDER"},
			Bet: betConfig{
				Strategy:        "SMART",
				Percentage:      ptr(9),
				PercentageGap:   ptr(20),
				MaxPoints:       ptr(50000),
				StealthMode:     ptr(false),
				DelayMode:       "FROM_END",
				Delay:           ptr(6.0),
				MinimumPoints:   ptr(0),
				FilterCondition: &filterConditionConfig{},
			},
			AnonymizeLogs: true,
			Discord:       discordConfig{Webhook: "https://discord.com/api/webhooks/1/a", Events: []string{"BET_WIN"}},
			Dashboard:     dashboardConfig{Enabled: false, Host: "0.0.0.0", Port: 8080},
			Telegram:      telegramConfig{ChatID: "42", Token: "123:abc", Events: []string{}, Silent: true},
			StreamerOverrides: map[string]streamerSettingsConfig{
				"one": {MakePredictions: ptr(false), Bet: betConfig{MaxPoints: ptr(500)}},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if written := readFile(t, path); written != olderConfigConverted {
			t.Errorf("wrote\n%s\nwant\n%s", written, olderConfigConverted)
		}
	})

	t.Run("the converted file a second time", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, olderConfig)
		first, err := loadOrCreateConfig(path)
		if err != nil {
			t.Fatal(err)
		}

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		if !reflect.DeepEqual(got, first) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, first)
		}
		if written := readFile(t, path); written != olderConfigConverted {
			t.Errorf("the file now holds\n%s\nwant\n%s", written, olderConfigConverted)
		}
	})

	t.Run("an older file with keys missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, `{"username": "me", "betting(make_predictions)": false, "privacy": {"anonymize_logs": true}}`)

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		want := defaultsDecoded()
		want.Username = "me"
		want.MakePredictions = false
		want.AnonymizeLogs = true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		wantSettings := defaultConfig()
		wantSettings["username"] = "me"
		wantSettings["make_predictions"] = false
		wantSettings["anonymize_logs"] = true
		if written, want := readFile(t, path), formatted(t, wantSettings); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
	})
}
