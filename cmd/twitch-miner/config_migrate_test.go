package main

import (
	"reflect"
	"testing"
)

func TestMigrateSettings(t *testing.T) {
	tests := []struct {
		name        string
		settings    map[string]interface{}
		want        map[string]interface{}
		wantChanged bool
	}{
		{
			name:        "password, removed",
			settings:    map[string]interface{}{"username": "me", "password": "secret"},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name:        "claim_drops_startup, removed",
			settings:    map[string]interface{}{"claim_drops": true, "claim_drops_startup": false},
			want:        map[string]interface{}{"claim_drops": true},
			wantChanged: true,
		},
		{
			name:        "watch_streak_warm_start_cache, removed",
			settings:    map[string]interface{}{"username": "me", "watch_streak_warm_start_cache": false},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name:        "disable_ssl_cert_verification, removed",
			settings:    map[string]interface{}{"username": "me", "disable_ssl_cert_verification": true},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name:        "bet.deduct_stake_on_place, removed",
			settings:    map[string]interface{}{"bet": map[string]interface{}{"percentage": 9.0, "deduct_stake_on_place": false}},
			want:        map[string]interface{}{"bet": map[string]interface{}{"percentage": 9.0}},
			wantChanged: true,
		},
		{
			name: "deduct_stake_on_place in a streamer's bet, removed",
			settings: map[string]interface{}{"streamer_overrides": map[string]interface{}{
				"one": map[string]interface{}{"bet": map[string]interface{}{"max_points": 5.0, "deduct_stake_on_place": false}},
				"two": map[string]interface{}{"claim_drops": false},
			}},
			want: map[string]interface{}{"streamer_overrides": map[string]interface{}{
				"one": map[string]interface{}{"bet": map[string]interface{}{"max_points": 5.0}},
				"two": map[string]interface{}{"claim_drops": false},
			}},
			wantChanged: true,
		},
		{
			name:        "betting(make_predictions), renamed",
			settings:    map[string]interface{}{"betting(make_predictions)": false},
			want:        map[string]interface{}{"make_predictions": false},
			wantChanged: true,
		},
		{
			name:        "analytics, renamed",
			settings:    map[string]interface{}{"analytics": map[string]interface{}{"enabled": false, "host": "0.0.0.0", "port": 8080.0}},
			want:        map[string]interface{}{"dashboard": map[string]interface{}{"enabled": false, "host": "0.0.0.0", "port": 8080.0}},
			wantChanged: true,
		},
		{
			name: "discord.webhook_api, renamed",
			settings: map[string]interface{}{"discord": map[string]interface{}{
				"webhook_api": "https://discord.com/api/webhooks/1/a", "events": []interface{}{"BET_WIN"},
			}},
			want: map[string]interface{}{"discord": map[string]interface{}{
				"webhook": "https://discord.com/api/webhooks/1/a", "events": []interface{}{"BET_WIN"},
			}},
			wantChanged: true,
		},
		{
			name:        "telegram.disable_notification, renamed",
			settings:    map[string]interface{}{"telegram": map[string]interface{}{"token": "123:abc", "disable_notification": true}},
			want:        map[string]interface{}{"telegram": map[string]interface{}{"token": "123:abc", "silent": true}},
			wantChanged: true,
		},
		{
			name:        "privacy.anonymize_logs, moved to the top level",
			settings:    map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": true}},
			want:        map[string]interface{}{"anonymize_logs": true},
			wantChanged: true,
		},
		{
			name:        "a privacy object with no keys",
			settings:    map[string]interface{}{"username": "me", "privacy": map[string]interface{}{}},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name:        "a privacy that is not an object",
			settings:    map[string]interface{}{"username": "me", "privacy": nil},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name:        "a privacy object with a key the program does not know",
			settings:    map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": true, "custom": 1.0}},
			want:        map[string]interface{}{"anonymize_logs": true, "privacy": map[string]interface{}{"custom": 1.0}},
			wantChanged: true,
		},
		{
			name:        "disable_at_in_nickname, renamed",
			settings:    map[string]interface{}{"disable_at_in_nickname": true},
			want:        map[string]interface{}{"mention_without_at": true},
			wantChanged: true,
		},
		{
			name:        "watch_queue_logging, renamed",
			settings:    map[string]interface{}{"watch_queue_logging": true},
			want:        map[string]interface{}{"show_watch_queue": true},
			wantChanged: true,
		},
		{
			name:        "debug and debug_deep both on, merged",
			settings:    map[string]interface{}{"debug": true, "debug_deep": true},
			want:        map[string]interface{}{"debug": "full"},
			wantChanged: true,
		},
		{
			name:        "debug on and debug_deep off, merged",
			settings:    map[string]interface{}{"debug": true, "debug_deep": false},
			want:        map[string]interface{}{"debug": true},
			wantChanged: true,
		},
		{
			name:        "debug off and debug_deep on, merged",
			settings:    map[string]interface{}{"debug": false, "debug_deep": true},
			want:        map[string]interface{}{"debug": false},
			wantChanged: true,
		},
		{
			name:        "debug and debug_deep both off, merged",
			settings:    map[string]interface{}{"debug": false, "debug_deep": false},
			want:        map[string]interface{}{"debug": false},
			wantChanged: true,
		},
		{
			name:        "debug_deep on with no debug key",
			settings:    map[string]interface{}{"username": "me", "debug_deep": true},
			want:        map[string]interface{}{"username": "me"},
			wantChanged: true,
		},
		{
			name: "a file already in the new shape",
			settings: map[string]interface{}{
				"username":           "me",
				"make_predictions":   false,
				"mention_without_at": true,
				"show_watch_queue":   true,
				"anonymize_logs":     true,
				"debug":              "full",
				"bet":                map[string]interface{}{"percentage": 9.0},
				"discord":            map[string]interface{}{"webhook": "https://discord.com/api/webhooks/1/a"},
				"telegram":           map[string]interface{}{"silent": true},
				"dashboard":          map[string]interface{}{"port": 8080.0},
				"streamer_overrides": map[string]interface{}{"one": map[string]interface{}{"bet": map[string]interface{}{"max_points": 5.0}}},
			},
			want: map[string]interface{}{
				"username":           "me",
				"make_predictions":   false,
				"mention_without_at": true,
				"show_watch_queue":   true,
				"anonymize_logs":     true,
				"debug":              "full",
				"bet":                map[string]interface{}{"percentage": 9.0},
				"discord":            map[string]interface{}{"webhook": "https://discord.com/api/webhooks/1/a"},
				"telegram":           map[string]interface{}{"silent": true},
				"dashboard":          map[string]interface{}{"port": 8080.0},
				"streamer_overrides": map[string]interface{}{"one": map[string]interface{}{"bet": map[string]interface{}{"max_points": 5.0}}},
			},
			wantChanged: false,
		},
		{
			name: "a key the program does not know",
			settings: map[string]interface{}{
				"an_old_key": true,
				"bet":        map[string]interface{}{"custom": 1.0},
			},
			want: map[string]interface{}{
				"an_old_key": true,
				"bet":        map[string]interface{}{"custom": 1.0},
			},
			wantChanged: false,
		},
		{
			name:        "no settings",
			settings:    map[string]interface{}{},
			want:        map[string]interface{}{},
			wantChanged: false,
		},
		{
			name: "groups and overrides that are not objects",
			settings: map[string]interface{}{
				"bet":                "SMART",
				"discord":            nil,
				"telegram":           5.0,
				"streamer_overrides": map[string]interface{}{"one": "yes", "two": map[string]interface{}{"bet": nil}},
			},
			want: map[string]interface{}{
				"bet":                "SMART",
				"discord":            nil,
				"telegram":           5.0,
				"streamer_overrides": map[string]interface{}{"one": "yes", "two": map[string]interface{}{"bet": nil}},
			},
			wantChanged: false,
		},
		{
			name: "both the old and the new name of every renamed key",
			settings: map[string]interface{}{
				"betting(make_predictions)": true,
				"make_predictions":          false,
				"analytics":                 map[string]interface{}{"port": 5000.0},
				"dashboard":                 map[string]interface{}{"port": 8080.0},
				"disable_at_in_nickname":    false,
				"mention_without_at":        true,
				"watch_queue_logging":       false,
				"show_watch_queue":          true,
				"privacy":                   map[string]interface{}{"anonymize_logs": false},
				"anonymize_logs":            true,
				"discord":                   map[string]interface{}{"webhook_api": "https://old.example", "webhook": "https://new.example"},
				"telegram":                  map[string]interface{}{"disable_notification": false, "silent": true},
			},
			want: map[string]interface{}{
				"make_predictions":   false,
				"dashboard":          map[string]interface{}{"port": 8080.0},
				"mention_without_at": true,
				"show_watch_queue":   true,
				"anonymize_logs":     true,
				"discord":            map[string]interface{}{"webhook": "https://new.example"},
				"telegram":           map[string]interface{}{"silent": true},
			},
			wantChanged: true,
		},
		{
			name:        "debug already full and debug_deep off",
			settings:    map[string]interface{}{"debug": "full", "debug_deep": false},
			want:        map[string]interface{}{"debug": "full"},
			wantChanged: true,
		},
		{
			name:        "debug already full and debug_deep on",
			settings:    map[string]interface{}{"debug": "full", "debug_deep": true},
			want:        map[string]interface{}{"debug": "full"},
			wantChanged: true,
		},
		{
			name: "every old key at once",
			settings: map[string]interface{}{
				"username":                      "me",
				"password":                      "secret",
				"claim_drops":                   true,
				"claim_drops_startup":           false,
				"watch_streak_warm_start_cache": false,
				"disable_ssl_cert_verification": true,
				"betting(make_predictions)":     false,
				"disable_at_in_nickname":        true,
				"watch_queue_logging":           true,
				"debug":                         true,
				"debug_deep":                    true,
				"privacy":                       map[string]interface{}{"anonymize_logs": true},
				"analytics":                     map[string]interface{}{"enabled": false, "host": "0.0.0.0", "port": 8080.0},
				"discord":                       map[string]interface{}{"webhook_api": "https://discord.com/api/webhooks/1/a", "events": []interface{}{"BET_WIN"}},
				"telegram":                      map[string]interface{}{"token": "123:abc", "chat_id": 42.0, "events": []interface{}{}, "disable_notification": true},
				"bet":                           map[string]interface{}{"percentage": 9.0, "deduct_stake_on_place": false},
				"streamer_overrides": map[string]interface{}{
					"one": map[string]interface{}{"make_predictions": false, "bet": map[string]interface{}{"deduct_stake_on_place": false}},
				},
			},
			want: map[string]interface{}{
				"username":           "me",
				"claim_drops":        true,
				"make_predictions":   false,
				"mention_without_at": true,
				"show_watch_queue":   true,
				"debug":              "full",
				"anonymize_logs":     true,
				"dashboard":          map[string]interface{}{"enabled": false, "host": "0.0.0.0", "port": 8080.0},
				"discord":            map[string]interface{}{"webhook": "https://discord.com/api/webhooks/1/a", "events": []interface{}{"BET_WIN"}},
				"telegram":           map[string]interface{}{"token": "123:abc", "chat_id": 42.0, "events": []interface{}{}, "silent": true},
				"bet":                map[string]interface{}{"percentage": 9.0},
				"streamer_overrides": map[string]interface{}{
					"one": map[string]interface{}{"make_predictions": false, "bet": map[string]interface{}{}},
				},
			},
			wantChanged: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := migrateSettings(tt.settings)
			if changed != tt.wantChanged {
				t.Errorf("migrateSettings() = %v, want %v", changed, tt.wantChanged)
			}
			if !reflect.DeepEqual(tt.settings, tt.want) {
				t.Errorf("migrateSettings() left %v, want %v", tt.settings, tt.want)
			}
		})
	}
}
