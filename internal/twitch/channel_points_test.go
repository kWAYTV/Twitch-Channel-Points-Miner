package twitch

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/auth"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func TestLoadChannelPointsContextWrapsMissingChannel(t *testing.T) {
	twitch := &Client{
		clientVersion:  twitchapi.ClientVersion,
		versionTTL:     time.Hour,
		versionFetched: time.Now(),
		twitchLogin:    &auth.Login{Token: "token"},
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusOK, `{"data":{"community":{"channel":null}}}`), nil
			}),
		},
	}
	streamer := &domain.Streamer{Username: "oldlogin", ChannelID: "123456"}

	_, err := twitch.LoadChannelPointsContext(streamer)
	if !errors.Is(err, ErrChannelNotFound) {
		t.Fatalf("expected ErrChannelNotFound, got %v", err)
	}
}
