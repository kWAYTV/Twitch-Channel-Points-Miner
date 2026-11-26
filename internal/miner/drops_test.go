package miner

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestMiner_shouldClaimDrops(t *testing.T) {
	claiming := domain.StreamerSettings{ClaimDrops: true}
	notClaiming := domain.StreamerSettings{ClaimDrops: false}
	tests := []struct {
		name      string
		global    domain.StreamerSettings
		streamers []*domain.Streamer
		want      bool
	}{
		{"on for everyone and no streamers", claiming, nil, true},
		{"on for everyone and off for a streamer", claiming, []*domain.Streamer{{Settings: notClaiming}}, true},
		{"off for everyone and no streamers", notClaiming, nil, false},
		{"off for everyone and for each streamer", notClaiming, []*domain.Streamer{{Settings: notClaiming}, {Settings: notClaiming}}, false},
		{"off for everyone and on for the last streamer", notClaiming, []*domain.Streamer{{Settings: notClaiming}, {Settings: claiming}}, true},
		{"off for everyone and a missing streamer", notClaiming, []*domain.Streamer{nil, {Settings: notClaiming}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, Options{StreamerSettings: tt.global})

			if got := h.miner.shouldClaimDrops(tt.streamers); got != tt.want {
				t.Errorf("shouldClaimDrops() = %t, want %t", got, tt.want)
			}
		})
	}
}
