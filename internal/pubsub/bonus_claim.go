package pubsub

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
)

func (p *Client) processClaimAvailable(payload map[string]interface{}, channelID string) error {
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	claim, _ := data["claim"].(map[string]interface{})
	claimID := jsonvalue.String(claim["id"])

	if channelID == "" && len(p.streamerMap) == 1 {
		for id := range p.streamerMap {
			channelID = id
		}
	}
	streamer := p.streamerMap[channelID]
	if streamer == nil {
		// A bonus on a channel this session doesn't mine, such as one another
		// session of the account is watching: nothing is wrong.
		p.debugf("claim-available ignored: not a mined channel")
		return nil
	}
	if claimID == "" {
		if p.anonymizeLogs() {
			p.logger.Errorf("claim-available ignored: [redacted]")
		} else {
			p.logger.Errorf("claim-available ignored: channel=%s claim=%s streamer=%v", channelID, claimID, streamer != nil)
			p.deepDebugf("claim-available payload: %v", payload)
		}
		return nil
	}
	// p.logger.EmojiPrintf(":gift:", "Claim bonus for %s (claim %s, channel %s)", streamer.Username, claimID, channelID)
	if err := p.account.ClaimBonus(streamer, claimID); err != nil {
		if p.anonymizeLogs() {
			p.logger.Errorf("claim bonus %s: %v", p.streamerName(streamer), err)
		} else {
			p.logger.Errorf("claim bonus %s (channel %s): %v", streamer.Username, channelID, err)
		}
		return nil
	}
	// p.logger.Printf("Claim bonus success for %s (claim %s)", streamer.Username, claimID)
	return nil
}

// pollPendingClaims proactively checks each streamer for any outstanding community-point bonuses.
func (p *Client) pollPendingClaims(stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Hour)
	defer ticker.Stop()
	p.checkPendingClaims()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			p.checkPendingClaims()
		}
	}
}

func (p *Client) checkPendingClaims() {
	for _, streamer := range p.streamers {
		if streamer == nil || streamer.Username == "" || streamer.ChannelID == "" {
			continue
		}
		if p.anonymizeLogs() {
			p.debugf("Checking pending bonus for %s", p.streamerName(streamer))
		} else {
			p.debugf("Checking pending bonus for %s (%s)", streamer.Username, streamer.ChannelID)
		}
		if _, err := p.account.LoadChannelPointsContext(streamer); err != nil {
			p.logger.Errorf("pending bonus check %s: %v", p.streamerName(streamer), err)
		}
	}
}
