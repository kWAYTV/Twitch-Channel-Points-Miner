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
	startDashboard(cfg.Dashboard, logger)
	startTelegram(cfg.Telegram)
	startMatrix(cfg.Matrix)

	if len(cfg.Streamers) > 0 {
		session.Mine(cfg.Streamers)
		return
	}
	session.MineFollowers(domain.FollowersOrderDESC)
}
