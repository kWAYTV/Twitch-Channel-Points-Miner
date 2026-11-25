package main

import (
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

func applyTimezoneOverride(raw *string, logger *logging.Logger) {
	if raw == nil {
		return
	}
	zone := strings.TrimSpace(*raw)
	if zone == "" || strings.EqualFold(zone, "auto") {
		return
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		logger.Errorf("%sTimezone override ignored; falling back to system time: %v%s", ansi.ColorRed, err, ansi.ColorReset)
	}
	time.Local = loc
}
