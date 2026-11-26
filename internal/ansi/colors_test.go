package ansi

import (
	"regexp"
	"testing"
)

// The log file and every notification drop colours by this pattern.
var colourEscape = regexp.MustCompile(`^\x1b\[[0-9;]+m$`)

func TestColors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color string
	}{
		{"green", ColorGreen},
		{"red", ColorRed},
		{"cyan", ColorCyan},
		{"purple", ColorPurple},
		{"yellow", ColorYellow},
		{"reset", ColorReset},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !colourEscape.MatchString(tc.color) {
				t.Errorf("%q is not a colour escape", tc.color)
			}
		})
	}
}
