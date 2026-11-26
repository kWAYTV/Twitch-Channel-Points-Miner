package pubsub

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
)

func (p *Client) handleMessage(raw []byte, onPong func()) error {
	var envelope map[string]interface{}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	typ, _ := envelope["type"].(string)
	switch typ {
	case "PONG":
		if onPong != nil {
			onPong()
		}
		return nil
	case "RESPONSE", "RECONNECT":
		return nil
	case "MESSAGE":
		return p.handleTopicMessage(envelope)
	default:
		return nil
	}
}

func (p *Client) handleTopicMessage(envelope map[string]interface{}) error {
	data, _ := envelope["data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	topic, _ := data["topic"].(string)
	messageStr, _ := data["message"].(string)
	if messageStr == "" {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(messageStr), &payload); err != nil {
		return err
	}
	msgType := payloadType(payload)
	family := topicFamily(topic)
	p.debugf("PubSub topic %s type=%s", family, msgType)
	p.deepDebugf("PubSub topic %s payload %s", topic, strings.TrimSpace(messageStr))
	channelID := channelIDFromPayload(payload, topic)

	switch {
	case msgType == "points-earned":
		return p.processPointsEarned(payload, channelID)
	case msgType == "claim-available":
		return p.processClaimAvailable(payload, channelID)
	case family == playbackTopic:
		return p.processPlaybackMessage(topic, payload)
	case family == raidTopic:
		return p.processRaidMessage(topic, payload)
	case family == momentsChannelTopic:
		return p.processMomentMessage(topic, payload)
	case family == predictionsChannelTopic:
		return p.processPredictionChannel(topic, payload)
	case family == predictionsUserTopic:
		return p.processPredictionUser(payload)
	case family == communityPointsChannelTopic:
		return p.processCommunityPointChannel(topic, payload)
	default:
		return nil
	}
}

// payloadType is the kind of message a topic delivered, in lower case.
func payloadType(payload map[string]interface{}) string {
	return strings.ToLower(fmt.Sprint(payload["type"]))
}

// claim/watch events are routed even when Twitch omits certain fields.
func channelIDFromPayload(payload map[string]interface{}, topic string) string {
	data, _ := payload["data"].(map[string]interface{})
	if data != nil {
		if prediction, ok := data["prediction"].(map[string]interface{}); ok {
			if id := jsonvalue.Text(prediction["channel_id"]); id != "" {
				return id
			}
		}
		if claim, ok := data["claim"].(map[string]interface{}); ok {
			if id := jsonvalue.Text(claim["channel_id"]); id != "" {
				return id
			}
		}
		if id := jsonvalue.Text(data["channel_id"]); id != "" {
			return id
		}
		if balance, ok := data["balance"].(map[string]interface{}); ok {
			if id := jsonvalue.Text(balance["channel_id"]); id != "" {
				return id
			}
		}
	}
	// A user topic ends in the account's own id, which is not a channel.
	if topic != "" && !strings.Contains(topicFamily(topic), "-user-") {
		if parts := strings.Split(topic, "."); len(parts) == 2 {
			if id := strings.TrimSpace(parts[1]); id != "" {
				return id
			}
		}
	}
	return ""
}
