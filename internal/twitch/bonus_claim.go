package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// ClaimBonus redeems the community points bonus.
func (t *Client) ClaimBonus(streamer *domain.Streamer, claimID string) error {
	if claimID == "" {
		return fmt.Errorf("missing claim id")
	}
	if streamer == nil || streamer.ChannelID == "" {
		return fmt.Errorf("missing streamer channel id")
	}
	return t.claimBonusTV(streamer, claimID)
}

func (t *Client) claimBonusTV(streamer *domain.Streamer, claimID string) error {
	if streamer == nil || streamer.ChannelID == "" {
		return fmt.Errorf("missing streamer/channel")
	}
	if claimID == "" {
		return fmt.Errorf("missing claim id")
	}

	op := twitchapi.GQLOperations.ClaimCommunityPoints.With(map[string]interface{}{
		"input": map[string]interface{}{
			"channelID": streamer.ChannelID,
			"claimID":   claimID,
		},
	})

	reqBody, _ := json.Marshal(op)
	req, _ := http.NewRequest(http.MethodPost, twitchapi.GQLOperations.URL, bytes.NewReader(reqBody))
	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", t.twitchLogin.AuthToken()))
	req.Header.Set("Client-Id", twitchapi.ClientID)
	req.Header.Set("Client-Session-Id", t.clientSession)
	req.Header.Set("Client-Version", t.UpdateClientVersion())
	req.Header.Set("User-Agent", t.userAgent) // Android TV UA
	req.Header.Set("X-Device-Id", t.deviceID)
	req.Header.Set("Content-Type", "application/json")
	authToken := t.twitchLogin.AuthToken()
	userID := t.twitchLogin.UserID()
	if authToken != "" && userID != "" {
		req.Header.Set("Cookie", fmt.Sprintf("auth-token=%s; persistent=%s", authToken, userID))
	} else if authToken != "" {
		req.Header.Set("Cookie", fmt.Sprintf("auth-token=%s", authToken))
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("claim bonus request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	t.debugf("ClaimCommunityPoints status=%d", resp.StatusCode)
	t.deepDebugf("ClaimCommunityPoints status=%d headers=%v req=%s resp=%s", resp.StatusCode, req.Header, strings.TrimSpace(string(reqBody)), strings.TrimSpace(string(respBody)))

	if resp.StatusCode != http.StatusOK {
		if t.anonymizer.Enabled() {
			return fmt.Errorf("claim bonus status %d", resp.StatusCode)
		}
		return fmt.Errorf("claim bonus status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("claim bonus decode: %w", err)
	}

	return claimBonusFailure(result, t.anonymizer.Enabled())
}

// claimBonusFailure reads Twitch's answer to a bonus claim and returns what went
// wrong, or nil when the bonus is claimed. An anonymized error leaves the answer out.
func claimBonusFailure(result map[string]interface{}, anonymized bool) error {
	if gqlErrs, ok := result["errors"]; ok {
		if anonymized {
			return fmt.Errorf("claim bonus gql errors")
		}
		return fmt.Errorf("claim bonus gql errors: %v", gqlErrs)
	}
	if status := jsonvalue.At(result, "data.claimCommunityPoints.status"); status != nil {
		statusStr := strings.ToUpper(fmt.Sprint(status))
		if statusStr != "" && statusStr != "SUCCESS" && statusStr != "ALREADY_CLAIMED" {
			if anonymized {
				return fmt.Errorf("claim bonus status %s", statusStr)
			}
			return fmt.Errorf("claim bonus status %s (resp=%v)", statusStr, result)
		}
	}
	if msg := jsonvalue.At(result, "data.claimCommunityPoints.error.message"); msg != nil {
		if anonymized {
			return fmt.Errorf("claim bonus error")
		}
		return fmt.Errorf("claim bonus error: %v (resp=%v)", msg, result)
	}
	return nil
}
