package miner

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

// twitchAPI is what the miner asks of Twitch for the logged-in account.
type twitchAPI interface {
	ChatToken() string
	GetChannelID(login string) (string, error)
	GetUserByID(channelID string) (*twitch.TwitchUserIdentity, error)
	LoadChannelPointsContext(streamer *domain.Streamer) (int, error)
	CheckStreamerOnline(streamer *domain.Streamer) (bool, error)
	IsStreamLive(channelID string) (bool, error)
	UpdateStream(streamer *domain.Streamer) error
	SendMinuteWatched(streamer *domain.Streamer) error
	RecoverStreak(streamer *domain.Streamer) (bool, error)
	LogExtraInfo(channelID string, online bool) string
}
