package main

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

func TestChatID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want chatID
	}{
		{name: "a number", json: `123456789`, want: "123456789"},
		{name: "a negative number", json: `-1001234567890`, want: "-1001234567890"},
		{name: "zero", json: `0`, want: "0"},
		{name: "a number with spaces around it", json: ` 42 `, want: "42"},
		{name: "a channel name", json: `"@mychannel"`, want: "@mychannel"},
		{name: "a padded string", json: `" 42 "`, want: "42"},
		{name: "an empty string", json: `""`, want: ""},
		{name: "a name outside ASCII", json: `"@canal_ñ"`, want: "@canal_ñ"},
		{name: "null", json: `null`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chatID("unset")
			if err := got.UnmarshalJSON([]byte(tt.json)); err != nil {
				t.Fatalf("UnmarshalJSON(%s) error = %v, want none", tt.json, err)
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %q, want %q", tt.json, got, tt.want)
			}
		})
	}

	failures := []struct {
		name string
		json string
		want *regexp.Regexp
	}{
		{name: "a boolean", json: `true`, want: regexp.MustCompile(`^json: cannot unmarshal bool into Go value of type json\.Number$`)},
		{name: "an object", json: `{}`, want: regexp.MustCompile(`^json: cannot unmarshal object into Go value of type json\.Number$`)},
		{name: "a string that never ends", json: `"abc`, want: regexp.MustCompile(`^unexpected end of JSON input$`)},
		{name: "nothing", json: ``, want: regexp.MustCompile(`^unexpected end of JSON input$`)},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			got := chatID("unset")
			err := got.UnmarshalJSON([]byte(tt.json))
			if err == nil || !tt.want.MatchString(err.Error()) {
				t.Errorf("UnmarshalJSON(%s) error = %v, want one matching %s", tt.json, err, tt.want)
			}
			if got != "unset" {
				t.Errorf("UnmarshalJSON(%s) left %q, want the value untouched", tt.json, got)
			}
		})
	}
}

func TestTelegramConfig_settings(t *testing.T) {
	t.Run("every field set", func(t *testing.T) {
		cfg := telegramConfig{
			ChatID: "-1001234567890",
			Token:  "123:abc",
			Events: []string{"BET_WIN", "STREAMER_ONLINE"},
			Silent: true,
		}
		want := notify.TelegramSettings{
			ChatID:              "-1001234567890",
			Token:               "123:abc",
			Events:              []string{"BET_WIN", "STREAMER_ONLINE"},
			DisableNotification: true,
		}
		if got := cfg.settings(); !reflect.DeepEqual(got, want) {
			t.Errorf("settings() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing set", func(t *testing.T) {
		cfg := telegramConfig{}
		want := notify.TelegramSettings{}
		if got := cfg.settings(); !reflect.DeepEqual(got, want) {
			t.Errorf("settings() = %+v, want %+v", got, want)
		}
	})
}
