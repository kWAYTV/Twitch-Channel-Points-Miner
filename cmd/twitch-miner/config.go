package main

import "encoding/json"

type discordConfig struct {
	Webhook string   `json:"webhook"`
	Events  []string `json:"events"`
}

type config struct {
	Username          string          `json:"username"`
	AutoUpdate        bool            `json:"auto_update"`
	Humanize          bool            `json:"humanize"`
	Debug             debugLevel      `json:"debug"`
	ShowWatchQueue    bool            `json:"show_watch_queue"`
	ShowSeconds       bool            `json:"show_seconds"`
	ShowUsername      bool            `json:"show_username"`
	ClaimDrops        bool            `json:"claim_drops"`
	ClaimMoments      bool            `json:"claim_moments"`
	MakePredictions   bool            `json:"make_predictions"`
	FollowRaid        bool            `json:"follow_raid"`
	CommunityGoals    bool            `json:"community_goals"`
	Emojis            bool            `json:"emojis"`
	SaveLogs          bool            `json:"save_logs"`
	ShowGame          bool            `json:"show_game"`
	IRCMode           string          `json:"chat_presence"`
	MentionWithoutAt  bool            `json:"mention_without_at"`
	ShowDropsProgress bool            `json:"show_drops_progress"`
	Streamers         []string        `json:"streamers"`
	StreamersExclude  []string        `json:"streamers_exclude"`
	GamePriority      []string        `json:"game_priority"`
	GameExclude       []string        `json:"game_exclude"`
	WatchPriority     []string        `json:"watch_priority"`
	Bet               betConfig       `json:"bet"`
	Timezone          *string         `json:"timezone"`
	AnonymizeLogs     bool            `json:"anonymize_logs"`
	Discord           discordConfig   `json:"discord"`
	Dashboard         dashboardConfig `json:"dashboard"`
	Telegram          telegramConfig  `json:"telegram"`
	Matrix            matrixConfig    `json:"matrix"`

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
