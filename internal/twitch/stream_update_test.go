package twitch

import (
	"fmt"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_UpdateStream(t *testing.T) {
	t.Run("a live stream", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		streamer := newStreamer()

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v", got, liveFacts)
		}
		wantPayload := []map[string]interface{}{{
			"event": "minute-watched",
			"properties": map[string]interface{}{
				"channel_id":   "123456",
				"broadcast_id": "broadcast-1",
				"user_id":      "user-id",
				"player":       "site",
				"live":         true,
				"channel":      "streamer",
				"game":         "Tanks",
				"game_id":      "game-1",
			},
		}}
		if !equal(streamer.Stream.Payload, wantPayload) {
			t.Errorf("payload = %v, want %v", streamer.Stream.Payload, wantPayload)
		}
		if want := []string{}; !equal(streamer.Stream.CampaignIDs, want) {
			t.Errorf("campaign ids = %#v, want %#v: claiming drops is off", streamer.Stream.CampaignIDs, want)
		}
		if !streamer.Stream.WatchStreakMissing || streamer.CompletedWatchStreak {
			t.Errorf("watch streak missing = %t, completed = %t, want it still missing: the setting is off", streamer.Stream.WatchStreakMissing, streamer.CompletedWatchStreak)
		}
	})

	t.Run("a streamer with no stream yet", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		streamer := newStreamer()
		streamer.Stream = nil

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v", got, liveFacts)
		}
	})

	t.Run("a stream updated a moment ago", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive), ok(overlayNoGame))
		streamer := newStreamer()
		if err := client.UpdateStream(streamer); err != nil {
			t.Fatalf("first UpdateStream error = %v, want none", err)
		}

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v, as first read", got, liveFacts)
		}
	})

	t.Run("a stream with no game", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoGame))
		streamer := newStreamer()

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		wantPayload := []map[string]interface{}{{
			"event": "minute-watched",
			"properties": map[string]interface{}{
				"channel_id":   "123456",
				"broadcast_id": "broadcast-1",
				"user_id":      "user-id",
				"player":       "site",
				"live":         true,
				"channel":      "streamer",
			},
		}}
		if !equal(streamer.Stream.Payload, wantPayload) {
			t.Errorf("payload = %v, want %v", streamer.Stream.Payload, wantPayload)
		}
	})

	t.Run("an offline streamer", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayOffline))
		streamer := newStreamer()

		err := client.UpdateStream(streamer)

		if err != ErrStreamerOffline {
			t.Errorf("UpdateStream error = %v, want ErrStreamerOffline", err)
		}
		if streamer.Stream.Payload != nil {
			t.Errorf("payload = %v, want none", streamer.Stream.Payload)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, down("connection refused"))

		err := client.UpdateStream(newStreamer())

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("UpdateStream error = %v, want %s", err, want)
		}
	})

	t.Run("watch streak on, reached during this stream", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if !streamer.CompletedWatchStreak {
			t.Errorf("watch streak is not marked completed")
		}
		if streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak is still marked missing")
		}
	})

	t.Run("watch streak on, reached before this stream", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtNine))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if streamer.CompletedWatchStreak {
			t.Errorf("watch streak is marked completed")
		}
		if !streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak is no longer marked missing")
		}
	})

	t.Run("watch streak on, no start time in the stream's info", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoStartTime))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(liveSinceTen))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if !streamer.CompletedWatchStreak {
			t.Errorf("watch streak is not marked completed")
		}
		if !streamer.Stream.CreatedAt.Equal(tenOClock) {
			t.Errorf("stream started at %s, want %s", streamer.Stream.CreatedAt, tenOClock)
		}
	})

	t.Run("watch streak on, reward list down", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", rewardListVariables, down("connection refused"))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if streamer.CompletedWatchStreak || !streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak missing = %t, completed = %t, want it still missing", streamer.Stream.WatchStreakMissing, streamer.CompletedWatchStreak)
		}
		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL VideoPlayerStreamInfoOverlayChannel (stream info) | Status 200",
			`debug: GQL request failed: Post "https://gql.twitch.tv/gql": connection refused`,
			`debug: RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`,
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("watch streak off, reached during this stream", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))
		streamer := newStreamer()

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if streamer.CompletedWatchStreak || !streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak missing = %t, completed = %t, want it still missing", streamer.Stream.WatchStreakMissing, streamer.CompletedWatchStreak)
		}
	})

	t.Run("watch streak on, a streamer with no channel id", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", "", ok(streakAtSixPastTen))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true
		streamer.ChannelID = ""

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if streamer.CompletedWatchStreak || !streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak missing = %t, completed = %t, want it still missing", streamer.Stream.WatchStreakMissing, streamer.CompletedWatchStreak)
		}
	})

	t.Run("claim drops on", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":[{"id":"campaign-1"}]}}}`))
		streamer := newStreamer()
		streamer.Settings.ClaimDrops = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if want := []string{"campaign-1"}; !equal(streamer.Stream.CampaignIDs, want) {
			t.Errorf("campaign ids = %#v, want %#v", streamer.Stream.CampaignIDs, want)
		}
	})

	t.Run("claim drops on, a stream with no game", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoGame))
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":[{"id":"campaign-1"}]}}}`))
		streamer := newStreamer()
		streamer.Settings.ClaimDrops = true

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if want := []string{}; !equal(streamer.Stream.CampaignIDs, want) {
			t.Errorf("campaign ids = %#v, want %#v", streamer.Stream.CampaignIDs, want)
		}
	})

	t.Run("claim drops on, campaigns down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, down("connection refused"))
		streamer := newStreamer()
		streamer.Settings.ClaimDrops = true
		streamer.Stream.CampaignIDs = []string{"campaign-0"}

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if want := []string{"campaign-0"}; !equal(streamer.Stream.CampaignIDs, want) {
			t.Errorf("campaign ids = %#v, want %#v, as before", streamer.Stream.CampaignIDs, want)
		}
	})

	gameChanges := func(client *Client) *[]string {
		var changes []string
		client.SetGameChangeHandler(func(streamer *domain.Streamer, previous, current string) {
			changes = append(changes, fmt.Sprintf("%s: %q to %q", streamer.Username, previous, current))
		})
		return &changes
	}

	t.Run("a game change handler, a first game", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		changes := gameChanges(client)

		err := client.UpdateStream(newStreamer())

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if want := []string{`streamer: "" to "Tanks!"`}; !equal(*changes, want) {
			t.Errorf("game changes = %q, want %q", *changes, want)
		}
	})

	t.Run("a game change handler, another game than before", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		changes := gameChanges(client)
		streamer := newStreamer()
		streamer.Stream.Game = map[string]interface{}{"displayName": " Chess "}

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if want := []string{`streamer: "Chess" to "Tanks!"`}; !equal(*changes, want) {
			t.Errorf("game changes = %q, want %q", *changes, want)
		}
	})

	t.Run("a game change handler, the same game as before", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		changes := gameChanges(client)
		streamer := newStreamer()
		streamer.Stream.Game = map[string]interface{}{"displayName": "Tanks!"}

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if len(*changes) != 0 {
			t.Errorf("game changes = %q, want none", *changes)
		}
	})

	t.Run("a game change handler, a stream with no game", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoGame))
		changes := gameChanges(client)
		streamer := newStreamer()
		streamer.Stream.Game = map[string]interface{}{"displayName": "Chess"}

		err := client.UpdateStream(streamer)

		if err != nil {
			t.Fatalf("UpdateStream error = %v, want none", err)
		}
		if len(*changes) != 0 {
			t.Errorf("game changes = %q, want none", *changes)
		}
	})
}

func TestMinuteWatchedProperties(t *testing.T) {
	tests := []struct {
		name string
		game map[string]interface{}
		want map[string]interface{}
	}{
		{
			"a game with a name and an id",
			map[string]interface{}{"id": "game-1", "name": "Tanks", "displayName": "Tanks!"},
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer", "game": "Tanks", "game_id": "game-1"},
		},
		{
			"a game with a name and no id",
			map[string]interface{}{"name": "Tanks"},
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer", "game": "Tanks"},
		},
		{
			"a game with only a display name and an id",
			map[string]interface{}{"id": "game-1", "displayName": "Tanks!"},
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer"},
		},
		{
			"a game with an empty name",
			map[string]interface{}{"id": "game-1", "name": ""},
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer"},
		},
		{
			"a game whose id is a number",
			map[string]interface{}{"id": 7, "name": "Tanks"},
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer", "game": "Tanks"},
		},
		{
			"no game",
			nil,
			map[string]interface{}{"channel_id": "123456", "broadcast_id": "broadcast-1", "user_id": "user-id", "player": "site", "live": true, "channel": "streamer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minuteWatchedProperties("123456", "broadcast-1", "user-id", "streamer", tt.game)
			if !equal(got, tt.want) {
				t.Errorf("minuteWatchedProperties(%v) = %v, want %v", tt.game, got, tt.want)
			}
		})
	}
}
