package main

import "github.com/kWAYTV/twitch-points-miner/internal/domain"

func defaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"username":            "your-twitch-username",
		"auto_update":         true,
		"humanize":            true,
		"debug":               false,
		"show_watch_queue":    false,
		"show_seconds":        false,
		"show_username":       false,
		"claim_drops":         true,
		"claim_moments":       true,
		"make_predictions":    true,
		"follow_raid":         true,
		"community_goals":     false,
		"emojis":              true,
		"save_logs":           false,
		"show_game":           true,
		"chat_presence":       "ONLINE",
		"mention_without_at":  false,
		"show_drops_progress": false,
		"timezone":            nil,
		"anonymize_logs":      false,
		"discord": map[string]interface{}{
			"webhook": "",
			"events":  []interface{}{},
		},
		"dashboard":         defaultDashboardConfig(),
		"telegram":          defaultTelegramConfig(),
		"matrix":            defaultMatrixConfig(),
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
			"strategy":       string(domain.DefaultBetStrategy),
			"percentage":     domain.DefaultBetPercentage,
			"percentage_gap": domain.DefaultBetPercentageGap,
			"max_points":     domain.DefaultBetMaxPoints,
			"stealth_mode":   domain.DefaultBetStealthMode,
			"delay_mode":     string(domain.DefaultBetDelayMode),
			"delay":          domain.DefaultBetDelay,
			"minimum_points": domain.DefaultBetMinimumPoints,
			"filter_condition": map[string]interface{}{
				"by":    nil,
				"where": nil,
				"value": nil,
			},
		},
	}
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
