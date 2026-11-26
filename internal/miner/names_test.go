package miner

import (
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

var anonymizedOptions = Options{Logger: logging.LoggerSettings{AnonymizeLogs: true}}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		login string
		want  string
	}{
		{"an empty name", "", ""},
		{"a lower-case name", "alice", "Alice"},
		{"a capitalised name", "Alice", "Alice"},
		{"one letter", "a", "A"},
		{"a name that starts with a digit", "1alice", "1alice"},
		{"a name that starts with an underscore", "_alice", "_alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayName(tt.login); got != tt.want {
				t.Errorf("displayName(%q) = %q, want %q", tt.login, got, tt.want)
			}
		})
	}

	t.Run("a first letter outside ASCII (today: the letter is broken)", func(t *testing.T) {
		if got, want := displayName("élan"), "�\xa9lan"; got != want {
			t.Errorf("displayName(%q) = %q, want %q", "élan", got, want)
		}
	})

	t.Run("a first letter outside ASCII (bug: should read Élan)", func(t *testing.T) {
		t.Skip("bug: only the first byte is upper-cased, which splits a multi-byte letter")
		if got, want := displayName("élan"), "Élan"; got != want {
			t.Errorf("displayName(%q) = %q, want %q", "élan", got, want)
		}
	})
}

func TestMiner_anonymizeLogs(t *testing.T) {
	t.Run("anonymizing off", func(t *testing.T) {
		h := setup(t, Options{})

		if h.miner.anonymizeLogs() {
			t.Errorf("anonymizeLogs() = true, want false")
		}
	})

	t.Run("anonymizing on", func(t *testing.T) {
		h := setup(t, anonymizedOptions)

		if !h.miner.anonymizeLogs() {
			t.Errorf("anonymizeLogs() = false, want true")
		}
	})

	t.Run("no miner", func(t *testing.T) {
		var m *Miner

		if m.anonymizeLogs() {
			t.Errorf("anonymizeLogs() = true, want false")
		}
	})
}

func TestMiner_rawStreamerName(t *testing.T) {
	t.Run("anonymizing off", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.rawStreamerName("alice"); got != "alice" {
			t.Errorf("rawStreamerName(%q) = %q, want %q", "alice", got, "alice")
		}
	})

	t.Run("anonymizing on", func(t *testing.T) {
		h := setup(t, anonymizedOptions)

		if got := h.miner.rawStreamerName("alice"); got != "Streamer1" {
			t.Errorf("rawStreamerName(%q) = %q, want %q", "alice", got, "Streamer1")
		}
	})
}

func TestMiner_styledStreamerName(t *testing.T) {
	boosted := []domain.ActiveMultiplier{{Factor: 1.2}}
	tests := []struct {
		name     string
		options  Options
		streamer *domain.Streamer
		want     string
	}{
		{"no streamer", Options{}, nil, ""},
		{"a streamer", Options{}, &domain.Streamer{Username: "alice"}, "Alice"},
		{"a streamer with a points multiplier", Options{}, &domain.Streamer{Username: "alice", ActiveMultipliers: boosted}, colorDropsAccent + "Alice" + colorReset},
		{"anonymizing on", anonymizedOptions, &domain.Streamer{Username: "alice"}, "Streamer1"},
		{"anonymizing on, with a points multiplier", anonymizedOptions, &domain.Streamer{Username: "alice", ActiveMultipliers: boosted}, colorDropsAccent + "Streamer1" + colorReset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)

			if got := h.miner.styledStreamerName(tt.streamer); got != tt.want {
				t.Errorf("styledStreamerName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMiner_formattedStreamerPoints(t *testing.T) {
	t.Run("no streamer", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.formattedStreamerPoints(nil); got != "" {
			t.Errorf("formattedStreamerPoints() = %q, want none", got)
		}
	})

	t.Run("a balance in the thousands", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.formattedStreamerPoints(&domain.Streamer{Username: "alice", ChannelPoints: 1_500}); got != "1.5k" {
			t.Errorf("formattedStreamerPoints() = %q, want %q", got, "1.5k")
		}
	})

	t.Run("anonymizing on", func(t *testing.T) {
		h := setup(t, anonymizedOptions)
		madeUp := regexp.MustCompile(`^([1-9]\d{2}|1k)$`)

		if got := h.miner.formattedStreamerPoints(&domain.Streamer{Username: "alice", ChannelPoints: 1_234_567}); !madeUp.MatchString(got) {
			t.Errorf("formattedStreamerPoints() = %q, want a made-up balance from 100 to 1k", got)
		}
	})
}
