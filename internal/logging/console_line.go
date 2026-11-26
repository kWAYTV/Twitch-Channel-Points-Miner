package logging

import (
	"fmt"
	"time"
)

// levelEnabled reports whether a line of this level is written: DEBUG needs
// debug on, DEEP needs debug and deep debug on, every other level is always written.
func levelEnabled(level string, settings LoggerSettings) bool {
	if level == "DEBUG" && !settings.Debug {
		return false
	}
	if level == "DEEP" && (!settings.Debug || !settings.DebugDeep) {
		return false
	}
	return true
}

// consoleLine writes one log line: the level, the time, then the message with
// its glyph in front when emojis are on.
func consoleLine(settings LoggerSettings, level, glyph, message string, at time.Time) string {
	if glyph != "" && settings.Emoji {
		message = fmt.Sprintf("%s %s", glyph, message)
	}
	timestampFormat := "15:04 02/01/06"
	if settings.ShowSeconds {
		timestampFormat = "15:04:05 02/01/06"
	}
	return fmt.Sprintf("[%s] %s: %s", level, at.Format(timestampFormat), message)
}
