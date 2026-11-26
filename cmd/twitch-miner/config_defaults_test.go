package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	// config.example.json is what a new user copies, so it has to be exactly
	// what the program writes on a first run.
	t.Run("formatted, it is config.example.json", func(t *testing.T) {
		example := strings.ReplaceAll(readFile(t, filepath.Join("..", "..", "config.example.json")), "\r\n", "\n")

		if got := formatted(t, defaultConfig()); got != example {
			t.Errorf("config.example.json is out of date: copy the config.json a first run writes over it.\nA first run writes:\n%s", got)
		}
	})
}

func TestFillMissing(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		defaults     map[string]interface{}
		wantSettings map[string]interface{}
		wantAdded    bool
	}{
		{
			name:         "nothing missing",
			settings:     map[string]interface{}{"username": "me", "bet": map[string]interface{}{"percentage": 9.0}},
			defaults:     map[string]interface{}{"username": "you", "bet": map[string]interface{}{"percentage": 5}},
			wantSettings: map[string]interface{}{"username": "me", "bet": map[string]interface{}{"percentage": 9.0}},
			wantAdded:    false,
		},
		{
			name:         "a key missing",
			settings:     map[string]interface{}{"username": "me"},
			defaults:     map[string]interface{}{"username": "you", "emojis": true},
			wantSettings: map[string]interface{}{"username": "me", "emojis": true},
			wantAdded:    true,
		},
		{
			name:         "a key missing from a group",
			settings:     map[string]interface{}{"bet": map[string]interface{}{"percentage": 9.0}},
			defaults:     map[string]interface{}{"bet": map[string]interface{}{"percentage": 5, "delay": 6}},
			wantSettings: map[string]interface{}{"bet": map[string]interface{}{"percentage": 9.0, "delay": 6}},
			wantAdded:    true,
		},
		{
			name:         "a whole group missing",
			settings:     map[string]interface{}{"username": "me"},
			defaults:     map[string]interface{}{"username": "you", "bet": map[string]interface{}{"percentage": 5}},
			wantSettings: map[string]interface{}{"username": "me", "bet": map[string]interface{}{"percentage": 5}},
			wantAdded:    true,
		},
		{
			name: "a key missing two groups deep",
			settings: map[string]interface{}{"bet": map[string]interface{}{
				"filter_condition": map[string]interface{}{"by": "ODDS"},
			}},
			defaults: map[string]interface{}{"bet": map[string]interface{}{
				"filter_condition": map[string]interface{}{"by": nil, "where": nil},
			}},
			wantSettings: map[string]interface{}{"bet": map[string]interface{}{
				"filter_condition": map[string]interface{}{"by": "ODDS", "where": nil},
			}},
			wantAdded: true,
		},
		{
			name:         "a group that is null",
			settings:     map[string]interface{}{"privacy": nil},
			defaults:     map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": false}},
			wantSettings: map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": false}},
			wantAdded:    true,
		},
		{
			name:         "a group that is a plain value",
			settings:     map[string]interface{}{"privacy": "yes"},
			defaults:     map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": false}},
			wantSettings: map[string]interface{}{"privacy": map[string]interface{}{"anonymize_logs": false}},
			wantAdded:    true,
		},
		{
			name:         "a value of another type than its default",
			settings:     map[string]interface{}{"emojis": "no", "timezone": nil},
			defaults:     map[string]interface{}{"emojis": true, "timezone": "auto"},
			wantSettings: map[string]interface{}{"emojis": "no", "timezone": nil},
			wantAdded:    false,
		},
		{
			name:         "a key the defaults do not have",
			settings:     map[string]interface{}{"an_old_key": true},
			defaults:     map[string]interface{}{"emojis": true},
			wantSettings: map[string]interface{}{"an_old_key": true, "emojis": true},
			wantAdded:    true,
		},
		{
			name:         "no settings",
			settings:     map[string]interface{}{},
			defaults:     map[string]interface{}{"emojis": true, "bet": map[string]interface{}{"percentage": 5}},
			wantSettings: map[string]interface{}{"emojis": true, "bet": map[string]interface{}{"percentage": 5}},
			wantAdded:    true,
		},
		{
			name:         "no defaults",
			settings:     map[string]interface{}{"username": "me"},
			defaults:     map[string]interface{}{},
			wantSettings: map[string]interface{}{"username": "me"},
			wantAdded:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added := fillMissing(tt.settings, tt.defaults)
			if added != tt.wantAdded {
				t.Errorf("fillMissing() = %v, want %v", added, tt.wantAdded)
			}
			if !reflect.DeepEqual(tt.settings, tt.wantSettings) {
				t.Errorf("fillMissing() left %v, want %v", tt.settings, tt.wantSettings)
			}
		})
	}
}
