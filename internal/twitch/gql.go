package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func (t *Client) PostGQL(payload interface{}) (map[string]interface{}, error) {
	return t.postGQLWithHeaders(payload, nil)
}

func (t *Client) PostGQLDecode(payload interface{}, out interface{}) error {
	return t.postGQLDecodeWithHeaders(payload, nil, out)
}

func (t *Client) postGQLWithHeaders(payload interface{}, extraHeaders map[string]string) (map[string]interface{}, error) {
	if payload == nil {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	if err := t.postGQLDecodeWithHeaders(payload, extraHeaders, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Client) postGQLDecodeWithHeaders(payload interface{}, extraHeaders map[string]string, out interface{}) error {
	if payload == nil {
		return nil
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, twitchapi.GQLOperations.URL, bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", t.twitchLogin.AuthToken()))
	req.Header.Set("Client-Id", twitchapi.ClientID)
	req.Header.Set("Client-Session-Id", t.clientSession)
	req.Header.Set("Client-Version", t.UpdateClientVersion())
	req.Header.Set("User-Agent", t.userAgent)
	req.Header.Set("X-Device-Id", t.deviceID)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		t.debugf("GQL request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	t.debugf("GQL %s | Status %d", operationLabel(payload, true), resp.StatusCode)

	deepEnabled := false
	if t.anonymizer == nil || !t.anonymizer.Enabled() {
		if deep, ok := t.logger.(deepDebugLogger); ok && deep.DeepDebugEnabled() {
			deepEnabled = true
		}
	}

	if deepEnabled {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		t.deepDebugf(
			"GQL %s | Status %d | Headers: %v | Request: %s | Response: %s",
			operationLabel(payload, false),
			resp.StatusCode,
			req.Header,
			strings.TrimSpace(string(body)),
			strings.TrimSpace(string(respBody)),
		)
		if out == nil {
			return nil
		}
		return json.Unmarshal(respBody, out)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}
