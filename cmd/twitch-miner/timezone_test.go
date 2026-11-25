package main

import (
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

func TestApplyTimezoneOverride(t *testing.T) {
	original := time.Local
	defer func() { time.Local = original }()

	zone := "UTC"
	logger := logging.NewLogger(logging.LoggerSettings{}, "")
	applyTimezoneOverride(&zone, logger)
	if time.Local.String() != "UTC" {
		t.Fatalf("expected time.Local set to UTC, got %s", time.Local.String())
	}

	bad := "Bad/Zone"
	applyTimezoneOverride(&bad, logger)
	if time.Local.String() != "UTC" {
		t.Fatalf("invalid timezone should not change location, got %s", time.Local.String())
	}
}
