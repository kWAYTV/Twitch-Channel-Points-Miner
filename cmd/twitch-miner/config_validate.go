package main

import (
	"fmt"
	"sort"
	"strings"
)

// The values each named setting accepts, in the order the docs list them.
var (
	strategyNames = []string{
		"KELLY", "SMART", "MOST_VOTED", "HIGH_ODDS", "PERCENTAGE", "SMART_MONEY",
		"NUMBER_1", "NUMBER_2", "NUMBER_3", "NUMBER_4", "NUMBER_5", "NUMBER_6", "NUMBER_7", "NUMBER_8",
	}
	delayModeNames     = []string{"PERCENTAGE", "FROM_START", "FROM_END"}
	chatPresenceNames  = []string{"ALWAYS", "NEVER", "ONLINE", "OFFLINE"}
	watchPriorityNames = []string{
		"ORDER", "STREAK", "DROPS", "SUBSCRIBED", "SUBS", "MULTIPLIER",
		"POINTS_ASC", "POINTS_ASCENDING", "POINTS_DESC", "POINTS_DESCENDING",
	}
	filterFigureNames = []string{
		"PERCENTAGE_USERS", "ODDS", "ODDS_PERCENTAGE", "TOP_POINTS",
		"TOTAL_USERS", "TOTAL_POINTS", "DECISION_USERS", "DECISION_POINTS",
	}
	filterComparisonNames = []string{"GT", "LT", "GTE", "LTE"}
)

// validateConfig rejects a setting whose value is not one the program knows.
// Without it a mistyped name is not an error anywhere: a strategy nobody
// recognises bets on the highest odds, and a watch priority is ignored. An
// empty value is the default and is fine; capitals and spaces do not matter.
func validateConfig(cfg config) error {
	if err := oneOf("chat_presence", cfg.IRCMode, chatPresenceNames); err != nil {
		return err
	}
	for _, priority := range cfg.WatchPriority {
		if err := oneOf("watch_priority", priority, watchPriorityNames); err != nil {
			return err
		}
	}
	if err := validateBet("bet", cfg.Bet); err != nil {
		return err
	}
	// In name order, so the same mistake is reported on every run.
	logins := make([]string, 0, len(cfg.StreamerOverrides))
	for login := range cfg.StreamerOverrides {
		logins = append(logins, login)
	}
	sort.Strings(logins)
	for _, login := range logins {
		override := cfg.StreamerOverrides[login]
		key := "streamer_overrides." + login
		if override.IRCMode != nil {
			if err := oneOf(key+".chat_presence", *override.IRCMode, chatPresenceNames); err != nil {
				return err
			}
		}
		if err := validateBet(key+".bet", override.Bet); err != nil {
			return err
		}
	}
	return nil
}

func validateBet(key string, bet betConfig) error {
	if err := oneOf(key+".strategy", bet.Strategy, strategyNames); err != nil {
		return err
	}
	if err := oneOf(key+".delay_mode", bet.DelayMode, delayModeNames); err != nil {
		return err
	}
	if bet.FilterCondition == nil {
		return nil
	}
	if err := oneOf(key+".filter_condition.by", bet.FilterCondition.By, filterFigureNames); err != nil {
		return err
	}
	return oneOf(key+".filter_condition.where", bet.FilterCondition.Where, filterComparisonNames)
}

// oneOf reports a value that is set and is none of the names.
func oneOf(key, value string, names []string) error {
	name := settingName(value)
	if name == "" {
		return nil
	}
	for _, known := range names {
		if name == known {
			return nil
		}
	}
	return fmt.Errorf("invalid config: %s %q is not one of %s", key, value, strings.Join(names, ", "))
}

// settingName is a named setting as the program compares it: upper case,
// without the spaces around it.
func settingName(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
