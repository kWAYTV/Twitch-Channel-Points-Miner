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
		ShowUsername:  cfg.ShowUsername,
		Debug:         cfg.Debug != debugOff,
		DebugDeep:     cfg.Debug == debugFull,
		AnonymizeLogs: cfg.AnonymizeLogs,
		Discord: notify.DiscordSettings{
			WebhookAPI: cfg.Discord.Webhook,
			Events:     cfg.Discord.Events,
		},
	}
}
