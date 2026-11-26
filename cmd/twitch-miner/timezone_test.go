package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

func TestRequestedTimezone(t *testing.T) {
	text := func(s string) *string { return &s }
	tests := []struct {
		name string
		raw  *string
		want string
	}{
		{name: "unset", raw: nil, want: ""},
		{name: "empty", raw: text(""), want: ""},
		{name: "blank", raw: text("   "), want: ""},
		{name: "auto", raw: text("auto"), want: ""},
		{name: "auto in upper case", raw: text("AUTO"), want: ""},
		{name: "auto padded", raw: text(" auto "), want: ""},
		{name: "a zone", raw: text("Europe/Madrid"), want: "Europe/Madrid"},
		{name: "a padded zone", raw: text(" Europe/Madrid "), want: "Europe/Madrid"},
		{name: "a word that starts with auto", raw: text("automatic"), want: "automatic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestedTimezone(tt.raw); got != tt.want {
				t.Errorf("requestedTimezone() = %q, want %q", got, tt.want)
			}
		})
	}
}

// setupTimezone makes the system zone one no test asks for, so a change
// shows, and puts the real one back when the test ends.
func setupTimezone(t *testing.T) (system *time.Location, logger *logging.Logger, logged *bytes.Buffer) {
	t.Helper()
	original := time.Local
	t.Cleanup(func() { time.Local = original })
	system = time.FixedZone("SYSTEM", 3600)
	time.Local = system
	logged = &bytes.Buffer{}
	return system, logging.NewLoggerTo(logged, logging.LoggerSettings{}), logged
}

func TestApplyTimezoneOverride(t *testing.T) {
	t.Run("a zone", func(t *testing.T) {
		_, logger, logged := setupTimezone(t)
		zone := "UTC"

		applyTimezoneOverride(&zone, logger)

		if time.Local != time.UTC {
			t.Errorf("time.Local = %v, want UTC", time.Local)
		}
		if logged.String() != "" {
			t.Errorf("logged %q, want nothing", logged.String())
		}
	})

	t.Run("unset", func(t *testing.T) {
		system, logger, logged := setupTimezone(t)

		applyTimezoneOverride(nil, logger)

		if time.Local != system {
			t.Errorf("time.Local = %v, want the system zone", time.Local)
		}
		if logged.String() != "" {
			t.Errorf("logged %q, want nothing", logged.String())
		}
	})

	t.Run("an unknown zone", func(t *testing.T) {
		system, logger, _ := setupTimezone(t)
		zone := "Bad/Zone"

		applyTimezoneOverride(&zone, logger)

		if time.Local != system {
			t.Errorf("time.Local = %v, want the system zone", time.Local)
		}
	})
}
