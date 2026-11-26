package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// userConfigFolder is where the config goes when the usual place is not writable.
const userConfigFolder = "TwitchChannelPointsMiner"

func shouldFallbackToUserConfig(err error) bool {
	if err == nil {
		return false
	}
	if os.IsPermission(err) {
		return true
	}
	return errors.Is(err, syscall.EROFS)
}

// locationChosenByUser reports whether a flag or environment variable
// (TCPM_CONFIG, TCPM_DATA_DIR) says where the config lives, in which case no
// other place is tried.
func locationChosenByUser(configFlag, dataDirFlag, configEnv, dataDirEnv string) bool {
	return configFlag != "" || dataDirFlag != "" ||
		strings.TrimSpace(configEnv) != "" ||
		strings.TrimSpace(dataDirEnv) != ""
}

// loadConfigFromUserFolder moves to the per-user config folder and loads the
// config there.
func loadConfigFromUserFolder() (config, bool) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return config{}, false
	}
	return loadConfigUnder(base)
}

// loadConfigUnder moves to the program's folder under base, creating it when
// needed, and loads the config there.
func loadConfigUnder(base string) (config, bool) {
	dir := filepath.Join(base, userConfigFolder)
	_ = os.MkdirAll(dir, 0o755)
	if err := os.Chdir(dir); err != nil {
		return config{}, false
	}
	cfg, err := loadOrCreateConfig(filepath.Join(dir, "config.json"))
	if err != nil {
		return config{}, false
	}
	log.Printf("using config directory %q", dir)
	return cfg, true
}
