package logging

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// dualWriter writes each line to the console as it is, and to the log file
// without its colour codes.
type dualWriter struct {
	console io.Writer
	file    io.Writer
}

func (w dualWriter) Write(p []byte) (int, error) {
	if w.console != nil {
		if _, err := w.console.Write(p); err != nil {
			return 0, err
		}
	}
	if w.file != nil {
		clean := ansiRegexp.ReplaceAll(p, nil)
		if _, err := w.file.Write(clean); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// logFileName names the log file after the user, or "miner" when logs are
// anonymous or there is no user name.
func logFileName(username string, anonymize bool) string {
	name := strings.TrimSpace(username)
	if anonymize || name == "" {
		name = "miner"
	}
	return fmt.Sprintf("%s.log", SanitizeFilename(name))
}

func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
