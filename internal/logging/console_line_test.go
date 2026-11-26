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
		level    string
		glyph    string
		message  string
		want     string
	}{
		{"a message alone", LoggerSettings{}, "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"another level", LoggerSettings{}, "ERROR", "", "hello", "[ERROR] 15:04 25/11/26: hello"},
		{"seconds turned on", LoggerSettings{ShowSeconds: true}, "INFO", "", "hello", "[INFO] 15:04:05 25/11/26: hello"},
		{"a glyph with emojis turned on", LoggerSettings{Emoji: true}, "INFO", "🟢", "hello", "[INFO] 15:04 25/11/26: 🟢 hello"},
		{"a glyph with emojis turned off", LoggerSettings{}, "INFO", "🟢", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"no glyph with emojis turned on", LoggerSettings{Emoji: true}, "INFO", "", "hello", "[INFO] 15:04 25/11/26: hello"},
		{"a percent sign in the message", LoggerSettings{}, "INFO", "", "100% done", "[INFO] 15:04 25/11/26: 100% done"},
		{"an empty message", LoggerSettings{}, "INFO", "", "", "[INFO] 15:04 25/11/26: "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := consoleLine(tc.settings, tc.level, tc.glyph, tc.message, now); got != tc.want {
				t.Errorf("consoleLine(%+v, %q, %q, %q) = %q, want %q", tc.settings, tc.level, tc.glyph, tc.message, got, tc.want)
			}
		})
	}
}
