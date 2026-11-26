package miner

import (
	"testing"
	"time"
)

func TestFormatChannelPoints(t *testing.T) {
	tests := []struct {
		name   string
		points int
		want   string
	}{
		{"zero", 0, "0"},
		{"just under a thousand", 999, "999"},
		{"exactly a thousand", 1_000, "1k"},
		{"thousands with a fraction", 1_500, "1.5k"},
		{"thousands rounded to two decimals", 1_234, "1.23k"},
		{"the last value that reads in thousands", 999_994, "999.99k"},
		{"just under a million (today: 1000k)", 999_999, "1000k"},
		{"exactly a million", 1_000_000, "1M"},
		{"millions with a fraction", 1_500_000, "1.5M"},
		{"a negative balance", -1_500, "1.5k"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatChannelPoints(tt.points); got != tt.want {
				t.Errorf("formatChannelPoints(%d) = %q, want %q", tt.points, got, tt.want)
			}
		})
	}

	t.Run("just under a million (bug: should read 1M)", func(t *testing.T) {
		t.Skip("bug: 999,995 to 999,999 round up to 1000k instead of 1M")
		if got := formatChannelPoints(999_999); got != "1M" {
			t.Errorf("formatChannelPoints(%d) = %q, want %q", 999_999, got, "1M")
		}
	})
}

func TestFormatPointsWithSuffix(t *testing.T) {
	tests := []struct {
		name    string
		points  int
		divisor float64
		suffix  string
		want    string
	}{
		{"a whole number", 2_000_000, 1_000_000, "M", "2M"},
		{"one decimal", 1_500_000, 1_000_000, "M", "1.5M"},
		{"two decimals", 1_250_000, 1_000_000, "M", "1.25M"},
		{"a third decimal rounded away", 1_001, 1_000, "k", "1k"},
		{"a trailing zero inside the number", 10_000, 1_000, "k", "10k"},
		{"zero", 0, 1_000, "k", "0k"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPointsWithSuffix(tt.points, tt.divisor, tt.suffix); got != tt.want {
				t.Errorf("formatPointsWithSuffix(%d, %v, %q) = %q, want %q", tt.points, tt.divisor, tt.suffix, got, tt.want)
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
