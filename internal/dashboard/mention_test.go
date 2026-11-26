package dashboard

import "testing"

func TestMentionedStreamer(t *testing.T) {
	known := map[string]*streamerState{"alice": {}, "alice_tv": {}, "bob": {}}

	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"one streamer", "Won a bet on bob.", "bob"},
		{"a name in capitals", "Alice is online.", "alice"},
		{"a longer name that starts with a shorter one", "Alice_tv is online.", "alice_tv"},
		{"two streamers", "Bet on bob: will alice win?", "bob"},
		{"two streamers the other way round", "Bet on alice: will bob win?", "alice"},
		{"a name inside a longer word", "Bobby is online.", ""},
		{"nobody known", "Claimed a drop.", ""},
		{"an empty message", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentionedStreamer(tt.message, known); got != tt.want {
				t.Errorf("mentionedStreamer(%q) = %q, want %q", tt.message, got, tt.want)
			}
		})
	}

	t.Run("no streamers known", func(t *testing.T) {
		if got := mentionedStreamer("Won a bet on bob.", map[string]*streamerState{}); got != "" {
			t.Errorf("mentionedStreamer with no streamers = %q, want %q", got, "")
		}
	})
}

func TestIndexWord(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		word string
		want int
	}{
		{"the whole message", "alice", "alice", 0},
		{"at the start", "alice is online.", "alice", 0},
		{"in the middle", "won a bet on alice today", "alice", 13},
		{"at the end", "won a bet on alice", "alice", 13},
		{"before punctuation", "bet on alice: yes", "alice", 7},
		{"in brackets", "(alice)", "alice", 1},
		{"after a letter outside ASCII", "éalice", "alice", 2},
		{"the start of a longer name", "alice_tv is online.", "alice", -1},
		{"the end of a longer name", "malice is online.", "alice", -1},
		{"followed by a digit", "alice2 is online.", "alice", -1},
		{"inside a longer name, then alone", "alice_tv raided alice", "alice", 16},
		{"absent", "bob is online.", "alice", -1},
		{"an empty message", "", "alice", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indexWord(tt.msg, tt.word); got != tt.want {
				t.Errorf("indexWord(%q, %q) = %d, want %d", tt.msg, tt.word, got, tt.want)
			}
		})
	}
}
