package main

import (
	"log"
	"os"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/update"
)

// installedUpdate reports whether a newer release replaced this binary, in
// which case the new one has already been started.
func installedUpdate(cfg config) bool {
	if !cfg.AutoUpdate || autoUpdateDisabled(os.Getenv("TCPM_DISABLE_AUTO_UPDATE")) {
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

// autoUpdateDisabled reports whether the value of TCPM_DISABLE_AUTO_UPDATE
// forces the self-updater off (container images must not replace their own binary).
func autoUpdateDisabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
