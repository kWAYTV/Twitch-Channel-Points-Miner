package miner

import (
	"testing"
	"time"
)

func TestSignedChange(t *testing.T) {
	tests := []struct {
		name          string
		delta         int
		wantColor     string
		wantSign      string
		wantMagnitude int
	}{
		{"a gain", 50, colorGreen, "+", 50},
		{"a loss", -50, colorRed, "-", 50},
		{"no change", 0, colorGreen, "+", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color, sign, magnitude := signedChange(tt.delta)

			if color != tt.wantColor || sign != tt.wantSign || magnitude != tt.wantMagnitude {
				t.Errorf("signedChange(%d) = %q, %q, %d, want %q, %q, %d", tt.delta, color, sign, magnitude, tt.wantColor, tt.wantSign, tt.wantMagnitude)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "00s"},
		{"seconds only", 45 * time.Second, "45s"},
		{"just under a minute", 59 * time.Second, "59s"},
		{"exactly a minute", time.Minute, "01m 00s"},
		{"minutes and seconds", 3*time.Minute + 4*time.Second, "03m 04s"},
		{"exactly an hour", time.Hour, "01h 00m 00s"},
		{"exactly a day", 24 * time.Hour, "1d 00h 00m 00s"},
		{"days, hours, minutes and seconds", 26*time.Hour + 3*time.Minute + 4*time.Second, "1d 02h 03m 04s"},
		{"half a second rounds up", 1500 * time.Millisecond, "02s"},
		{"under half a second rounds down", 1400 * time.Millisecond, "01s"},
		{"a negative duration", -45 * time.Second, "45s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.duration); got != tt.want {
				t.Errorf("formatDuration(%s) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}
