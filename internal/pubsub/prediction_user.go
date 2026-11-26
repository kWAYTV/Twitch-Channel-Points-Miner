package pubsub

import (
	"fmt"
	"strings"
)

func (p *Client) processPredictionUser(payload map[string]interface{}) error {
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	predictionData, _ := data["prediction"].(map[string]interface{})
	if predictionData == nil {
		return nil
	}
	eventID := fmt.Sprint(predictionData["event_id"])
	p.predMu.Lock()
	event, ok := p.predictions[eventID]
	p.predMu.Unlock()
	if !ok || event == nil {
		return nil
	}

	switch strings.ToLower(fmt.Sprint(payload["type"])) {
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
		p.predMu.Lock()
		delete(p.predictions, eventID)
		p.predMu.Unlock()
	}
	return nil
}
