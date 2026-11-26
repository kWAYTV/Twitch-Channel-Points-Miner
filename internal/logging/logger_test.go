package logging

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
	"github.com/kWAYTV/twitch-points-miner/internal/notify/notifytest"
)

var now = time.Date(2026, 11, 25, 15, 4, 5, 0, time.UTC)

// lineAtAnyTime is a debug line that says hello, stamped with the real clock.
var lineAtAnyTime = regexp.MustCompile(`^\[DEBUG\] \d\d:\d\d \d\d/\d\d/\d\d: hello\n$`)

func setup(settings LoggerSettings) (*Logger, *bytes.Buffer) {
	console := &bytes.Buffer{}
	return newLoggerTo(console, settings, "", func() time.Time { return now }), console
}

// setupSaving builds a logger that saves to <the returned directory>/log.
func setupSaving(t *testing.T, settings LoggerSettings, username string) (*Logger, *bytes.Buffer, string) {
	t.Helper()
	dir := t.TempDir()
	console := &bytes.Buffer{}
	logger, file := newLogger(console, filepath.Join(dir, "log"), settings, username, func() time.Time { return now })
	if file != nil {
		t.Cleanup(func() { _ = file.Close() })
	}
	return logger, console, dir
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func TestNewLogger(t *testing.T) {
	t.Run("saving turned on", func(t *testing.T) {
		logger, console, dir := setupSaving(t, LoggerSettings{Save: true}, "tester")

		logger.Printf("hello")

		if got, want := readFile(t, filepath.Join(dir, "log", "tester.log")), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
		if got, want := console.String(), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("saving turned off", func(t *testing.T) {
		logger, console, dir := setupSaving(t, LoggerSettings{}, "tester")

		logger.Printf("hello")

		if _, err := os.Stat(filepath.Join(dir, "log")); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("stat of the log directory = %v, want it not to exist", err)
		}
		if got, want := console.String(), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("saving anonymous logs", func(t *testing.T) {
		logger, _, dir := setupSaving(t, LoggerSettings{Save: true, AnonymizeLogs: true}, "tester")

		logger.Printf("hello")

		if got, want := readFile(t, filepath.Join(dir, "log", "miner.log")), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
	})

	t.Run("the username turned on", func(t *testing.T) {
		logger, console, dir := setupSaving(t, LoggerSettings{Save: true, ShowUsername: true}, "tester")

		logger.Printf("hello")

		if got, want := readFile(t, filepath.Join(dir, "log", "tester.log")), "[INFO] 15:04 25/11/26: [tester] hello\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
		if got, want := console.String(), "[INFO] 15:04 25/11/26: [tester] hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("the username turned on in anonymous logs", func(t *testing.T) {
		logger, console, dir := setupSaving(t, LoggerSettings{Save: true, ShowUsername: true, AnonymizeLogs: true}, "tester")

		logger.Printf("hello")

		if got, want := readFile(t, filepath.Join(dir, "log", "miner.log")), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
		if got, want := console.String(), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("saving a line with colours", func(t *testing.T) {
		logger, console, dir := setupSaving(t, LoggerSettings{Save: true}, "tester")

		logger.Printf("%sAlice%s is live", ansi.ColorGreen, ansi.ColorReset)

		if got, want := readFile(t, filepath.Join(dir, "log", "tester.log")), "[INFO] 15:04 25/11/26: Alice is live\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
		if got, want := console.String(), "[INFO] 15:04 25/11/26: \x1b[38;5;46mAlice\x1b[0m is live\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("saving to a file that already has lines", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "log"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "log", "tester.log"), []byte("an earlier line\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		logger, file := newLogger(&bytes.Buffer{}, filepath.Join(dir, "log"), LoggerSettings{Save: true}, "tester", func() time.Time { return now })
		t.Cleanup(func() { _ = file.Close() })

		logger.Printf("hello")

		if got, want := readFile(t, filepath.Join(dir, "log", "tester.log")), "an earlier line\n[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("log file holds %q, want %q", got, want)
		}
	})

	t.Run("saving where the log directory can't be made", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "log"), []byte("in the way"), 0o644); err != nil {
			t.Fatal(err)
		}
		console := &bytes.Buffer{}
		logger, _ := newLogger(console, filepath.Join(dir, "log"), LoggerSettings{Save: true}, "tester", func() time.Time { return now })

		logger.Printf("hello")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		if got, want := readFile(t, filepath.Join(dir, "log")), "in the way"; got != want {
			t.Errorf("the file in the way holds %q, want %q", got, want)
		}
	})
}

func TestLogger_Printf(t *testing.T) {
	t.Run("a format and its arguments", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.Printf("loaded %d streamers in %s", 3, "2s")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: loaded 3 streamers in 2s\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("seconds turned on", func(t *testing.T) {
		logger, console := setup(LoggerSettings{ShowSeconds: true})

		logger.Printf("hello")

		if got, want := console.String(), "[INFO] 15:04:05 25/11/26: hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("two lines", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})
		logger.Printf("first")

		logger.Printf("second")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: first\n[INFO] 15:04 25/11/26: second\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("a line that is no event", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, _ := setup(LoggerSettings{})

		logger.Printf("Alice is Online!")

		if len(subscriber.Received) != 0 {
			t.Errorf("subscriber received %+v, want nothing", subscriber.Received)
		}
	})
}

func TestLogger_Println(t *testing.T) {
	t.Run("one value", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.Println("https://example.com/100%")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: https://example.com/100%\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("two words", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.Println("hello", "world")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: hello world\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})
}

func TestLogger_Errorf(t *testing.T) {
	t.Run("a format and its arguments", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.Errorf("login failed: %s", "timeout")

		if got, want := console.String(), "[ERROR] 15:04 25/11/26: login failed: timeout\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})
}

func TestLogger_EmojiPrintf(t *testing.T) {
	t.Run("emojis turned on", func(t *testing.T) {
		logger, console := setup(LoggerSettings{Emoji: true})

		logger.EmojiPrintf(":rocket:", "started %d", 1)

		if got, want := console.String(), "[INFO] 15:04 25/11/26: 🚀 started 1\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("emojis turned off", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.EmojiPrintf(":rocket:", "started %d", 1)

		if got, want := console.String(), "[INFO] 15:04 25/11/26: started 1\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})
}

func TestLogger_Eventf(t *testing.T) {
	t.Run("an event with colours", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{})

		logger.Eventf(events.EventStreamerOnline, "%sAlice%s is Online!", ansi.ColorGreen, ansi.ColorReset)

		if got, want := console.String(), "[INFO] 15:04 25/11/26: \x1b[38;5;46mAlice\x1b[0m is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		want := []publication{{events.EventStreamerOnline, notify.Message{Headline: "Alice is online"}}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("emojis turned on", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{Emoji: true})

		logger.Eventf(events.EventStreamerOnline, "Alice is Online!")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: 🟢 Alice is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		want := []publication{{events.EventStreamerOnline, notify.Message{Headline: "Alice is online"}}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("no event", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{})

		logger.Eventf("", "Alice is Online!")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: Alice is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		if len(subscriber.Received) != 0 {
			t.Errorf("subscriber received %+v, want nothing", subscriber.Received)
		}
	})

	t.Run("a Discord webhook set", func(t *testing.T) {
		discord := notifytest.NewFakeDiscord(t)
		logger, _ := setup(LoggerSettings{Discord: notify.DiscordSettings{WebhookAPI: discord.URL}})

		logger.Eventf(events.EventStreamerOnline, "%sAlice%s (1.2k points) is Online!", ansi.ColorGreen, ansi.ColorReset)

		want := []notifytest.DiscordMessage{{
			ContentType: "application/json",
			Username:    "Twitch Channel Points Miner",
			AvatarURL:   "https://raw.githubusercontent.com/kWAYTV/twitch-points-miner/main/docs/assets/gopher.png",
			Embeds:      []notifytest.DiscordEmbed{{Title: "🟢 Alice is online", Description: "Balance 1.2k points", Color: 0x0A7A2F}},
		}}
		if got := discord.Messages(); !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	// The login is for telling console lines apart. A notification never carries it.
	t.Run("the username turned on", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console, _ := setupSaving(t, LoggerSettings{ShowUsername: true}, "tester")

		logger.Eventf(events.EventStreamerOnline, "Alice (1.2k points) is Online!")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: [tester] Alice (1.2k points) is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		want := []publication{{events.EventStreamerOnline, notify.Message{Headline: "Alice is online", Details: []string{"Balance 1.2k points"}}}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("the username turned on and a Discord webhook set", func(t *testing.T) {
		discord := notifytest.NewFakeDiscord(t)
		logger, _, _ := setupSaving(t, LoggerSettings{ShowUsername: true, Discord: notify.DiscordSettings{WebhookAPI: discord.URL}}, "tester")

		logger.Eventf(events.EventStreamerOnline, "Alice (1.2k points) is Online!")

		want := []notifytest.DiscordMessage{{
			ContentType: "application/json",
			Username:    "Twitch Channel Points Miner",
			AvatarURL:   "https://raw.githubusercontent.com/kWAYTV/twitch-points-miner/main/docs/assets/gopher.png",
			Embeds:      []notifytest.DiscordEmbed{{Title: "🟢 Alice is online", Description: "Balance 1.2k points", Color: 0x0A7A2F}},
		}}
		if got := discord.Messages(); !reflect.DeepEqual(got, want) {
			t.Errorf("discord received %+v, want %+v", got, want)
		}
	})

	t.Run("a Discord webhook set and a line that is no event", func(t *testing.T) {
		discord := notifytest.NewFakeDiscord(t)
		logger, _ := setup(LoggerSettings{Discord: notify.DiscordSettings{WebhookAPI: discord.URL}})

		logger.Printf("Alice (1.2k points) is Online!")

		if got := discord.Messages(); len(got) != 0 {
			t.Errorf("discord received %+v, want nothing", got)
		}
	})
}

func TestLogger_EmojiEventf(t *testing.T) {
	t.Run("emojis turned on", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{Emoji: true})

		logger.EmojiEventf(":bar_chart:", events.EventBetLose, "EventPrediction: bob - overtime? - Decision: No - Result: %sLOSE%s, Gained: -500", ansi.ColorRed, ansi.ColorReset)

		if got, want := console.String(), "[INFO] 15:04 25/11/26: 📉 EventPrediction: bob - overtime? - Decision: No - Result: \x1b[38;5;196mLOSE\x1b[0m, Gained: -500\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		want := []publication{{events.EventBetLose, notify.Message{Headline: "Lost 500 points on bob", Details: []string{`"overtime?"`, "Chose No"}}}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("emojis turned off", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.EmojiEventf(":partying_face:", events.EventStreamerOnline, "Alice is Online!")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: Alice is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("no event", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{Emoji: true})

		logger.EmojiEventf(":partying_face:", "", "Alice is Online!")

		if got, want := console.String(), "[INFO] 15:04 25/11/26: 🥳 Alice is Online!\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		if len(subscriber.Received) != 0 {
			t.Errorf("subscriber received %+v, want nothing", subscriber.Received)
		}
	})
}

func TestLogger_ErrorEventf(t *testing.T) {
	t.Run("a failed prediction", func(t *testing.T) {
		subscriber := subscribeFake(t)
		logger, console := setup(LoggerSettings{})

		logger.ErrorEventf(events.EventBetFailed, "prediction %s: %s", "Alice", "timeout")

		if got, want := console.String(), "[ERROR] 15:04 25/11/26: prediction Alice: timeout\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		want := []publication{{events.EventBetFailed, notify.Message{Headline: "Failed prediction on Alice", Details: []string{"timeout"}}}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})
}

func TestLogger_Debugf(t *testing.T) {
	t.Run("debug turned on", func(t *testing.T) {
		logger, console := setup(LoggerSettings{Debug: true})

		logger.Debugf("socket %d opened", 2)

		if got, want := console.String(), "[DEBUG] 15:04 25/11/26: socket 2 opened\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("debug turned off", func(t *testing.T) {
		logger, console := setup(LoggerSettings{})

		logger.Debugf("socket %d opened", 2)

		if got := console.String(); got != "" {
			t.Errorf("console shows %q, want nothing", got)
		}
	})
}

func TestLogger_DeepDebugf(t *testing.T) {
	t.Run("debug and deep debug turned on", func(t *testing.T) {
		logger, console := setup(LoggerSettings{Debug: true, DebugDeep: true})

		logger.DeepDebugf("frame %d", 7)

		if got, want := console.String(), "[DEEP] 15:04 25/11/26: frame 7\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("only debug turned on", func(t *testing.T) {
		logger, console := setup(LoggerSettings{Debug: true})

		logger.DeepDebugf("frame %d", 7)

		if got := console.String(); got != "" {
			t.Errorf("console shows %q, want nothing", got)
		}
	})
}

func TestLogger_DebugEnabled(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings LoggerSettings
		want     bool
	}{
		{"debug turned on", LoggerSettings{Debug: true}, true},
		{"debug turned off", LoggerSettings{}, false},
		{"only deep debug turned on", LoggerSettings{DebugDeep: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := setup(tc.settings)
			if got := logger.DebugEnabled(); got != tc.want {
				t.Errorf("DebugEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLogger_DeepDebugEnabled(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings LoggerSettings
		want     bool
	}{
		{"debug and deep debug turned on", LoggerSettings{Debug: true, DebugDeep: true}, true},
		{"only debug turned on", LoggerSettings{Debug: true}, false},
		{"only deep debug turned on", LoggerSettings{DebugDeep: true}, false},
		{"both turned off", LoggerSettings{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := setup(tc.settings)
			if got := logger.DeepDebugEnabled(); got != tc.want {
				t.Errorf("DeepDebugEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewLoggerTo(t *testing.T) {
	t.Run("a writer of the caller's", func(t *testing.T) {
		output := &bytes.Buffer{}
		logger := NewLoggerTo(output, LoggerSettings{Debug: true})

		logger.Debugf("hello")

		if got := output.String(); !lineAtAnyTime.MatchString(got) {
			t.Errorf("output holds %q, want it to match %s", got, lineAtAnyTime)
		}
	})
}
