package main

import "encoding/json"

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
