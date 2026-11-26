package notify

import (
	"strings"
)

// describePresence covers "Alice (1.2k points) is Online!" and its offline twin.
func describePresence(head, game, streak, verb string) (Message, bool) {
	i, n := findLastFold(head, " is "+verb)
	if i <= 0 {
		return Message{}, false
	}
	if tail := head[i+n:]; tail != "" && tail != "!" {
		return Message{}, false
	}
	name, points := splitPoints(strings.TrimSpace(head[:i]))
	if name == "" {
		name = "The channel"
	}
	return newMessage(name+" is "+verb, playing(game), balance(points), watchStreak(streak)), true
}
