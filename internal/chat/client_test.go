package chat

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
)

const (
	mention     = ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :hello @target there"
	bareMention = ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :target is cool"
	authFailed  = ":tmi.twitch.tv NOTICE * :Login authentication failed"
)

var mentionEntry = logEntry{Level: "info", Emoji: ":speech_balloon:", Event: "CHAT_MENTION", Text: "nick at #chan wrote: hello @target there"}

type clientOptions struct {
	token      string
	bareNames  bool
	anonymizer *privacy.Anonymizer
}

// setup builds the chat client of the account "target" in the channel "chan",
// on a fake Twitch chat.
func setup(options clientOptions) (*Client, *fakeIRC, *fakeLogger) {
	irc := newFakeIRC()
	logger := &fakeLogger{}
	client := newClient("target", options.token, "chan", logger, options.bareNames, options.anonymizer, irc.link())
	return client, irc, logger
}

// connected starts the client and returns its connection once it has joined.
func connected(t *testing.T, client *Client, irc *fakeIRC) *ircConn {
	t.Helper()
	client.Start()
	conn := irc.Accept(t)
	conn.Lines(t, 3)
	return conn
}

func checkEntries(t *testing.T, logger *fakeLogger, want []logEntry) {
	t.Helper()
	if got := logger.Entries(); !reflect.DeepEqual(got, want) {
		t.Errorf("logged %+v, want %+v", got, want)
	}
}

func checkSlept(t *testing.T, irc *fakeIRC, want time.Duration) {
	t.Helper()
	if got := irc.Slept(t); got != want {
		t.Errorf("waited %v before retrying, want %v", got, want)
	}
}

func TestClient_Start(t *testing.T) {
	t.Run("a name, token and channel in capitals and spaces", func(t *testing.T) {
		irc := newFakeIRC()
		client := newClient(" Target ", " token ", " Chan ", &fakeLogger{}, false, nil, irc.link())

		client.Start()

		got := irc.Accept(t).Lines(t, 3)
		client.Stop()
		want := []string{"PASS oauth:token", "NICK target", "JOIN #chan"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("client sent %q, want %q", got, want)
		}
	})

	t.Run("a ping", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, "PING :tmi.twitch.tv")

		got := conn.Line(t)
		client.Stop()
		if got != "PONG :tmi.twitch.tv" {
			t.Errorf("client answered %q, want %q", got, "PONG :tmi.twitch.tv")
		}
		checkEntries(t, logger, nil)
	})

	t.Run("a mention", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, mention)

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, []logEntry{mentionEntry})
	})

	t.Run("a message for someone else", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, ":nick!nick@nick.tmi.twitch.tv PRIVMSG #chan :hello @other there")

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, nil)
	})

	t.Run("a blank line", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, "")

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, nil)
	})

	t.Run("the bare name", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, bareMention)

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, nil)
	})

	t.Run("the bare name, with bare names counted", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token", bareNames: true})
		conn := connected(t, client, irc)

		conn.Send(t, bareMention)

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, []logEntry{
			{Level: "info", Emoji: ":speech_balloon:", Event: "CHAT_MENTION", Text: "nick at #chan wrote: target is cool"},
		})
	})

	t.Run("a mention, with names hidden", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token", anonymizer: privacy.New(true)})
		conn := connected(t, client, irc)

		conn.Send(t, mention)

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, []logEntry{
			{Level: "info", Emoji: ":speech_balloon:", Event: "CHAT_MENTION", Text: "Chat mention in Streamer1"},
		})
	})

	t.Run("a mention, with the anonymizer switched off", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token", anonymizer: privacy.New(false)})
		conn := connected(t, client, irc)

		conn.Send(t, mention)

		conn.Settle(t)
		client.Stop()
		checkEntries(t, logger, []logEntry{mentionEntry})
	})

	t.Run("a mention, with no logger", func(t *testing.T) {
		irc := newFakeIRC()
		client := newClient("target", "token", "chan", nil, false, nil, irc.link())
		conn := connected(t, client, irc)

		conn.Send(t, mention)

		conn.Settle(t)
		client.Stop()
	})

	t.Run("a login Twitch refuses", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Send(t, authFailed)

		checkSlept(t, irc, 5*time.Second)
		checkEntries(t, logger, []logEntry{
			{Level: "info", Text: "chat #chan authentication failed"},
			{Level: "error", Text: "chat #chan error: io: read/write on closed pipe"},
		})
		if !conn.ClosedByClient() {
			t.Errorf("the connection is still open, want it closed")
		}
		irc.StopParked(client.Stop)
	})

	t.Run("a login Twitch refuses, with names hidden", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token", anonymizer: privacy.New(true)})
		conn := connected(t, client, irc)

		conn.Send(t, authFailed)

		checkSlept(t, irc, 5*time.Second)
		checkEntries(t, logger, []logEntry{
			{Level: "info", Text: "chat authentication failed"},
			{Level: "error", Text: "chat error: io: read/write on closed pipe"},
		})
		irc.StopParked(client.Stop)
	})

	t.Run("Twitch hanging up", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		conn.Close()

		checkSlept(t, irc, 5*time.Second)
		shape := regexp.MustCompile(`^chat #chan error: (EOF|io: read/write on closed pipe)$`)
		got := logger.Entries()
		if len(got) != 1 || got[0].Level != "error" || !shape.MatchString(got[0].Text) {
			t.Errorf("logged %+v, want one error matching %s", got, shape)
		}
		irc.StopParked(client.Stop)
	})

	t.Run("Twitch unreachable", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		irc.Refuse(errors.New("dial tcp: connection refused"))

		client.Start()

		checkSlept(t, irc, 5*time.Second)
		checkEntries(t, logger, []logEntry{{Level: "error", Text: "chat #chan error: dial tcp: connection refused"}})
		irc.StopParked(client.Stop)
	})

	t.Run("Twitch unreachable, with no logger", func(t *testing.T) {
		irc := newFakeIRC()
		client := newClient("target", "token", "chan", nil, false, nil, irc.link())
		irc.Refuse(errors.New("dial tcp: connection refused"))

		client.Start()

		checkSlept(t, irc, 5*time.Second)
		irc.StopParked(client.Stop)
	})

	t.Run("Twitch reachable on the second try", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		irc.RefuseOnce(errors.New("dial tcp: connection refused"))
		client.Start()
		irc.Slept(t)

		irc.Wake()

		got := irc.Accept(t).Lines(t, 3)
		client.Stop()
		want := []string{"PASS oauth:token", "NICK target", "JOIN #chan"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("client sent %q, want %q", got, want)
		}
		checkEntries(t, logger, []logEntry{{Level: "error", Text: "chat #chan error: dial tcp: connection refused"}})
	})

	t.Run("no token", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: ""})

		client.Start()

		client.Stop()
		checkEntries(t, logger, []logEntry{{Level: "error", Text: "chat #chan error: missing chat credentials"}})
		if got := irc.Unaccepted(); got != 0 {
			t.Errorf("%d connections opened, want none", got)
		}
	})
}

func TestClient_Stop(t *testing.T) {
	t.Run("before it was started", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})

		client.Stop()

		checkEntries(t, logger, nil)
		if got := irc.Unaccepted(); got != 0 {
			t.Errorf("%d connections opened, want none", got)
		}
	})

	t.Run("while waiting to try again", func(t *testing.T) {
		client, irc, _ := setup(clientOptions{token: "token"})
		irc.Refuse(errors.New("dial tcp: connection refused"))
		client.Start()
		checkSlept(t, irc, 5*time.Second)

		client.Stop()

		if got := irc.Unaccepted(); got != 0 {
			t.Errorf("%d connections opened, want none: the wait ends without another try", got)
		}
	})

	t.Run("while connected", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		conn := connected(t, client, irc)

		client.Stop()

		if !conn.ClosedByClient() {
			t.Errorf("the connection is still open, want it closed")
		}
		checkEntries(t, logger, nil)
	})

	t.Run("a second time", func(t *testing.T) {
		client, irc, logger := setup(clientOptions{token: "token"})
		connected(t, client, irc)
		client.Stop()

		client.Stop()

		checkEntries(t, logger, nil)
		if got := irc.Unaccepted(); got != 0 {
			t.Errorf("%d more connections opened, want none", got)
		}
	})
}
