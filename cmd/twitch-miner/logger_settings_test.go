package main

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

func TestLoggerSettingsFrom(t *testing.T) {
	t.Run("every setting set", func(t *testing.T) {
		cfg := config{
			SaveLogs:      true,
			Emojis:        true,
			ShowSeconds:   true,
			Debug:         debugFull,
			AnonymizeLogs: true,
			Discord:       discordConfig{Webhook: "https://discord.com/api/webhooks/1/a", Events: []string{"BET_WIN"}},
		}
		want := logging.LoggerSettings{
			Save:          true,
			Emoji:         true,
			ShowSeconds:   true,
			Debug:         true,
			DebugDeep:     true,
			AnonymizeLogs: true,
			Discord:       notify.DiscordSettings{WebhookAPI: "https://discord.com/api/webhooks/1/a", Events: []string{"BET_WIN"}},
		}
		if got := loggerSettingsFrom(cfg); !reflect.DeepEqual(got, want) {
			t.Errorf("loggerSettingsFrom() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing set", func(t *testing.T) {
		want := logging.LoggerSettings{}
		if got := loggerSettingsFrom(config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("loggerSettingsFrom() = %+v, want %+v", got, want)
		}
	})

	// Each row turns on one switch and expects only its counterpart on.
	oneSwitch := []struct {
		name string
		cfg  config
		want logging.LoggerSettings
	}{
		{name: "only saving logs", cfg: config{SaveLogs: true}, want: logging.LoggerSettings{Save: true}},
		{name: "only emojis", cfg: config{Emojis: true}, want: logging.LoggerSettings{Emoji: true}},
		{name: "only seconds", cfg: config{ShowSeconds: true}, want: logging.LoggerSettings{ShowSeconds: true}},
		{name: "only debug", cfg: config{Debug: debugSummary}, want: logging.LoggerSettings{Debug: true}},
		{name: "only full debug", cfg: config{Debug: debugFull}, want: logging.LoggerSettings{Debug: true, DebugDeep: true}},
		{name: "only anonymous logs", cfg: config{AnonymizeLogs: true}, want: logging.LoggerSettings{AnonymizeLogs: true}},
	}
	for _, tt := range oneSwitch {
		t.Run(tt.name, func(t *testing.T) {
			if got := loggerSettingsFrom(tt.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loggerSettingsFrom() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
