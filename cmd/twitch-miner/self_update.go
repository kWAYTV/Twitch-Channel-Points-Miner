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

// autoUpdateDisabledByEnv reports whether TCPM_DISABLE_AUTO_UPDATE forces the
// self-updater off (container images must not replace their own binary).
func autoUpdateDisabledByEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TCPM_DISABLE_AUTO_UPDATE"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
