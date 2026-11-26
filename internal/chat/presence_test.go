package chat

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

var (
	joinedEntry = logEntry{Level: "info", Emoji: ":speech_balloon:", Text: "Join IRC Chat: Streamer"}
	leftEntry   = logEntry{Level: "info", Emoji: ":speech_balloon:", Text: "Leave IRC Chat: Streamer"}
)

// setupPresence builds the presence of the account "miner", logged in with
// the given chat token, on a fake Twitch chat.
func setupPresence(token string) (*Presence, *fakeIRC, *fakeLogger) {
	irc := newFakeIRC()
	logger := &fakeLogger{}
	displayName := func(streamer *domain.Streamer) string { return streamer.Username }
	presence := newPresence("miner", func() string { return token }, logger, false, nil, displayName, irc.link())
	return presence, irc, logger
}

func streamerIn(mode domain.IRCMode) *domain.Streamer {
	streamer := &domain.Streamer{Username: "Streamer"}
	streamer.Settings.IRCMode = mode
	return streamer
}

// joined puts the presence in the streamer's chat and returns the connection.
func joined(t *testing.T, presence *Presence, irc *fakeIRC, streamer *domain.Streamer) *ircConn {
	t.Helper()
	presence.Update(streamer, true)
	conn := irc.Accept(t)
	conn.Lines(t, 3)
	return conn
}

func checkNoConnections(t *testing.T, irc *fakeIRC) {
	t.Helper()
	if got := irc.Unaccepted(); got != 0 {
		t.Errorf("%d connections opened, want none", got)
	}
}

func TestPresence_Update(t *testing.T) {
	t.Run("a streamer who is online", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")

		presence.Update(streamerIn(domain.IRCModeOnline), true)

		got := irc.Accept(t).Lines(t, 3)
		presence.LeaveAll()
		want := []string{"PASS oauth:token", "NICK miner", "JOIN #streamer"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("client sent %q, want %q", got, want)
		}
		checkEntries(t, logger, []logEntry{joinedEntry})
	})

	t.Run("a streamer who is offline", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")

		presence.Update(streamerIn(domain.IRCModeOnline), false)

		checkEntries(t, logger, nil)
		checkNoConnections(t, irc)
	})

	t.Run("a streamer whose chat is already open", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		joined(t, presence, irc, streamer)

		presence.Update(streamer, true)

		presence.LeaveAll()
		checkEntries(t, logger, []logEntry{joinedEntry})
		checkNoConnections(t, irc)
	})

	t.Run("a streamer who goes offline", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		conn := joined(t, presence, irc, streamer)

		presence.Update(streamer, false)

		if !conn.ClosedByClient() {
			t.Errorf("the connection is still open, want it closed")
		}
		checkEntries(t, logger, []logEntry{joinedEntry, leftEntry})
	})

	t.Run("a streamer who comes back online", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		joined(t, presence, irc, streamer)
		presence.Update(streamer, false)

		presence.Update(streamer, true)

		got := irc.Accept(t).Lines(t, 3)
		presence.LeaveAll()
		want := []string{"PASS oauth:token", "NICK miner", "JOIN #streamer"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("client sent %q, want %q", got, want)
		}
		checkEntries(t, logger, []logEntry{joinedEntry, leftEntry, joinedEntry})
	})

	t.Run("a streamer who is online, before the account is logged in", func(t *testing.T) {
		presence, irc, logger := setupPresence("")

		presence.Update(streamerIn(domain.IRCModeOnline), true)

		checkEntries(t, logger, nil)
		checkNoConnections(t, irc)
	})

	t.Run("no streamer", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")

		presence.Update(nil, true)

		checkEntries(t, logger, nil)
		checkNoConnections(t, irc)
	})

	t.Run("no presence", func(t *testing.T) {
		var presence *Presence

		presence.Update(streamerIn(domain.IRCModeOnline), true)
	})
}

func TestPresence_Leave(t *testing.T) {
	t.Run("a streamer whose chat is open", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		conn := joined(t, presence, irc, streamer)

		presence.Leave(streamer)

		if !conn.ClosedByClient() {
			t.Errorf("the connection is still open, want it closed")
		}
		checkEntries(t, logger, []logEntry{joinedEntry, leftEntry})
	})

	t.Run("the same streamer named in other capitals", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		conn := joined(t, presence, irc, streamerIn(domain.IRCModeOnline))

		presence.Leave(&domain.Streamer{Username: "STREAMER"})

		if !conn.ClosedByClient() {
			t.Errorf("the connection is still open, want it closed")
		}
		checkEntries(t, logger, []logEntry{
			joinedEntry,
			{Level: "info", Emoji: ":speech_balloon:", Text: "Leave IRC Chat: STREAMER"},
		})
	})

	t.Run("a streamer whose chat was never opened", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")

		presence.Leave(streamerIn(domain.IRCModeOnline))

		checkEntries(t, logger, nil)
		checkNoConnections(t, irc)
	})

	t.Run("a streamer whose chat is already closed", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		joined(t, presence, irc, streamer)
		presence.Leave(streamer)

		presence.Leave(streamer)

		checkEntries(t, logger, []logEntry{joinedEntry, leftEntry})
	})

	t.Run("no streamer", func(t *testing.T) {
		presence, _, logger := setupPresence("token")

		presence.Leave(nil)

		checkEntries(t, logger, nil)
	})

	t.Run("no presence", func(t *testing.T) {
		var presence *Presence

		presence.Leave(streamerIn(domain.IRCModeOnline))
	})
}

func TestPresence_LeaveAll(t *testing.T) {
	t.Run("two open chats", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		first := joined(t, presence, irc, &domain.Streamer{Username: "Streamer"})
		second := joined(t, presence, irc, &domain.Streamer{Username: "Other"})

		presence.LeaveAll()

		if !first.ClosedByClient() || !second.ClosedByClient() {
			t.Errorf("a connection is still open, want both closed")
		}
		checkEntries(t, logger, []logEntry{
			joinedEntry,
			{Level: "info", Emoji: ":speech_balloon:", Text: "Join IRC Chat: Other"},
		})
	})

	t.Run("a chat opened again afterwards", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")
		streamer := streamerIn(domain.IRCModeOnline)
		joined(t, presence, irc, streamer)
		presence.LeaveAll()

		presence.Update(streamer, true)

		got := irc.Accept(t).Lines(t, 3)
		presence.LeaveAll()
		want := []string{"PASS oauth:token", "NICK miner", "JOIN #streamer"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("client sent %q, want %q", got, want)
		}
		checkEntries(t, logger, []logEntry{joinedEntry, joinedEntry})
	})

	t.Run("no open chats", func(t *testing.T) {
		presence, irc, logger := setupPresence("token")

		presence.LeaveAll()

		checkEntries(t, logger, nil)
		checkNoConnections(t, irc)
	})

	t.Run("no presence", func(t *testing.T) {
		var presence *Presence

		presence.LeaveAll()
	})
}

func TestShouldJoin(t *testing.T) {
	tests := []struct {
		name   string
		mode   domain.IRCMode
		online bool
		want   bool
	}{
		{"always, online", domain.IRCModeAlways, true, true},
		{"always, offline", domain.IRCModeAlways, false, true},
		{"never, online", domain.IRCModeNever, true, false},
		{"never, offline", domain.IRCModeNever, false, false},
		{"online only, online", domain.IRCModeOnline, true, true},
		{"online only, offline", domain.IRCModeOnline, false, false},
		{"offline only, online", domain.IRCModeOffline, true, false},
		{"offline only, offline", domain.IRCModeOffline, false, true},
		{"no mode, online", domain.IRCMode(""), true, true},
		{"no mode, offline", domain.IRCMode(""), false, false},
		{"an unknown mode, online", domain.IRCMode("unknown"), true, true},
		{"an unknown mode, offline", domain.IRCMode("unknown"), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldJoin(tt.mode, tt.online); got != tt.want {
				t.Errorf("ShouldJoin(%q, %t) = %t, want %t", tt.mode, tt.online, got, tt.want)
			}
		})
	}
}
