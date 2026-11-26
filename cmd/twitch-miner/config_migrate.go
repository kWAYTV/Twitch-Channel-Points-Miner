package main

import "github.com/kWAYTV/twitch-points-miner/internal/domain"

// Keys whose behaviour is now fixed. They are dropped from the file.
var removedSettings = []string{
	"password",
	"claim_drops_startup",
	"watch_streak_warm_start_cache",
	"disable_ssl_cert_verification",
}

// removedBetSetting is dropped from bet and from every streamer's bet.
const removedBetSetting = "deduct_stake_on_place"

// renamedSetting is a key that kept its meaning under another name. An empty
// group is the top level of the file.
type renamedSetting struct {
	group, from, to string
}

var renamedSettings = []renamedSetting{
	{"", "betting(make_predictions)", "make_predictions"},
	{"", "analytics", "dashboard"},
	{"", "disable_at_in_nickname", "mention_without_at"},
	{"", "watch_queue_logging", "show_watch_queue"},
	{"", "show_username_in_console", "show_username"},
	{"discord", "webhook_api", "webhook"},
	{"telegram", "disable_notification", "silent"},
}

// migrateSettings turns the settings of an older version into the current
// ones, in place: removed keys are dropped, renamed keys move to their new
// name, and debug_deep is folded into debug. When a file has both the old and
// the new name of a key, the new one is kept. It reports whether it changed
// anything.
func migrateSettings(settings map[string]interface{}) bool {
	changed := false
	for _, key := range removedSettings {
		changed = dropKey(settings, key) || changed
	}
	changed = dropKey(group(settings, "bet"), removedBetSetting) || changed
	for _, override := range group(settings, "streamer_overrides") {
		streamer, _ := override.(map[string]interface{})
		changed = dropKey(group(streamer, "bet"), removedBetSetting) || changed
	}
	for _, renamed := range renamedSettings {
		object := settings
		if renamed.group != "" {
			object = group(settings, renamed.group)
		}
		changed = renameKey(object, renamed.from, renamed.to) || changed
	}
	changed = liftAnonymizeLogs(settings) || changed
	changed = mergeDebugDeep(settings) || changed
	if changed {
		moveOffOldBetTiming(group(settings, "bet"))
		moveOffOldBetStrategy(group(settings, "bet"))
	}
	return changed
}

// moveOffOldBetTiming gives a file from an older version the current default
// bet timing when it still has the old default, six seconds before the end,
// which lost the bet whenever a channel closed betting early. It runs only
// while such a file is converted, so the old timing can still be chosen by
// writing it in a current file.
func moveOffOldBetTiming(bet map[string]interface{}) {
	if bet == nil || bet["delay_mode"] != "FROM_END" || bet["delay"] != 6.0 {
		return
	}
	bet["delay_mode"] = string(domain.DefaultBetDelayMode)
	bet["delay"] = domain.DefaultBetDelay
}

// moveOffOldBetStrategy gives a file from an older version the current
// default strategy when it still has the old default, SMART, which bet the
// same share of the balance on every prediction, however few viewers had
// voted. Like the timing, it is changed only while such a file is converted.
func moveOffOldBetStrategy(bet map[string]interface{}) {
	if bet == nil || bet["strategy"] != "SMART" {
		return
	}
	bet["strategy"] = string(domain.DefaultBetStrategy)
}

// group returns the object under key, or nil when there is none.
func group(settings map[string]interface{}, key string) map[string]interface{} {
	object, _ := settings[key].(map[string]interface{})
	return object
}

func dropKey(object map[string]interface{}, key string) bool {
	if _, ok := object[key]; !ok {
		return false
	}
	delete(object, key)
	return true
}

// renameKey moves the value under from to to, unless to is already set.
func renameKey(object map[string]interface{}, from, to string) bool {
	value, ok := object[from]
	if !ok {
		return false
	}
	delete(object, from)
	if _, taken := object[to]; !taken {
		object[to] = value
	}
	return true
}

// liftAnonymizeLogs moves privacy.anonymize_logs to the top level and drops
// the privacy object. A privacy object holding anything else keeps the rest.
func liftAnonymizeLogs(settings map[string]interface{}) bool {
	old, ok := settings["privacy"]
	if !ok {
		return false
	}
	privacy, _ := old.(map[string]interface{})
	changed := false
	if value, set := privacy["anonymize_logs"]; set {
		delete(privacy, "anonymize_logs")
		if _, taken := settings["anonymize_logs"]; !taken {
			settings["anonymize_logs"] = value
		}
		changed = true
	}
	if len(privacy) == 0 {
		delete(settings, "privacy")
		changed = true
	}
	return changed
}

// mergeDebugDeep folds debug_deep into debug. Deep debug only ever logged
// while debug was on too, so only that pair becomes "full".
func mergeDebugDeep(settings map[string]interface{}) bool {
	deep, ok := settings["debug_deep"]
	if !ok {
		return false
	}
	delete(settings, "debug_deep")
	if deep == true && settings["debug"] == true {
		settings["debug"] = "full"
	}
	return true
}
