package twitchapi

import "testing"

func TestGetUserAgent(t *testing.T) {
	const androidTV = "Mozilla/5.0 (Linux; Android 7.1; Smart Box C1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
	tests := []struct {
		name    string
		browser string
	}{
		{"Chrome", "CHROME"},
		{"Firefox", "FIREFOX"},
		{"an unknown browser", "LYNX"},
		{"no browser", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserAgent(tt.browser); got != androidTV {
				t.Errorf("GetUserAgent(%q) = %q, want the Android TV agent, %q", tt.browser, got, androidTV)
			}
		})
	}
}
