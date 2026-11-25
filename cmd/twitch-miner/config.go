package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonfile"
)

type privacyConfig struct {
	AnonymizeLogs bool `json:"anonymize_logs"`
}

type discordConfig struct {
	WebhookAPI string   `json:"webhook_api"`
	Events     []string `json:"events"`
}

type config struct {
	Username                   string          `json:"username"`
	Password                   string          `json:"password"`
	AutoUpdate                 bool            `json:"auto_update"`
	Debug                      bool            `json:"debug"`
	DebugDeep                  bool            `json:"debug_deep"`
	WatchQueueLogging          bool            `json:"watch_queue_logging"`
	DisableSSLCertVerification bool            `json:"disable_ssl_cert_verification"`
	ShowSeconds                bool            `json:"show_seconds"`
	ClaimDropsStartup          bool            `json:"claim_drops_startup"`
	ClaimDrops                 bool            `json:"claim_drops"`
	ClaimMoments               bool            `json:"claim_moments"`
	BettingMakePredictions     bool            `json:"betting(make_predictions)"`
	FollowRaid                 bool            `json:"follow_raid"`
	CommunityGoals             bool            `json:"community_goals"`
	Emojis                     bool            `json:"emojis"`
	SaveLogs                   bool            `json:"save_logs"`
	ShowGame                   bool            `json:"show_game"`
	WatchStreakWarmStartCache  bool            `json:"watch_streak_warm_start_cache"`
	IRCMode                    string          `json:"chat_presence"`
	DisableAtInNickname        bool            `json:"disable_at_in_nickname"`
	ShowDropsProgress          bool            `json:"show_drops_progress"`
	Streamers                  []string        `json:"streamers"`
	StreamersExclude           []string        `json:"streamers_exclude"`
	GamePriority               []string        `json:"game_priority"`
	GameExclude                []string        `json:"game_exclude"`
	WatchPriority              []string        `json:"watch_priority"`
	Bet                        betConfig       `json:"bet"`
	Timezone                   *string         `json:"timezone"`
	Privacy                    privacyConfig   `json:"privacy"`
	Discord                    discordConfig   `json:"discord"`
	Analytics                  analyticsConfig `json:"analytics"`
	Telegram                   telegramConfig  `json:"telegram"`

	StreamerOverrides map[string]streamerSettingsConfig `json:"streamer_overrides"`
}

func defaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"username":                      "your-twitch-username",
		"password":                      "",
		"auto_update":                   true,
		"debug":                         false,
		"debug_deep":                    false,
		"watch_queue_logging":           false,
		"disable_ssl_cert_verification": false,
		"show_seconds":                  false,
		"claim_drops_startup":           true,
		"claim_drops":                   true,
		"claim_moments":                 true,
		"betting(make_predictions)":     true,
		"follow_raid":                   true,
		"community_goals":               false,
		"emojis":                        true,
		"save_logs":                     false,
		"show_game":                     true,
		"watch_streak_warm_start_cache": true,
		"chat_presence":                 "ONLINE",
		"disable_at_in_nickname":        false,
		"show_drops_progress":           false,
		"timezone":                      nil,
		"privacy": map[string]interface{}{
			"anonymize_logs": false,
		},
		"discord": map[string]interface{}{
			"webhook_api": "",
			"events":      []interface{}{},
		},
		"analytics":         defaultAnalyticsConfig(),
		"telegram":          defaultTelegramConfig(),
		"streamers":         []interface{}{},
		"streamers_exclude": []interface{}{},
		"game_priority":     []interface{}{},
		"game_exclude":      []interface{}{},
		"watch_priority": []interface{}{
			"STREAK",
			"DROPS",
			"ORDER",
		},
		"streamer_overrides": map[string]interface{}{},
		"bet": map[string]interface{}{
			"strategy":              "SMART",
			"percentage":            5,
			"percentage_gap":        20,
			"max_points":            50000,
			"stealth_mode":          false,
			"deduct_stake_on_place": true,
			"delay_mode":            "FROM_END",
			"delay":                 6,
			"minimum_points":        0,
			"filter_condition": map[string]interface{}{
				"by":    nil,
				"where": nil,
				"value": nil,
			},
		},
	}
}

// loadOrCreateConfig reads the config, adds any key a newer version introduced,
// and writes the file back when it was missing or incomplete.
func loadOrCreateConfig(path string) (config, error) {
	settings, found, err := readSettings(path)
	if err != nil {
		return config{}, err
	}
	added := fillMissing(settings, defaultConfig())
	if !found || added {
		if err := writeSettings(path, settings); err != nil {
			return config{}, err
		}
	}
	return decodeConfig(settings)
}

// readSettings returns the file as a map. A file that cannot be read counts as
// not found; one that is not valid JSON is an error.
func readSettings(path string) (settings map[string]interface{}, found bool, err error) {
	settings = map[string]interface{}{}
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return settings, false, nil
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, false, fmt.Errorf("invalid config: %w", err)
	}
	return settings, true, nil
}

// fillMissing copies every key settings lacks from defaults, at any depth, and
// reports whether it added anything. Values the user set are never changed,
// except a group that is not an object, which is replaced by its defaults.
func fillMissing(settings, defaults map[string]interface{}) bool {
	added := false
	for key, value := range defaults {
		group, isGroup := value.(map[string]interface{})
		current, exists := settings[key]
		if !exists {
			settings[key] = value
			added = true
			continue
		}
		if !isGroup {
			continue
		}
		nested, ok := current.(map[string]interface{})
		if !ok {
			settings[key] = group
			added = true
			continue
		}
		if fillMissing(nested, group) {
			added = true
		}
	}
	return added
}

func writeSettings(path string, settings map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return jsonfile.SaveJSON(path, settings)
}

func decodeConfig(settings map[string]interface{}) (config, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return config{}, err
	}
	var cfg config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}
