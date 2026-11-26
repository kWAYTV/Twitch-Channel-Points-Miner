package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

type LoggerSettings struct {
	Save          bool                   `json:"save"`
	Emoji         bool                   `json:"emoji"`
	ShowSeconds   bool                   `json:"show_seconds"`
	ShowUsername  bool                   `json:"show_username"`
	Debug         bool                   `json:"debug"`
	DebugDeep     bool                   `json:"debug_deep"`
	AnonymizeLogs bool                   `json:"anonymize_logs"`
	Discord       notify.DiscordSettings `json:"discord"`
}

type Logger struct {
	base     *log.Logger
	settings LoggerSettings
	username string
	discord  *notify.DiscordWebhook
	now      func() time.Time
}

// NewLogger returns a logger that writes to the console and, when settings.Save
// is on, to a file named after the user in the "log" directory.
func NewLogger(settings LoggerSettings, username string) *Logger {
	logger, _ := newLogger(os.Stdout, "log", settings, username, time.Now)
	return logger
}

// NewLoggerTo returns a logger that writes to output only, without a log file.
// It is given no account, so its lines never carry a login.
func NewLoggerTo(output io.Writer, settings LoggerSettings) *Logger {
	return newLoggerTo(output, settings, "", time.Now)
}

// newLogger also returns the log file it opened, nil when there is none. The
// file stays open for as long as the logger is used.
func newLogger(console io.Writer, logDir string, settings LoggerSettings, username string, now func() time.Time) (*Logger, *os.File) {
	output := console
	var file *os.File
	if settings.Save {
		if err := os.MkdirAll(logDir, 0o755); err == nil {
			logPath := filepath.Join(logDir, logFileName(username, settings.AnonymizeLogs))
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err == nil {
				file = f
				output = dualWriter{
					console: console,
					file:    f,
				}
			}
		}
	}
	return newLoggerTo(output, settings, username, now), file
}

func newLoggerTo(output io.Writer, settings LoggerSettings, username string, now func() time.Time) *Logger {
	return &Logger{
		base:     log.New(output, "", 0),
		settings: settings,
		username: username,
		discord:  notify.NewDiscordWebhook(settings.Discord),
		now:      now,
	}
}

func (l *Logger) log(level, emoji, format string, args ...interface{}) {
	l.logEvent(level, emoji, "", format, args...)
}

func (l *Logger) logEvent(level, emoji string, event events.Event, format string, args ...interface{}) {
	if !levelEnabled(level, l.settings) {
		return
	}
	message := fmt.Sprintf(format, args...)
	if event != "" {
		// The console keeps the full line. Discord and every subscriber get the
		// event described from the account's side, taken before the line gets its
		// time and login.
		described := notify.Describe(message)
		if l.discord != nil {
			l.discord.Send(described, event)
		}
		publish(event, described)
	}
	l.base.Print(consoleLine(l.settings, l.username, level, emojiFor(emoji, event), message, l.now()))
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.log("INFO", "", format, args...)
}

func (l *Logger) Println(v ...interface{}) {
	l.log("INFO", "", "%s", strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log("ERROR", "", format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log("ERROR", "", format, args...)
	os.Exit(1)
}

func (l *Logger) EmojiPrintf(emoji, format string, args ...interface{}) {
	l.log("INFO", emoji, format, args...)
}

func (l *Logger) Eventf(event events.Event, format string, args ...interface{}) {
	l.logEvent("INFO", "", event, format, args...)
}

func (l *Logger) EmojiEventf(emoji string, event events.Event, format string, args ...interface{}) {
	l.logEvent("INFO", emoji, event, format, args...)
}

func (l *Logger) ErrorEventf(event events.Event, format string, args ...interface{}) {
	l.logEvent("ERROR", "", event, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log("DEBUG", "", format, args...)
}

func (l *Logger) DebugEnabled() bool {
	return l.settings.Debug
}

func (l *Logger) DeepDebugf(format string, args ...interface{}) {
	l.log("DEEP", "", format, args...)
}

func (l *Logger) DeepDebugEnabled() bool {
	return l.settings.Debug && l.settings.DebugDeep
}
