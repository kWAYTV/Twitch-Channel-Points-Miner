package twitch

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// UpdateStream refreshes metadata and payload required for minute-watched events.
func (t *Client) UpdateStream(streamer *domain.Streamer) error {
	if streamer.Stream == nil {
		streamer.Stream = domain.NewStream()
	}
	if !streamer.Stream.UpdateRequired() {
		return nil
	}
	prevGame := streamer.Stream.GameName()
	info, err := t.streamInfo(streamer)
	if err != nil {
		return err
	}
	game := info.Game
	streamer.Stream.Update(
		info.StreamID,
		info.Title,
		game,
		info.Tags,
		info.ViewersCount,
		info.CreatedAt,
		twitchapi.DropID,
	)
	if info.WatchStreakComplete {
		streak.MarkCompleted(streamer)
	}

	eventProps := minuteWatchedProperties(streamer.ChannelID, streamer.Stream.BroadcastID, t.twitchLogin.UserID(), streamer.Username, game)
	gameName, hasGameName := game["name"].(string)
	if hasGameName && gameName != "" && streamer.Settings.ClaimDrops {
		// campaigns, hasGameDrops, err := t.CampaignIDsForStreamer(streamer)
		campaigns, err := t.CampaignIDsForStreamer(streamer)
		if err == nil {
			streamer.Stream.CampaignIDs = campaigns
			// streamer.Stream.CampaignsResolved = true
			// streamer.Stream.DropsActive = hasGameDrops
		}
	}
	streamer.Stream.Payload = []map[string]interface{}{
		{
			"event":      "minute-watched",
			"properties": eventProps,
		},
	}
	if t.onGameChange != nil {
		if gameName := streamer.Stream.GameName(); gameName != "" && gameName != prevGame {
			t.onGameChange(streamer, prevGame, gameName)
		}
	}
	return nil
}

// minuteWatchedProperties describes one minute of a live stream as watched by the
// user. The game is named only when Twitch gave it a name.
func minuteWatchedProperties(channelID, broadcastID, userID, username string, game map[string]interface{}) map[string]interface{} {
	eventProps := map[string]interface{}{
		"channel_id":   channelID,
		"broadcast_id": broadcastID,
		"user_id":      userID,
		"player":       "site",
		"live":         true,
		"channel":      username,
	}
	gameName, hasGameName := game["name"].(string)
	if hasGameName && gameName != "" {
		eventProps["game"] = gameName
		if id, ok := game["id"].(string); ok {
			eventProps["game_id"] = id
		}
	}
	return eventProps
}
