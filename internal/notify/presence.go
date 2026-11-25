package notify

import (
	"strings"
)

// describePresence covers "Alice (1.2k points) is Online!" and its offline twin.
func describePresence(head, game, streak, verb string) (Message, bool) {
	lower := strings.ToLower(head)
	marker := " is " + verb
	i := strings.LastIndex(lower, marker)
	if i <= 0 {
		return Message{}, false
	}
	if tail := lower[i+len(marker):]; tail != "" && tail != "!" {
		return Message{}, false
	}
	name, points := splitPoints(strings.TrimSpace(head[:i]))
	if name == "" {
		name = "The channel"
	}
	var details []string
	if game != "" {
		details = append(details, "Playing "+game)
	}
	details = append(details, balance(points))
	if streak != "" {
		details = append(details, "Watch streak "+streak)
	}
	return newMessage(name+" is "+verb, details...), true
}
