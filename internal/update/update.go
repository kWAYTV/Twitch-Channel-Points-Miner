package update

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

func RunAutoUpdate() (bool, error) {
	exePath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("locate executable: %w", err)
	}
	u := updater{
		version:     project.Version,
		exePath:     exePath,
		tempDir:     os.TempDir(),
		releasesURL: releasesURL,
		goos:        runtime.GOOS,
		arch:        runtime.GOARCH,
		logf:        log.Printf,
		install:     installAndRelaunch,
	}
	return u.run()
}

// updater replaces the executable at exePath with the latest release, when
// that release is newer than version.
type updater struct {
	version     string
	exePath     string
	tempDir     string
	releasesURL string
	goos        string
	arch        string
	logf        func(format string, args ...interface{})
	// install swaps the downloaded file in and starts it, and reports whether
	// the running program should now exit.
	install func(exePath, newPath string) (bool, error)
}

func (u updater) run() (bool, error) {
	// A build with no version stamped in was made from source: it is whatever
	// its builder checked out, and a release must not replace it.
	if u.version == project.DevVersion {
		return false, nil
	}

	devRun := isGoRunExecutable(u.exePath, u.tempDir)

	release, err := fetchLatestRelease(u.releasesURL)
	if errors.Is(err, errNoRelease) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	currentVersion := normalizeVersion(u.version)
	latestVersion := normalizeVersion(release.TagName)
	if compareVersions(latestVersion, currentVersion) <= 0 {
		return false, nil
	}

	if devRun {
		u.logf("\033[31mauto-update: newer version available (%s). Rebuild binary to update.\033[0m", release.TagName)
		return false, nil
	}

	asset, err := pickAsset(release.Assets, u.goos, u.arch)
	if err != nil {
		return false, err
	}

	u.logf("auto-update: found newer version %s (asset %s)", release.TagName, asset.Name)
	tempPath, err := downloadAsset(asset.BrowserDownloadURL, filepath.Dir(u.exePath), u.logf)
	if err != nil {
		return false, fmt.Errorf("download update: %w", err)
	}

	return u.install(u.exePath, tempPath)
}

// isGoRunExecutable reports whether path is a binary "go run" built, which
// lives in a go-build folder or under the temp directory.
func isGoRunExecutable(path, tempDir string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "go-build") {
		return true
	}
	temp := strings.ToLower(tempDir)
	return strings.HasPrefix(lower, temp)
}
