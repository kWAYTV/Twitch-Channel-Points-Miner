package miner

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/heartbeat"
)

func TestMiner_beat(t *testing.T) {
	noon := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)

	t.Run("a heartbeat file set", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "heartbeat")
		m := New(Options{HeartbeatPath: path})
		m.clock = func() time.Time { return noon }

		m.beat()

		if got, err := heartbeat.Age(path, noon.Add(time.Minute)); err != nil || got != time.Minute {
			t.Errorf("Age = %v, %v, want 1m0s", got, err)
		}
	})

	t.Run("no heartbeat file set", func(t *testing.T) {
		New(Options{}).beat()
	})
}
