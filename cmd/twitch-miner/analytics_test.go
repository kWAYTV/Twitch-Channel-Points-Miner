package main

import (
	"path/filepath"
	"testing"
)

func TestAnalyticsConfigDefaultsAndEnvOverride(t *testing.T) {
	t.Setenv("TCPM_ANALYTICS_ADDR", "")
	cfg, err := loadOrCreateConfig(filepath.Join(t.TempDir(), "cfg.json"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Analytics.Enabled || cfg.Analytics.Host != "127.0.0.1" || cfg.Analytics.Port != 5000 {
		t.Fatalf("unexpected analytics defaults %+v", cfg.Analytics)
	}
	if got := analyticsAddr(cfg.Analytics); got != "127.0.0.1:5000" {
		t.Fatalf("addr got %q", got)
	}
	if got := analyticsAddr(analyticsConfig{}); got != "127.0.0.1:5000" {
		t.Fatalf("empty config addr got %q", got)
	}
	t.Setenv("TCPM_ANALYTICS_ADDR", "127.0.0.1:8080")
	if got := analyticsAddr(cfg.Analytics); got != "127.0.0.1:8080" {
		t.Fatalf("env override got %q", got)
	}
}
