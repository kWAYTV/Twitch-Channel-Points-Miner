package main

import (
	"os"
	"path/filepath"
	"strings"
)

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && !info.IsDir()
}

type appPaths struct {
	WorkDir    string
	ConfigPath string
}

func resolveAppPaths(configFlag, dataDirFlag string) (appPaths, error) {
	if dataDirFlag != "" {
		abs, err := filepath.Abs(dataDirFlag)
		if err != nil {
			return appPaths{}, err
		}
		return appPaths{
			WorkDir:    abs,
			ConfigPath: filepath.Join(abs, "config.json"),
		}, nil
	}

	if configFlag != "" {
		abs, err := filepath.Abs(configFlag)
		if err != nil {
			return appPaths{}, err
		}
		return appPaths{
			WorkDir:    filepath.Dir(abs),
			ConfigPath: abs,
		}, nil
	}

	// Environment overrides (useful for non-interactive launches).
	if raw := strings.TrimSpace(os.Getenv("TCPM_DATA_DIR")); raw != "" {
		abs, err := filepath.Abs(raw)
		if err != nil {
			return appPaths{}, err
		}
		return appPaths{
			WorkDir:    abs,
			ConfigPath: filepath.Join(abs, "config.json"),
		}, nil
	}

	if raw := strings.TrimSpace(os.Getenv("TCPM_CONFIG")); raw != "" {
		abs, err := filepath.Abs(raw)
		if err != nil {
			return appPaths{}, err
		}
		return appPaths{
			WorkDir:    filepath.Dir(abs),
			ConfigPath: abs,
		}, nil
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
		return appPaths{
			WorkDir:    cwd,
			ConfigPath: filepath.Join(cwd, "config.json"),
		}
	}

	if exePath != "" && !isGoRunExecutable(exePath, tempDir) {
		exeDir := filepath.Dir(exePath)
		return appPaths{
			WorkDir:    exeDir,
			ConfigPath: filepath.Join(exeDir, "config.json"),
		}
	}

	// Fallback (dev runs, restricted environments).
	if cwd == "" {
		cwd = "."
	}
	return appPaths{
		WorkDir:    cwd,
		ConfigPath: filepath.Join(cwd, "config.json"),
	}
}
