package twitch

import (
	"errors"
	"testing"
)

func TestClient_CheckStreamerOnline(t *testing.T) {
	t.Run("a live streamer", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		streamer := newStreamer()

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if !online {
			t.Errorf("CheckStreamerOnline = false, want true")
		}
		if !streamer.IsOnline {
			t.Errorf("streamer is marked offline, want online")
		}
		if !streamer.OnlineAt.Equal(now) {
			t.Errorf("online since %s, want %s", streamer.OnlineAt, now)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v", got, liveFacts)
		}
	})

	t.Run("a login in upper case with padding", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		streamer := newStreamer()
		streamer.Username = " Streamer "

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if !online {
			t.Errorf("CheckStreamerOnline = false, want true")
		}
	})

	t.Run("a live stream with the drops tag", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(`{"data":{"user":{"stream":{"id":"broadcast-1","createdAt":"2026-11-26T10:00:00Z","viewersCount":42,"tags":[{"id":"c2542d6d-cd10-4532-919b-3d19f30a768b"}]},"broadcastSettings":{"title":"Speedrun night","game":{"id":"game-1","name":"Tanks","displayName":"Tanks!"}}}}}`))
		streamer := newStreamer()

		_, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if !streamer.Stream.DropsTags {
			t.Errorf("stream is not marked as tagged for drops")
		}
	})

	t.Run("a streamer with no stream yet", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		streamer := newStreamer()
		streamer.Stream = nil

		_, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v", got, liveFacts)
		}
	})

	t.Run("a stream with no start time in its info", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoStartTime))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(liveSinceTen))
		streamer := newStreamer()

		_, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if got := factsOf(streamer.Stream); !equal(got, liveFacts) {
			t.Errorf("stream = %+v, want %+v, started at the time the live check gives", got, liveFacts)
		}
	})

	t.Run("a stream with no start time anywhere", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoStartTime))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(`{"data":{"user":{"stream":null}}}`))
		streamer := newStreamer()

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if !online {
			t.Errorf("CheckStreamerOnline = false, want true")
		}
		if !streamer.Stream.CreatedAt.IsZero() {
			t.Errorf("stream started at %s, want no start time", streamer.Stream.CreatedAt)
		}
	})

	t.Run("a stream with no start time, live check down", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoStartTime))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, down("connection refused"))
		streamer := newStreamer()

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if !online {
			t.Errorf("CheckStreamerOnline = false, want true")
		}
		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL VideoPlayerStreamInfoOverlayChannel (stream info) | Status 200",
			`debug: GQL request failed: Post "https://gql.twitch.tv/gql": connection refused`,
			`debug: fallback createdAt lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`,
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("watch streak on, reached during this stream", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))
		streamer := newStreamer()
		streamer.Settings.WatchStreak = true

		_, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if streamer.CompletedWatchStreak || !streamer.Stream.WatchStreakMissing {
			t.Errorf("watch streak missing = %t, completed = %t, want it still missing: only a stream update reads the streak", streamer.Stream.WatchStreakMissing, streamer.CompletedWatchStreak)
		}
	})

	t.Run("an offline streamer", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayOffline))
		streamer := newStreamer()
		streamer.IsOnline = true

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if online {
			t.Errorf("CheckStreamerOnline = true, want false")
		}
		if streamer.IsOnline {
			t.Errorf("streamer is marked online, want offline")
		}
		if !streamer.OfflineAt.Equal(now) {
			t.Errorf("offline since %s, want %s", streamer.OfflineAt, now)
		}
	})

	t.Run("a stream with no broadcast settings", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(`{"data":{"user":{"stream":{"id":"broadcast-1"},"broadcastSettings":null}}}`))
		streamer := newStreamer()

		online, err := client.CheckStreamerOnline(streamer)

		if err != nil {
			t.Fatalf("CheckStreamerOnline error = %v, want none", err)
		}
		if online {
			t.Errorf("CheckStreamerOnline = true, want false")
		}
	})

	t.Run("an unknown channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayNoUser))
		streamer := newStreamer()
		streamer.IsOnline = true

		online, err := client.CheckStreamerOnline(streamer)

		if !errors.Is(err, ErrChannelNotFound) {
			t.Errorf("CheckStreamerOnline error = %v, want it to wrap ErrChannelNotFound", err)
		}
		if want := "missing user data for streamer: channel not found"; errorText(err) != want {
			t.Errorf("CheckStreamerOnline error = %v, want %s", err, want)
		}
		if !online {
			t.Errorf("CheckStreamerOnline = false, want true, as the streamer was before")
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, down("connection refused"))
		streamer := newStreamer()

		online, err := client.CheckStreamerOnline(streamer)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("CheckStreamerOnline error = %v, want %s", err, want)
		}
		if online {
			t.Errorf("CheckStreamerOnline = true, want false, as the streamer was before")
		}
		if !streamer.OfflineAt.IsZero() {
			t.Errorf("offline since %s, want it left unset", streamer.OfflineAt)
		}
	})
}

func TestClient_IsStreamLive(t *testing.T) {
	t.Run("a live channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(liveSinceTen))

		live, err := client.IsStreamLive("123456")

		if err != nil {
			t.Fatalf("IsStreamLive error = %v, want none", err)
		}
		if !live {
			t.Errorf("IsStreamLive = false, want true")
		}
	})

	t.Run("an offline channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(`{"data":{"user":{"stream":null}}}`))

		live, err := client.IsStreamLive("123456")

		if err != nil {
			t.Fatalf("IsStreamLive error = %v, want none", err)
		}
		if live {
			t.Errorf("IsStreamLive = true, want false")
		}
	})

	t.Run("an unknown channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(overlayNoUser))

		live, err := client.IsStreamLive("123456")

		if err != nil {
			t.Fatalf("IsStreamLive error = %v, want none", err)
		}
		if live {
			t.Errorf("IsStreamLive = true, want false")
		}
	})

	t.Run("no channel id", func(t *testing.T) {
		client, fake := setup(t)

		live, err := client.IsStreamLive("")

		if want := "missing channel id"; errorText(err) != want {
			t.Errorf("IsStreamLive error = %v, want %s", err, want)
		}
		if live {
			t.Errorf("IsStreamLive = true, want false")
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, down("connection refused"))

		live, err := client.IsStreamLive("123456")

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("IsStreamLive error = %v, want %s", err, want)
		}
		if live {
			t.Errorf("IsStreamLive = true, want false")
		}
	})
}
