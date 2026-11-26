package update

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

var downloadName = regexp.MustCompile(`^miner-update-\d+$`)

// checkDownloaded checks that path is a download in dir holding content.
func checkDownloaded(t *testing.T, path, dir, content string) {
	t.Helper()
	if filepath.Dir(path) != dir || !downloadName.MatchString(filepath.Base(path)) {
		t.Errorf("downloaded to %q, want a file matching %s in %q", path, downloadName, dir)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the download: %v", err)
	}
	if string(got) != content {
		t.Errorf("downloaded %q, want %q", got, content)
	}
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("list %s: %v", dir, err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestDownloadAsset(t *testing.T) {
	t.Run("an asset of known size", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = "new binary"
		dir := t.TempDir()
		log := &logLines{}

		path, err := downloadAsset(github.AssetURL("miner"), dir, log.Logf)

		if err != nil {
			t.Fatalf("downloadAsset: %v", err)
		}
		checkDownloaded(t, path, dir, "new binary")
		wantLog := []string{
			"auto-update: downloading update (0.0 MB)...",
			"auto-update: download 100% (0.0/0.0 MB)",
		}
		if !reflect.DeepEqual(log.Lines, wantLog) {
			t.Errorf("logged %q, want %q", log.Lines, wantLog)
		}
		wantRequests := []githubRequest{{Path: "/assets/miner", UserAgent: "TwitchChannelPointsMiner-Updater"}}
		if got := github.Requests(); !reflect.DeepEqual(got, wantRequests) {
			t.Errorf("GitHub received %+v, want %+v", got, wantRequests)
		}
	})

	t.Run("an asset of unknown size", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = "new binary"
		github.Unsized = true
		dir := t.TempDir()
		log := &logLines{}

		path, err := downloadAsset(github.AssetURL("miner"), dir, log.Logf)

		if err != nil {
			t.Fatalf("downloadAsset: %v", err)
		}
		checkDownloaded(t, path, dir, "new binary")
		if log.Lines != nil {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})

	t.Run("an empty asset", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = ""
		dir := t.TempDir()
		log := &logLines{}

		path, err := downloadAsset(github.AssetURL("miner"), dir, log.Logf)

		if err != nil {
			t.Fatalf("downloadAsset: %v", err)
		}
		checkDownloaded(t, path, dir, "")
		if log.Lines != nil {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})

	t.Run("an asset that is gone", func(t *testing.T) {
		github := startFakeGitHub(t)
		dir := t.TempDir()

		path, err := downloadAsset(github.AssetURL("miner"), dir, (&logLines{}).Logf)

		want := "download asset: unexpected status 404 Not Found"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
		if path != "" {
			t.Errorf("path = %q, want none", path)
		}
		if got := filesIn(t, dir); len(got) != 0 {
			t.Errorf("files left behind: %q, want none", got)
		}
	})

	t.Run("GitHub answering 503", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = "new binary"
		github.AssetStatus = http.StatusServiceUnavailable

		_, err := downloadAsset(github.AssetURL("miner"), t.TempDir(), (&logLines{}).Logf)

		want := "download asset: unexpected status 503 Service Unavailable"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})

	t.Run("a dropped connection", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Dropped = true

		_, err := downloadAsset(github.AssetURL("miner"), t.TempDir(), (&logLines{}).Logf)

		shape := regexp.MustCompile(`^Get "http://127\.0\.0\.1:\d+/assets/miner": EOF$`)
		if err == nil || !shape.MatchString(err.Error()) {
			t.Errorf("error = %v, want it to match %s", err, shape)
		}
	})

	t.Run("a download cut short", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = "new b"
		github.CutShort = true
		dir := t.TempDir()

		downloadAsset(github.AssetURL("miner"), dir, (&logLines{}).Logf)

		if got := filesIn(t, dir); len(got) != 0 {
			t.Errorf("files left behind: %q, want none", got)
		}
	})

	t.Run("a folder that does not exist", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Assets["miner"] = "new binary"
		dir := filepath.Join(t.TempDir(), "missing")

		_, err := downloadAsset(github.AssetURL("miner"), dir, (&logLines{}).Logf)

		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want %v", err, fs.ErrNotExist)
		}
	})

	t.Run("an address that is not a URL", func(t *testing.T) {
		_, err := downloadAsset("://", t.TempDir(), (&logLines{}).Logf)

		want := `parse "://": missing protocol scheme`
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})
}
