package logging

import (
	"testing"
)

func TestLevelEnabled(t *testing.T) {
	for _, tc := range []struct {
		name     string
		level    string
		settings LoggerSettings
		want     bool
	}{
		{"info with debug turned off", "INFO", LoggerSettings{}, true},
		{"error with debug turned off", "ERROR", LoggerSettings{}, true},
		{"debug with debug turned off", "DEBUG", LoggerSettings{}, false},
		{"debug with debug turned on", "DEBUG", LoggerSettings{Debug: true}, true},
		{"debug with only deep debug turned on", "DEBUG", LoggerSettings{DebugDeep: true}, false},
		{"deep with both turned off", "DEEP", LoggerSettings{}, false},
		{"deep with only debug turned on", "DEEP", LoggerSettings{Debug: true}, false},
		{"deep with only deep debug turned on", "DEEP", LoggerSettings{DebugDeep: true}, false},
		{"deep with both turned on", "DEEP", LoggerSettings{Debug: true, DebugDeep: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := levelEnabled(tc.level, tc.settings); got != tc.want {
				t.Errorf("levelEnabled(%q, %+v) = %v, want %v", tc.level, tc.settings, got, tc.want)
			}
		})
	}
}

func TestConsoleLine(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings LoggerSettings
		username string
		level    string
		glyph    string
		message  string
		want     string
	}{
		{"a message alone", LoggerSettings{}, "my_username", "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"another level", LoggerSettings{}, "my_username", "ERROR", "", "hello", "[ERROR] 15:04 25/11/26: hello"},
		{"seconds turned on", LoggerSettings{ShowSeconds: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04:05 25/11/26: hello"},
		{"a glyph with emojis turned on", LoggerSettings{Emoji: true}, "my_username", "INFO", "🟢", "hello", "[INFO] 15:04 25/11/26: 🟢 hello"},
		{"a glyph with emojis turned off", LoggerSettings{}, "my_username", "INFO", "🟢", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"no glyph with emojis turned on", LoggerSettings{Emoji: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"a percent sign in the message", LoggerSettings{}, "my_username", "INFO", "", "100% done", "[INFO] 15:04 25/11/26: 100% done"},
		{"an empty message", LoggerSettings{}, "my_username", "INFO", "", "", "[INFO] 15:04 25/11/26: "},
		{"the username turned on", LoggerSettings{ShowUsername: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04 25/11/26: [my_username] hello"},
		{"the username at another level", LoggerSettings{ShowUsername: true}, "my_username", "ERROR", "", "hello", "[ERROR] 15:04 25/11/26: [my_username] hello"},
		{"the username with seconds", LoggerSettings{ShowUsername: true, ShowSeconds: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04:05 25/11/26: [my_username] hello"},
		{"the username with a glyph", LoggerSettings{ShowUsername: true, Emoji: true}, "my_username", "INFO", "🟢", "hello", "[INFO] 15:04 25/11/26: [my_username] 🟢 hello"},
		{"the username with a percent sign in it", LoggerSettings{ShowUsername: true}, "100%d", "INFO", "", "hello", "[INFO] 15:04 25/11/26: [100%d] hello"},
		{"the username with an empty message", LoggerSettings{ShowUsername: true}, "my_username", "INFO", "", "", "[INFO] 15:04 25/11/26: [my_username] "},
		{"the username in anonymous logs", LoggerSettings{ShowUsername: true, AnonymizeLogs: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"anonymous logs alone", LoggerSettings{AnonymizeLogs: true}, "my_username", "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"the username turned on with no account", LoggerSettings{ShowUsername: true}, "", "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := consoleLine(tc.settings, tc.username, tc.level, tc.glyph, tc.message, now); got != tc.want {
				t.Errorf("consoleLine(%+v, %q, %q, %q, %q) = %q, want %q", tc.settings, tc.username, tc.level, tc.glyph, tc.message, got, tc.want)
			}
		})
	}
}
