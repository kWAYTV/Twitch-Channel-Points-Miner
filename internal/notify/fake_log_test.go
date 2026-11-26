package notify

import (
	"fmt"
)

// fakeLog stands in for the process log a notifier warns on. It keeps every line.
type fakeLog struct {
	Lines []string
}

func (l *fakeLog) Printf(format string, args ...any) {
	l.Lines = append(l.Lines, fmt.Sprintf(format, args...))
}
