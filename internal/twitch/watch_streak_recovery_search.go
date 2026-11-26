package twitch

import (
	"fmt"
	"slices"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// listing returns the operation that lists the streamer's clips for a period
// ("LAST_DAY", "LAST_WEEK") or their vods ("videos"), and the key its results sit under.
func (r *streakRecovery) listing(source string) (mode string, op twitchapi.GQLPersistedOperation) {
	if source == "videos" {
		op = twitchapi.GQLOperations.FilterableVideoTowerVideos.With(map[string]interface{}{"channelOwnerLogin": r.streamer.Username})
		return "videos", op
	}
	op = twitchapi.GQLOperations.ClipsCardsUser.With(map[string]interface{}{
		"login":    r.streamer.Username,
		"criteria": map[string]interface{}{"filter": source},
	})
	return "clips", op
}

// watchMissedFrom pages through one listing and watches every clip or vod that
// belongs to a missed stream. It stops as soon as the streak is settled.
func (r *streakRecovery) watchMissedFrom(source string) (saved, settled bool, err error) {
	mode, op := r.listing(source)
	cursor := ""
	end, err := r.client.eachEdge(op, "data.user."+mode, func(edge map[string]interface{}) (bool, error) {
		if c, ok := edge["cursor"].(string); ok {
			cursor = c
		}
		node, _ := edge["node"].(map[string]interface{})
		broadcastIdentifier, _ := node["broadcastIdentifier"].(map[string]interface{})
		id, _ := broadcastIdentifier["id"].(string)
		if !slices.Contains(r.missed, id) {
			return false, nil
		}
		var err error
		if mode == "clips" {
			saved, settled, err = r.watchClip(node)
		} else {
			saved, settled, err = r.watchVOD(node)
		}
		return settled, err
	})
	if err != nil || settled {
		return saved, settled, err
	}
	switch end {
	case pagesMissing:
		return false, false, fmt.Errorf("missing %s data for channel %s", mode, r.streamer.ChannelID)
	case pagesCutShort:
		r.logf("empty %s edges but hasNextPage == true with cursor %s", mode, cursor)
	}
	return false, false, nil
}
