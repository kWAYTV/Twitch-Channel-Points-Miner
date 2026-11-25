package notify

// describeRaid covers "Joining raid from Alice to Bob".
func describeRaid(head string) (Message, bool) {
	from, to, ok := splitOnceFold(head, "joining raid from ", " to ")
	if !ok {
		return Message{}, false
	}
	return newMessage("Joined the raid from " + from + " to " + to), true
}
