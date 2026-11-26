package miner

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

var cacheOptions = Options{Username: "account", WatchStreakWarmStartCache: true}

// resolvedAlice is onlineStreamer("alice", "1", "Poker") with its watch streak resolved.
func resolvedAlice() *domain.Streamer {
	alice := onlineStreamer("alice", "1", "Poker")
	alice.Stream.WatchStreakMissing = false
	return alice
}

func TestWarmStartCachePathFor(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"a lower-case name", "account", filepath.Join("log", "watch_streak_cache.account.json")},
		{"capitals with spaces around", "  Account ", filepath.Join("log", "watch_streak_cache.account.json")},
		{"characters a file name cannot hold", `a/b\c:d*e?f"g<h>i|j`, filepath.Join("log", "watch_streak_cache.a_b_c_d_e_f_g_h_i_j.json")},
		{"an empty name", "", ""},
		{"a blank name", "   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := warmStartCachePathFor(tt.username); got != tt.want {
				t.Errorf("warmStartCachePathFor(%q) = %q, want %q", tt.username, got, tt.want)
			}
		})
	}
}

func TestMiner_loadWarmStartCache(t *testing.T) {
	t.Run("a streak resolved in an earlier session", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "log", "watch_streak_cache.account.json")
		earlier := setup(t, cacheOptions)
		earlier.miner.warmStartCachePath = path
		earlier.miner.loadWarmStartCache()
		earlier.miner.saveWarmStartCacheSnapshot([]*domain.Streamer{resolvedAlice()})
		h := setup(t, cacheOptions)
		h.miner.warmStartCachePath = path
		alice := onlineStreamer("alice", "1", "Poker")

		h.miner.loadWarmStartCache()

		if applied := h.miner.applyWarmStartCache(alice); !applied {
			t.Errorf("applyWarmStartCache() = false, want true")
		}
		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
	})

	t.Run("no earlier session", func(t *testing.T) {
		h := setup(t, cacheOptions)
		h.miner.warmStartCachePath = filepath.Join(t.TempDir(), "log", "watch_streak_cache.account.json")
		alice := onlineStreamer("alice", "1", "Poker")

		h.miner.loadWarmStartCache()

		if applied := h.miner.applyWarmStartCache(alice); applied {
			t.Errorf("applyWarmStartCache() = true, want false")
		}
	})

	t.Run("the warm-start cache turned off", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "log", "watch_streak_cache.account.json")
		earlier := setup(t, cacheOptions)
		earlier.miner.warmStartCachePath = path
		earlier.miner.loadWarmStartCache()
		earlier.miner.saveWarmStartCacheSnapshot([]*domain.Streamer{resolvedAlice()})
		h := setup(t, Options{Username: "account"})
		h.miner.warmStartCachePath = path
		alice := onlineStreamer("alice", "1", "Poker")

		h.miner.loadWarmStartCache()

		if applied := h.miner.applyWarmStartCache(alice); applied {
			t.Errorf("applyWarmStartCache() = true, want false")
		}
	})
}

func TestMiner_saveWarmStartCacheSnapshot(t *testing.T) {
	t.Run("an online and an offline streamer", func(t *testing.T) {
		h := setup(t, cacheOptions)
		alice := onlineStreamer("alice", "1", "Poker")
		bob := &domain.Streamer{Username: "bob", ChannelID: "2"}

		h.miner.saveWarmStartCacheSnapshot([]*domain.Streamer{alice, bob})

		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(true, true) {
			t.Errorf("cached entry for alice = %+v, want %+v", entry, aliceEntry(true, true))
		}
		entry, _ := h.cache.Get("bob")
		wantBob := aliceEntry(false, true)
		wantBob.StreamerLogin, wantBob.ChannelID, wantBob.BroadcastID, wantBob.StreamCreatedAt = "bob", "2", "", time.Time{}
		if entry != wantBob {
			t.Errorf("cached entry for bob = %+v, want %+v", entry, wantBob)
		}
	})

	t.Run("the warm-start cache turned off", func(t *testing.T) {
		h := setup(t, Options{Username: "account"})

		h.miner.saveWarmStartCacheSnapshot([]*domain.Streamer{onlineStreamer("alice", "1", "Poker")})

		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("a cache file that cannot be written", func(t *testing.T) {
		debug := cacheOptions
		debug.Logger = logging.LoggerSettings{Debug: true}
		h := unwritableCache(t, debug)

		h.miner.saveWarmStartCacheSnapshot([]*domain.Streamer{onlineStreamer("alice", "1", "Poker")})

		got := h.logLines(t)
		if len(got) != 1 || !regexp.MustCompile(`^\[DEBUG\] warm-start cache save failed: .+`).MatchString(got[0]) {
			t.Errorf("log = %q, want one debug line saying the save failed and why", got)
		}
	})
}

// unwritableCache is a miner whose warm-start cache file sits under a regular file.
func unwritableCache(t *testing.T, options Options) *harness {
	t.Helper()
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatalf("write %s: %v", blocker, err)
	}
	h := setup(t, options)
	h.miner.warmStartCachePath = filepath.Join(blocker, "watch_streak_cache.account.json")
	h.miner.loadWarmStartCache()
	return h
}

func TestMiner_syncWarmStartCacheFromStreamer(t *testing.T) {
	t.Run("an online streamer", func(t *testing.T) {
		h := setup(t, cacheOptions)

		h.miner.syncWarmStartCacheFromStreamer(onlineStreamer("alice", "1", "Poker"))

		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(true, true) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(true, true))
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, cacheOptions)

		h.miner.syncWarmStartCacheFromStreamer(nil)
	})

	t.Run("the warm-start cache turned off", func(t *testing.T) {
		h := setup(t, Options{Username: "account"})

		h.miner.syncWarmStartCacheFromStreamer(onlineStreamer("alice", "1", "Poker"))

		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("a cache file that cannot be written", func(t *testing.T) {
		debug := cacheOptions
		debug.Logger = logging.LoggerSettings{Debug: true}
		h := unwritableCache(t, debug)

		h.miner.syncWarmStartCacheFromStreamer(onlineStreamer("alice", "1", "Poker"))

		got := h.logLines(t)
		if len(got) != 1 || !regexp.MustCompile(`^\[DEBUG\] warm-start cache save failed for alice: .+`).MatchString(got[0]) {
			t.Errorf("log = %q, want one debug line saying the save for alice failed and why", got)
		}
	})
}

func TestMiner_applyWarmStartCache(t *testing.T) {
	tests := []struct {
		name        string
		cached      func() *domain.Streamer
		cachedAt    time.Time
		streamer    func() *domain.Streamer
		want        bool
		wantMissing bool
	}{
		{
			name:     "this stream's streak resolved an hour ago",
			cached:   resolvedAlice,
			cachedAt: testNow.Add(-time.Hour),
			streamer: func() *domain.Streamer { return onlineStreamer("alice", "1", "Poker") },
			want:     true,
		},
		{
			name:        "this stream's streak resolved seven hours ago",
			cached:      resolvedAlice,
			cachedAt:    testNow.Add(-7 * time.Hour),
			streamer:    func() *domain.Streamer { return onlineStreamer("alice", "1", "Poker") },
			wantMissing: true,
		},
		{
			name:        "this stream's streak still pending in the cache",
			cached:      func() *domain.Streamer { return onlineStreamer("alice", "1", "Poker") },
			cachedAt:    testNow.Add(-time.Hour),
			streamer:    func() *domain.Streamer { return onlineStreamer("alice", "1", "Poker") },
			wantMissing: true,
		},
		{
			name:     "another broadcast's streak resolved",
			cached:   resolvedAlice,
			cachedAt: testNow.Add(-time.Hour),
			streamer: func() *domain.Streamer {
				alice := onlineStreamer("alice", "1", "Poker")
				alice.Stream.BroadcastID = "alice-broadcast-2"
				return alice
			},
			wantMissing: true,
		},
		{
			name:        "nothing cached for the streamer",
			cached:      func() *domain.Streamer { return nil },
			cachedAt:    testNow.Add(-time.Hour),
			streamer:    func() *domain.Streamer { return onlineStreamer("alice", "1", "Poker") },
			wantMissing: true,
		},
		{
			name:     "an offline streamer",
			cached:   resolvedAlice,
			cachedAt: testNow.Add(-time.Hour),
			streamer: func() *domain.Streamer {
				alice := onlineStreamer("alice", "1", "Poker")
				alice.IsOnline = false
				return alice
			},
			wantMissing: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, cacheOptions)
			h.cache.UpdateFromStreamer(tt.cached(), tt.cachedAt)
			streamer := tt.streamer()

			got := h.miner.applyWarmStartCache(streamer)

			if got != tt.want {
				t.Errorf("applyWarmStartCache() = %t, want %t", got, tt.want)
			}
			if streamer.Stream.WatchStreakMissing != tt.wantMissing {
				t.Errorf("WatchStreakMissing = %t, want %t", streamer.Stream.WatchStreakMissing, tt.wantMissing)
			}
		})
	}

	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, cacheOptions)

		if got := h.miner.applyWarmStartCache(nil); got {
			t.Errorf("applyWarmStartCache() = true, want false")
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		h := setup(t, cacheOptions)
		h.cache.UpdateFromStreamer(resolvedAlice(), testNow.Add(-time.Hour))

		if got := h.miner.applyWarmStartCache(&domain.Streamer{Username: "alice", ChannelID: "1", IsOnline: true}); got {
			t.Errorf("applyWarmStartCache() = true, want false")
		}
	})

	t.Run("the warm-start cache turned off", func(t *testing.T) {
		h := setup(t, Options{Username: "account"})

		if got := h.miner.applyWarmStartCache(onlineStreamer("alice", "1", "Poker")); got {
			t.Errorf("applyWarmStartCache() = true, want false")
		}
	})
}
