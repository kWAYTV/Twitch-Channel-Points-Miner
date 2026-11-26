package dashboard

// RecordEvent stores a miner log event, attributed to the streamer named in the message.
func (s *Store) RecordEvent(event, message string) {
	typ, ok := storedType(event)
	if !ok {
		return
	}
	msg := clipMessage(message)
	s.mu.Lock()
	defer s.mu.Unlock()
	name := mentionedStreamer(msg, s.streamers)
	switch typ {
	case "STREAMER_ONLINE", "STREAMER_OFFLINE":
		if name == "" {
			return
		}
		s.streamers[name].online = typ == "STREAMER_ONLINE"
	}
	if name == "" {
		name = globalName
	}
	s.appendLine(name+eventsSuffix, Event{T: s.now().UnixMilli(), Type: typ, Msg: msg})
}
