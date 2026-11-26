package pubsub

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestPongOverdue(t *testing.T) {
	for _, row := range []struct {
		name     string
		lastPong time.Time
		want     bool
	}{
		{"a PONG this instant", noon, false},
		{"a PONG just under five minutes ago", noon.Add(-5*time.Minute + time.Nanosecond), false},
		{"a PONG exactly five minutes ago", noon.Add(-5 * time.Minute), false},
		{"a PONG just over five minutes ago", noon.Add(-5*time.Minute - time.Nanosecond), true},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := pongOverdue(row.lastPong, noon); got != row.want {
				t.Errorf("pongOverdue(%v, noon) = %t, want %t", row.lastPong, got, row.want)
			}
		})
	}
}

func TestClient_randomPingInterval(t *testing.T) {
	t.Run("a whole number of seconds from 25 to 30", func(t *testing.T) {
		h := setup(world{})

		got := h.client.randomPingInterval()

		if got < 25*time.Second || got > 30*time.Second || got%time.Second != 0 {
			t.Errorf("randomPingInterval = %v, want a whole number of seconds from 25s to 30s", got)
		}
	})
}

// listen runs connectAndListen as connection 1 until it returns.
func listen(t *testing.T, h *harness, topics []string, stop <-chan struct{}) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- h.client.connectAndListen(1, topics, stop) }()
	return within(t, done)
}

func TestClient_connectAndListen(t *testing.T) {
	t.Run("the server asks to reconnect", func(t *testing.T) {
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			receive(conn)
			send(conn, `{"type":"RECONNECT"}`)
			untilClosed(conn)
		})
		h := setup(world{endpoint: endpoint})

		err := listen(t, h, []string{"video-playback-by-id.123"}, make(chan struct{}))

		if got, want := errorText(err), "PubSub[1] server requested reconnect: pubsub reconnect requested"; got != want {
			t.Errorf("connectAndListen = %q, want %q", got, want)
		}
		if !errors.Is(err, ErrReconnectRequested) {
			t.Errorf("connectAndListen = %v, want it to be ErrReconnectRequested", err)
		}
		if got, want := h.logger.Lines, []string{"info: Connected to Twitch PubSub (conn #1) with 1 topic(s)"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("the client is stopped", func(t *testing.T) {
		listening := make(chan sent, 1)
		endpoint := fakePubSub(t, func(conn *websocket.Conn) { forward(conn, listening) })
		h := setup(world{endpoint: endpoint})
		stop := make(chan struct{})
		done := make(chan error, 1)
		go func() { done <- h.client.connectAndListen(1, []string{"video-playback-by-id.123"}, stop) }()
		within(t, listening)

		close(stop)

		if err := within(t, done); err != nil {
			t.Errorf("connectAndListen = %v, want nil", err)
		}
	})

	t.Run("the server hangs up", func(t *testing.T) {
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			receive(conn)
			hangUp(conn)
		})
		h := setup(world{endpoint: endpoint})

		err := listen(t, h, []string{"video-playback-by-id.123"}, make(chan struct{}))

		if shape := regexp.MustCompile(`status = StatusNormalClosure and reason = "bye"$`); !shape.MatchString(errorText(err)) {
			t.Errorf("connectAndListen = %q, want it to match %s", errorText(err), shape)
		}
	})

	t.Run("a server that is not PubSub", func(t *testing.T) {
		h := setup(world{endpoint: notPubSub(t)})

		err := listen(t, h, []string{"video-playback-by-id.123"}, make(chan struct{}))

		if shape := regexp.MustCompile(`^failed to WebSocket dial: expected handshake response status code 101 but got 404$`); !shape.MatchString(errorText(err)) {
			t.Errorf("connectAndListen = %q, want it to match %s", errorText(err), shape)
		}
		if h.logger.Lines != nil {
			t.Errorf("lines = %q, want none", h.logger.Lines)
		}
	})

	t.Run("a bonus for a followed channel", func(t *testing.T) {
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			receive(conn)
			send(conn, `{"type":"MESSAGE","data":{"topic":"community-points-user-v1.999","message":"{\"type\":\"claim-available\",\"data\":{\"claim\":{\"id\":\"claim-1\",\"channel_id\":\"123\"}}}"}}`)
			send(conn, `{"type":"RECONNECT"}`)
			untilClosed(conn)
		})
		h := setup(world{endpoint: endpoint, streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		listen(t, h, []string{"community-points-user-v1.999"}, make(chan struct{}))

		if got, want := h.account.BonusClaims, []claimed{{"tester", "claim-1"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("bonus claims = %+v, want %+v", got, want)
		}
	})

	t.Run("a message whose contents are not JSON", func(t *testing.T) {
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			receive(conn)
			send(conn, `{"type":"MESSAGE","data":{"topic":"community-points-user-v1.999","message":"claim-available"}}`)
			send(conn, `{"type":"RECONNECT"}`)
			untilClosed(conn)
		})
		h := setup(world{endpoint: endpoint})

		listen(t, h, []string{"community-points-user-v1.999"}, make(chan struct{}))

		want := []string{
			"info: Connected to Twitch PubSub (conn #1) with 1 topic(s)",
			"error: PubSub message error: invalid character 'c' looking for beginning of value",
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a self-signed server", func(t *testing.T) {
		endpoint := fakePubSubTLS(t, untilClosed)
		h := setup(world{endpoint: endpoint})

		err := listen(t, h, []string{"video-playback-by-id.123"}, make(chan struct{}))

		if shape := regexp.MustCompile(`x509: `); !shape.MatchString(errorText(err)) {
			t.Errorf("connectAndListen = %q, want it to match %s", errorText(err), shape)
		}
	})
}

// keptAlive runs keepAlive on conn with a PING due after pingIn, and returns
// what it returned.
func keptAlive(t *testing.T, h *harness, ctx context.Context, conn *websocket.Conn, pingIn time.Duration, readErr <-chan error, lastPong time.Time) <-chan error {
	t.Helper()
	pingTimer := time.NewTimer(pingIn)
	t.Cleanup(func() { pingTimer.Stop() })
	var pong atomic.Pointer[time.Time]
	pong.Store(&lastPong)
	done := make(chan error, 1)
	go func() { done <- h.client.keepAlive(ctx, conn, pingTimer, readErr, &pong) }()
	return done
}

func TestClient_keepAlive(t *testing.T) {
	errLost := errors.New("connection lost")

	t.Run("the client is stopped", func(t *testing.T) {
		conn := connected(t, untilClosed)
		h := setup(world{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := within(t, keptAlive(t, h, ctx, conn, time.Hour, make(chan error), noon))

		if err != nil {
			t.Errorf("keepAlive = %v, want nil", err)
		}
	})

	t.Run("reading fails", func(t *testing.T) {
		conn := connected(t, untilClosed)
		h := setup(world{})
		readErr := make(chan error, 1)
		readErr <- errLost

		err := within(t, keptAlive(t, h, context.Background(), conn, time.Hour, readErr, noon))

		if err != errLost {
			t.Errorf("keepAlive = %v, want %v", err, errLost)
		}
	})

	t.Run("reading fails because the client is stopped", func(t *testing.T) {
		conn := connected(t, untilClosed)
		h := setup(world{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		readErr := make(chan error, 1)
		readErr <- errLost

		err := within(t, keptAlive(t, h, ctx, conn, time.Hour, readErr, noon))

		if err != nil {
			t.Errorf("keepAlive = %v, want nil", err)
		}
	})

	t.Run("a PING due, the last PONG a minute old", func(t *testing.T) {
		received := make(chan sent, 1)
		conn := connected(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{})
		readErr := make(chan error, 1)

		done := keptAlive(t, h, context.Background(), conn, 0, readErr, noon.Add(-time.Minute))

		if got, want := within(t, received), (sent{"type": "PING"}); !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
		readErr <- errLost
		if err := within(t, done); err != errLost {
			t.Errorf("keepAlive = %v, want %v", err, errLost)
		}
	})

	t.Run("a PING due, the last PONG six minutes old", func(t *testing.T) {
		received := make(chan sent, 1)
		conn := connected(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{})

		done := keptAlive(t, h, context.Background(), conn, 0, make(chan error), noon.Add(-6*time.Minute))

		if got, want := errorText(within(t, done)), "last PONG >5m ago, reconnecting"; got != want {
			t.Errorf("keepAlive = %q, want %q", got, want)
		}
		if got, want := within(t, received), (sent{"type": "PING"}); !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
	})

	t.Run("a PING due on a closed connection", func(t *testing.T) {
		conn := connected(t, untilClosed)
		conn.CloseNow()
		h := setup(world{})

		err := within(t, keptAlive(t, h, context.Background(), conn, 0, make(chan error), noon))

		if shape := regexp.MustCompile(`^failed to write JSON message: `); !shape.MatchString(errorText(err)) {
			t.Errorf("keepAlive = %q, want it to match %s", errorText(err), shape)
		}
	})
}
