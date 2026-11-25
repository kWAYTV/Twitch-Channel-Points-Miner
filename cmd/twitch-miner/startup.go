package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

// userConfigFolder is where the config goes when the usual place is not writable.
const userConfigFolder = "TwitchChannelPointsMiner"

// loadConfig finds the data folder, moves into it and reads config.json,
// creating it on the first run. It ends the program when no config can be loaded.
func loadConfig() config {
	configFlag := flag.String("config", "", "Path to config.json (default: ./config.json or next to the executable)")
	dataDirFlag := flag.String("data-dir", "", "Directory for config/cookies/log (default: current directory if config.json exists; otherwise the executable directory)")
	flag.Parse()

	paths, err := resolveAppPaths(*configFlag, *dataDirFlag)
	if err != nil {
		log.Fatalf("failed to resolve config paths: %v", err)
	}
	enterWorkDir(paths.WorkDir)

	setConsoleTitle(project.Name)
	clearConsole()

	cfg, err := loadOrCreateConfig(paths.ConfigPath)
	if err != nil && !locationChosenByUser(*configFlag, *dataDirFlag) && shouldFallbackToUserConfig(err) {
		if fallback, ok := loadConfigFromUserFolder(); ok {
			return fallback
		}
	}
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}

func enterWorkDir(dir string) {
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	if err := os.Chdir(dir); err != nil {
		log.Printf("warning: failed to change working directory to %q: %v", dir, err)
	}
}

// locationChosenByUser reports whether a flag or environment variable says where
// the config lives, in which case no other place is tried.
func locationChosenByUser(configFlag, dataDirFlag string) bool {
	return configFlag != "" || dataDirFlag != "" ||
		strings.TrimSpace(os.Getenv("TCPM_CONFIG")) != "" ||
		strings.TrimSpace(os.Getenv("TCPM_DATA_DIR")) != ""
}

// loadConfigFromUserFolder moves to the per-user config folder and loads the
// config there.
func loadConfigFromUserFolder() (config, bool) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return config{}, false
	}
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
