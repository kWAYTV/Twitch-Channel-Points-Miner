package main

import (
	"os"
	"path/filepath"
	"testing"
)

// sameFolder reports whether two paths name one folder, however each is spelled.
func sameFolder(t *testing.T, a, b string) bool {
	t.Helper()
	infoA, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	infoB, err := os.Stat(b)
	if err != nil {
		t.Fatal(err)
	}
	return os.SameFile(infoA, infoB)
}

func TestEnterWorkDir(t *testing.T) {
	t.Run("a folder that exists", func(t *testing.T) {
		enterTempFolder(t)
		dir := t.TempDir()

		enterWorkDir(dir)

		if cwd := currentFolder(t); !sameFolder(t, cwd, dir) {
			t.Errorf("the current folder is %q, want %q", cwd, dir)
		}
	})

	t.Run("a folder that does not exist yet", func(t *testing.T) {
		enterTempFolder(t)
		dir := filepath.Join(t.TempDir(), "data", "miner")

		enterWorkDir(dir)

		if cwd := currentFolder(t); !sameFolder(t, cwd, dir) {
			t.Errorf("the current folder is %q, want %q", cwd, dir)
		}
	})

	t.Run("no folder", func(t *testing.T) {
		start := enterTempFolder(t)

		enterWorkDir("")

		if cwd := currentFolder(t); cwd != start {
			t.Errorf("the current folder is %q, want %q", cwd, start)
		}
	})

	t.Run("a path that is a file", func(t *testing.T) {
		start := enterTempFolder(t)
		file := filepath.Join(t.TempDir(), "config.json")
		writeFile(t, file, "{}")

		enterWorkDir(file)

		if cwd := currentFolder(t); cwd != start {
			t.Errorf("the current folder is %q, want %q", cwd, start)
		}
	})
}
