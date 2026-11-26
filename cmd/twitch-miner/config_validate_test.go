package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestValidateConfig(t *testing.T) {
	const strategies = "KELLY, SMART, MOST_VOTED, HIGH_ODDS, PERCENTAGE, SMART_MONEY, NUMBER_1, NUMBER_2, NUMBER_3, NUMBER_4, NUMBER_5, NUMBER_6, NUMBER_7, NUMBER_8"

	tests := []struct {
		name string
		cfg  config
		want string
	}{
		{"nothing set", config{}, ""},
		{"every checked setting valid", config{
			IRCMode:       "ALWAYS",
			WatchPriority: []string{"STREAK", "DROPS", "ORDER", "SUBSCRIBED", "POINTS_ASCENDING", "POINTS_DESCENDING"},
			Bet: betConfig{
				Strategy:        "SMART_MONEY",
				DelayMode:       "FROM_START",
				FilterCondition: &filterConditionConfig{By: "TOTAL_USERS", Where: "GTE"},
			},
		}, ""},
		{"a strategy in lower case with spaces around it", config{Bet: betConfig{Strategy: " kelly "}}, ""},
		{"the short names of the watch priorities", config{WatchPriority: []string{"subs", "MULTIPLIER", "POINTS_ASC", "POINTS_DESC"}}, ""},
		{"a filter with nothing in it", config{Bet: betConfig{FilterCondition: &filterConditionConfig{}}}, ""},

		{"a mistyped strategy", config{Bet: betConfig{Strategy: "KELY"}},
			`invalid config: bet.strategy "KELY" is not one of ` + strategies},
		{"a mistyped delay mode", config{Bet: betConfig{DelayMode: "FROM_THE_END"}},
			`invalid config: bet.delay_mode "FROM_THE_END" is not one of PERCENTAGE, FROM_START, FROM_END`},
		{"a mistyped chat presence", config{IRCMode: "SOMETIMES"},
			`invalid config: chat_presence "SOMETIMES" is not one of ALWAYS, NEVER, ONLINE, OFFLINE`},
		{"a mistyped watch priority among good ones", config{WatchPriority: []string{"STREAK", "DROP", "ORDER"}},
			`invalid config: watch_priority "DROP" is not one of ORDER, STREAK, DROPS, SUBSCRIBED, SUBS, MULTIPLIER, POINTS_ASC, POINTS_ASCENDING, POINTS_DESC, POINTS_DESCENDING`},
		{"a mistyped filter figure", config{Bet: betConfig{FilterCondition: &filterConditionConfig{By: "USERS", Where: "GT"}}},
			`invalid config: bet.filter_condition.by "USERS" is not one of PERCENTAGE_USERS, ODDS, ODDS_PERCENTAGE, TOP_POINTS, TOTAL_USERS, TOTAL_POINTS, DECISION_USERS, DECISION_POINTS`},
		{"a mistyped filter comparison", config{Bet: betConfig{FilterCondition: &filterConditionConfig{By: "ODDS", Where: "MORE"}}},
			`invalid config: bet.filter_condition.where "MORE" is not one of GT, LT, GTE, LTE`},

		{"a mistyped strategy for one streamer", config{StreamerOverrides: map[string]streamerSettingsConfig{
			"alice": {Bet: betConfig{Strategy: "SMRT"}},
		}}, `invalid config: streamer_overrides.alice.bet.strategy "SMRT" is not one of ` + strategies},
		{"a mistyped chat presence for one streamer", config{StreamerOverrides: map[string]streamerSettingsConfig{
			"alice": {IRCMode: ptr("ON")},
		}}, `invalid config: streamer_overrides.alice.chat_presence "ON" is not one of ALWAYS, NEVER, ONLINE, OFFLINE`},
		{"a mistyped filter for one streamer", config{StreamerOverrides: map[string]streamerSettingsConfig{
			"alice": {Bet: betConfig{FilterCondition: &filterConditionConfig{Where: "ABOVE"}}},
		}}, `invalid config: streamer_overrides.alice.bet.filter_condition.where "ABOVE" is not one of GT, LT, GTE, LTE`},
		{"a valid override", config{StreamerOverrides: map[string]streamerSettingsConfig{
			"alice": {IRCMode: ptr("never"), Bet: betConfig{Strategy: "high_odds", DelayMode: "percentage"}},
		}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorText(validateConfig(tt.cfg)); got != tt.want {
				t.Errorf("validateConfig() error = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("the defaults a first run writes", func(t *testing.T) {
		if err := validateConfig(defaultsDecoded()); err != nil {
			t.Errorf("validateConfig(defaults) error = %v, want none", err)
		}
	})

	t.Run("two streamers with mistakes: the first by name", func(t *testing.T) {
		cfg := config{StreamerOverrides: map[string]streamerSettingsConfig{
			"zed":   {Bet: betConfig{Strategy: "X"}},
			"alice": {Bet: betConfig{DelayMode: "Y"}},
		}}
		want := `invalid config: streamer_overrides.alice.bet.delay_mode "Y" is not one of PERCENTAGE, FROM_START, FROM_END`
		if got := errorText(validateConfig(cfg)); got != want {
			t.Errorf("validateConfig() error = %q, want %q", got, want)
		}
	})
}

func TestMergeBetSettings_normalisesNames(t *testing.T) {
	got := mergeBetSettings(domain.BetSettings{}, betConfig{Strategy: " kelly ", DelayMode: "from_end"})

	if got.Strategy != domain.StrategyKelly {
		t.Errorf("Strategy = %q, want %q", got.Strategy, domain.StrategyKelly)
	}
	if got.DelayMode != domain.DelayModeFromEnd {
		t.Errorf("DelayMode = %q, want %q", got.DelayMode, domain.DelayModeFromEnd)
	}
}

func TestLoadOrCreateConfig_mistypedSetting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"username": "me", "bet": {"strategy": "KELY"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadOrCreateConfig(path)

	want := `invalid config: bet.strategy "KELY" is not one of KELLY, SMART, MOST_VOTED, HIGH_ODDS, PERCENTAGE, SMART_MONEY, NUMBER_1, NUMBER_2, NUMBER_3, NUMBER_4, NUMBER_5, NUMBER_6, NUMBER_7, NUMBER_8`
	if got := errorText(err); got != want {
		t.Errorf("loadOrCreateConfig() error = %q, want %q", got, want)
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
