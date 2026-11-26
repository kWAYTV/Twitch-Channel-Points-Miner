package domain

import "testing"

func TestLoginKey(t *testing.T) {
	tests := []struct {
		name  string
		login string
		want  string
	}{
		{"a lower-case login", "alice", "alice"},
		{"capitals", "Alice", "alice"},
		{"spaces around it", "  Alice ", "alice"},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
		{"an underscore and digits", "The_Real_Fer2", "the_real_fer2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoginKey(tt.login); got != tt.want {
				t.Errorf("LoginKey(%q) = %q, want %q", tt.login, got, tt.want)
			}
		})
	}
}

func TestGameKey(t *testing.T) {
	tests := []struct {
		name string
		game string
		want string
	}{
		{"a game as Twitch shows it", "Counter-Strike", "counter-strike"},
		{"spaces around it", " Just Chatting ", "just chatting"},
		{"all capitals", "VALORANT", "valorant"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GameKey(tt.game); got != tt.want {
				t.Errorf("GameKey(%q) = %q, want %q", tt.game, got, tt.want)
			}
		})
	}
}
