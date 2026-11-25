package notify

import (
	"strings"
)

// describeMoment covers "Alice Claimed Moment".
func describeMoment(head string) (Message, bool) {
	name, ok := cutSuffixFold(head, " claimed moment")
	if !ok {
		return Message{}, false
	}
	if name = strings.TrimSpace(name); name == "" {
		return newMessage("Claimed a moment"), true
	}
	return newMessage("Claimed a moment on " + name), true
}
