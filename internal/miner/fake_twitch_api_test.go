package miner

import (
	"fmt"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

// fakeChannel is one channel as Twitch knows it, under its current login.
type fakeChannel struct {
	ID             string
	Points         int
	Online         bool
	BroadcastID    string
	Game           string
	CreatedAt      time.Time
	StreakComplete bool
	StreakSaved    bool
}

// fakeTwitchAPI answers from Channels, keyed by login, and records what the miner sends.
type fakeTwitchAPI struct {
	mu sync.Mutex

	Token     string
	Channels  map[string]*fakeChannel
	Fail      map[string]error
	ExtraInfo string

	MinutesWatched []string
	Recoveries     []string
}

var _ twitchAPI = (*fakeTwitchAPI)(nil)

func newFakeTwitchAPI() *fakeTwitchAPI {
	return &fakeTwitchAPI{
		Channels: make(map[string]*fakeChannel),
		Fail:     make(map[string]error),
	}
}

func (f *fakeTwitchAPI) channel(login string) (*fakeChannel, error) {
	channel, ok := f.Channels[login]
	if !ok {
		return nil, fmt.Errorf("channel missing for %s: %w", login, twitch.ErrChannelNotFound)
	}
	return channel, nil
}

func (f *fakeTwitchAPI) ChatToken() string {
	return f.Token
}

func (f *fakeTwitchAPI) GetChannelID(login string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["GetChannelID"]; err != nil {
		return "", err
	}
	channel, err := f.channel(login)
	if err != nil {
		return "", err
	}
	return channel.ID, nil
}

func (f *fakeTwitchAPI) GetUserByID(channelID string) (*twitch.TwitchUserIdentity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["GetUserByID"]; err != nil {
		return nil, err
	}
	for login, channel := range f.Channels {
		if channel.ID == channelID {
			return &twitch.TwitchUserIdentity{ID: channelID, Login: login, DisplayName: login}, nil
		}
	}
	return nil, nil
}

func (f *fakeTwitchAPI) LoadChannelPointsContext(streamer *domain.Streamer) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["LoadChannelPointsContext"]; err != nil {
		return 0, err
	}
	channel, err := f.channel(streamer.Username)
	if err != nil {
		return 0, err
	}
	streamer.ChannelPoints = channel.Points
	return channel.Points, nil
}

func (f *fakeTwitchAPI) CheckStreamerOnline(streamer *domain.Streamer) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["CheckStreamerOnline"]; err != nil {
		return streamer.IsOnline, err
	}
	channel, err := f.channel(streamer.Username)
	if err != nil {
		return streamer.IsOnline, err
	}
	if !channel.Online {
		streamer.IsOnline = false
		return false, nil
	}
	f.updateStream(streamer, channel)
	streamer.IsOnline = true
	return true, nil
}

func (f *fakeTwitchAPI) IsStreamLive(channelID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["IsStreamLive"]; err != nil {
		return false, err
	}
	for _, channel := range f.Channels {
		if channel.ID == channelID {
			return channel.Online, nil
		}
	}
	return false, nil
}

func (f *fakeTwitchAPI) UpdateStream(streamer *domain.Streamer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.refreshStream(streamer)
}

func (f *fakeTwitchAPI) SendMinuteWatched(streamer *domain.Streamer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["SendMinuteWatched"]; err != nil {
		return err
	}
	if err := f.refreshStream(streamer); err != nil {
		return err
	}
	f.MinutesWatched = append(f.MinutesWatched, streamer.Username)
	return nil
}

func (f *fakeTwitchAPI) RecoverStreak(streamer *domain.Streamer) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.Fail["RecoverStreak"]; err != nil {
		return false, err
	}
	channel, err := f.channel(streamer.Username)
	if err != nil {
		return false, err
	}
	f.Recoveries = append(f.Recoveries, streamer.Username)
	return channel.StreakSaved, nil
}

func (f *fakeTwitchAPI) LogExtraInfo(channelID string, online bool) string {
	return f.ExtraInfo
}

func (f *fakeTwitchAPI) refreshStream(streamer *domain.Streamer) error {
	if err := f.Fail["UpdateStream"]; err != nil {
		return err
	}
	if streamer.Stream == nil {
		streamer.Stream = domain.NewStream()
	}
	if !streamer.Stream.UpdateRequired() {
		return nil
	}
	channel, err := f.channel(streamer.Username)
	if err != nil {
		return err
	}
	if !channel.Online {
		return twitch.ErrStreamerOffline
	}
	f.updateStream(streamer, channel)
	if channel.StreakComplete {
		streamer.CompletedWatchStreak = true
		streamer.Stream.WatchStreakMissing = false
	}
	return nil
}

func (f *fakeTwitchAPI) updateStream(streamer *domain.Streamer, channel *fakeChannel) {
	if streamer.Stream == nil {
		streamer.Stream = domain.NewStream()
	}
	game := map[string]interface{}{}
	if channel.Game != "" {
		game["displayName"] = channel.Game
	}
	streamer.Stream.Update(channel.BroadcastID, "", game, nil, 0, channel.CreatedAt, "")
}
