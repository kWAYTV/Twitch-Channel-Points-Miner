package domain

import (
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/textcase"
)

type Streamer struct {
	Username                     string             `json:"username"`
	ChannelID                    string             `json:"channel_id"`
	ChannelPoints                int                `json:"channel_points"`
	Settings                     StreamerSettings   `json:"settings"`
	StreamerURL                  string             `json:"-"`
	IsOnline                     bool               `json:"-"`
	PresenceKnown                bool               `json:"-"`
	OnlineAt                     time.Time          `json:"-"`
	OfflineAt                    time.Time          `json:"-"`
	CompletedWatchStreak         bool               `json:"-"`
	ResolvedStreakCarryover      bool               `json:"-"`
	ResolvedStreakCarryoverUntil time.Time          `json:"-"`
	WatchFailureCount            int                `json:"-"`
	WatchBackoffUntil            time.Time          `json:"-"`
	Stream                       *Stream            `json:"-"`
	PointsInit                   bool               `json:"-"`
	ActiveMultipliers            []ActiveMultiplier `json:"-"`
	LastRaidID                   string             `json:"-"`
	History                      map[string]*HistoryEntry
	HistoryMu                    sync.Mutex
	CommunityGoals               map[string]*CommunityGoal `json:"-"`
}

type ActiveMultiplier struct {
	Factor float64 `json:"factor"`
}

type HistoryEntry struct {
	Count  int
	Amount int
}

// DisplayName is the streamer's login with its first letter in upper case, as
// it is shown in logs.
func (s *Streamer) DisplayName() string {
	return textcase.UpperFirst(s.Username)
}

// Spend takes points off the balance, which never goes below zero.
func (s *Streamer) Spend(points int) {
	s.ChannelPoints -= points
	if s.ChannelPoints < 0 {
		s.ChannelPoints = 0
	}
}

func (s *Streamer) HasActiveMultipliers() bool {
	return len(s.ActiveMultipliers) > 0
}

func (s *Streamer) TotalMultiplier() float64 {
	total := 0.0
	for _, mult := range s.ActiveMultipliers {
		total += mult.Factor
	}
	return total
}

func (s *Streamer) PredictionWindowSeconds(predictionWindow float64) float64 {
	delay := 0.0
	if s.Settings.Bet.Delay != nil {
		delay = *s.Settings.Bet.Delay
	}
	switch s.Settings.Bet.DelayMode {
	case DelayModeFromStart:
		if delay < predictionWindow {
			return delay
		}
		return predictionWindow
	case DelayModeFromEnd:
		if predictionWindow-delay > 0 {
			return predictionWindow - delay
		}
		return 0
	case DelayModePercentage:
		return predictionWindow * delay
	default:
		return predictionWindow
	}
}
