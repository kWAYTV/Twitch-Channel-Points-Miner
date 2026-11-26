package pubsub

import (
	"fmt"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func (p *Client) processPredictionChannel(topic string, payload map[string]interface{}) error {
	channelID := strings.TrimPrefix(topic, "predictions-channel-v1.")
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

	msgType := strings.ToLower(fmt.Sprint(payload["type"]))

	switch msgType {
	case "event-created":
		if status != "ACTIVE" || eventID == "" {
			return nil
		}
		window := eventMap["prediction_window_seconds"]
		eventMap["prediction_window_seconds"] = streamer.PredictionWindowSeconds(jsonvalue.Float(window))
		event := prediction.NewEvent(streamer, eventMap)
		if event == nil {
			return nil
		}
		if streamer.Settings.Bet.MinimumPoints != nil && streamer.ChannelPoints <= *streamer.Settings.Bet.MinimumPoints {
			return nil
		}
		wait := event.ClosingAfter(p.env.now())
		p.predMu.Lock()
		_, alreadyScheduled := p.predictions[event.EventID]
		if !alreadyScheduled {
			p.predictions[event.EventID] = event
		}
		p.predMu.Unlock()
		// Twitch can announce an event more than once; it gets one bet.
		if alreadyScheduled {
			return nil
		}
		p.env.after(wait, func() {
			p.placePrediction(event.EventID)
		})
		p.logger.EmojiEventf(":alarm_clock:", events.EventBetStart, "Place bet after %s for %s", wait.Truncate(time.Second), p.streamerName(streamer))
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
			p.resolvePredictionFromChannel(existing, eventMap)
		}
	}
	return nil
}
