package main

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/miner"
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
