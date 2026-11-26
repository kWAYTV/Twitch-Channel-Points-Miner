package pubsub

import (
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func (p *Client) processPredictionChannel(topic string, payload map[string]interface{}) error {
	channelID := topicChannelID(topic, predictionsChannelTopic)
	streamer := p.streamerMap[channelID]
	if streamer == nil || !streamer.Settings.MakePredictions {
		return nil
	}
	data, _ := payload["data"].(map[string]interface{})
	eventMap, _ := data["event"].(map[string]interface{})
	if eventMap == nil {
		return nil
	}
	eventID, _ := eventMap["id"].(string)
	status := strings.ToUpper(jsonvalue.String(eventMap["status"]))

	msgType := payloadType(payload)

	switch msgType {
	case "event-created":
		p.scheduleBet(streamer, eventID, status, eventMap)
	case "event-updated":
		var existing *prediction.Event
		p.predMu.Lock()
		if ev, ok := p.predictions[eventID]; ok {
			existing = ev
			existing.Status = status
			if outcomes, ok := eventMap["outcomes"].([]interface{}); ok {
				existing.UpdateOutcomes(outcomes)
			}
		}
		p.predMu.Unlock()
		if existing != nil {
			p.resolvePredictionFromChannel(existing, status, eventMap)
			return nil
		}
		// An event that was already open when the miner started, or whose
		// creation was missed, is first heard of here. The update carries the
		// whole event, so it is picked up as if it had just been created.
		if describesWholeEvent(eventMap) {
			p.scheduleBet(streamer, eventID, status, eventMap)
		}
	}
	return nil
}

// describesWholeEvent reports whether a message says when the event opened
// and for how long, which is what scheduling a bet needs.
func describesWholeEvent(event map[string]interface{}) bool {
	return jsonvalue.String(event["created_at"]) != "" && jsonvalue.Float(event["prediction_window_seconds"]) > 0
}

// scheduleBet starts following an open event and sets the bet for the end of
// its window. An event is scheduled once: Twitch can announce one more than
// once, and updates keep arriving after a bet was placed, skipped or made
// from another session of the account.
func (p *Client) scheduleBet(streamer *domain.Streamer, eventID, status string, eventMap map[string]interface{}) {
	if status != "ACTIVE" || eventID == "" {
		return
	}
	p.predMu.Lock()
	_, seen := p.seenPredictions[eventID]
	p.predMu.Unlock()
	if seen {
		return
	}
	window := jsonvalue.Float(eventMap["prediction_window_seconds"])
	eventMap["prediction_window_seconds"] = p.env.human.BetMoment(streamer.PredictionWindowSeconds(window), window)
	event := prediction.NewEvent(streamer, eventMap)
	if event == nil {
		return
	}
	if belowMinimumPoints(streamer) {
		return
	}
	wait := event.ClosingAfter(p.env.now())
	p.predMu.Lock()
	_, seen = p.seenPredictions[event.EventID]
	if !seen {
		p.seenPredictions[event.EventID] = struct{}{}
		p.predictions[event.EventID] = event
	}
	p.predMu.Unlock()
	if seen {
		return
	}
	p.env.after(wait, func() {
		p.placePrediction(event.EventID)
	})
	p.logger.EmojiEventf(":alarm_clock:", events.EventBetStart, "Place bet after %s for %s", wait.Truncate(time.Second), p.anonymizer.Shown(streamer))
}
