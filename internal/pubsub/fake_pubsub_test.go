package pubsub

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// fakePubSub starts a PubSub server in the test's process and returns its
// address. serve runs once per connection, with the server's end of it, and
// the connection is dropped when serve returns.
func fakePubSub(t *testing.T, serve func(conn *websocket.Conn)) string {
	t.Helper()
	server := httptest.NewServer(accepting(serve))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

// fakePubSubTLS is fakePubSub behind a certificate no client trusts.
func fakePubSubTLS(t *testing.T, serve func(conn *websocket.Conn)) string {
	t.Helper()
	server := httptest.NewUnstartedServer(accepting(serve))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	t.Cleanup(server.Close)
	return "wss" + strings.TrimPrefix(server.URL, "https")
}

// notPubSub starts a server that answers every request with 404 and returns
// the address a client would dial.
func notPubSub(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func accepting(serve func(conn *websocket.Conn)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		serve(conn)
	})
}

// connected opens a connection to a fake PubSub and returns the client's end.
func connected(t *testing.T, serve func(conn *websocket.Conn)) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(context.Background(), fakePubSub(t, serve), nil)
	if err != nil {
		t.Fatalf("connecting to the fake PubSub: %v", err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

type sent = map[string]interface{}

// receive returns the next message the client sends, and false once the
// connection is gone.
func receive(conn *websocket.Conn) (sent, bool) {
	var message sent
	if err := wsjson.Read(context.Background(), conn, &message); err != nil {
		return nil, false
	}
	return message, true
}

// forward passes on every message the client sends until the connection is gone.
func forward(conn *websocket.Conn, received chan<- sent) {
	for {
		message, ok := receive(conn)
		if !ok {
			return
		}
		received <- message
	}
}

func untilClosed(conn *websocket.Conn) {
	for {
		if _, ok := receive(conn); !ok {
			return
		}
	}
}

func send(conn *websocket.Conn, message string) {
	_ = conn.Write(context.Background(), websocket.MessageText, []byte(message))
}

// hangUp closes the connection once the client has read everything sent on it.
func hangUp(conn *websocket.Conn) {
	_ = conn.Close(websocket.StatusNormalClosure, "bye")
}

// within waits for the next value on a channel, and fails the test if the
// client never sends one.
func within[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(5 * time.Second):
		t.Fatalf("nothing arrived within 5 seconds")
		panic("unreachable")
	}
}

var nonceShape = regexp.MustCompile(`^[A-Za-z0-9]{16}$`)

// withoutNonce checks that a LISTEN carries a random nonce and returns the rest of it.
func withoutNonce(t *testing.T, listen sent) sent {
	t.Helper()
	nonce, _ := listen["nonce"].(string)
	if !nonceShape.MatchString(nonce) {
		t.Errorf("nonce = %q, want 16 letters and digits", nonce)
	}
	rest := sent{}
	for key, value := range listen {
		if key != "nonce" {
			rest[key] = value
		}
	}
	return rest
}
