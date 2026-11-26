package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestShouldFallbackToUserConfig(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "no error", err: nil, want: false},
		{name: "permission denied", err: fs.ErrPermission, want: true},
		{name: "permission denied opening a file", err: &fs.PathError{Op: "open", Path: "config.json", Err: fs.ErrPermission}, want: true},
		{name: "a read-only file system", err: syscall.EROFS, want: true},
		{name: "a read-only file system opening a file", err: &fs.PathError{Op: "open", Path: "config.json", Err: syscall.EROFS}, want: true},
		{name: "a read-only file system, wrapped", err: fmt.Errorf("write config: %w", syscall.EROFS), want: true},
		{name: "a missing file", err: fs.ErrNotExist, want: false},
		{name: "another error", err: errors.New("disk full"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldFallbackToUserConfig(tt.err); got != tt.want {
				t.Errorf("shouldFallbackToUserConfig(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestLocationChosenByUser(t *testing.T) {
	tests := []struct {
		name        string
		configFlag  string
		dataDirFlag string
		configEnv   string
		dataDirEnv  string
		want        bool
	}{
		{name: "nothing chosen", want: false},
		{name: "a config flag", configFlag: "custom.json", want: true},
		{name: "a data folder flag", dataDirFlag: "data", want: true},
		{name: "TCPM_CONFIG", configEnv: "custom.json", want: true},
		{name: "TCPM_DATA_DIR", dataDirEnv: "data", want: true},
		{name: "everything chosen", configFlag: "custom.json", dataDirFlag: "data", configEnv: "other.json", dataDirEnv: "other", want: true},
		{name: "blank variables", configEnv: "  ", dataDirEnv: "  ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := locationChosenByUser(tt.configFlag, tt.dataDirFlag, tt.configEnv, tt.dataDirEnv); got != tt.want {
				t.Errorf("locationChosenByUser(%q, %q, %q, %q) = %v, want %v", tt.configFlag, tt.dataDirFlag, tt.configEnv, tt.dataDirEnv, got, tt.want)
			}
		})
	}
}

func TestLoadConfigUnder(t *testing.T) {
	t.Run("a base with no folder for the program yet", func(t *testing.T) {
		enterTempFolder(t)
		base := t.TempDir()

		got, ok := loadConfigUnder(base)

		if want := defaultsDecoded(); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("loadConfigUnder() = %+v, %v, want %+v, true", got, ok, want)
		}
		folder := filepath.Join(base, "TwitchChannelPointsMiner")
		if written, want := readFile(t, filepath.Join(folder, "config.json")), formatted(t, defaultConfig()); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
		if cwd := currentFolder(t); !sameFolder(t, cwd, folder) {
			t.Errorf("the current folder is %q, want %q", cwd, folder)
		}
	})

	t.Run("a config already there", func(t *testing.T) {
		enterTempFolder(t)
		base := t.TempDir()
		folder := filepath.Join(base, "TwitchChannelPointsMiner")
		if err := os.Mkdir(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		settings := defaultConfig()
		settings["username"] = "me"
		writeFile(t, filepath.Join(folder, "config.json"), formatted(t, settings))

		got, ok := loadConfigUnder(base)

		want := defaultsDecoded()
		want.Username = "me"
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("loadConfigUnder() = %+v, %v, want %+v, true", got, ok, want)
		}
		if cwd := currentFolder(t); !sameFolder(t, cwd, folder) {
			t.Errorf("the current folder is %q, want %q", cwd, folder)
		}
	})

	t.Run("a config there that is not JSON", func(t *testing.T) {
		enterTempFolder(t)
		base := t.TempDir()
		folder := filepath.Join(base, "TwitchChannelPointsMiner")
		if err := os.Mkdir(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(folder, "config.json"), `{nope`)

		got, ok := loadConfigUnder(base)

		if want := (config{}); ok || !reflect.DeepEqual(got, want) {
			t.Errorf("loadConfigUnder() = %+v, %v, want %+v, false", got, ok, want)
		}
		if after := readFile(t, filepath.Join(folder, "config.json")); after != `{nope` {
			t.Errorf("the file now holds %q, want it untouched", after)
		}
	})

	t.Run("a base that is a file", func(t *testing.T) {
		start := enterTempFolder(t)
		base := filepath.Join(t.TempDir(), "not-a-folder")
		writeFile(t, base, "")

		got, ok := loadConfigUnder(base)

		if want := (config{}); ok || !reflect.DeepEqual(got, want) {
			t.Errorf("loadConfigUnder() = %+v, %v, want %+v, false", got, ok, want)
		}
		if cwd := currentFolder(t); cwd != start {
			t.Errorf("the current folder is %q, want %q", cwd, start)
		}
	})
}
