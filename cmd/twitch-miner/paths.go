package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/devrun"
)

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && !info.IsDir()
}

type appPaths struct {
	WorkDir    string
	ConfigPath string
}

// pathsInFolder keeps the config in dir, under its usual name.
func pathsInFolder(dir string) appPaths {
	return appPaths{
		WorkDir:    dir,
		ConfigPath: filepath.Join(dir, "config.json"),
	}
}

func pathsForDataDir(dir string) (appPaths, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return appPaths{}, err
	}
	return pathsInFolder(abs), nil
}

// pathsForConfigFile works in the folder the config file is in.
func pathsForConfigFile(path string) (appPaths, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return appPaths{}, err
	}
	return appPaths{
		WorkDir:    filepath.Dir(abs),
		ConfigPath: abs,
	}, nil
}

func resolveAppPaths(configFlag, dataDirFlag string) (appPaths, error) {
	if dataDirFlag != "" {
		return pathsForDataDir(dataDirFlag)
	}
	if configFlag != "" {
		return pathsForConfigFile(configFlag)
	}

	// Environment overrides (useful for non-interactive launches).
	if raw := strings.TrimSpace(os.Getenv("TCPM_DATA_DIR")); raw != "" {
		return pathsForDataDir(raw)
	}
	if raw := strings.TrimSpace(os.Getenv("TCPM_CONFIG")); raw != "" {
		return pathsForConfigFile(raw)
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	cwdHasConfig := cwd != "" && fileExists(filepath.Join(cwd, "config.json"))
	exePath, err := os.Executable()
	if err != nil {
		exePath = ""
	}
	return defaultAppPaths(cwd, cwdHasConfig, exePath, os.TempDir()), nil
}

// defaultAppPaths picks the data folder when no flag or variable names one.
// cwd and exePath are empty when the system could not tell them.
func defaultAppPaths(cwd string, cwdHasConfig bool, exePath, tempDir string) appPaths {
	// Preserve the historical behavior when the user runs from a folder that already has config.json.
	if cwd != "" && cwdHasConfig {
		return pathsInFolder(cwd)
	}

	if exePath != "" && !devrun.IsGoRunExecutable(exePath, tempDir) {
		return pathsInFolder(filepath.Dir(exePath))
	}

	// Fallback (dev runs, restricted environments).
	if cwd == "" {
		cwd = "."
	}
	return pathsInFolder(cwd)
}
