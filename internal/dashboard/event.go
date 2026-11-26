package dashboard

import "strings"

// Event types kept from the miner's log events. Gains (watch streak, claim, raid)
// are not listed here: they are derived from the recorded points instead.
var storedEvents = map[string]bool{
	"STREAMER_ONLINE":  true,
	"STREAMER_OFFLINE": true,
	"BET_WIN":          true,
	"BET_LOSE":         true,
	"BET_REFUND":       true,
	"DROP_CLAIM":       true,
	"MOMENT_CLAIM":     true,
}

// storedType normalises a miner event name and reports whether it is kept.
func storedType(event string) (string, bool) {
	typ := strings.ToUpper(strings.TrimSpace(event))
	return typ, storedEvents[typ]
}

// clipMessage trims a message and cuts it to maxEventMessage characters.
func clipMessage(message string) string {
	msg := strings.TrimSpace(message)
	if r := []rune(msg); len(r) > maxEventMessage {
		msg = string(r[:maxEventMessage]) + "…"
	}
	return msg
}
