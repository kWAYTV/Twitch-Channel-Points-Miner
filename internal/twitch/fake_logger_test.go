package twitch

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// fakeLogger keeps every line logged, each starting with its level.
type fakeLogger struct {
	Debug bool
	Deep  bool
	Lines []string
}

func (l *fakeLogger) Printf(format string, args ...interface{}) {
	l.Lines = append(l.Lines, "print: "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) Errorf(format string, args ...interface{}) {
	l.Lines = append(l.Lines, "error: "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) ErrorEventf(event events.Event, format string, args ...interface{}) {
	l.Lines = append(l.Lines, "error "+string(event)+": "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) EmojiPrintf(emoji, format string, args ...interface{}) {
	l.Lines = append(l.Lines, "emoji "+emoji+" "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) Debugf(format string, args ...interface{}) {
	l.Lines = append(l.Lines, "debug: "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) DeepDebugf(format string, args ...interface{}) {
	l.Lines = append(l.Lines, "deep: "+fmt.Sprintf(format, args...))
}

func (l *fakeLogger) DebugEnabled() bool     { return l.Debug }
func (l *fakeLogger) DeepDebugEnabled() bool { return l.Deep }
