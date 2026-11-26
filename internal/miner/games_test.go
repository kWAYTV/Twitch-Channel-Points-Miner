package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func TestMiner_resolveGameName(t *testing.T) {
	showGame := Options{Username: "account", ShowGame: true, WatchStreakWarmStartCache: true}

	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, showGame)

		if got := h.miner.resolveGameName(nil); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
	})

	t.Run("a game already known, with showing games off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", " Poker ")

		if got := h.miner.resolveGameName(alice); got != "Poker" {
			t.Errorf("resolveGameName() = %q, want %q", got, "Poker")
		}
	})

	t.Run("no game known, with showing games off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "")
		h.twitch.Channels["alice"].Game = "Chess"

		if got := h.miner.resolveGameName(alice); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
	})

	t.Run("no game known and the streamer offline", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "")
		h.twitch.Channels["alice"].Game = "Chess"
		alice.IsOnline = false

		if got := h.miner.resolveGameName(alice); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
	})

	t.Run("no game known, before login", func(t *testing.T) {
		m := New(Options{ShowGame: true})
		alice := onlineStreamer("alice", "1", "")

		if got := m.resolveGameName(alice); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
	})

	t.Run("no game known and Twitch has one", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "")
		h.twitch.Channels["alice"].Game = "Chess"

		if got := h.miner.resolveGameName(alice); got != "Chess" {
			t.Errorf("resolveGameName() = %q, want %q", got, "Chess")
		}
	})

	t.Run("no stream yet and Twitch has a game", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "Chess")
		alice.Stream = nil

		if got := h.miner.resolveGameName(alice); got != "Chess" {
			t.Errorf("resolveGameName() = %q, want %q", got, "Chess")
		}
	})

	t.Run("no game known and Twitch has none", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "")

		if got := h.miner.resolveGameName(alice); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
		if entry, ok := h.cache.Get("alice"); ok {
			t.Errorf("cached entry = %+v, want none", entry)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "")
		h.twitch.Fail["UpdateStream"] = errors.New("503")

		if got := h.miner.resolveGameName(alice); got != "" {
			t.Errorf("resolveGameName() = %q, want none", got)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("Twitch failing, with debug logging", func(t *testing.T) {
		debug := showGame
		debug.Logger = logging.LoggerSettings{Debug: true}
		h := setup(t, debug)
		alice := h.live("alice", "1", "")
		h.twitch.Fail["UpdateStream"] = errors.New("503")

		h.miner.resolveGameName(alice)

		wantLog := []string{"[DEBUG] update stream Alice for game: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch reporting the streak complete", func(t *testing.T) {
		h := setup(t, showGame)
		alice := h.live("alice", "1", "")
		h.twitch.Channels["alice"].StreakComplete = true

		h.miner.resolveGameName(alice)

		entry, _ := h.cache.Get("alice")
		wantEntry := streak.WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "alice",
			ChannelID:       "1",
			BroadcastID:     "alice-broadcast",
			StreamCreatedAt: testNow.Add(-5 * time.Minute),
			CheckedAt:       testNow,
			IsOnline:        true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})
}

func TestMiner_gameInfo(t *testing.T) {
	t.Run("showing games on", func(t *testing.T) {
		h := setup(t, Options{ShowGame: true})
		alice := freshStreamer("alice", "1", "Poker")

		if got := h.miner.gameInfo(alice); got != "Poker" {
			t.Errorf("gameInfo() = %q, want %q", got, "Poker")
		}
	})

	t.Run("showing games off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		if got := h.miner.gameInfo(alice); got != "" {
			t.Errorf("gameInfo() = %q, want none", got)
		}
	})

	t.Run("no miner", func(t *testing.T) {
		var m *Miner
		alice := freshStreamer("alice", "1", "Poker")

		if got := m.gameInfo(alice); got != "" {
			t.Errorf("gameInfo() = %q, want none", got)
		}
	})
}

func TestMiner_gameSuffix(t *testing.T) {
	t.Run("a streamer with a game", func(t *testing.T) {
		h := setup(t, Options{ShowGame: true})
		alice := freshStreamer("alice", "1", "Poker")

		if got := h.miner.gameSuffix(alice); got != "Poker" {
			t.Errorf("gameSuffix() = %q, want %q", got, "Poker")
		}
	})

	t.Run("showing games off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		if got := h.miner.gameSuffix(alice); got != "" {
			t.Errorf("gameSuffix() = %q, want none", got)
		}
	})
}

func TestMiner_watchContext(t *testing.T) {
	t.Run("a streamer with a game", func(t *testing.T) {
		h := setup(t, Options{ShowGame: true})
		alice := freshStreamer("alice", "1", "Poker")

		want := "| " + colorGameLabel + "Game:" + colorReset + " Poker"
		if got := h.miner.watchContext(alice); got != want {
			t.Errorf("watchContext() = %q, want %q", got, want)
		}
	})

	t.Run("showing games off", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		if got := h.miner.watchContext(alice); got != "" {
			t.Errorf("watchContext() = %q, want none", got)
		}
	})
}

func TestMiner_handleGameChange(t *testing.T) {
	announced := []string{"[INFO] Alice now playing: Chess!"}
	tests := []struct {
		name     string
		options  Options
		previous string
		current  string
		wantLog  []string
	}{
		{"a new game", Options{ShowGame: true}, "Poker", "Chess", announced},
		{"a new game with spaces around", Options{ShowGame: true}, " Poker ", "  Chess ", announced},
		{"the same game in another case", Options{ShowGame: true}, "chess", "Chess", nil},
		{"a first game", Options{ShowGame: true}, "", "Chess", nil},
		{"no game any more", Options{ShowGame: true}, "Poker", "  ", nil},
		{"showing games off", Options{}, "Poker", "Chess", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)
			alice := &domain.Streamer{Username: "alice"}

			h.miner.handleGameChange(alice, tt.previous, tt.current)

			if got := h.logLines(t); !reflect.DeepEqual(got, tt.wantLog) {
				t.Errorf("log = %q, want %q", got, tt.wantLog)
			}
		})
	}

	t.Run("no miner", func(t *testing.T) {
		var m *Miner

		m.handleGameChange(&domain.Streamer{Username: "alice"}, "Poker", "Chess")
	})
}
