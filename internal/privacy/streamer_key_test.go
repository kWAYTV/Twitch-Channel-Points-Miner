package privacy

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestStreamerKey(t *testing.T) {
	tests := []struct {
		name     string
		streamer *domain.Streamer
		want     string
	}{
		{"a channel ID and a username", &domain.Streamer{ChannelID: "123", Username: "pewdiepie"}, "id:123"},
		{"a channel ID in spaces", &domain.Streamer{ChannelID: " 123 "}, "id:123"},
		{"only a username", &domain.Streamer{Username: "pewdiepie"}, "name:pewdiepie"},
		{"a username in capitals and spaces", &domain.Streamer{Username: " PewDiePie "}, "name:pewdiepie"},
		{"a channel ID of spaces", &domain.Streamer{ChannelID: "  ", Username: "pewdiepie"}, "name:pewdiepie"},
		{"neither", &domain.Streamer{}, ""},
		{"no streamer", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streamerKey(tt.streamer); got != tt.want {
				t.Errorf("streamerKey(%+v) = %q, want %q", tt.streamer, got, tt.want)
			}
		})
	}
}
