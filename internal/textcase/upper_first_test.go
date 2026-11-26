package textcase

import "testing"

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"a login", "alice", "Alice"},
		{"already capitalised", "Alice", "Alice"},
		{"one letter", "a", "A"},
		{"empty", "", ""},
		{"a first letter outside ASCII", "élan", "Élan"},
		{"a digit first", "1up", "1up"},
		{"an underscore first", "_alice", "_alice"},
		{"a sentence", "the stake is too small", "The stake is too small"},
		{"a leading space", " alice", " alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpperFirst(tt.in); got != tt.want {
				t.Errorf("UpperFirst(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
