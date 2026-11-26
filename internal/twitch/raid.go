package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// JoinRaid follows a raid target to mimic viewer behavior.
func (t *Client) JoinRaid(streamer *domain.Streamer, raidID string) error {
	if raidID == "" {
		return fmt.Errorf("missing raid id")
	}
	op := twitchapi.GQLOperations.JoinRaid.With(map[string]interface{}{"input": map[string]interface{}{"raidID": raidID}})
	_, err := t.PostGQL(op)
	return err
}
