package pubsub

import "github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"

func (p *Client) processPredictionUser(payload map[string]interface{}) error {
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	predictionData, _ := data["prediction"].(map[string]interface{})
	if predictionData == nil {
		return nil
	}
	eventID := jsonvalue.Text(predictionData["event_id"])
	p.predMu.Lock()
	event, ok := p.predictions[eventID]
	if eventID != "" {
		// A bet on an event the miner isn't following was made from another
		// session of the account: the event is not to be bet on again.
		p.seenPredictions[eventID] = struct{}{}
	}
	p.predMu.Unlock()
	if !ok || event == nil {
		return nil
	}

	switch payloadType(payload) {
	case "prediction-made":
		event.BetConfirmed = true
	case "prediction-result":
		result, _ := predictionData["result"].(map[string]interface{})
		if result == nil {
			return nil
		}
		if !event.BetConfirmed {
			// Assume confirmation if Twitch skipped sending prediction-made
			event.BetConfirmed = true
		}
		p.logPredictionResult(event, result)
		p.forget(eventID)
	}
	return nil
}
