package main

import (
	"log"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/miner"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
	"github.com/kWAYTV/twitch-points-miner/internal/update"
)

func main() {
	cfg := loadConfig()
	if installedUpdate(cfg) {
		return
	}

	loggerSettings := loggerSettingsFrom(cfg)
	logger := logging.NewLogger(loggerSettings, cfg.Username)
	applyTimezoneOverride(cfg.Timezone, logger)

	session := miner.New(minerOptionsFrom(cfg, loggerSettings))
	startAnalytics(cfg.Analytics, logger)
	startTelegram(cfg.Telegram)

	if len(cfg.Streamers) > 0 {
		session.Mine(cfg.Streamers)
		return
	}
	session.MineFollowers(domain.FollowersOrderDESC)
}

// installedUpdate reports whether a newer release replaced this binary, in
// which case the new one has already been started.
func installedUpdate(cfg config) bool {
	if !cfg.AutoUpdate || autoUpdateDisabledByEnv() {
		return false
	}
	updated, err := update.RunAutoUpdate()
	if err != nil {
		log.Printf("auto-update failed: %v", err)
	}
	if updated {
		log.Printf("auto-update installed a newer version; restarting...")
	}
	return updated
}

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

func minerOptionsFrom(cfg config, loggerSettings logging.LoggerSettings) miner.Options {
	streamerSettings := buildBaseStreamerSettings(cfg)
	return miner.Options{
		Username:                   cfg.Username,
		Password:                   cfg.Password,
		ClaimDropsStartup:          cfg.ClaimDropsStartup,
		DisableSSLCertVerification: cfg.DisableSSLCertVerification,
		Logger:                     loggerSettings,
		StreamerSettings:           streamerSettings,
		StreamerOverrides:          buildOverrideSettings(streamerSettings, cfg.StreamerOverrides),
		WatchPriority:              cfg.WatchPriority,
		StreamersExclude:           cfg.StreamersExclude,
		GamePriority:               cfg.GamePriority,
		GameExclude:                cfg.GameExclude,
		DisableAtInNickname:        cfg.DisableAtInNickname,
		ShowGame:                   cfg.ShowGame,
		WatchQueueLogging:          cfg.WatchQueueLogging,
		WatchStreakWarmStartCache:  cfg.WatchStreakWarmStartCache,
		ShowDropsProgress:          cfg.ShowDropsProgress,
	}
}
