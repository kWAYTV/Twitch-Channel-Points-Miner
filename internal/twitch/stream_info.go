package twitch

import (
	"fmt"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

type streamInfoResult struct {
	StreamID            string
	Title               string
	Game                map[string]interface{}
	Tags                []map[string]interface{}
	ViewersCount        int
	CreatedAt           time.Time
	WatchStreakComplete bool
}

func parseRFC3339Timestamp(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func (t *Client) fallbackStreamCreatedAt(channelID string) time.Time {
	if channelID == "" {
		return time.Time{}
	}
	op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.WithIsStreamLiveQuery)
	if op.Variables == nil {
		op.Variables = map[string]interface{}{}
	}
	op.Variables["id"] = channelID
	var resp gqlIsStreamLiveResponse
	if err := t.PostGQLDecode(op, &resp); err != nil {
		t.debugf("fallback createdAt lookup failed for channel %s: %v", channelID, err)
		return time.Time{}
	}
	if resp.Data.User == nil || resp.Data.User.Stream == nil {
		return time.Time{}
	}
	return parseRFC3339Timestamp(resp.Data.User.Stream.CreatedAt)
}

func (t *Client) streamInfoOverlay(username, channelID string) (*streamInfoResult, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.VideoPlayerStreamInfoOverlay)
	if op.Variables == nil {
		op.Variables = map[string]interface{}{}
	}
	op.Variables["channel"] = username
	var resp gqlStreamInfoOverlayResponse
	if err := t.PostGQLDecode(op, &resp); err != nil {
		return nil, err
	}
	if resp.Data.User == nil {
		return nil, fmt.Errorf("missing user data for %s: %w", username, ErrChannelNotFound)
	}
	if resp.Data.User.Stream == nil || resp.Data.User.BroadcastSettings == nil {
		return nil, ErrStreamerOffline
	}
	createdAt := parseRFC3339Timestamp(resp.Data.User.Stream.CreatedAt)
	if createdAt.IsZero() {
		createdAt = t.fallbackStreamCreatedAt(channelID)
	}
	return &streamInfoResult{
		StreamID:     resp.Data.User.Stream.ID,
		Title:        resp.Data.User.BroadcastSettings.Title,
		Game:         gameToInterfaceMap(resp.Data.User.BroadcastSettings.Game),
		Tags:         tagsToInterfaceMaps(resp.Data.User.Stream.Tags),
		ViewersCount: resp.Data.User.Stream.ViewersCount,
		CreatedAt:    createdAt,
	}, nil
}

func (t *Client) streamInfo(streamer *domain.Streamer) (*streamInfoResult, error) {
	result, err := t.streamInfoOverlay(streamer.Username, streamer.ChannelID)
	if err != nil {
		return nil, err
	}
	if streamer.Settings.WatchStreak && streamer.ChannelID != "" {
		achievementAt := t.rewardListAchievementAt(streamer.ChannelID)
		result.WatchStreakComplete = currentStreamHasCompletedWatchStreak(result.CreatedAt, achievementAt)
	}
	return result, nil
}
