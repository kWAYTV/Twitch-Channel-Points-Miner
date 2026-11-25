package notify

import (
	"regexp"
	"strings"
)

var reLoaded = regexp.MustCompile(`(?i)^(\d+ streamers?) loaded in ([^:]+): (\d+) online(?: \((.*)\))?, (\d+) offline\.?$`)

// describeSession covers the session starting, ending and the streamer list loading.
func describeSession(head string) (Message, bool) {
	lower := strings.ToLower(head)
	switch {
	case strings.HasPrefix(lower, "start session:"):
		return newMessage("Session started"), true
	case strings.HasPrefix(lower, "ending session:"):
		return newMessage("Session ended"), true
	}
	if m := reLoaded.FindStringSubmatch(head); m != nil {
		online := m[3] + " online"
		if m[4] != "" {
			online += ": " + m[4]
		}
		return newMessage(m[1]+" loaded", online, m[5]+" offline", "Took "+strings.TrimSpace(m[2])), true
	}
	return Message{}, false
}

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
