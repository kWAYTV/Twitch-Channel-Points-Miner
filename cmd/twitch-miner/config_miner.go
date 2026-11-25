package main

import (
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/miner"
)

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
