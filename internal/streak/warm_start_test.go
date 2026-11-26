package streak

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

var checkedAt = time.Date(2026, time.April, 10, 11, 45, 0, 0, time.UTC)

// liveStreamer is on broadcast-1 with its watch streak collected.
func liveStreamer(login string) *domain.Streamer {
	return &domain.Streamer{
		Username:  login,
		ChannelID: "123",
		IsOnline:  true,
		Stream:    &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt},
	}
}

func cachePath(t *testing.T) string {
	return filepath.Join(t.TempDir(), "watch_streak_cache.json")
}

func writeCacheFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestWarmStartCache_entryRoundTrip(t *testing.T) {
	t.Run("a live streamer whose streak is collected", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		want := WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "streamer",
			ChannelID:       "123",
			BroadcastID:     "broadcast-1",
			StreamCreatedAt: startedAt,
			CheckedAt:       checkedAt,
			IsOnline:        true,
		}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("an offline streamer with every optional field empty", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("a login read back in capitals with spaces around it", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get(" STREAMER ")

		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt}
		if !ok || got != want {
			t.Errorf("Get( STREAMER ) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("a later check of the same streamer replaces the earlier one", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt.Add(time.Minute))
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt.Add(time.Minute)}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("a login nobody saved", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("someone-else")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(someone-else) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("another account's file", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("someone-else", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})
}

func TestWarmStartCache_SaveIfDirty(t *testing.T) {
	t.Run("two streamers, written in order of login", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "zulu"}, checkedAt)
		cache.UpdateFromStreamer(liveStreamer("alpha"), checkedAt)

		err := cache.SaveIfDirty()

		if err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		want := `{
  "version": 1,
  "entries": [
    {
      "account_name": "account",
      "streamer_login": "alpha",
      "channel_id": "123",
      "broadcast_id": "broadcast-1",
      "stream_created_at": "2026-04-10T11:30:00Z",
      "watch_streak_missing": false,
      "checked_at": "2026-04-10T11:45:00Z",
      "is_online": true
    },
    {
      "account_name": "account",
      "streamer_login": "zulu",
      "stream_created_at": "0001-01-01T00:00:00Z",
      "watch_streak_missing": true,
      "checked_at": "2026-04-10T11:45:00Z",
      "is_online": false
    }
  ]
}`
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		if string(raw) != want {
			t.Errorf("file = %s, want %s", raw, want)
		}
	})

	t.Run("a folder that does not exist yet is created", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache", "account", "watch_streak_cache.json")
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)

		err := cache.SaveIfDirty()

		if err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		got, ok := LoadWarmStartCache(path, "account").Get("streamer")
		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("nothing to save writes no file", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)

		err := cache.SaveIfDirty()

		if err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("Stat(%s) error = %v, want %v", path, statErr, os.ErrNotExist)
		}
	})

	t.Run("no streamer to remember writes no file", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(nil, checkedAt)

		err := cache.SaveIfDirty()

		if err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("Stat(%s) error = %v, want %v", path, statErr, os.ErrNotExist)
		}
	})

	t.Run("a second save with nothing new writes no file", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove %s: %v", path, err)
		}

		err := cache.SaveIfDirty()

		if err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("Stat(%s) error = %v, want %v", path, statErr, os.ErrNotExist)
		}
	})

	t.Run("the same check again after a save writes no file", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove %s: %v", path, err)
		}
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)

		err := cache.SaveIfDirty()

		if err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("Stat(%s) error = %v, want %v", path, statErr, os.ErrNotExist)
		}
	})

	t.Run("a new check after a save is written", func(t *testing.T) {
		path := cachePath(t)
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt.Add(time.Minute))

		err := cache.SaveIfDirty()

		if err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		got, ok := LoadWarmStartCache(path, "account").Get("streamer")
		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt.Add(time.Minute)}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("a folder that cannot be created", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "not-a-folder")
		writeCacheFile(t, blocker, "")
		cache := NewWarmStartCache("account", filepath.Join(blocker, "watch_streak_cache.json"))
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)

		err := cache.SaveIfDirty()

		if !errors.Is(err, syscall.ENOTDIR) {
			t.Errorf("SaveIfDirty() error = %v, want %v", err, syscall.ENOTDIR)
		}
	})

	t.Run("a save that failed is tried again", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "cache")
		writeCacheFile(t, blocker, "")
		path := filepath.Join(blocker, "watch_streak_cache.json")
		cache := NewWarmStartCache("account", path)
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); !errors.Is(err, syscall.ENOTDIR) {
			t.Fatalf("SaveIfDirty() error = %v, want %v", err, syscall.ENOTDIR)
		}
		if err := os.Remove(blocker); err != nil {
			t.Fatalf("remove %s: %v", blocker, err)
		}

		err := cache.SaveIfDirty()

		if err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}
		got, ok := LoadWarmStartCache(path, "account").Get("streamer")
		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt}
		if !ok || got != want {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, want)
		}
	})

	t.Run("a cache with no path", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)

		if err := cache.SaveIfDirty(); err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
	})

	t.Run("no cache", func(t *testing.T) {
		var cache *WarmStartCache
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)

		if err := cache.SaveIfDirty(); err != nil {
			t.Errorf("SaveIfDirty() error = %v, want nil", err)
		}
	})
}

func TestLoadWarmStartCache(t *testing.T) {
	const oneEntry = `"entries": [{"account_name": "account", "streamer_login": "streamer", "watch_streak_missing": true, "checked_at": "2026-04-10T11:45:00Z", "is_online": false}]`
	saved := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: checkedAt}

	t.Run("a file of this version", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, `{"version": 1, `+oneEntry+`}`)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if !ok || got != saved {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, saved)
		}
	})

	t.Run("a file with no version", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, `{`+oneEntry+`}`)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if !ok || got != saved {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, saved)
		}
	})

	t.Run("a file of a later version", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, `{"version": 2, `+oneEntry+`}`)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("a file that is not JSON", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, `{not-json`)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("a file with a version that is not a number", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, `{"version": "one", `+oneEntry+`}`)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("an empty file", func(t *testing.T) {
		path := cachePath(t)
		writeCacheFile(t, path, ``)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("no file", func(t *testing.T) {
		path := cachePath(t)

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("no path", func(t *testing.T) {
		got, ok := LoadWarmStartCache("", "account").Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("a cache loaded from no file saves to the same path", func(t *testing.T) {
		path := cachePath(t)
		cache := LoadWarmStartCache(path, "Account")
		cache.UpdateFromStreamer(&domain.Streamer{Username: "streamer"}, checkedAt)
		if err := cache.SaveIfDirty(); err != nil {
			t.Fatalf("SaveIfDirty() error = %v, want nil", err)
		}

		got, ok := LoadWarmStartCache(path, "account").Get("streamer")

		if !ok || got != saved {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, true", got, ok, saved)
		}
	})
}

func TestWarmStartCache_Get(t *testing.T) {
	t.Run("no cache", func(t *testing.T) {
		var cache *WarmStartCache

		got, ok := cache.Get("streamer")

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("Get(streamer) = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})
}

func TestWarmStartCache_ResolvedEntryFor(t *testing.T) {
	resolved := WarmStartEntry{
		AccountName:     "account",
		StreamerLogin:   "streamer",
		ChannelID:       "123",
		BroadcastID:     "broadcast-1",
		StreamCreatedAt: startedAt,
		CheckedAt:       checkedAt,
		IsOnline:        true,
	}

	t.Run("a collected streak on the stream still live", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)

		got, ok := cache.ResolvedEntryFor(liveStreamer("streamer"), now)

		if !ok || got != resolved {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, true", got, ok, resolved)
		}
	})

	t.Run("a login in capitals with spaces around it", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)

		got, ok := cache.ResolvedEntryFor(liveStreamer(" STREAMER "), now)

		if !ok || got != resolved {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, true", got, ok, resolved)
		}
	})

	t.Run("seven hours after the check", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)

		got, ok := cache.ResolvedEntryFor(liveStreamer("streamer"), checkedAt.Add(7*time.Hour))

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("the streamer has moved to another broadcast", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)
		streamer := &domain.Streamer{Username: "streamer", IsOnline: true, Stream: &domain.Stream{BroadcastID: "broadcast-2", CreatedAt: startedAt}}

		got, ok := cache.ResolvedEntryFor(streamer, now)

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("a streamer the cache has never seen", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)

		got, ok := cache.ResolvedEntryFor(liveStreamer("someone-else"), now)

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		cache := NewWarmStartCache("account", "")
		cache.UpdateFromStreamer(liveStreamer("streamer"), checkedAt)

		got, ok := cache.ResolvedEntryFor(nil, now)

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})

	t.Run("no cache", func(t *testing.T) {
		var cache *WarmStartCache

		got, ok := cache.ResolvedEntryFor(liveStreamer("streamer"), now)

		if ok || got != (WarmStartEntry{}) {
			t.Errorf("ResolvedEntryFor() = %+v, %t, want %+v, false", got, ok, WarmStartEntry{})
		}
	})
}
