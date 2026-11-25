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
