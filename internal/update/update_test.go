package update

import (
	"errors"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

const linuxAsset = "TwitchChannelPointsMiner-linux-amd64"

type installCall struct {
	ExePath string
	NewPath string
}

// fakeInstaller keeps the installs asked of it and answers with Exit and Err.
type fakeInstaller struct {
	Exit  bool
	Err   error
	Calls []installCall
}

func (i *fakeInstaller) Install(exePath, newPath string) (bool, error) {
	i.Calls = append(i.Calls, installCall{ExePath: exePath, NewPath: newPath})
	return i.Exit, i.Err
}

type fixture struct {
	updater   updater
	github    *fakeGitHub
	installer *fakeInstaller
	log       *logLines
	dir       string
}

// setup builds the updater of a 1.6.0 release for linux/amd64, installed in a
// folder of its own.
func setup(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	f := &fixture{
		github:    startFakeGitHub(t),
		installer: &fakeInstaller{Exit: true},
		log:       &logLines{},
		dir:       dir,
	}
	f.updater = updater{
		version:     "1.6.0",
		exePath:     filepath.Join(dir, "miner"),
		tempDir:     filepath.Join(dir, "tmp"),
		releasesURL: f.github.ReleasesURL(),
		goos:        "linux",
		arch:        "amd64",
		logf:        f.log.Logf,
		install:     f.installer.Install,
	}
	return f
}

func checkNotInstalled(t *testing.T, f *fixture) {
	t.Helper()
	if f.installer.Calls != nil {
		t.Errorf("installed %+v, want nothing", f.installer.Calls)
	}
	if got := filesIn(t, f.dir); len(got) != 0 {
		t.Errorf("files next to the executable: %q, want none", got)
	}
}

func TestRunAutoUpdate(t *testing.T) {
	t.Run("a build from source", func(t *testing.T) {
		if project.Version != project.DevVersion {
			t.Skip("this test binary has a release version stamped in")
		}

		updated, err := RunAutoUpdate()

		if updated || err != nil {
			t.Errorf("RunAutoUpdate() = %t, %v, want false, <nil>", updated, err)
		}
	})
}

func TestUpdater_run(t *testing.T) {
	t.Run("a newer release", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.7.0", fakeAsset{Name: linuxAsset, Content: "new binary"})

		updated, err := f.updater.run()

		if !updated || err != nil {
			t.Fatalf("run() = %t, %v, want true, <nil>", updated, err)
		}
		if len(f.installer.Calls) != 1 || f.installer.Calls[0].ExePath != f.updater.exePath {
			t.Fatalf("installed %+v, want one install over %q", f.installer.Calls, f.updater.exePath)
		}
		checkDownloaded(t, f.installer.Calls[0].NewPath, f.dir, "new binary")
		checkLog(t, f.log, []string{
			"auto-update: found newer version v1.7.0 (asset " + linuxAsset + ")",
			"auto-update: downloading update (0.0 MB)...",
			"auto-update: download 100% (0.0/0.0 MB)",
			"auto-update: download 100% (0.0/0.0 MB)",
		})
	})

	t.Run("the same release", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.6.0", fakeAsset{Name: linuxAsset, Content: "same binary"})

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkNotInstalled(t, f)
		checkLog(t, f.log, nil)
	})

	t.Run("an older release", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.5.9", fakeAsset{Name: linuxAsset, Content: "old binary"})

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkNotInstalled(t, f)
	})

	t.Run("a build from source", func(t *testing.T) {
		f := setup(t)
		f.updater.version = "dev"
		f.github.Publish("v1.7.0", fakeAsset{Name: linuxAsset, Content: "new binary"})

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkNotInstalled(t, f)
		checkLog(t, f.log, nil)
	})

	t.Run("a newer release, started by go run", func(t *testing.T) {
		f := setup(t)
		f.updater.exePath = filepath.Join(f.dir, "go-build123", "miner")
		f.github.Publish("v1.7.0", fakeAsset{Name: linuxAsset, Content: "new binary"})

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkNotInstalled(t, f)
		checkLog(t, f.log, []string{
			"\033[31mauto-update: newer version available (v1.7.0). Rebuild binary to update.\033[0m",
		})
	})

	t.Run("the same release, started by go run", func(t *testing.T) {
		f := setup(t)
		f.updater.exePath = filepath.Join(f.dir, "go-build123", "miner")
		f.github.Publish("v1.6.0", fakeAsset{Name: linuxAsset, Content: "same binary"})

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkLog(t, f.log, nil)
	})

	t.Run("no release published", func(t *testing.T) {
		f := setup(t)

		updated, err := f.updater.run()

		if updated || err != nil {
			t.Errorf("run() = %t, %v, want false, <nil>", updated, err)
		}
		checkNotInstalled(t, f)
	})

	t.Run("GitHub answering 503", func(t *testing.T) {
		f := setup(t)
		f.github.ReleaseStatus = http.StatusServiceUnavailable

		updated, err := f.updater.run()

		want := "fetch release: unexpected status 503 Service Unavailable"
		if updated || err == nil || err.Error() != want {
			t.Errorf("run() = %t, %v, want false, %q", updated, err, want)
		}
		checkNotInstalled(t, f)
	})

	t.Run("a newer release with nothing for this platform", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.7.0", fakeAsset{Name: "TwitchChannelPointsMiner-windows-amd64.exe", Content: "new binary"})

		updated, err := f.updater.run()

		want := "no release asset for linux/amd64"
		if updated || err == nil || err.Error() != want {
			t.Errorf("run() = %t, %v, want false, %q", updated, err, want)
		}
		checkNotInstalled(t, f)
		checkLog(t, f.log, nil)
	})

	t.Run("a newer release whose download fails", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.7.0", fakeAsset{Name: linuxAsset, Content: "new binary"})
		f.github.AssetStatus = http.StatusServiceUnavailable

		updated, err := f.updater.run()

		want := "download update: download asset: unexpected status 503 Service Unavailable"
		if updated || err == nil || err.Error() != want {
			t.Errorf("run() = %t, %v, want false, %q", updated, err, want)
		}
		checkNotInstalled(t, f)
		checkLog(t, f.log, []string{"auto-update: found newer version v1.7.0 (asset " + linuxAsset + ")"})
	})

	t.Run("a newer release whose install fails", func(t *testing.T) {
		f := setup(t)
		f.github.Publish("v1.7.0", fakeAsset{Name: linuxAsset, Content: "new binary"})
		f.installer.Exit = false
		f.installer.Err = errors.New("swap executable: access denied")

		updated, err := f.updater.run()

		if updated || !reflect.DeepEqual(err, f.installer.Err) {
			t.Errorf("run() = %t, %v, want false, %v", updated, err, f.installer.Err)
		}
	})
}

func TestIsGoRunExecutable(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		tempDir string
		want    bool
	}{
		{"an installed binary", "/opt/miner/miner", "/tmp", false},
		{"a binary in a go-build folder", "/home/me/.cache/go-build/ab/miner", "/tmp", true},
		{"a binary in a GO-BUILD folder", `D:\Cache\GO-BUILD123\miner.exe`, `C:\Temp`, true},
		{"a binary under the temp folder", "/tmp/run/miner", "/tmp", true},
		{"a binary under the temp folder in other capitals", `c:\temp\run\miner.exe`, `C:\Temp`, true},
		{"a binary beside the temp folder", "/var/tmp/miner", "/tmp", false},
		{"no temp folder", "/opt/miner/miner", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGoRunExecutable(tt.path, tt.tempDir); got != tt.want {
				t.Errorf("isGoRunExecutable(%q, %q) = %t, want %t", tt.path, tt.tempDir, got, tt.want)
			}
		})
	}
}
