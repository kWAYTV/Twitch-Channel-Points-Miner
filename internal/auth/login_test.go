package auth

import (
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

var now = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

const (
	activationPrompt = "Open https://www.twitch.tv/activate and enter code: ABCD1234 (expires in 30 minutes)\n"
	deviceForm       = "client_id=client-1&scopes=channel_read+chat%3Aread+user_blocks_edit+user_blocks_read+user_follows_edit+user_read"
	tokenForm        = "client_id=client-1&device_code=device-code-1&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
	userLookup       = `{"operationName":"GetIDFromLogin","variables":{"login":"miner"},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"94e82a7b1e3c21e186daa73ee2afc4b8f23bade1fbbff6fe8ac133f50a2f58ca"}}}`

	savedToken = `{
  "auth-token": {
    "value": "saved-token"
  }
}`
	newToken = `{
  "auth-token": {
    "value": "new-token"
  }
}`
)

type fixture struct {
	login       *Login
	twitch      *fakeTwitch
	clock       *fakeClock
	prompt      *bytes.Buffer
	cookiesPath string
}

// setup builds the login of the account "miner" against a fake Twitch, with
// no cookies saved yet.
func setup(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		twitch:      startFakeTwitch(t),
		clock:       &fakeClock{Time: now},
		prompt:      &bytes.Buffer{},
		cookiesPath: filepath.Join(t.TempDir(), "cookies", "miner.json"),
	}
	f.login = newLogin("client-1", "device-1", "miner", "TestAgent/1.0", twitchSide{
		gqlURL:    f.twitch.URL + "/gql",
		deviceURL: f.twitch.URL + "/oauth2/device",
		tokenURL:  f.twitch.URL + "/oauth2/token",
		now:       f.clock.Now,
		sleep:     f.clock.Sleep,
		prompt:    f.prompt,
	})
	return f
}

func (f *fixture) saveCookieFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(f.cookiesPath), 0o755); err != nil {
		t.Fatalf("make the cookies folder: %v", err)
	}
	if err := os.WriteFile(f.cookiesPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write the cookies: %v", err)
	}
}

func (f *fixture) checkCookieFile(t *testing.T, want string) {
	t.Helper()
	got, err := os.ReadFile(f.cookiesPath)
	if err != nil {
		t.Fatalf("read the cookies: %v", err)
	}
	if string(got) != want {
		t.Errorf("cookies file holds %s, want %s", got, want)
	}
}

func (f *fixture) checkNoCookieFile(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(f.cookiesPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("cookies file: %v, want %v", err, fs.ErrNotExist)
	}
}

func (f *fixture) checkToken(t *testing.T, want string) {
	t.Helper()
	if got := f.login.AuthToken(); got != want {
		t.Errorf("AuthToken() = %q, want %q", got, want)
	}
}

func (f *fixture) checkDeviceFlow(t *testing.T, starts, polls int) {
	t.Helper()
	if got := len(f.twitch.DeviceRequests()); got != starts {
		t.Errorf("device flow started %d times, want %d", got, starts)
	}
	if got := len(f.twitch.TokenRequests()); got != polls {
		t.Errorf("token asked for %d times, want %d", got, polls)
	}
}

func checkError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func oauthRequestWith(form string) oauthRequest {
	return oauthRequest{
		ContentType: "application/x-www-form-urlencoded",
		ClientID:    "client-1",
		DeviceID:    "device-1",
		UserAgent:   twitchapi.AndroidTVUserAgent,
		Origin:      "https://android.tv.twitch.tv",
		Referer:     "https://android.tv.twitch.tv/",
		Form:        form,
	}
}

func TestLogin_Login(t *testing.T) {
	t.Run("a saved token Twitch accepts", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, savedToken)
		f.twitch.Users["saved-token"] = "42"

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "saved-token")
		wantGQL := []gqlRequest{{
			ContentType:   "application/json",
			Authorization: "OAuth saved-token",
			ClientID:      "client-1",
			DeviceID:      "device-1",
			UserAgent:     "TestAgent/1.0",
			Body:          userLookup,
		}}
		if got := f.twitch.GQLRequests(); !reflect.DeepEqual(got, wantGQL) {
			t.Errorf("GQL received %+v, want %+v", got, wantGQL)
		}
		f.checkDeviceFlow(t, 0, 0)
		if got := f.prompt.String(); got != "" {
			t.Errorf("prompted %q, want nothing", got)
		}
		f.checkCookieFile(t, savedToken)
	})

	t.Run("no saved cookies", func(t *testing.T) {
		f := setup(t)
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		if got := f.prompt.String(); got != activationPrompt {
			t.Errorf("prompted %q, want %q", got, activationPrompt)
		}
		wantDevice := []oauthRequest{oauthRequestWith(deviceForm)}
		if got := f.twitch.DeviceRequests(); !reflect.DeepEqual(got, wantDevice) {
			t.Errorf("device endpoint received %+v, want %+v", got, wantDevice)
		}
		wantToken := []oauthRequest{oauthRequestWith(tokenForm)}
		if got := f.twitch.TokenRequests(); !reflect.DeepEqual(got, wantToken) {
			t.Errorf("token endpoint received %+v, want %+v", got, wantToken)
		}
		if want := []time.Duration{5 * time.Second}; !reflect.DeepEqual(f.clock.Slept, want) {
			t.Errorf("slept %v, want %v", f.clock.Slept, want)
		}
		f.checkCookieFile(t, newToken)
	})

	t.Run("a saved token Twitch rejects", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, savedToken)
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 1)
		f.checkCookieFile(t, newToken)
	})

	t.Run("saved cookies without a token", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, `{"unique_id": {"value": "abc"}}`)
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		if got := f.twitch.GQLRequests(); got != nil {
			t.Errorf("GQL received %+v, want nothing", got)
		}
		f.checkDeviceFlow(t, 1, 1)
	})

	t.Run("saved cookies that are not JSON", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, "<html>")
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkCookieFile(t, newToken)
	})

	t.Run("a saved token, GQL answering 503", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, savedToken)
		f.twitch.Users["saved-token"] = "42"
		f.twitch.GQL = &answer{Status: http.StatusServiceUnavailable, Body: "unavailable"}
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 1)
	})

	t.Run("a saved token, GQL dropping the connection", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, savedToken)
		f.twitch.Users["saved-token"] = "42"
		f.twitch.Dropped["/gql"] = true
		f.twitch.Grant("new-token")

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 1)
	})

	t.Run("the device flow refused", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device = answer{Status: http.StatusBadRequest, Body: `{"message": "invalid client"}`}

		err := f.login.Login(f.cookiesPath)

		checkError(t, err, `device flow start failed: {"message": "invalid client"}`)
		f.checkToken(t, "")
		f.checkNoCookieFile(t)
	})

	t.Run("a cookies path that cannot be written", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, "in the way")
		blocked := filepath.Join(f.cookiesPath, "miner.json")
		f.twitch.Grant("new-token")

		err := f.login.Login(blocked)

		if err == nil {
			t.Errorf("error = nil, want one")
		}
		f.checkToken(t, "new-token")
	})

	t.Run("a second login", func(t *testing.T) {
		f := setup(t)
		f.twitch.Grant("new-token")
		f.twitch.Users["new-token"] = "42"
		if err := f.login.Login(f.cookiesPath); err != nil {
			t.Fatalf("first Login: %v", err)
		}

		err := f.login.Login(f.cookiesPath)

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 1)
		f.checkCookieFile(t, newToken)
	})
}

func TestLogin_UserID(t *testing.T) {
	t.Run("after a login from a saved token", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, savedToken)
		f.twitch.Users["saved-token"] = "42"
		if err := f.login.Login(f.cookiesPath); err != nil {
			t.Fatalf("Login: %v", err)
		}

		if got := f.login.UserID(); got != "42" {
			t.Errorf("UserID() = %q, want %q", got, "42")
		}
	})

	t.Run("after a login by device code", func(t *testing.T) {
		f := setup(t)
		f.twitch.Grant("new-token")
		f.twitch.Users["new-token"] = "42"
		if err := f.login.Login(f.cookiesPath); err != nil {
			t.Fatalf("Login: %v", err)
		}

		if got := f.login.UserID(); got != "42" {
			t.Errorf("UserID() = %q, want %q", got, "42")
		}
	})

	t.Run("a saved user ID, with Twitch unreachable", func(t *testing.T) {
		f := setup(t)
		f.saveCookieFile(t, `{"auth-token": {"value": "saved-token"}, "persistent": {"value": "42"}}`)
		f.twitch.Dropped["/gql"] = true
		f.twitch.Grant("new-token")
		if err := f.login.Login(f.cookiesPath); err != nil {
			t.Fatalf("Login: %v", err)
		}

		if got := f.login.UserID(); got != "42" {
			t.Errorf("UserID() = %q, want %q", got, "42")
		}
	})

	t.Run("a user Twitch does not know", func(t *testing.T) {
		f := setup(t)
		f.twitch.Grant("new-token")
		if err := f.login.Login(f.cookiesPath); err != nil {
			t.Fatalf("Login: %v", err)
		}

		if got := f.login.UserID(); got != "" {
			t.Errorf("UserID() = %q, want %q", got, "")
		}
	})

	t.Run("before any login", func(t *testing.T) {
		f := setup(t)

		if got := f.login.UserID(); got != "" {
			t.Errorf("UserID() = %q, want %q", got, "")
		}
	})
}
