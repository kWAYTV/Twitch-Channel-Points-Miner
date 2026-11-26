package pubsub

import (
	"fmt"
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// fakeLogger is a Logger that keeps every line it is given: Lines for what a
// user would see, Debug for debug output.
type fakeLogger struct {
	mu sync.Mutex

	DebugOn bool

	Lines []string
	Debug []string
}

func (l *fakeLogger) line(prefix, format string, args []interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Lines = append(l.Lines, prefix+": "+fmt.Sprintf(format, args...))
}

// Shown is a copy of Lines, safe to take while the client is still running.
func (l *fakeLogger) Shown() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.Lines...)
}

func (l *fakeLogger) Printf(format string, args ...interface{}) {
	l.line("info", format, args)
}

func (l *fakeLogger) Errorf(format string, args ...interface{}) {
	l.line("error", format, args)
}

func (l *fakeLogger) EmojiPrintf(emoji, format string, args ...interface{}) {
	l.line(emoji+" info", format, args)
}

func (l *fakeLogger) Eventf(event events.Event, format string, args ...interface{}) {
	l.line(string(event), format, args)
}

func (l *fakeLogger) EmojiEventf(emoji string, event events.Event, format string, args ...interface{}) {
	l.line(emoji+" "+string(event), format, args)
}

func (l *fakeLogger) ErrorEventf(event events.Event, format string, args ...interface{}) {
	l.line("error "+string(event), format, args)
}

func (l *fakeLogger) Debugf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Debug = append(l.Debug, fmt.Sprintf(format, args...))
}

func (l *fakeLogger) DebugEnabled() bool {
	return l.DebugOn
}

// fakeDeepLogger is a fakeLogger that also offers deep debug output.
type fakeDeepLogger struct {
	*fakeLogger

	DeepOn bool

	Deep []string
}

func (l *fakeDeepLogger) DeepDebugf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Deep = append(l.Deep, fmt.Sprintf(format, args...))
}

func (l *fakeDeepLogger) DeepDebugEnabled() bool {
	return l.DeepOn
}
