package auth

import (
	"net/http"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestLogin_runDeviceFlow(t *testing.T) {
	t.Run("authorized on the first poll", func(t *testing.T) {
		f := setup(t)
		f.twitch.Grant("new-token")

		err := f.login.runDeviceFlow()

		if err != nil {
			t.Fatalf("runDeviceFlow: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 1)
		if want := []time.Duration{5 * time.Second}; !reflect.DeepEqual(f.clock.Slept, want) {
			t.Errorf("slept %v, want %v", f.clock.Slept, want)
		}
	})

	t.Run("authorized on the third poll", func(t *testing.T) {
		f := setup(t)
		f.twitch.Pend()
		f.twitch.Pend()
		f.twitch.Grant("new-token")

		err := f.login.runDeviceFlow()

		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		f.checkToken(t, "new-token")
		f.checkDeviceFlow(t, 1, 3)
		if want := []time.Duration{5 * time.Second, 5 * time.Second, 5 * time.Second}; !reflect.DeepEqual(f.clock.Slept, want) {
			t.Errorf("slept %v, want %v", f.clock.Slept, want)
		}
	})

	t.Run("never authorized before the code expires", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device.Body = `{"device_code": "device-code-1", "user_code": "ABCD1234", "interval": 5, "expires_in": 12}`

		err := f.login.runDeviceFlow()

		checkError(t, err, "device code expired before authorization")
		f.checkToken(t, "")
		f.checkDeviceFlow(t, 1, 3)
		if want := []time.Duration{5 * time.Second, 5 * time.Second, 5 * time.Second}; !reflect.DeepEqual(f.clock.Slept, want) {
			t.Errorf("slept %v, want %v", f.clock.Slept, want)
		}
	})

	t.Run("a code that expires on a poll", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device.Body = `{"device_code": "device-code-1", "user_code": "ABCD1234", "interval": 5, "expires_in": 10}`

		err := f.login.runDeviceFlow()

		checkError(t, err, "device code expired before authorization")
		f.checkDeviceFlow(t, 1, 2)
	})

	t.Run("a code already expired", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device.Body = `{"device_code": "device-code-1", "user_code": "ABCD1234", "interval": 5, "expires_in": 0}`

		err := f.login.runDeviceFlow()

		checkError(t, err, "device code expired before authorization")
		f.checkDeviceFlow(t, 1, 0)
		if got, want := f.prompt.String(), "Open https://www.twitch.tv/activate and enter code: ABCD1234 (expires in 0 minutes)\n"; got != want {
			t.Errorf("prompted %q, want %q", got, want)
		}
	})

	t.Run("the device flow refused", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device = answer{Status: http.StatusBadRequest, Body: `{"message": "invalid client"}`}

		err := f.login.runDeviceFlow()

		checkError(t, err, `device flow start failed: {"message": "invalid client"}`)
		f.checkToken(t, "")
		f.checkDeviceFlow(t, 1, 0)
	})

	t.Run("a device answer that is not JSON", func(t *testing.T) {
		f := setup(t)
		f.twitch.Device = answer{Status: http.StatusOK, Body: "<html>"}

		err := f.login.runDeviceFlow()

		checkError(t, err, "invalid character '<' looking for beginning of value")
		f.checkDeviceFlow(t, 1, 0)
	})

	t.Run("the device endpoint dropping the connection", func(t *testing.T) {
		f := setup(t)
		f.twitch.Dropped["/oauth2/device"] = true

		err := f.login.runDeviceFlow()

		shape := regexp.MustCompile(`^Post "http://127\.0\.0\.1:\d+/oauth2/device": EOF$`)
		if err == nil || !shape.MatchString(err.Error()) {
			t.Errorf("error = %v, want it to match %s", err, shape)
		}
	})

	t.Run("the token endpoint dropping the connection", func(t *testing.T) {
		f := setup(t)
		f.twitch.Dropped["/oauth2/token"] = true

		err := f.login.runDeviceFlow()

		shape := regexp.MustCompile(`^Post "http://127\.0\.0\.1:\d+/oauth2/token": EOF$`)
		if err == nil || !shape.MatchString(err.Error()) {
			t.Errorf("error = %v, want it to match %s", err, shape)
		}
	})

	t.Run("a grant with no access token", func(t *testing.T) {
		f := setup(t)
		f.twitch.Tokens = []answer{{Status: http.StatusOK, Body: `{}`}}

		err := f.login.runDeviceFlow()

		checkError(t, err, "no access token received")
		f.checkToken(t, "")
	})

	t.Run("a grant that is not JSON", func(t *testing.T) {
		f := setup(t)
		f.twitch.Tokens = []answer{{Status: http.StatusOK, Body: "<html>"}}

		err := f.login.runDeviceFlow()

		checkError(t, err, "invalid character '<' looking for beginning of value")
	})
}
