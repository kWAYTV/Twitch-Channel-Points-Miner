package auth

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// offlineLogin is a login with no Twitch to reach: enough to load and save cookies.
func offlineLogin() *Login {
	return newLogin("client-1", "device-1", "miner", "TestAgent/1.0", twitchSide{})
}

func cookieFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cookies.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write the cookies: %v", err)
	}
	return path
}

// loggedIn is a login holding the given saved cookies.
func loggedIn(t *testing.T, cookies string) *Login {
	t.Helper()
	login := offlineLogin()
	if err := login.loadCookies(cookieFile(t, cookies)); err != nil {
		t.Fatalf("load the cookies: %v", err)
	}
	return login
}

// sentTo lists the cookies the login sends to a host, by name.
func sentTo(login *Login, host string) map[string]string {
	cookies := map[string]string{}
	for _, c := range login.Client().Jar.Cookies(&url.URL{Scheme: "https", Host: host}) {
		cookies[c.Name] = c.Value
	}
	return cookies
}

type session struct {
	Token      string
	UserID     string
	TwitchTV   map[string]string
	IDTwitchTV map[string]string
}

func sessionOf(login *Login) session {
	return session{
		Token:      login.AuthToken(),
		UserID:     login.UserID(),
		TwitchTV:   sentTo(login, "twitch.tv"),
		IDTwitchTV: sentTo(login, "id.twitch.tv"),
	}
}

func roundTrip(t *testing.T, login *Login) session {
	t.Helper()
	path := filepath.Join(t.TempDir(), "saved", "cookies.json")
	if err := login.saveCookies(path); err != nil {
		t.Fatalf("saveCookies: %v", err)
	}
	restored := offlineLogin()
	if err := restored.loadCookies(path); err != nil {
		t.Fatalf("loadCookies: %v", err)
	}
	return sessionOf(restored)
}

func TestLogin_cookiesRoundTrip(t *testing.T) {
	t.Run("a token and a user ID", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}, "persistent": {"value": "42"}}`)

		got := roundTrip(t, login)

		want := session{
			Token:      "tok",
			UserID:     "42",
			TwitchTV:   map[string]string{"auth-token": "tok", "persistent": "42"},
			IDTwitchTV: map[string]string{"auth-token": "tok", "persistent": "42"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("read back %+v, want %+v", got, want)
		}
	})

	t.Run("a token only", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}}`)

		got := roundTrip(t, login)

		want := session{
			Token:      "tok",
			TwitchTV:   map[string]string{"auth-token": "tok"},
			IDTwitchTV: map[string]string{"auth-token": "tok"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("read back %+v, want %+v", got, want)
		}
	})

	t.Run("a token and another cookie", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}, "unique_id": {"value": "abc"}}`)

		got := roundTrip(t, login)

		want := session{
			Token:      "tok",
			TwitchTV:   map[string]string{"auth-token": "tok", "unique_id": "abc"},
			IDTwitchTV: map[string]string{"auth-token": "tok", "unique_id": "abc"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("read back %+v, want %+v", got, want)
		}
	})

	t.Run("no cookies", func(t *testing.T) {
		login := loggedIn(t, `{}`)

		got := roundTrip(t, login)

		want := session{TwitchTV: map[string]string{}, IDTwitchTV: map[string]string{}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("read back %+v, want %+v", got, want)
		}
	})

	t.Run("a cookie for id.twitch.tv alone", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}, "unique_id": {"value": "abc", "domain": "id.twitch.tv"}}`)

		got := roundTrip(t, login)

		want := session{
			Token:      "tok",
			TwitchTV:   map[string]string{"auth-token": "tok"},
			IDTwitchTV: map[string]string{"auth-token": "tok", "unique_id": "abc"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("read back %+v, want %+v", got, want)
		}
	})
}

func TestLogin_saveCookies(t *testing.T) {
	t.Run("a token and a user ID", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}, "persistent": {"value": "42"}}`)
		path := filepath.Join(t.TempDir(), "cookies.json")

		err := login.saveCookies(path)

		if err != nil {
			t.Fatalf("saveCookies: %v", err)
		}
		want := `{
  "auth-token": {
    "value": "tok"
  },
  "persistent": {
    "value": "42"
  }
}`
		got, _ := os.ReadFile(path)
		if string(got) != want {
			t.Errorf("file holds %s, want %s", got, want)
		}
	})

	t.Run("no cookies", func(t *testing.T) {
		login := offlineLogin()
		path := filepath.Join(t.TempDir(), "cookies.json")

		err := login.saveCookies(path)

		if err != nil {
			t.Fatalf("saveCookies: %v", err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != "{}" {
			t.Errorf("file holds %s, want {}", got)
		}
	})

	t.Run("a folder that does not exist yet", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}}`)
		path := filepath.Join(t.TempDir(), "accounts", "miner", "cookies.json")

		err := login.saveCookies(path)

		if err != nil {
			t.Fatalf("saveCookies: %v", err)
		}
		want := `{
  "auth-token": {
    "value": "tok"
  }
}`
		got, _ := os.ReadFile(path)
		if string(got) != want {
			t.Errorf("file holds %s, want %s", got, want)
		}
	})

	t.Run("a file where the folder should be", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}}`)
		path := filepath.Join(cookieFile(t, "in the way"), "cookies.json")

		err := login.saveCookies(path)

		if err == nil {
			t.Errorf("error = nil, want one")
		}
	})
}

func TestLogin_loadCookies(t *testing.T) {
	t.Run("a token and a user ID", func(t *testing.T) {
		login := offlineLogin()
		path := cookieFile(t, `{"auth-token": {"value": "tok"}, "persistent": {"value": "42"}}`)

		err := login.loadCookies(path)

		if err != nil {
			t.Fatalf("loadCookies: %v", err)
		}
		want := session{
			Token:      "tok",
			UserID:     "42",
			TwitchTV:   map[string]string{"auth-token": "tok", "persistent": "42"},
			IDTwitchTV: map[string]string{"auth-token": "tok", "persistent": "42"},
		}
		if got := sessionOf(login); !reflect.DeepEqual(got, want) {
			t.Errorf("session = %+v, want %+v", got, want)
		}
	})

	t.Run("a legacy list", func(t *testing.T) {
		login := offlineLogin()
		path := cookieFile(t, `[{"name": "auth-token", "value": "tok"}, {"name": "persistent", "value": "42"}]`)

		err := login.loadCookies(path)

		if err != nil {
			t.Fatalf("loadCookies: %v", err)
		}
		want := session{
			Token:      "tok",
			UserID:     "42",
			TwitchTV:   map[string]string{"auth-token": "tok", "persistent": "42"},
			IDTwitchTV: map[string]string{"auth-token": "tok", "persistent": "42"},
		}
		if got := sessionOf(login); !reflect.DeepEqual(got, want) {
			t.Errorf("session = %+v, want %+v", got, want)
		}
	})

	t.Run("a cookie for id.twitch.tv alone", func(t *testing.T) {
		login := offlineLogin()
		path := cookieFile(t, `{"unique_id": {"value": "abc", "domain": "id.twitch.tv"}}`)

		err := login.loadCookies(path)

		if err != nil {
			t.Fatalf("loadCookies: %v", err)
		}
		want := session{
			TwitchTV:   map[string]string{},
			IDTwitchTV: map[string]string{"unique_id": "abc"},
		}
		if got := sessionOf(login); !reflect.DeepEqual(got, want) {
			t.Errorf("session = %+v, want %+v", got, want)
		}
	})

	t.Run("cookies with empty values", func(t *testing.T) {
		login := offlineLogin()
		path := cookieFile(t, `{"auth-token": {"value": ""}, "persistent": {"value": ""}, "unique_id": {"value": ""}}`)

		err := login.loadCookies(path)

		if err != nil {
			t.Fatalf("loadCookies: %v", err)
		}
		want := session{TwitchTV: map[string]string{}, IDTwitchTV: map[string]string{}}
		if got := sessionOf(login); !reflect.DeepEqual(got, want) {
			t.Errorf("session = %+v, want %+v", got, want)
		}
	})

	t.Run("a second file with another user ID", func(t *testing.T) {
		login := loggedIn(t, `{"auth-token": {"value": "tok"}, "persistent": {"value": "42"}}`)
		path := cookieFile(t, `{"auth-token": {"value": "other"}, "persistent": {"value": "99"}}`)

		err := login.loadCookies(path)

		if err != nil {
			t.Fatalf("loadCookies: %v", err)
		}
		want := session{
			Token:      "other",
			UserID:     "42",
			TwitchTV:   map[string]string{"auth-token": "other", "persistent": "42"},
			IDTwitchTV: map[string]string{"auth-token": "other", "persistent": "42"},
		}
		if got := sessionOf(login); !reflect.DeepEqual(got, want) {
			t.Errorf("session = %+v, want %+v", got, want)
		}
	})

	t.Run("no file", func(t *testing.T) {
		login := offlineLogin()

		err := login.loadCookies(filepath.Join(t.TempDir(), "cookies.json"))

		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want %v", err, fs.ErrNotExist)
		}
	})

	t.Run("a file that is not JSON", func(t *testing.T) {
		login := offlineLogin()
		path := cookieFile(t, "<html>")

		err := login.loadCookies(path)

		want := "invalid character '<' looking for beginning of value"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})
}
