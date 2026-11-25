package main

import (
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

func loggerSettingsFrom(cfg config) logging.LoggerSettings {
	return logging.LoggerSettings{
		Save:          cfg.SaveLogs,
		Emoji:         cfg.Emojis,
		ShowSeconds:   cfg.ShowSeconds,
		Debug:         cfg.Debug,
		DebugDeep:     cfg.DebugDeep,
		AnonymizeLogs: cfg.Privacy.AnonymizeLogs,
		Discord: notify.DiscordSettings{
			WebhookAPI: cfg.Discord.WebhookAPI,
			Events:     cfg.Discord.Events,
		},
	}
}
