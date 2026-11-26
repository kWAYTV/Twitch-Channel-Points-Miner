package main

import (
	"os"
	"path/filepath"
	"testing"
)

// enterTempFolder moves the test into a new empty folder, puts it back where
// it was when the test ends, and returns the folder as the system names it.
func enterTempFolder(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatal(err)
		}
	})
	return currentFolder(t)
}

func currentFolder(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}

func TestFileExists(t *testing.T) {
	t.Run("a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, "{}")
		if !fileExists(path) {
			t.Errorf("fileExists(%q) = false, want true", path)
		}
	})

	t.Run("an empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, "")
		if !fileExists(path) {
			t.Errorf("fileExists(%q) = false, want true", path)
		}
	})

	t.Run("a folder", func(t *testing.T) {
		path := t.TempDir()
		if fileExists(path) {
			t.Errorf("fileExists(%q) = true, want false", path)
		}
	})

	t.Run("a missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		if fileExists(path) {
			t.Errorf("fileExists(%q) = true, want false", path)
		}
	})

	t.Run("no path", func(t *testing.T) {
		if fileExists("") {
			t.Errorf(`fileExists("") = true, want false`)
		}
	})
}

func TestDefaultAppPaths(t *testing.T) {
	p := filepath.FromSlash
	tests := []struct {
		name         string
		cwd          string
		cwdHasConfig bool
		exePath      string
		want         appPaths
	}{
		{
			name: "a config in the current folder",
			cwd:  p("/work"), cwdHasConfig: true, exePath: p("/opt/miner/twitch-miner"),
			want: appPaths{WorkDir: p("/work"), ConfigPath: p("/work/config.json")},
		},
		{
			name: "no config in the current folder",
			cwd:  p("/work"), cwdHasConfig: false, exePath: p("/opt/miner/twitch-miner"),
			want: appPaths{WorkDir: p("/opt/miner"), ConfigPath: p("/opt/miner/config.json")},
		},
		{
			name: "no config in the current folder and a dev run",
			cwd:  p("/work"), cwdHasConfig: false, exePath: p("/tmp/go-build1/exe/twitch-miner"),
			want: appPaths{WorkDir: p("/work"), ConfigPath: p("/work/config.json")},
		},
		{
			name: "no config in the current folder and the program's path unknown",
			cwd:  p("/work"), cwdHasConfig: false, exePath: "",
			want: appPaths{WorkDir: p("/work"), ConfigPath: p("/work/config.json")},
		},
		{
			name: "the current folder unknown",
			cwd:  "", cwdHasConfig: false, exePath: p("/opt/miner/twitch-miner"),
			want: appPaths{WorkDir: p("/opt/miner"), ConfigPath: p("/opt/miner/config.json")},
		},
		{
			name: "the current folder unknown and a dev run",
			cwd:  "", cwdHasConfig: false, exePath: p("/tmp/go-build1/exe/twitch-miner"),
			want: appPaths{WorkDir: ".", ConfigPath: "config.json"},
		},
		{
			name: "nothing known",
			cwd:  "", cwdHasConfig: false, exePath: "",
			want: appPaths{WorkDir: ".", ConfigPath: "config.json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultAppPaths(tt.cwd, tt.cwdHasConfig, tt.exePath, p("/tmp")); got != tt.want {
				t.Errorf("defaultAppPaths() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// setupAppPaths gives the test an empty current folder and no TCPM_ variables.
func setupAppPaths(t *testing.T) (cwd string) {
	t.Helper()
	t.Setenv("TCPM_DATA_DIR", "")
	t.Setenv("TCPM_CONFIG", "")
	return enterTempFolder(t)
}

func TestResolveAppPaths(t *testing.T) {
	t.Run("a data folder flag", func(t *testing.T) {
		setupAppPaths(t)
		data := t.TempDir()

		got, err := resolveAppPaths("", data)

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: data, ConfigPath: filepath.Join(data, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("a data folder flag relative to the current folder", func(t *testing.T) {
		cwd := setupAppPaths(t)

		got, err := resolveAppPaths("", "data")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: filepath.Join(cwd, "data"), ConfigPath: filepath.Join(cwd, "data", "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("a config flag", func(t *testing.T) {
		setupAppPaths(t)
		folder := t.TempDir()

		got, err := resolveAppPaths(filepath.Join(folder, "custom.json"), "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: folder, ConfigPath: filepath.Join(folder, "custom.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("a config flag relative to the current folder", func(t *testing.T) {
		cwd := setupAppPaths(t)

		got, err := resolveAppPaths(filepath.Join("conf", "custom.json"), "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: filepath.Join(cwd, "conf"), ConfigPath: filepath.Join(cwd, "conf", "custom.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("both flags", func(t *testing.T) {
		setupAppPaths(t)
		data := t.TempDir()
		folder := t.TempDir()

		got, err := resolveAppPaths(filepath.Join(folder, "custom.json"), data)

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: data, ConfigPath: filepath.Join(data, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("TCPM_DATA_DIR", func(t *testing.T) {
		setupAppPaths(t)
		data := t.TempDir()
		t.Setenv("TCPM_DATA_DIR", data)

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: data, ConfigPath: filepath.Join(data, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("TCPM_DATA_DIR with spaces around it", func(t *testing.T) {
		setupAppPaths(t)
		data := t.TempDir()
		t.Setenv("TCPM_DATA_DIR", "  "+data+"  ")

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: data, ConfigPath: filepath.Join(data, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("a config flag and TCPM_DATA_DIR", func(t *testing.T) {
		setupAppPaths(t)
		folder := t.TempDir()
		t.Setenv("TCPM_DATA_DIR", t.TempDir())

		got, err := resolveAppPaths(filepath.Join(folder, "custom.json"), "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: folder, ConfigPath: filepath.Join(folder, "custom.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("TCPM_CONFIG", func(t *testing.T) {
		setupAppPaths(t)
		folder := t.TempDir()
		t.Setenv("TCPM_CONFIG", filepath.Join(folder, "custom.json"))

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: folder, ConfigPath: filepath.Join(folder, "custom.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("both variables", func(t *testing.T) {
		setupAppPaths(t)
		data := t.TempDir()
		t.Setenv("TCPM_DATA_DIR", data)
		t.Setenv("TCPM_CONFIG", filepath.Join(t.TempDir(), "custom.json"))

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: data, ConfigPath: filepath.Join(data, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	t.Run("blank variables", func(t *testing.T) {
		cwd := setupAppPaths(t)
		t.Setenv("TCPM_DATA_DIR", "   ")
		t.Setenv("TCPM_CONFIG", "   ")

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: cwd, ConfigPath: filepath.Join(cwd, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})

	// The test binary is a dev build, so with nothing chosen the current
	// folder is used whether or not it holds a config.
	t.Run("nothing chosen", func(t *testing.T) {
		cwd := setupAppPaths(t)

		got, err := resolveAppPaths("", "")

		if err != nil {
			t.Fatalf("resolveAppPaths() error = %v, want none", err)
		}
		if want := (appPaths{WorkDir: cwd, ConfigPath: filepath.Join(cwd, "config.json")}); got != want {
			t.Errorf("resolveAppPaths() = %+v, want %+v", got, want)
		}
	})
}
