package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestDefaultConfigIncludesExpectedKeys(t *testing.T) {
	cfg := defaultConfig()
	required := []string{
		"username",
		"password",
		"claim_moments",
		"streamers",
		"game_priority",
		"chat_presence",
		"disable_at_in_nickname",
		"show_drops_progress",
		"bet",
		"watch_priority",
	}
	for _, key := range required {
		if _, ok := cfg[key]; !ok {
			t.Fatalf("missing key %q in default config", key)
		}
	}
	if got, ok := cfg["show_drops_progress"].(bool); !ok || got {
		t.Fatalf("show_drops_progress default got %#v, want false", cfg["show_drops_progress"])
	}
}

func TestLoadOrCreateConfigCreatesFileAndDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")

	cfg, err := loadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("load/create config error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file should be created: %v", err)
	}
	if cfg.WatchPriority == nil || len(cfg.WatchPriority) == 0 {
		t.Fatalf("watch priority should have defaults: %#v", cfg)
	}
}

func TestBuildBaseStreamerSettingsAppliesGlobalFilterCondition(t *testing.T) {
	cfg := config{
		BettingMakePredictions: true,
		FollowRaid:             true,
		ClaimDrops:             true,
		CommunityGoals:         false,
		IRCMode:                "ONLINE",
		Bet: betConfig{
			FilterCondition: &filterConditionConfig{
				By:    "TOTAL_USERS",
				Where: "GTE",
				Value: func() *float64 { v := 500000.0; return &v }(),
			},
		},
	}

	base := buildBaseStreamerSettings(cfg)
	if base.Bet.FilterCondition == nil {
		t.Fatalf("expected global filter_condition applied to base streamer settings")
	}
	if base.Bet.FilterCondition.By != "TOTAL_USERS" {
		t.Fatalf("expected By TOTAL_USERS, got %s", base.Bet.FilterCondition.By)
	}
	if base.Bet.FilterCondition.Where != "GTE" {
		t.Fatalf("expected Where GTE, got %s", base.Bet.FilterCondition.Where)
	}
	if base.Bet.FilterCondition.Value == nil || *base.Bet.FilterCondition.Value != 500000.0 {
		t.Fatalf("expected Value 500000, got %#v", base.Bet.FilterCondition.Value)
	}
}

func TestBuildBaseStreamerSettingsUsesGlobalClaimMoments(t *testing.T) {
	cfg := config{
		BettingMakePredictions: true,
		FollowRaid:             true,
		ClaimDrops:             true,
		ClaimMoments:           false,
		CommunityGoals:         false,
		IRCMode:                "ONLINE",
	}

	base := buildBaseStreamerSettings(cfg)
	if base.ClaimMoments {
		t.Fatalf("expected base claim_moments false from global config")
	}
}

func TestBuildOverrideSettingsMergesFilterCondition(t *testing.T) {
	base := domain.StreamerSettings{
		MakePredictions: true,
		Bet: domain.BetSettings{
			Strategy: domain.StrategySmart,
		},
	}
	base.Default()

	overrides := map[string]streamerSettingsConfig{
		"SomeStreamer": {
			Bet: betConfig{
				FilterCondition: &filterConditionConfig{
					By:    "TOTAL_POINTS",
					Where: "GT",
					Value: func() *float64 { v := 999999.0; return &v }(),
				},
			},
		},
	}

	merged := buildOverrideSettings(base, overrides)
	override, ok := merged["somestreamer"]
	if !ok {
		t.Fatalf("expected override settings keyed by lowercased streamer name")
	}
	if override.Bet.FilterCondition == nil {
		t.Fatalf("expected override filter_condition merged into settings")
	}
	if override.Bet.FilterCondition.By != "TOTAL_POINTS" || override.Bet.FilterCondition.Where != "GT" {
		t.Fatalf("expected override filter_condition TOTAL_POINTS GT, got %v %v", override.Bet.FilterCondition.By, override.Bet.FilterCondition.Where)
	}
	if override.Bet.FilterCondition.Value == nil || *override.Bet.FilterCondition.Value != 999999.0 {
		t.Fatalf("expected override Value 999999, got %#v", override.Bet.FilterCondition.Value)
	}
}

func TestBuildOverrideSettingsCanEnableClaimMomentsOverGlobalDefault(t *testing.T) {
	base := domain.StreamerSettings{
		ClaimMoments: false,
	}
	base.Default()

	enable := true
	overrides := map[string]streamerSettingsConfig{
		"SomeStreamer": {
			ClaimMoments: &enable,
		},
	}

	merged := buildOverrideSettings(base, overrides)
	got, ok := merged["somestreamer"]
	if !ok {
		t.Fatalf("expected override for somestreamer")
	}
	if !got.ClaimMoments {
		t.Fatalf("expected streamer override to enable claim_moments")
	}
}

func TestFillMissingAddsOnlyWhatIsAbsent(t *testing.T) {
	settings := map[string]interface{}{
		"username": "me",
		"bet":      map[string]interface{}{"percentage": 9.0},
		"privacy":  nil,
	}
	defaults := map[string]interface{}{
		"username": "your-twitch-username",
		"emojis":   true,
		"bet": map[string]interface{}{
			"percentage":       5,
			"filter_condition": map[string]interface{}{"by": nil},
		},
		"privacy": map[string]interface{}{"anonymize_logs": false},
	}

	if !fillMissing(settings, defaults) {
		t.Fatalf("missing keys should be reported as added")
	}
	bet := settings["bet"].(map[string]interface{})
	if settings["username"] != "me" || bet["percentage"] != 9.0 {
		t.Fatalf("values the user set must not change: %v", settings)
	}
	if settings["emojis"] != true {
		t.Fatalf("a missing top-level key should be added: %v", settings)
	}
	if _, ok := bet["filter_condition"].(map[string]interface{}); !ok {
		t.Fatalf("a missing nested group should be added: %v", bet)
	}
	if privacy, ok := settings["privacy"].(map[string]interface{}); !ok || privacy["anonymize_logs"] != false {
		t.Fatalf("a group that is not an object should be replaced by its defaults: %v", settings["privacy"])
	}
	if fillMissing(settings, defaults) {
		t.Fatalf("a complete config should need nothing added")
	}
}
