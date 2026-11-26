package update

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func TestReplaceExecutable(t *testing.T) {
	t.Run("a new file over the old one", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "miner")
		downloaded := filepath.Join(dir, "miner-update-1")
		writeFile(t, target, "old binary")
		writeFile(t, downloaded, "new binary")

		err := replaceExecutable(target, downloaded)

		if err != nil {
			t.Fatalf("replaceExecutable: %v", err)
		}
		if got := readFile(t, target); got != "new binary" {
			t.Errorf("executable holds %q, want %q", got, "new binary")
		}
		if got, want := filesIn(t, dir), []string{"miner"}; !reflect.DeepEqual(got, want) {
			t.Errorf("folder holds %q, want %q", got, want)
		}
	})

	t.Run("no old file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "miner")
		downloaded := filepath.Join(dir, "miner-update-1")
		writeFile(t, downloaded, "new binary")

		err := replaceExecutable(target, downloaded)

		if err != nil {
			t.Fatalf("replaceExecutable: %v", err)
		}
		if got := readFile(t, target); got != "new binary" {
			t.Errorf("executable holds %q, want %q", got, "new binary")
		}
	})

	t.Run("no new file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "miner")
		writeFile(t, target, "old binary")

		err := replaceExecutable(target, filepath.Join(dir, "miner-update-1"))

		if err == nil || !strings.HasPrefix(err.Error(), "chmod new executable: ") || !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want \"chmod new executable: \" wrapping %v", err, fs.ErrNotExist)
		}
		if got := readFile(t, target); got != "old binary" {
			t.Errorf("executable holds %q, want %q", got, "old binary")
		}
	})

	t.Run("a target folder that does not exist", func(t *testing.T) {
		dir := t.TempDir()
		downloaded := filepath.Join(dir, "miner-update-1")
		writeFile(t, downloaded, "new binary")

		err := replaceExecutable(filepath.Join(dir, "missing", "miner"), downloaded)

		if err == nil || !strings.HasPrefix(err.Error(), "swap executable: ") || !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want \"swap executable: \" wrapping %v", err, fs.ErrNotExist)
		}
		if got := readFile(t, downloaded); got != "new binary" {
			t.Errorf("download holds %q, want %q", got, "new binary")
		}
	})
}
