package notify

import (
	"strings"
)

// describeMention covers "nick at #channel wrote: text".
func describeMention(head string) (Message, bool) {
	const wroteMark = " wrote: "
	i := indexFold(head, wroteMark)
	if i < 0 {
		return Message{}, false
	}
	left := head[:i]
	wrote := strings.TrimSpace(head[i+len(wroteMark):])
	at := indexFold(left, " at #")
	if at <= 0 {
		return Message{}, false
	}
	nick := strings.TrimSpace(left[:at])
	channel := strings.TrimSpace(left[at+len(" at #"):])
	if nick == "" || channel == "" {
		return Message{}, false
	}
	return newMessage(nick+" mentioned you in #"+channel, `"`+wrote+`"`), true
}

// describeRaid covers "Joining raid from Alice to Bob".
func describeRaid(head string) (Message, bool) {
	from, to, ok := splitOnceFold(head, "joining raid from ", " to ")
	if !ok {
		return Message{}, false
	}
	return newMessage("Joined the raid from " + from + " to " + to), true
}

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
