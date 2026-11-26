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
		"delay_mode", "delay", "filter_condition",
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
		{"username", "auto_update", "humanize"},
		// Who to mine and in which order
		{"streamers", "streamers_exclude", "streamer_overrides", "watch_priority", "game_priority", "game_exclude"},
		// What to claim and join
		{
			"claim_drops", "claim_moments", "follow_raid", "community_goals", "chat_presence", "mention_without_at",
		},
		// Predictions
		{"make_predictions", "bet"},
		// Notifications
		{"discord", "telegram", "matrix"},
		// Dashboard
		{"dashboard"},
		// Console and logs
		{
			"emojis", "show_seconds", "show_username", "show_game", "show_drops_progress", "show_watch_queue", "save_logs",
			"timezone", "anonymize_logs", "debug",
		},
	},
	children: map[string]*layout{
		"streamer_overrides": {each: streamerOverrideLayout},
		"bet":                betLayout,
		"discord":            {groups: [][]string{{"webhook", "events"}}},
		"telegram":           {groups: [][]string{{"token", "chat_id", "events", "silent"}}},
		"matrix":             {groups: [][]string{{"homeserver", "access_token", "room_id", "events"}}},
		"dashboard":          {groups: [][]string{{"enabled", "host", "port"}}},
	},
}
