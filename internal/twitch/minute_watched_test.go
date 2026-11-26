package twitch

import (
	"net/http"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const (
	channelPage    = "GET https://www.twitch.tv/streamer"
	settingsScript = "GET https://static.twitchcdn.net/config/settings.abc123.js"
	masterPlaylist = "GET https://usher.ttvnw.net/api/channel/hls/streamer.m3u8?sig=signature&token=token%2B%26%3D"
	mediaPlaylist  = "GET https://video.example/media/playlist.m3u8"
	firstSegment   = "HEAD https://video.example/media/one.ts"
	spadeTrack     = "POST https://spade.example/track"

	playbackVariables = `{"isLive":true,"isVod":false,"login":"streamer","platform":"web","playerType":"site","vodID":""}`

	channelPageBody    = `<script src="https://static.twitchcdn.net/config/settings.abc123.js"></script>`
	settingsScriptBody = `window.settings={"spade_url":"https://spade.example/track","other":"x"}`
	playbackToken      = `{"data":{"streamPlaybackAccessToken":{"value":"token+&=","signature":"signature"}}}`
	masterPlaylistBody = "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1000000\nhttps://video.example/high.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=64000\nhttps://video.example/media/playlist.m3u8\n"
	mediaPlaylistBody  = "#EXTM3U\n#EXTINF:2.0,\none.ts\n#EXTINF:2.0,\ntwo.ts\n#EXTINF:2.0,\nthree.ts\n#EXT-X-ENDLIST\n"

	minuteWatchedEvents = `[{"event":"minute-watched","properties":{"broadcast_id":"broadcast-1","channel":"streamer","channel_id":"123456","game":"Tanks","game_id":"game-1","live":true,"player":"site","user_id":"user-id"}}]`
)

// serveLiveStream makes the fake serve everything one watched minute of
// "streamer" asks for: stream info, a playback token, a master playlist of two
// variants, the media playlist of three segments, the segments and spade.
func serveLiveStream(twitch *fakeTwitch) {
	twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayLive))
	twitch.answer("PlaybackAccessToken", playbackVariables, ok(playbackToken))
	twitch.serve(masterPlaylist, ok(masterPlaylistBody))
	twitch.serve(mediaPlaylist, ok(mediaPlaylistBody))
	twitch.serve(firstSegment, code(200, ""))
	twitch.serve("HEAD https://video.example/media/two.ts", code(200, ""))
	twitch.serve("HEAD https://video.example/media/three.ts", code(200, ""))
	twitch.serve(spadeTrack, code(204, ""))
}

// newWatchedStreamer is newStreamer with its spade URL already known.
func newWatchedStreamer() *domain.Streamer {
	streamer := newStreamer()
	streamer.Stream.SpadeURL = "https://spade.example/track"
	return streamer
}

func TestClient_SendMinuteWatched(t *testing.T) {
	t.Run("a live stream", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		streamer := newWatchedStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []spadePost{{URL: "https://spade.example/track", Events: minuteWatchedEvents}}; !equal(fake.twitch.Tracked, want) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, want)
		}
		wantHeaders := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}, "User-Agent": {"agent"}}
		if !equal(fake.twitch.Headers, wantHeaders) {
			t.Errorf("spade request headers = %v, want %v", fake.twitch.Headers, wantHeaders)
		}
		wantWatched := []string{"https://video.example/media/one.ts", "https://video.example/media/two.ts", "https://video.example/media/three.ts"}
		if !equal(fake.twitch.Watched, wantWatched) {
			t.Errorf("segments asked for = %v, want %v", fake.twitch.Watched, wantWatched)
		}
		if want := "https://video.example/media/playlist.m3u8"; streamer.Stream.PlaylistURL != want {
			t.Errorf("playlist URL = %q, want the last variant, %q", streamer.Stream.PlaylistURL, want)
		}
		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("a playlist already known, playback token down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.answer("PlaybackAccessToken", playbackVariables, down("connection refused"))
		streamer := newWatchedStreamer()
		streamer.Stream.PlaylistURL = "https://video.example/media/playlist.m3u8"

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		wantWatched := []string{"https://video.example/media/one.ts", "https://video.example/media/two.ts", "https://video.example/media/three.ts"}
		if !equal(fake.twitch.Watched, wantWatched) {
			t.Errorf("segments asked for = %v, want %v", fake.twitch.Watched, wantWatched)
		}
		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("no spade URL yet", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))
		streamer := newStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []spadePost{{URL: "https://spade.example/track", Events: minuteWatchedEvents}}; !equal(fake.twitch.Tracked, want) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, want)
		}
	})

	t.Run("no spade URL yet, channel page down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(channelPage, down("connection refused"))
		streamer := newStreamer()

		err := client.SendMinuteWatched(streamer)

		if want := `Get "https://www.twitch.tv/streamer": connection refused`; errorText(err) != want {
			t.Errorf("SendMinuteWatched error = %v, want %s", err, want)
		}
		if len(fake.twitch.Tracked) != 0 {
			t.Errorf("events tracked = %v, want none", fake.twitch.Tracked)
		}
	})

	t.Run("an offline streamer", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.answer("VideoPlayerStreamInfoOverlayChannel", overlayVariables, ok(overlayOffline))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != ErrStreamerOffline {
			t.Errorf("SendMinuteWatched error = %v, want ErrStreamerOffline", err)
		}
		if len(fake.twitch.Tracked) != 0 {
			t.Errorf("events tracked = %v, want none", fake.twitch.Tracked)
		}
		if len(fake.twitch.Watched) != 0 {
			t.Errorf("segments asked for = %v, want none", fake.twitch.Watched)
		}
	})

	t.Run("spade answering 500", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(spadeTrack, code(500, "nope"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if want := "minute watched failed: 500 nope"; errorText(err) != want {
			t.Errorf("SendMinuteWatched error = %v, want %s", err, want)
		}
	})

	t.Run("spade answering 500, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(spadeTrack, code(500, "nope"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if want := "minute watched failed: 500"; errorText(err) != want {
			t.Errorf("SendMinuteWatched error = %v, want %s", err, want)
		}
	})

	t.Run("spade answering 200", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(spadeTrack, code(200, ""))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if want := "minute watched failed: 200 "; errorText(err) != want {
			t.Errorf("SendMinuteWatched error = %v, want %q: only 204 counts", err, want)
		}
	})

	t.Run("spade down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(spadeTrack, down("connection refused"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if want := `Post "https://spade.example/track": connection refused`; errorText(err) != want {
			t.Errorf("SendMinuteWatched error = %v, want %s", err, want)
		}
	})

	t.Run("debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		serveLiveStream(fake.twitch)

		_ = client.SendMinuteWatched(newWatchedStreamer())

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL VideoPlayerStreamInfoOverlayChannel (stream info) | Status 200",
			"debug: GQL PlaybackAccessToken (playback token) | Status 200",
			"debug: Send minute watched payload to streamer (https://spade.example/track)",
			"debug: Minute watched response for streamer: 204 ",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("debug logging on, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.logger.Debug = true
		serveLiveStream(fake.twitch)

		_ = client.SendMinuteWatched(newWatchedStreamer())

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL VideoPlayerStreamInfoOverlayChannel (stream info) | Status 200",
			"debug: GQL PlaybackAccessToken (playback token) | Status 200",
			"debug: Send minute watched payload",
			"debug: Minute watched response: 204",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})
}
