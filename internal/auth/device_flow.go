package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// runDeviceFlow shows an activation code, then asks Twitch for a token until
// the code is used or expires.
func (t *Login) runDeviceFlow() error {
	postData := url.Values{
		"client_id": {t.ClientID},
		"scopes":    {("channel_read chat:read user_blocks_edit user_blocks_read user_follows_edit user_read")},
	}

	addDeviceHeaders := func(req *http.Request) {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Accept-Language", "en-US")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Client-Id", t.ClientID)
		req.Header.Set("Host", "id.twitch.tv")
		req.Header.Set("Origin", "https://android.tv.twitch.tv")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Referer", "https://android.tv.twitch.tv/")
		req.Header.Set("User-Agent", twitchapi.AndroidTVUserAgent)
		req.Header.Set("X-Device-Id", t.DeviceID)
	}

	req, _ := http.NewRequest(http.MethodPost, t.twitch.deviceURL, bytes.NewBufferString(postData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addDeviceHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("device flow start failed: %s", string(body))
	}
	var payload struct {
		DeviceCode string `json:"device_code"`
		UserCode   string `json:"user_code"`
		Interval   int    `json:"interval"`
		ExpiresIn  int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	fmt.Fprintf(t.twitch.prompt, "Open https://www.twitch.tv/activate and enter code: %s (expires in %d minutes)\n", payload.UserCode, payload.ExpiresIn/60)

	tokenData := url.Values{
		"client_id":   {t.ClientID},
		"device_code": {payload.DeviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	deadline := t.twitch.now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	for t.twitch.now().Before(deadline) {
		t.twitch.sleep(time.Duration(payload.Interval) * time.Second)
		req, _ := http.NewRequest(http.MethodPost, t.twitch.tokenURL, bytes.NewBufferString(tokenData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addDeviceHeaders(req)
		resp, err := t.client.Do(req)
		if err != nil {
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var tok struct {
				AccessToken string `json:"access_token"`
			}
			if err := json.Unmarshal(body, &tok); err != nil {
				return err
			}
			if tok.AccessToken == "" {
				return errors.New("no access token received")
			}
			t.setToken(tok.AccessToken)
			return nil
		}
	}
	return errors.New("device code expired before authorization")
}
