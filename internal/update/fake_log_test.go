package update

import "fmt"

// logLines collects what the updater logs.
type logLines struct {
	Lines []string
}

func (l *logLines) Logf(format string, args ...interface{}) {
	l.Lines = append(l.Lines, fmt.Sprintf(format, args...))
}
