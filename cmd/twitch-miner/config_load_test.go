package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func formatted(t *testing.T, settings map[string]interface{}) string {
	t.Helper()
	raw, err := formatSettings(settings)
	if err != nil {
		t.Fatalf("formatSettings() error = %v", err)
	}
	return string(raw)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestLoadOrCreateConfig(t *testing.T) {
	t.Run("a missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		if want := defaultsDecoded(); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if written, want := readFile(t, path), formatted(t, defaultConfig()); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
	})

	t.Run("a missing file in a missing folder", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "data", "miner", "config.json")

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		if want := defaultsDecoded(); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if written, want := readFile(t, path), formatted(t, defaultConfig()); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
	})

	t.Run("a file with keys missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, `{"username": "me", "bet": {"percentage": 9}}`)

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		want := defaultsDecoded()
		want.Username = "me"
		want.Bet.Percentage = ptr(9)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		wantSettings := defaultConfig()
		wantSettings["username"] = "me"
		wantSettings["bet"].(map[string]interface{})["percentage"] = 9
		if written, want := readFile(t, path), formatted(t, wantSettings); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
	})

	t.Run("a complete file in another layout", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		settings := defaultConfig()
		settings["username"] = "me"
		alphabetical, err := json.MarshalIndent(settings, "", "    ")
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, path, string(alphabetical))

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		want := defaultsDecoded()
		want.Username = "me"
		if !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if written, want := readFile(t, path), formatted(t, settings); written != want {
			t.Errorf("wrote\n%s\nwant\n%s", written, want)
		}
	})

	t.Run("a complete file in the program's layout with Windows line endings", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		onDisk := strings.ReplaceAll(formatted(t, defaultConfig()), "\n", "\r\n")
		writeFile(t, path, onDisk)

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		if want := defaultsDecoded(); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if after := readFile(t, path); after != onDisk {
			t.Errorf("the file was written again:\n%q", after)
		}
	})

	t.Run("the same file a second time", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		if _, err := loadOrCreateConfig(path); err != nil {
			t.Fatal(err)
		}

		got, err := loadOrCreateConfig(path)

		if err != nil {
			t.Fatalf("loadOrCreateConfig() error = %v, want none", err)
		}
		if want := defaultsDecoded(); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if written, want := readFile(t, path), formatted(t, defaultConfig()); written != want {
			t.Errorf("the file now holds\n%s\nwant\n%s", written, want)
		}
	})

	t.Run("a file that is not JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, path, `{nope`)

		got, err := loadOrCreateConfig(path)

		if err == nil || err.Error() != "invalid config: invalid character 'n' looking for beginning of object key string" {
			t.Errorf("loadOrCreateConfig() error = %v, want invalid config: invalid character 'n' looking for beginning of object key string", err)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if after := readFile(t, path); after != `{nope` {
			t.Errorf("the file now holds %q, want it untouched", after)
		}
	})

	t.Run("a complete file with a number where text belongs", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		settings := defaultConfig()
		settings["username"] = 5
		onDisk := formatted(t, settings)
		writeFile(t, path, onDisk)

		got, err := loadOrCreateConfig(path)

		wantErr := regexp.MustCompile(`^json: cannot unmarshal number into Go struct field .*username of type string$`)
		if err == nil || !wantErr.MatchString(err.Error()) {
			t.Errorf("loadOrCreateConfig() error = %v, want one matching %s", err, wantErr)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if after := readFile(t, path); after != onDisk {
			t.Errorf("the file now holds\n%s\nwant it untouched", after)
		}
	})

	t.Run("a path that cannot be written, being a folder", func(t *testing.T) {
		path := t.TempDir()

		got, err := loadOrCreateConfig(path)

		if err == nil || !strings.Contains(err.Error(), path) {
			t.Errorf("loadOrCreateConfig() error = %v, want a failure that names %s", err, path)
		}
		if want := (config{}); !reflect.DeepEqual(got, want) {
			t.Errorf("loadOrCreateConfig() = %+v, want %+v", got, want)
		}
		if info, statErr := os.Stat(path); statErr != nil || !info.IsDir() {
			t.Errorf("the folder is gone or no longer a folder: %v", statErr)
		}
		leftBehind, _ := filepath.Glob(path + "*.tmp")
		if len(leftBehind) != 0 {
			t.Errorf("files left behind: %v, want none", leftBehind)
		}
	})
}
