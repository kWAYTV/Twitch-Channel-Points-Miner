package main

import (
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

// requestedTimezone returns the zone name the config asks for, or "" when it
// leaves the system's time zone in place (unset, blank or "auto").
func requestedTimezone(raw *string) string {
	if raw == nil {
		return ""
	}
	zone := strings.TrimSpace(*raw)
	if zone == "" || strings.EqualFold(zone, "auto") {
		return ""
	}
	return zone
}

func applyTimezoneOverride(raw *string, logger *logging.Logger) {
	zone := requestedTimezone(raw)
	if zone == "" {
		return
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		logger.Errorf("%sTimezone override ignored; falling back to system time: %v%s", ansi.ColorRed, err, ansi.ColorReset)
	}
	time.Local = loc
}
