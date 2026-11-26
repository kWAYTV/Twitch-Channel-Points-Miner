package main

import "github.com/kWAYTV/twitch-points-miner/internal/domain"

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
			"strategy":              string(domain.DefaultBetStrategy),
			"percentage":            domain.DefaultBetPercentage,
			"percentage_gap":        domain.DefaultBetPercentageGap,
			"max_points":            domain.DefaultBetMaxPoints,
			"stealth_mode":          domain.DefaultBetStealthMode,
			"deduct_stake_on_place": domain.DefaultBetDeductStakeOnPlace,
			"delay_mode":            string(domain.DefaultBetDelayMode),
			"delay":                 domain.DefaultBetDelay,
			"minimum_points":        domain.DefaultBetMinimumPoints,
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
