package twitch

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func (t *Client) GetSpadeURL(streamer *domain.Streamer) error {
	if streamer.Stream == nil {
		streamer.Stream = domain.NewStream()
	}
	headers := map[string]string{
		"User-Agent": twitchapi.LinuxFirefoxUserAgent,
	}
	pageURL := streamer.StreamerURL
	if pageURL == "" {
		pageURL = fmt.Sprintf("%s/%s", twitchapi.URL, streamer.Username)
	}
	mainReq, _ := http.NewRequest(http.MethodGet, pageURL, nil)
	for k, v := range headers {
		mainReq.Header.Set(k, v)
	}
	resp, err := t.client.Do(mainReq)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		t.debugf("GetSpadeURL main page status %d", resp.StatusCode)
	} else {
		t.debugf("GetSpadeURL main page for %s status %d", streamer.Username, resp.StatusCode)
	}
	match := t.settingsRegex.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return errors.New("settings script not found")
	}
	settingsReq, _ := http.NewRequest(http.MethodGet, match[1], nil)
	for k, v := range headers {
		settingsReq.Header.Set(k, v)
	}
	settingsResp, err := t.client.Do(settingsReq)
	if err != nil {
		return err
	}
	defer settingsResp.Body.Close()
	settingsBody, err := io.ReadAll(settingsResp.Body)
	if err != nil {
		return err
	}
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		t.debugf("GetSpadeURL settings status %d", settingsResp.StatusCode)
	} else {
		t.debugf("GetSpadeURL settings for %s status %d", streamer.Username, settingsResp.StatusCode)
	}
	spade := t.spadeRegex.FindStringSubmatch(string(settingsBody))
	if len(spade) < 2 {
		return errors.New("spade url not found")
	}
	streamer.Stream.SpadeURL = spade[1]
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		t.debugf("Spade URL resolved")
	} else {
		t.debugf("Spade URL for %s resolved to %s", streamer.Username, streamer.Stream.SpadeURL)
	}
	return nil
}

func (t *Client) sendSpadePayload(spadeurl string, username string, rawpayload []map[string]interface{}) error {
	raw, err := json.Marshal(rawpayload)
	if err != nil {
		return err
	}
	payload := map[string]string{
		"data": base64.StdEncoding.EncodeToString(raw),
	}
	form := url.Values{}
	form.Set("data", payload["data"])
	req, _ := http.NewRequest(http.MethodPost, spadeurl, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", t.userAgent)
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		t.debugf("Send minute watched payload")
	} else {
		t.debugf("Send minute watched payload to %s (%s)", username, spadeurl)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		t.debugf("Minute watched response: %d", resp.StatusCode)
	} else {
		t.debugf("Minute watched response for %s: %d %s", username, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if t.anonymizer != nil && t.anonymizer.Enabled() {
		return fmt.Errorf("minute watched failed: %d", resp.StatusCode)
	}
	return fmt.Errorf("minute watched failed: %d %s", resp.StatusCode, string(bodyBytes))
}
