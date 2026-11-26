package chat

import "testing"

func TestIsAuthFailure(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"Twitch's refusal notice", ":tmi.twitch.tv NOTICE * :Login authentication failed", true},
		{"the refusal in capitals", ":tmi.twitch.tv NOTICE * :LOGIN AUTHENTICATION FAILED", true},
		{"another notice", ":tmi.twitch.tv NOTICE * :Improperly formatted auth", false},
		{"an empty line", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAuthFailure(tt.line); got != tt.want {
				t.Errorf("isAuthFailure(%q) = %t, want %t", tt.line, got, tt.want)
			}
		})
	}

	t.Run("a chat message quoting the refusal", func(t *testing.T) {
		line := ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :my authentication failed"
		if got := isAuthFailure(line); got != false {
			t.Errorf("isAuthFailure(%q) = %t, want false", line, got)
		}
	})
}

func TestParsePrivmsg(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantNick string
		wantMsg  string
		wantOK   bool
	}{
		{"a chat message", ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :hello there", "nick", "hello there", true},
		{"a message with a colon in it", ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :look :here", "nick", "look :here", true},
		{"an empty message", ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :", "nick", "", true},
		{"a message outside ASCII", ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :cześć 你好", "nick", "cześć 你好", true},
		{"a sender with no user part", ":tmi.twitch.tv PRIVMSG #chan :hello", "", "hello", true},
		{"a sender that starts with the bang", ":!nick@nick.tmi.twitch.tv PRIVMSG #chan :hello", "", "hello", true},
		{"no sender", "PRIVMSG #chan :hello", "", "hello", true},
		{"a message with no text part", ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan", "", "", false},
		{"a join", ":nick!nick@nick.tmi.twitch.tv JOIN #chan", "", "", false},
		{"a notice that names PRIVMSG", ":tmi.twitch.tv NOTICE * :PRIVMSG is rate limited", "", "PRIVMSG is rate limited", true},
		{"an empty line", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nick, msg, ok := parsePrivmsg(tt.line)

			if nick != tt.wantNick || msg != tt.wantMsg || ok != tt.wantOK {
				t.Errorf("parsePrivmsg(%q) = %q, %q, %t, want %q, %q, %t", tt.line, nick, msg, ok, tt.wantNick, tt.wantMsg, tt.wantOK)
			}
		})
	}
}

func TestMentions(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		username string
		bareName bool
		want     bool
	}{
		{"an @ mention", "hello @target there", "target", false, true},
		{"an @ mention in capitals", "hello @TaRgEt there", "target", false, true},
		{"an @ mention, bare names counted", "hello @target there", "target", true, true},
		{"the bare name", "target is cool", "target", false, false},
		{"the bare name, bare names counted", "target is cool", "target", true, true},
		{"the bare name in capitals, bare names counted", "TARGET is cool", "target", true, true},
		{"someone else", "hello @other there", "target", false, false},
		{"someone else, bare names counted", "hello @other there", "target", true, false},
		{"a longer name that starts with it", "hello @target2", "target", false, true},
		{"the name inside a word, bare names counted", "retargeting ads", "target", true, true},
		{"the whole message is the mention", "@target", "target", false, true},
		{"an empty message", "", "target", false, false},
		{"an empty message, bare names counted", "", "target", true, false},
		{"a name outside ASCII", "cześć @żółw", "żółw", false, true},
		{"no username, a message with an @", "hello @other", "", false, true},
		{"no username, bare names counted", "hello", "", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentions(tt.msg, tt.username, tt.bareName); got != tt.want {
				t.Errorf("mentions(%q, %q, %t) = %t, want %t", tt.msg, tt.username, tt.bareName, got, tt.want)
			}
		})
	}
}

func TestMentionLine(t *testing.T) {
	tests := []struct {
		name    string
		nick    string
		channel string
		msg     string
		want    string
	}{
		{"a known sender", "nick", "chan", "hello @target", "nick at #chan wrote: hello @target"},
		{"no sender", "", "chan", "hello @target", "unknown at #chan wrote: hello @target"},
		{"an empty message", "nick", "chan", "", "nick at #chan wrote: "},
		{"a message outside ASCII", "nick", "chan", "cześć @target 你好", "nick at #chan wrote: cześć @target 你好"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentionLine(tt.nick, tt.channel, tt.msg); got != tt.want {
				t.Errorf("mentionLine(%q, %q, %q) = %q, want %q", tt.nick, tt.channel, tt.msg, got, tt.want)
			}
		})
	}
}
