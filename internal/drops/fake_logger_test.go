package drops

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

type logEntry struct {
	Emoji string
	Event events.Event
	Text  string
}

// fakeLogger keeps every line logged, in order.
type fakeLogger struct {
	Entries []logEntry
}

func (l *fakeLogger) Printf(format string, args ...interface{}) {
	l.Entries = append(l.Entries, logEntry{Text: fmt.Sprintf(format, args...)})
}

func (l *fakeLogger) Eventf(event events.Event, format string, args ...interface{}) {
	l.Entries = append(l.Entries, logEntry{Event: event, Text: fmt.Sprintf(format, args...)})
}

func (l *fakeLogger) EmojiEventf(emoji string, event events.Event, format string, args ...interface{}) {
	l.Entries = append(l.Entries, logEntry{Emoji: emoji, Event: event, Text: fmt.Sprintf(format, args...)})
}
