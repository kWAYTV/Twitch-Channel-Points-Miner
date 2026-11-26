package heartbeat

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var noon = time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)

func TestAge(t *testing.T) {
	t.Run("a beat three minutes ago", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "heartbeat")
		if err := Write(path, noon); err != nil {
			t.Fatalf("Write = %v", err)
		}

		got, err := Age(path, noon.Add(3*time.Minute))

		if err != nil || got != 3*time.Minute {
			t.Errorf("Age = %v, %v, want 3m0s", got, err)
		}
	})

	t.Run("a beat this instant", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "heartbeat")
		Write(path, noon)

		got, err := Age(path, noon)

		if err != nil || got != 0 {
			t.Errorf("Age = %v, %v, want 0s", got, err)
		}
	})

	t.Run("the latest of two beats", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "heartbeat")
		Write(path, noon)
		Write(path, noon.Add(time.Minute))

		got, err := Age(path, noon.Add(time.Minute+10*time.Second))

		if err != nil || got != 10*time.Second {
			t.Errorf("Age = %v, %v, want 10s", got, err)
		}
	})

	t.Run("no beat yet", func(t *testing.T) {
		_, err := Age(filepath.Join(t.TempDir(), "heartbeat"), noon)

		if !os.IsNotExist(err) {
			t.Errorf("Age error = %v, want a missing file", err)
		}
	})

	t.Run("a file that holds no time", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "heartbeat")
		os.WriteFile(path, []byte("hello"), 0o644)

		if _, err := Age(path, noon); err == nil {
			t.Errorf("Age error = nil, want an error")
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("a folder that does not exist yet", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "data", "heartbeat")

		if err := Write(path, noon); err != nil {
			t.Fatalf("Write = %v", err)
		}

		if got, err := Age(path, noon.Add(time.Second)); err != nil || got != time.Second {
			t.Errorf("Age = %v, %v, want 1s", got, err)
		}
	})

	t.Run("a path that is a folder", func(t *testing.T) {
		if err := Write(t.TempDir(), noon); err == nil {
			t.Errorf("Write error = nil, want an error")
		}
	})
}
