package update

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// installAndRelaunch puts the downloaded file in the executable's place and
// starts it with this process's arguments.
func installAndRelaunch(exePath, newPath string) (bool, error) {
	if runtime.GOOS == "windows" {
		return true, launchWindowsUpdater(exePath, newPath, os.Args[1:])
	}

	if err := replaceExecutable(exePath, newPath); err != nil {
		return false, err
	}
	if err := relaunch(exePath, os.Args[1:]); err != nil {
		return false, err
	}
	return true, nil
}

func replaceExecutable(targetPath, newPath string) error {
	if err := os.Chmod(newPath, 0o755); err != nil && !errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("chmod new executable: %w", err)
	}
	if err := os.Rename(newPath, targetPath); err != nil {
		return fmt.Errorf("swap executable: %w", err)
	}
	return nil
}

func relaunch(targetPath string, args []string) error {
	cmd := exec.Command(targetPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Start()
}

func launchWindowsUpdater(targetPath, newPath string, args []string) error {
	scriptPath := filepath.Join(filepath.Dir(targetPath), fmt.Sprintf("update-%d.bat", time.Now().UnixNano()))
	script := windowsUpdaterScript(targetPath, newPath, filepath.Dir(targetPath), args)

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		return fmt.Errorf("write updater script: %w", err)
	}

	cmd := exec.Command("cmd", "/C", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
