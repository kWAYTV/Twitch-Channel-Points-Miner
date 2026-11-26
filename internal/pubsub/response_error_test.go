package pubsub

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCookieFileOf(t *testing.T) {
	for _, row := range []struct {
		name     string
		username string
		want     string
	}{
		{"a login", "miner", filepath.FromSlash("cookies/miner.json")},
		{"no login", "", filepath.FromSlash("cookies/<username>.json")},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := cookieFileOf(row.username); got != row.want {
				t.Errorf("cookieFileOf(%q) = %q, want %q", row.username, got, row.want)
			}
		})
	}
}

func TestClient_logResponseError(t *testing.T) {
	t.Run("an error with its nonce", func(t *testing.T) {
		h := setup(world{})

		h.client.logResponseError(1, map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADTOPIC", "nonce": "abc"})

		if got, want := h.logger.Lines, []string{"error: PubSub[1] RESPONSE error nonce=abc: ERR_BADTOPIC"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an error with a blank nonce", func(t *testing.T) {
		h := setup(world{})

		h.client.logResponseError(1, map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADTOPIC", "nonce": " "})

		if got, want := h.logger.Lines, []string{"error: PubSub[1] RESPONSE error: ERR_BADTOPIC"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an error without a nonce", func(t *testing.T) {
		h := setup(world{})

		h.client.logResponseError(1, map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADTOPIC"})

		if got, want := h.logger.Lines, []string{"error: PubSub[1] RESPONSE error: ERR_BADTOPIC"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a rejected token", func(t *testing.T) {
		h := setup(world{})

		h.client.logResponseError(1, map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADAUTH", "nonce": "abc"})

		want := []string{
			"error: PubSub[1] RESPONSE error nonce=abc: ERR_BADAUTH",
			fmt.Sprintf("error: PubSub[1] ERR_BADAUTH: most likely you have an outdated cookie file %q. Delete this file and try again.", filepath.FromSlash("cookies/miner.json")),
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a rejected token, login unknown", func(t *testing.T) {
		h := setup(world{})
		h.account.Login = ""

		h.client.logResponseError(1, map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADAUTH", "nonce": "abc"})

		want := []string{
			"error: PubSub[1] RESPONSE error nonce=abc: ERR_BADAUTH",
			fmt.Sprintf("error: PubSub[1] ERR_BADAUTH: most likely you have an outdated cookie file %q. Delete this file and try again.", filepath.FromSlash("cookies/<username>.json")),
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	for _, row := range []struct {
		name     string
		envelope map[string]interface{}
	}{
		{"an empty error", map[string]interface{}{"type": "RESPONSE", "error": "", "nonce": "abc"}},
		{"an error of spaces", map[string]interface{}{"type": "RESPONSE", "error": "  ", "nonce": "abc"}},
		{"an error that is not text", map[string]interface{}{"type": "RESPONSE", "error": 5.0, "nonce": "abc"}},
		{"no error", map[string]interface{}{"type": "RESPONSE", "nonce": "abc"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			h := setup(world{})

			h.client.logResponseError(1, row.envelope)

			if got := h.logger.Lines; got != nil {
				t.Errorf("lines = %q, want none", got)
			}
		})
	}
}
