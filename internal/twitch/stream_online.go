package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func (t *Client) CheckStreamerOnline(streamer *domain.Streamer) (bool, error) {
	info, err := t.streamInfoOverlay(streamer.Username, streamer.ChannelID)
	if err == ErrStreamerOffline {
		streamer.IsOnline = false
		streamer.OfflineAt = t.timeNow()
		return false, nil
	}
	if err != nil {
		return streamer.IsOnline, err
	}
	if streamer.Stream == nil {
		streamer.Stream = domain.NewStream()
	}
	streamer.Stream.Update(
		info.StreamID,
		info.Title,
		info.Game,
		info.Tags,
		info.ViewersCount,
		info.CreatedAt,
		twitchapi.DropID,
	)
	streamer.IsOnline = true
	streamer.OnlineAt = t.timeNow()
	return true, nil
}

// IsStreamLive performs a lightweight live check without refreshing stream metadata.
func (t *Client) IsStreamLive(channelID string) (bool, error) {
	if channelID == "" {
		return false, fmt.Errorf("missing channel id")
	}
	resp, err := t.liveStream(channelID)
	if err != nil {
		return false, err
	}
	if resp.Data.User == nil {
		return false, nil
	}
	return resp.Data.User.Stream != nil, nil
}

// liveStream asks Twitch for the channel's stream, which the answer has only
// while the channel is live.
func (t *Client) liveStream(channelID string) (*gqlIsStreamLiveResponse, error) {
	op := twitchapi.GQLOperations.WithIsStreamLiveQuery.With(map[string]interface{}{"id": channelID})
	var resp gqlIsStreamLiveResponse
	if err := t.PostGQLDecode(op, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
