package twitch

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseStreamPlaylist(t *testing.T) {
	tests := []struct {
		name    string
		address string
		body    string
		want    []string
		wantErr string
	}{
		{
			name:    "a master playlist of two variants",
			address: "https://usher.example/api/channel/hls/streamer.m3u8?sig=s",
			body:    "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1000000\nhttps://video.example/high.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=64000\nhttps://video.example/low.m3u8\n",
			want:    []string{"https://video.example/high.m3u8", "https://video.example/low.m3u8"},
			wantErr: "<nil>",
		},
		{
			name:    "segments named by file",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\n#EXTINF:2.0,\none.ts\n#EXTINF:2.0,\ntwo.ts\n#EXT-X-ENDLIST\n",
			want:    []string{"https://video.example/media/one.ts", "https://video.example/media/two.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "a segment named by path",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\n/other/one.ts\n",
			want:    []string{"https://video.example/other/one.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "a segment on plain http",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nhttp://plain.example/one.ts\n",
			want:    []string{"http://plain.example/one.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "blank lines, padding and Windows line ends",
			address: "https://video.example/media/playlist.m3u8",
			body:    "\r\n  #EXTM3U\r\n\r\n  one.ts  \r\n",
			want:    []string{"https://video.example/media/one.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "no line end after the last segment",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\none.ts",
			want:    []string{"https://video.example/media/one.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "a segment name outside ASCII",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nséance.ts\n",
			want:    []string{"https://video.example/media/s%C3%A9ance.ts"},
			wantErr: "<nil>",
		},
		{
			name:    "no header",
			address: "https://video.example/media/playlist.m3u8",
			body:    "one.ts\n",
			wantErr: "invalid stream playlist",
		},
		{
			name:    "a header in lower case",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#extm3u\none.ts\n",
			wantErr: "invalid stream playlist",
		},
		{
			name:    "a header and nothing else",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\n#EXT-X-ENDLIST\n",
			wantErr: "unreadable or empty stream playlist",
		},
		{
			name:    "an empty body",
			address: "https://video.example/media/playlist.m3u8",
			body:    "",
			wantErr: "unreadable or empty stream playlist",
		},
		{
			name:    "a line longer than the scanner reads",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\none.ts\n" + strings.Repeat("a", 70000) + "\n",
			wantErr: "unreadable or empty stream playlist",
		},
		{
			name:    "a segment that is not a URL",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nhttp://[::1\n",
			wantErr: "invalid stream playlist entry",
		},
		{
			name:    "a segment on ftp",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nftp://video.example/one.ts\n",
			wantErr: "invalid stream playlist entry",
		},
		{
			name:    "a segment with no host",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nhttps:///one.ts\n",
			wantErr: "invalid stream playlist entry",
		},
		{
			name:    "a segment with credentials",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\nhttps://user:secret@video.example/one.ts\n",
			wantErr: "invalid stream playlist entry",
		},
		{
			name:    "a bad segment after a good one",
			address: "https://video.example/media/playlist.m3u8",
			body:    "#EXTM3U\none.ts\nftp://video.example/two.ts\n",
			wantErr: "invalid stream playlist entry",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := url.Parse(tt.address)
			if err != nil {
				t.Fatalf("address %q: %v", tt.address, err)
			}

			entries, err := parseStreamPlaylist(address, strings.NewReader(tt.body))

			var got []string
			for _, entry := range entries {
				got = append(got, entry.String())
			}
			if !equal(got, tt.want) || errorText(err) != tt.wantErr {
				t.Errorf("parseStreamPlaylist(%q, %.60q) = %q, %v, want %q, %s", tt.address, tt.body, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// A stream that can't be watched segment by segment is logged, and the minute is
// still reported.
func TestClient_SendMinuteWatched_Playlist(t *testing.T) {
	tracked := []spadePost{{URL: "https://spade.example/track", Events: minuteWatchedEvents}}

	t.Run("a segment answering 503", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(firstSegment, code(503, ""))
		streamer := newWatchedStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream segment returned status 503"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if want := []string{"https://video.example/media/one.ts"}; !equal(fake.twitch.Watched, want) {
			t.Errorf("segments asked for = %v, want %v", fake.twitch.Watched, want)
		}
		if streamer.Stream.PlaylistURL != "" {
			t.Errorf("playlist URL = %q, want it forgotten", streamer.Stream.PlaylistURL)
		}
		if !equal(fake.twitch.Tracked, tracked) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, tracked)
		}
	})

	t.Run("a segment answering 503, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(firstSegment, code(503, ""))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer1: stream segment returned status 503"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a segment down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(firstSegment, down("connection refused"))
		streamer := newWatchedStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream segment request failed"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if want := "https://video.example/media/playlist.m3u8"; streamer.Stream.PlaylistURL != want {
			t.Errorf("playlist URL = %q, want it kept as %q", streamer.Stream.PlaylistURL, want)
		}
		if !equal(fake.twitch.Tracked, tracked) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, tracked)
		}
	})

	t.Run("playback token down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.answer("PlaybackAccessToken", playbackVariables, down("connection refused"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{`error: minute watch Streamer: Post "https://gql.twitch.tv/gql": connection refused`}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if !equal(fake.twitch.Tracked, tracked) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, tracked)
		}
	})

	t.Run("no playback token", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.answer("PlaybackAccessToken", playbackVariables, ok(`{"data":{"streamPlaybackAccessToken":null}}`))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream playback token unavailable"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a playback token with no signature", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.answer("PlaybackAccessToken", playbackVariables, ok(`{"data":{"streamPlaybackAccessToken":{"value":"token+&=","signature":""}}}`))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream playback token unavailable"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("master playlist down", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(masterPlaylist, down("connection refused"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream playlist request failed"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("master playlist answering 403", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(masterPlaylist, code(403, ""))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream playlist returned status 403"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a master playlist whose last variant is not a playlist", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(masterPlaylist, ok("#EXTM3U\nhttps://video.example/media/playlist.m3u8\nhttps://video.example/low.ts\n"))
		streamer := newWatchedStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: invalid stream master playlist"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if streamer.Stream.PlaylistURL != "" {
			t.Errorf("playlist URL = %q, want none", streamer.Stream.PlaylistURL)
		}
	})

	t.Run("media playlist answering 404", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(mediaPlaylist, code(404, ""))
		streamer := newWatchedStreamer()

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: stream playlist returned status 404"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if streamer.Stream.PlaylistURL != "" {
			t.Errorf("playlist URL = %q, want it forgotten", streamer.Stream.PlaylistURL)
		}
	})

	t.Run("a media playlist that is not a playlist", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		fake.twitch.serve(mediaPlaylist, ok("<html>"))

		err := client.SendMinuteWatched(newWatchedStreamer())

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: invalid stream playlist"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a known playlist URL that is not a URL", func(t *testing.T) {
		client, fake := setup(t)
		serveLiveStream(fake.twitch)
		streamer := newWatchedStreamer()
		streamer.Stream.PlaylistURL = "http://[::1"

		err := client.SendMinuteWatched(streamer)

		if err != nil {
			t.Fatalf("SendMinuteWatched error = %v, want none", err)
		}
		if want := []string{"error: minute watch Streamer: invalid stream playlist URL"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})
}
