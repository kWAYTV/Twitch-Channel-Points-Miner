package update

import (
	"errors"
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

func TestFetchLatestRelease(t *testing.T) {
	t.Run("a release with two assets", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Publish("v1.7.0", fakeAsset{Name: "miner-linux"}, fakeAsset{Name: "miner-windows.exe"})

		release, err := fetchLatestRelease(github.ReleasesURL())

		if err != nil {
			t.Fatalf("fetchLatestRelease: %v", err)
		}
		want := githubRelease{TagName: "v1.7.0", Assets: []releaseAsset{
			{Name: "miner-linux", BrowserDownloadURL: github.URL + "/assets/miner-linux"},
			{Name: "miner-windows.exe", BrowserDownloadURL: github.URL + "/assets/miner-windows.exe"},
		}}
		if !reflect.DeepEqual(release, want) {
			t.Errorf("release = %+v, want %+v", release, want)
		}
		wantRequests := []githubRequest{{
			Path:      "/releases/latest",
			Accept:    "application/vnd.github+json",
			UserAgent: "TwitchChannelPointsMiner-Updater",
		}}
		if got := github.Requests(); !reflect.DeepEqual(got, wantRequests) {
			t.Errorf("GitHub received %+v, want %+v", got, wantRequests)
		}
	})

	t.Run("a release with no assets", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Publish("v1.7.0")

		release, err := fetchLatestRelease(github.ReleasesURL())

		if err != nil {
			t.Fatalf("fetchLatestRelease: %v", err)
		}
		want := githubRelease{TagName: "v1.7.0", Assets: []releaseAsset{}}
		if !reflect.DeepEqual(release, want) {
			t.Errorf("release = %+v, want %+v", release, want)
		}
	})

	t.Run("no release published", func(t *testing.T) {
		github := startFakeGitHub(t)

		release, err := fetchLatestRelease(github.ReleasesURL())

		if !errors.Is(err, errNoRelease) {
			t.Errorf("error = %v, want %v", err, errNoRelease)
		}
		if !reflect.DeepEqual(release, githubRelease{}) {
			t.Errorf("release = %+v, want none", release)
		}
	})

	t.Run("GitHub answering 503", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.ReleaseStatus = http.StatusServiceUnavailable

		_, err := fetchLatestRelease(github.ReleasesURL())

		want := "fetch release: unexpected status 503 Service Unavailable"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})

	t.Run("an answer that is not JSON", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.ReleaseStatus = http.StatusOK
		github.ReleaseBody = "<html>"

		_, err := fetchLatestRelease(github.ReleasesURL())

		want := "parse release: invalid character '<' looking for beginning of value"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})

	t.Run("a dropped connection", func(t *testing.T) {
		github := startFakeGitHub(t)
		github.Dropped = true

		_, err := fetchLatestRelease(github.ReleasesURL())

		shape := regexp.MustCompile(`^fetch release: Get "http://127\.0\.0\.1:\d+/releases/latest": EOF$`)
		if err == nil || !shape.MatchString(err.Error()) {
			t.Errorf("error = %v, want it to match %s", err, shape)
		}
	})

	t.Run("an address that is not a URL", func(t *testing.T) {
		_, err := fetchLatestRelease("://")

		want := `parse "://": missing protocol scheme`
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})
}

func TestPickAsset(t *testing.T) {
	linux := releaseAsset{Name: "TwitchChannelPointsMiner-linux-amd64", BrowserDownloadURL: "https://example.test/linux"}
	linuxArm := releaseAsset{Name: "TwitchChannelPointsMiner-linux-arm64", BrowserDownloadURL: "https://example.test/linux-arm"}
	windows := releaseAsset{Name: "TwitchChannelPointsMiner-windows-amd64.exe", BrowserDownloadURL: "https://example.test/windows"}
	windowsBare := releaseAsset{Name: "TwitchChannelPointsMiner-windows-amd64", BrowserDownloadURL: "https://example.test/windows-bare"}
	darwin := releaseAsset{Name: "TwitchChannelPointsMiner-darwin-arm64", BrowserDownloadURL: "https://example.test/darwin"}
	macos := releaseAsset{Name: "TwitchChannelPointsMiner-macos-arm64", BrowserDownloadURL: "https://example.test/macos"}
	osx := releaseAsset{Name: "TwitchChannelPointsMiner-osx-arm64", BrowserDownloadURL: "https://example.test/osx"}
	shouting := releaseAsset{Name: "TWITCHCHANNELPOINTSMINER-LINUX-AMD64", BrowserDownloadURL: "https://example.test/shouting"}
	checksums := releaseAsset{Name: "checksums.txt", BrowserDownloadURL: "https://example.test/checksums"}

	tests := []struct {
		name       string
		assets     []releaseAsset
		goos, arch string
		want       releaseAsset
		wantErr    string
	}{
		{"linux among other platforms", []releaseAsset{checksums, windows, linuxArm, linux}, "linux", "amd64", linux, ""},
		{"linux on another architecture", []releaseAsset{linux, linuxArm}, "linux", "arm64", linuxArm, ""},
		{"windows with the .exe", []releaseAsset{linux, windows}, "windows", "amd64", windows, ""},
		{"windows without the .exe", []releaseAsset{linux, windowsBare}, "windows", "amd64", releaseAsset{}, "no release asset for windows/amd64"},
		{"mac named darwin", []releaseAsset{linux, darwin}, "darwin", "arm64", darwin, ""},
		{"mac named macos", []releaseAsset{linux, macos}, "darwin", "arm64", macos, ""},
		{"mac named osx", []releaseAsset{linux, osx}, "darwin", "arm64", osx, ""},
		{"mac under two names", []releaseAsset{osx, darwin}, "darwin", "arm64", osx, ""},
		{"a name in capitals", []releaseAsset{shouting}, "linux", "amd64", shouting, ""},
		{"no asset for the platform", []releaseAsset{checksums, windows}, "linux", "amd64", releaseAsset{}, "no release asset for linux/amd64"},
		{"no assets", nil, "linux", "amd64", releaseAsset{}, "no release asset for linux/amd64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pickAsset(tt.assets, tt.goos, tt.arch)

			if got != tt.want {
				t.Errorf("pickAsset = %+v, want %+v", got, tt.want)
			}
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}

func TestIsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"199, just below", 199, false},
		{"200, the first", 200, true},
		{"204, in the middle", 204, true},
		{"299, the last", 299, true},
		{"300, just above", 300, false},
		{"404", 404, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSuccess(tt.statusCode); got != tt.want {
				t.Errorf("isSuccess(%d) = %t, want %t", tt.statusCode, got, tt.want)
			}
		})
	}
}
