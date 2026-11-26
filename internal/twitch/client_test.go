package twitch

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func TestNewClient(t *testing.T) {
	t.Run("an account that has not logged in", func(t *testing.T) {
		client, err := NewClient("miner", "agent", &fakeLogger{}, nil)

		if err != nil {
			t.Fatalf("NewClient error = %v, want none", err)
		}
		if got := client.LoginName(); got != "miner" {
			t.Errorf("LoginName() = %q, want %q", got, "miner")
		}
		if got := client.ChatToken(); got != "" {
			t.Errorf("ChatToken() = %q, want none before login", got)
		}
	})
}

func TestClient_Login(t *testing.T) {
	t.Run("a session that logs in", func(t *testing.T) {
		client, fake := setup(t)

		err := client.Login("miner")

		if err != nil {
			t.Errorf("Login error = %v, want none", err)
		}
		if want := []string{filepath.Join("cookies", "miner.json")}; !equal(fake.session.LoginPaths, want) {
			t.Errorf("cookie files = %q, want %q", fake.session.LoginPaths, want)
		}
	})

	t.Run("a session that fails to log in", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.LoginErr = errors.New("device flow refused")

		err := client.Login("miner")

		if errorText(err) != "device flow refused" {
			t.Errorf("Login error = %v, want device flow refused", err)
		}
	})
}

func TestClient_ChatToken(t *testing.T) {
	t.Run("a logged-in account", func(t *testing.T) {
		client, _ := setup(t)

		if got := client.ChatToken(); got != "token" {
			t.Errorf("ChatToken() = %q, want %q", got, "token")
		}
	})

	t.Run("a client with no session", func(t *testing.T) {
		client := newClient(clientParts{})

		if got := client.ChatToken(); got != "" {
			t.Errorf("ChatToken() = %q, want none", got)
		}
	})

	t.Run("no client", func(t *testing.T) {
		var client *Client

		if got := client.ChatToken(); got != "" {
			t.Errorf("ChatToken() = %q, want none", got)
		}
	})
}

func TestClient_UpdateClientVersion(t *testing.T) {
	const laterBuildID = "11111111-2222-4333-8444-555555555555"
	const laterHomePage = `<script>window.__twilightBuildID = "11111111-2222-4333-8444-555555555555";</script>`

	t.Run("the first call", func(t *testing.T) {
		client, _ := setup(t)

		if got := client.UpdateClientVersion(); got != buildID {
			t.Errorf("UpdateClientVersion() = %q, want %q", got, buildID)
		}
	})

	t.Run("the first call, debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true

		client.UpdateClientVersion()

		if want := []string{"debug: Client version updated to " + buildID}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a second call just under ten hours later", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", ok(homePage), ok(laterHomePage))
		client.UpdateClientVersion()
		fake.clock.Now = now.Add(10*time.Hour - time.Second)

		if got := client.UpdateClientVersion(); got != buildID {
			t.Errorf("UpdateClientVersion() = %q, want the build read before, %q", got, buildID)
		}
	})

	t.Run("a second call ten hours later", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", ok(homePage), ok(laterHomePage))
		client.UpdateClientVersion()
		fake.clock.Now = now.Add(10 * time.Hour)

		if got := client.UpdateClientVersion(); got != laterBuildID {
			t.Errorf("UpdateClientVersion() = %q, want the new build, %q", got, laterBuildID)
		}
	})

	t.Run("home page down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", down("connection refused"))

		if got := client.UpdateClientVersion(); got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q", got, twitchapi.ClientVersion)
		}
	})

	t.Run("a second call just under five minutes after the home page was down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", down("connection refused"), ok(homePage))
		client.UpdateClientVersion()
		fake.clock.Now = now.Add(5*time.Minute - time.Second)

		if got := client.UpdateClientVersion(); got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q until five minutes pass", got, twitchapi.ClientVersion)
		}
	})

	t.Run("a second call five minutes after the home page was down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", down("connection refused"), ok(homePage))
		client.UpdateClientVersion()
		fake.clock.Now = now.Add(5 * time.Minute)

		if got := client.UpdateClientVersion(); got != buildID {
			t.Errorf("UpdateClientVersion() = %q, want it read again: %q", got, buildID)
		}
	})

	t.Run("a second call five minutes after the home page answered 503", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://www.twitch.tv", code(503, ""), ok(homePage))
		client.UpdateClientVersion()
		fake.clock.Now = now.Add(5 * time.Minute)

		if got := client.UpdateClientVersion(); got != buildID {
			t.Errorf("UpdateClientVersion() = %q, want it read again: %q", got, buildID)
		}
	})

	t.Run("home page answering 503", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.serve("GET https://www.twitch.tv", code(503, homePage))

		got := client.UpdateClientVersion()

		if got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q", got, twitchapi.ClientVersion)
		}
		if want := []string{"debug: UpdateClientVersion request failed with status 503"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a home page with no build id", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.serve("GET https://www.twitch.tv", ok("<html></html>"))

		got := client.UpdateClientVersion()

		if got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q", got, twitchapi.ClientVersion)
		}
		if want := []string{"debug: UpdateClientVersion: unable to extract build id"}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a client with no HTTP client", func(t *testing.T) {
		client := newClient(clientParts{})

		if got := client.UpdateClientVersion(); got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q", got, twitchapi.ClientVersion)
		}
	})

	t.Run("no client", func(t *testing.T) {
		var client *Client

		if got := client.UpdateClientVersion(); got != twitchapi.ClientVersion {
			t.Errorf("UpdateClientVersion() = %q, want the built-in %q", got, twitchapi.ClientVersion)
		}
	})
}
