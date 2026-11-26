package main

import (
	"flag"
	"log"
	"os"

	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

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
	chosen := locationChosenByUser(*configFlag, *dataDirFlag, os.Getenv("TCPM_CONFIG"), os.Getenv("TCPM_DATA_DIR"))
	if err != nil && !chosen && shouldFallbackToUserConfig(err) {
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
