package chat

import (
	"fmt"
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

type logEntry struct {
	Level string
	Emoji string
	Event events.Event
	Text  string
}

// fakeLogger keeps every line logged, in order.
type fakeLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

func (l *fakeLogger) add(entry logEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

// Entries returns what has been logged so far.
func (l *fakeLogger) Entries() []logEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]logEntry(nil), l.entries...)
}

func (l *fakeLogger) Printf(format string, args ...interface{}) {
	l.add(logEntry{Level: "info", Text: fmt.Sprintf(format, args...)})
}

func (l *fakeLogger) Errorf(format string, args ...interface{}) {
	l.add(logEntry{Level: "error", Text: fmt.Sprintf(format, args...)})
}

func (l *fakeLogger) EmojiPrintf(emoji, format string, args ...interface{}) {
	l.add(logEntry{Level: "info", Emoji: emoji, Text: fmt.Sprintf(format, args...)})
}

func (l *fakeLogger) EmojiEventf(emoji string, event events.Event, format string, args ...interface{}) {
	l.add(logEntry{Level: "info", Emoji: emoji, Event: event, Text: fmt.Sprintf(format, args...)})
}
