package pubsub

import (
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
)

func (p *Client) processPointsEarned(payload map[string]interface{}, channelID string) error {
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	streamer := p.streamerMap[channelID]
	if streamer == nil {
		return nil
	}
	gain, ok := pointGainOf(data, streamer.ChannelPoints)
	if !ok {
		return nil
	}
	if p.onGain != nil {
		p.onGain(streamer, gain.earned, gain.reason, gain.balance)
	}
	return nil
}

// pointGain is what a points-earned message says: why points came in, how many,
// and the balance after them.
type pointGain struct {
	reason  string
	earned  int
	balance int
}

// pointGainOf reads the gain out of a points-earned message's data. known is
// the balance to keep when the message leaves it out. ok is false when the
// message carries no gain.
func pointGainOf(data map[string]interface{}, known int) (gain pointGain, ok bool) {
	pointGainVal := jsonvalue.At(data, "point_gain")
	pointGain, _ := pointGainVal.(map[string]interface{})
	if pointGain == nil {
		return gain, false
	}
	gain.reason = strings.ToUpper(stringValue(pointGain["reason_code"]))
	gain.earned = int(jsonvalue.Float(pointGain["total_points"]))
	gain.balance = known
	if balanceValue := jsonvalue.At(data, "balance.balance"); balanceValue != nil {
		gain.balance = int(jsonvalue.Float(balanceValue))
	}
	return gain, true
}
