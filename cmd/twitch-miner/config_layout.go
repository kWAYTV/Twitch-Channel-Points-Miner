package main

// layout is the order config.json is written in. Keys that belong together sit
// together, and each group of top-level keys is separated by a blank line.
type layout struct {
	// groups lists the keys in order. A key the layout does not know is written
	// after them, in alphabetical order.
	groups [][]string
	// children are the layouts of nested objects, by key.
	children map[string]*layout
	// each is the layout of every nested object when the keys are free-form,
	// as the streamer names under streamer_overrides are.
	each *layout
}

var filterConditionLayout = &layout{
	groups: [][]string{{"by", "where", "value"}},
}

var betLayout = &layout{
	groups: [][]string{{
		"strategy", "percentage", "percentage_gap", "max_points", "minimum_points", "stealth_mode",
		"deduct_stake_on_place", "delay_mode", "delay", "filter_condition",
	}},
	children: map[string]*layout{"filter_condition": filterConditionLayout},
}

var streamerOverrideLayout = &layout{
	groups: [][]string{{
		"make_predictions", "follow_raid", "claim_drops", "claim_moments", "watch_streak", "community_goals",
		"chat_presence", "bet",
	}},
	children: map[string]*layout{"bet": betLayout},
}

var configLayout = &layout{
	groups: [][]string{
		// Account
		{"username", "password", "auto_update"},
		// Who to mine and in which order
		{"streamers", "streamers_exclude", "streamer_overrides", "watch_priority", "game_priority", "game_exclude"},
		// What to claim and join
		{
			"claim_drops", "claim_drops_startup", "claim_moments", "follow_raid", "community_goals", "chat_presence",
			"disable_at_in_nickname", "watch_streak_warm_start_cache",
		},
		// Predictions
		{"betting(make_predictions)", "bet"},
		// Notifications
		{"discord", "telegram"},
		// Dashboard
		{"analytics"},
		// Console and logs
		{
			"emojis", "show_seconds", "show_game", "show_drops_progress", "watch_queue_logging", "save_logs", "timezone",
			"privacy", "debug", "debug_deep", "disable_ssl_cert_verification",
		},
	},
	children: map[string]*layout{
		"streamer_overrides": {each: streamerOverrideLayout},
		"bet":                betLayout,
		"discord":            {groups: [][]string{{"webhook_api", "events"}}},
		"telegram":           {groups: [][]string{{"token", "chat_id", "events", "disable_notification"}}},
		"analytics":          {groups: [][]string{{"enabled", "host", "port"}}},
		"privacy":            {groups: [][]string{{"anonymize_logs"}}},
	},
}
