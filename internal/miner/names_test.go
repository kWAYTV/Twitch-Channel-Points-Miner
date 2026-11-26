package miner

import (
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

var anonymizedOptions = Options{Logger: logging.LoggerSettings{AnonymizeLogs: true}}

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

	t.Run("a balance below zero is shown without its sign", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.formattedStreamerPoints(&domain.Streamer{Username: "alice", ChannelPoints: -1_500}); got != "1.5k" {
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

func TestMiner_nameWithPoints(t *testing.T) {
	t.Run("a streamer", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_500}

		want := "Alice (" + colorCyan + "1.5k" + colorReset + " points)"
		if got := h.miner.nameWithPoints(alice); got != want {
			t.Errorf("nameWithPoints() = %q, want %q", got, want)
		}
	})

	t.Run("a streamer with a points multiplier", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 950, ActiveMultipliers: []domain.ActiveMultiplier{{Factor: 1.2}}}

		want := colorDropsAccent + "Alice" + colorReset + " (" + colorCyan + "950" + colorReset + " points)"
		if got := h.miner.nameWithPoints(alice); got != want {
			t.Errorf("nameWithPoints() = %q, want %q", got, want)
		}
	})

	t.Run("anonymizing on", func(t *testing.T) {
		h := setup(t, anonymizedOptions)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 1_234_567}
		madeUp := regexp.MustCompile(`^Streamer1 \(` + regexp.QuoteMeta(colorCyan) + `([1-9]\d{2}|1k)` + regexp.QuoteMeta(colorReset) + ` points\)$`)

		if got := h.miner.nameWithPoints(alice); !madeUp.MatchString(got) {
			t.Errorf("nameWithPoints() = %q, want Streamer1 with a made-up balance from 100 to 1k", got)
		}
	})
}
