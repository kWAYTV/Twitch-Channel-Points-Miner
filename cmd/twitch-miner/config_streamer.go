package main

import (
	"sort"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

type streamerSettingsConfig struct {
	MakePredictions *bool     `json:"make_predictions"`
	FollowRaid      *bool     `json:"follow_raid"`
	ClaimDrops      *bool     `json:"claim_drops"`
	ClaimMoments    *bool     `json:"claim_moments"`
	WatchStreak     *bool     `json:"watch_streak"`
	CommunityGoals  *bool     `json:"community_goals"`
	Bet             betConfig `json:"bet"`
	IRCMode         *string   `json:"chat_presence"`
}

func mergeStreamerSettings(base domain.StreamerSettings, override streamerSettingsConfig) domain.StreamerSettings {
	out := base
	if override.MakePredictions != nil {
		out.MakePredictions = *override.MakePredictions
	}
	if override.FollowRaid != nil {
		out.FollowRaid = *override.FollowRaid
	}
	if override.ClaimDrops != nil {
		out.ClaimDrops = *override.ClaimDrops
	}
	if override.ClaimMoments != nil {
		out.ClaimMoments = *override.ClaimMoments
	}
	if override.WatchStreak != nil {
		out.WatchStreak = *override.WatchStreak
	}
	if override.CommunityGoals != nil {
		out.CommunityGoals = *override.CommunityGoals
	}
	out.Bet = mergeBetSettings(out.Bet, override.Bet)
	if override.IRCMode != nil {
		out.IRCMode = parseChatPresence(*override.IRCMode, out.IRCMode)
	}
	out.Default()
	return out
}

func parseChatPresence(mode string, fallback domain.IRCMode) domain.IRCMode {
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case string(domain.IRCModeAlways):
		return domain.IRCModeAlways
	case string(domain.IRCModeNever):
		return domain.IRCModeNever
	case string(domain.IRCModeOffline):
		return domain.IRCModeOffline
	case string(domain.IRCModeOnline):
		return domain.IRCModeOnline
	default:
		return fallback
	}
}

func buildBaseStreamerSettings(cfg config) domain.StreamerSettings {
	streamerSettings := domain.StreamerSettings{
		MakePredictions: cfg.MakePredictions,
		FollowRaid:      cfg.FollowRaid,
		ClaimDrops:      cfg.ClaimDrops,
		ClaimMoments:    cfg.ClaimMoments,
		WatchStreak:     true,
		CommunityGoals:  cfg.CommunityGoals,
		Bet:             mergeBetSettings(domain.BetSettings{}, cfg.Bet),
		IRCMode:         parseChatPresence(cfg.IRCMode, domain.IRCModeOnline),
	}
	streamerSettings.Default()
	return streamerSettings
}

func buildOverrideSettings(base domain.StreamerSettings, overrides map[string]streamerSettingsConfig) map[string]domain.StreamerSettings {
	// Two names that differ only in capitals are one streamer. They are taken
	// in sorted order, where capitals come first, so the same one wins on
	// every run: the spelling in lower case.
	names := make([]string, 0, len(overrides))
	for name := range overrides {
		names = append(names, name)
	}
	sort.Strings(names)
	overrideSettings := make(map[string]domain.StreamerSettings, len(overrides))
	for _, name := range names {
		key := domain.LoginKey(name)
		if key == "" {
			continue
		}
		overrideSettings[key] = mergeStreamerSettings(base, overrides[name])
	}
	return overrideSettings
}
