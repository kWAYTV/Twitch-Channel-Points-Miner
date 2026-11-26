package notify

import "encoding/json"

// statedFailure words why a service refused a message: the text its JSON answer
// holds under field, or the HTTP status when it holds none.
func statedFailure(status string, body []byte, field string) string {
	var answer map[string]any
	_ = json.Unmarshal(body, &answer)
	if reason, _ := answer[field].(string); reason != "" {
		return reason
	}
	return status
}
