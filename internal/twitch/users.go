package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func (t *Client) GetUserByID(channelID string) (*TwitchUserIdentity, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return nil, fmt.Errorf("missing channel id: %w", ErrChannelNotFound)
	}
	if t == nil || t.client == nil || t.twitchLogin == nil {
		return nil, fmt.Errorf("missing twitch client")
	}

	endpoint := "https://api.twitch.tv/helix/users?id=" + url.QueryEscape(channelID)
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+t.twitchLogin.AuthToken())
	req.Header.Set("Client-Id", twitchapi.ClientID)
	if t.userAgent != "" {
		req.Header.Set("User-Agent", t.userAgent)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("helix user lookup failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out helixUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("channel id %s not found: %w", channelID, ErrChannelNotFound)
	}

	user := out.Data[0]
	return &TwitchUserIdentity{
		ID:          user.ID,
		Login:       domain.LoginKey(user.Login),
		DisplayName: strings.TrimSpace(user.DisplayName),
	}, nil
}

func (t *Client) GetChannelID(login string) (string, error) {
	op := twitchapi.GQLOperations.GetIDFromLogin.With(map[string]interface{}{"login": domain.LoginKey(login)})
	resp, err := t.PostGQL(op)
	if err != nil {
		return "", err
	}
	user := jsonvalue.At(resp, "data.user.id")
	if s, ok := user.(string); ok && s != "" {
		return s, nil
	}
	return "", fmt.Errorf("user %s not found", login)
}

func (t *Client) GetFollowers(limit int, order domain.FollowersOrder) ([]string, error) {
	op := twitchapi.GQLOperations.ChannelFollows.With(map[string]interface{}{
		"limit": limit,
		"order": string(order),
	})
	var follows []string
	_, err := t.eachEdge(op, "data.user.follows", func(edge map[string]interface{}) (bool, error) {
		node, _ := edge["node"].(map[string]interface{})
		if login, _ := node["login"].(string); login != "" {
			follows = append(follows, domain.LoginKey(login))
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return follows, nil
}
