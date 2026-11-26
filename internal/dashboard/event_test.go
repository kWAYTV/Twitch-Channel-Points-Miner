package dashboard

import (
	"strings"
	"testing"
)

func TestStoredType(t *testing.T) {
	tests := []struct {
		name   string
		event  string
		want   string
		wantOK bool
	}{
		{"a streamer going online", "STREAMER_ONLINE", "STREAMER_ONLINE", true},
		{"a streamer going offline", "STREAMER_OFFLINE", "STREAMER_OFFLINE", true},
		{"a bet won", "BET_WIN", "BET_WIN", true},
		{"a bet lost", "BET_LOSE", "BET_LOSE", true},
		{"a bet refunded", "BET_REFUND", "BET_REFUND", true},
		{"a drop claimed", "DROP_CLAIM", "DROP_CLAIM", true},
		{"a moment claimed", "MOMENT_CLAIM", "MOMENT_CLAIM", true},
		{"lower case", "bet_win", "BET_WIN", true},
		{"spaces around it", " BET_WIN\n", "BET_WIN", true},
		{"a watch gain", "GAIN_FOR_WATCH", "GAIN_FOR_WATCH", false},
		{"a watch streak", "WATCH_STREAK", "WATCH_STREAK", false},
		{"an empty name", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := storedType(tt.event)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("storedType(%q) = %q, %v, want %q, %v", tt.event, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestClipMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"a short sentence", "Won a bet on alice.", "Won a bet on alice."},
		{"spaces around it", "  Won a bet on alice.\n", "Won a bet on alice."},
		{"an empty message", "", ""},
		{"only spaces", "   ", ""},
		{"299 characters", strings.Repeat("x", 299), strings.Repeat("x", 299)},
		{"300 characters", strings.Repeat("x", 300), strings.Repeat("x", 300)},
		{"301 characters", strings.Repeat("x", 301), strings.Repeat("x", 300) + "…"},
		{"300 characters outside ASCII", strings.Repeat("é", 300), strings.Repeat("é", 300)},
		{"301 characters outside ASCII", strings.Repeat("é", 301), strings.Repeat("é", 300) + "…"},
		{"300 characters with spaces around them", " " + strings.Repeat("x", 300) + " ", strings.Repeat("x", 300)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clipMessage(tt.message); got != tt.want {
				t.Errorf("clipMessage(%q) = %q, want %q", tt.message, got, tt.want)
			}
		})
	}
}
