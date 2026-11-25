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
