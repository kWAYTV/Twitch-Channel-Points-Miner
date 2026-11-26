package twitch

import (
	"net/http"
	"testing"
)

const firefoxOnLinux = "Mozilla/5.0 (X11; Linux x86_64; rv:85.0) Gecko/20100101 Firefox/85.0"

func TestClient_GetSpadeURL(t *testing.T) {
	t.Run("a channel page", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))
		streamer := newStreamer()

		err := client.GetSpadeURL(streamer)

		if err != nil {
			t.Fatalf("GetSpadeURL error = %v, want none", err)
		}
		if want := "https://spade.example/track"; streamer.Stream.SpadeURL != want {
			t.Errorf("spade URL = %q, want %q", streamer.Stream.SpadeURL, want)
		}
		if want := (http.Header{"User-Agent": {firefoxOnLinux}}); !equal(fake.twitch.Headers, want) {
			t.Errorf("request headers = %v, want %v", fake.twitch.Headers, want)
		}
	})

	t.Run("a settings script on assets.twitch.tv", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(`<script src="https://assets.twitch.tv/config/settings.abc123.js"></script>`))
		fake.twitch.serve("GET https://assets.twitch.tv/config/settings.abc123.js", ok(settingsScriptBody))
		streamer := newStreamer()

		err := client.GetSpadeURL(streamer)

		if err != nil {
			t.Fatalf("GetSpadeURL error = %v, want none", err)
		}
		if want := "https://spade.example/track"; streamer.Stream.SpadeURL != want {
			t.Errorf("spade URL = %q, want %q", streamer.Stream.SpadeURL, want)
		}
	})

	t.Run("a streamer with a page URL of its own", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv/popout/streamer", ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))
		streamer := newStreamer()
		streamer.StreamerURL = "https://www.twitch.tv/popout/streamer"

		err := client.GetSpadeURL(streamer)

		if err != nil {
			t.Fatalf("GetSpadeURL error = %v, want none", err)
		}
		if want := "https://spade.example/track"; streamer.Stream.SpadeURL != want {
			t.Errorf("spade URL = %q, want %q", streamer.Stream.SpadeURL, want)
		}
	})

	t.Run("a streamer with no stream yet", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))
		streamer := newStreamer()
		streamer.Stream = nil

		err := client.GetSpadeURL(streamer)

		if err != nil {
			t.Fatalf("GetSpadeURL error = %v, want none", err)
		}
		if want := "https://spade.example/track"; streamer.Stream.SpadeURL != want {
			t.Errorf("spade URL = %q, want %q", streamer.Stream.SpadeURL, want)
		}
	})

	t.Run("channel page down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, down("connection refused"))
		streamer := newStreamer()

		err := client.GetSpadeURL(streamer)

		if want := `Get "https://www.twitch.tv/streamer": connection refused`; errorText(err) != want {
			t.Errorf("GetSpadeURL error = %v, want %s", err, want)
		}
		if streamer.Stream.SpadeURL != "" {
			t.Errorf("spade URL = %q, want none", streamer.Stream.SpadeURL)
		}
	})

	t.Run("a channel page with no settings script", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(`<html></html>`))

		err := client.GetSpadeURL(newStreamer())

		if want := "settings script not found"; errorText(err) != want {
			t.Errorf("GetSpadeURL error = %v, want %s", err, want)
		}
	})

	t.Run("settings script down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, down("connection refused"))

		err := client.GetSpadeURL(newStreamer())

		if want := `Get "https://static.twitchcdn.net/config/settings.abc123.js": connection refused`; errorText(err) != want {
			t.Errorf("GetSpadeURL error = %v, want %s", err, want)
		}
	})

	t.Run("a settings script with no spade URL", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(`window.settings={}`))

		err := client.GetSpadeURL(newStreamer())

		if want := "spade url not found"; errorText(err) != want {
			t.Errorf("GetSpadeURL error = %v, want %s", err, want)
		}
	})

	t.Run("debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))

		_ = client.GetSpadeURL(newStreamer())

		wantLines := []string{
			"debug: GetSpadeURL main page for streamer status 200",
			"debug: GetSpadeURL settings for streamer status 200",
			"debug: Spade URL for streamer resolved to https://spade.example/track",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("debug logging on, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.logger.Debug = true
		fake.twitch.serve(channelPage, ok(channelPageBody))
		fake.twitch.serve(settingsScript, ok(settingsScriptBody))

		_ = client.GetSpadeURL(newStreamer())

		wantLines := []string{
			"debug: GetSpadeURL main page status 200",
			"debug: GetSpadeURL settings status 200",
			"debug: Spade URL resolved",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})
}
