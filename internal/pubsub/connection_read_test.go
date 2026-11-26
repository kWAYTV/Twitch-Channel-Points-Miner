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
)

func TestMessageType(t *testing.T) {
	for _, row := range []struct {
		name         string
		message      string
		wantType     string
		wantEnvelope map[string]interface{}
	}{
		{"a PONG", `{"type":"PONG"}`, "PONG", map[string]interface{}{"type": "PONG"}},
		{
			"a type in lower case with spaces around it",
			`{"type":" reconnect "}`,
			"RECONNECT",
			map[string]interface{}{"type": " reconnect "},
		},
		{
			"a RESPONSE with every field",
			`{"type":"RESPONSE","error":"ERR_BADAUTH","nonce":"abc"}`,
			"RESPONSE",
			map[string]interface{}{"type": "RESPONSE", "error": "ERR_BADAUTH", "nonce": "abc"},
		},
		{"no type", `{"data":{}}`, "", map[string]interface{}{"data": map[string]interface{}{}}},
		{"a type that is a number", `{"type":5}`, "", map[string]interface{}{"type": 5.0}},
		{"an empty object", `{}`, "", map[string]interface{}{}},
		{"not JSON", `PONG`, "", nil},
		{"nothing", ``, "", nil},
	} {
		t.Run(row.name, func(t *testing.T) {
			gotType, gotEnvelope := messageType([]byte(row.message))
			if gotType != row.wantType || !reflect.DeepEqual(gotEnvelope, row.wantEnvelope) {
				t.Errorf("messageType(%s) = %q, %v, want %q, %v", row.message, gotType, gotEnvelope, row.wantType, row.wantEnvelope)
			}
		})
	}
}

// reading is what readMessages left behind once the server hung up.
type reading struct {
	queued   []string
	err      error
	lastPong *time.Time
}

// read runs readMessages on a connection whose server sends messages and then
// hangs up. queue is how many messages the worker queue holds.
func read(t *testing.T, h *harness, queue int, messages ...string) reading {
	t.Helper()
	conn := connected(t, func(conn *websocket.Conn) {
		for _, message := range messages {
			send(conn, message)
		}
		hangUp(conn)
	})
	queued := make(chan []byte, queue)
	readErr := make(chan error, 1)
	var lastPong atomic.Pointer[time.Time]

	h.client.readMessages(context.Background(), conn, 1, queued, readErr, &lastPong)

	result := reading{lastPong: lastPong.Load()}
	close(queued)
	for message := range queued {
		result.queued = append(result.queued, string(message))
	}
	select {
	case result.err = <-readErr:
	default:
	}
	return result
}

func TestClient_readMessages(t *testing.T) {
	const streamDown = `{"type":"MESSAGE","data":{"topic":"video-playback-by-id.123","message":"{\"type\":\"stream-down\"}"}}`

	t.Run("a message", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256, streamDown)

		if want := []string{streamDown}; !reflect.DeepEqual(got.queued, want) {
			t.Errorf("queued %q, want %q", got.queued, want)
		}
		if got.lastPong != nil {
			t.Errorf("last PONG = %v, want none", got.lastPong)
		}
		if h.logger.Lines != nil {
			t.Errorf("lines = %q, want none", h.logger.Lines)
		}
	})

	t.Run("a PONG", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256, `{"type":"PONG"}`)

		if got.lastPong == nil || !got.lastPong.Equal(noon) {
			t.Errorf("last PONG = %v, want %v", got.lastPong, noon)
		}
		if got.queued != nil {
			t.Errorf("queued %q, want nothing", got.queued)
		}
	})

	t.Run("the server hangs up", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256)

		if shape := regexp.MustCompile(`status = StatusNormalClosure and reason = "bye"$`); !shape.MatchString(errorText(got.err)) {
			t.Errorf("reported %q, want it to match %s", errorText(got.err), shape)
		}
	})

	t.Run("a RECONNECT, then a message", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256, `{"type":"RECONNECT"}`, streamDown)

		if got, want := errorText(got.err), "PubSub[1] server requested reconnect: pubsub reconnect requested"; got != want {
			t.Errorf("reported %q, want %q", got, want)
		}
		if !errors.Is(got.err, ErrReconnectRequested) {
			t.Errorf("reported %v, want it to be ErrReconnectRequested", got.err)
		}
		if got.queued != nil {
			t.Errorf("queued %q, want nothing", got.queued)
		}
	})

	t.Run("a RESPONSE with an error", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256, `{"type":"RESPONSE","error":"ERR_BADTOPIC","nonce":"abc"}`)

		if got, want := h.logger.Lines, []string{"error: PubSub[1] RESPONSE error nonce=abc: ERR_BADTOPIC"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
		if want := []string{`{"type":"RESPONSE","error":"ERR_BADTOPIC","nonce":"abc"}`}; !reflect.DeepEqual(got.queued, want) {
			t.Errorf("queued %q, want %q", got.queued, want)
		}
	})

	t.Run("a RESPONSE without an error", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 256, `{"type":"RESPONSE","error":"","nonce":"abc"}`)

		if h.logger.Lines != nil {
			t.Errorf("lines = %q, want none", h.logger.Lines)
		}
		if want := []string{`{"type":"RESPONSE","error":"","nonce":"abc"}`}; !reflect.DeepEqual(got.queued, want) {
			t.Errorf("queued %q, want %q", got.queued, want)
		}
	})

	t.Run("two messages for a queue that holds one", func(t *testing.T) {
		h := setup(world{})

		got := read(t, h, 1, streamDown, `{"type":"MESSAGE","data":{"topic":"raid.123","message":"{}"}}`)

		if want := []string{streamDown}; !reflect.DeepEqual(got.queued, want) {
			t.Errorf("queued %q, want %q", got.queued, want)
		}
		if got, want := h.logger.Lines, []string{"error: PubSub[1] dropping message: worker queue full"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("debug on", func(t *testing.T) {
		h := setup(world{})
		h.logger.DebugOn = true

		read(t, h, 256, `{"type":"PONG"}`)

		if got, want := h.logger.Debug, []string{"PubSub[1] recv type=PONG (15 bytes)"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("debug on, a message without a type", func(t *testing.T) {
		h := setup(world{})
		h.logger.DebugOn = true

		read(t, h, 256, `{}`)

		if got, want := h.logger.Debug, []string{"PubSub[1] recv (2 bytes)"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("deep debug on", func(t *testing.T) {
		h := setup(world{deepDebug: true})
		h.deep.DeepOn = true

		read(t, h, 256, ` {"type":"PONG"} `)

		if got, want := h.deep.Deep, []string{`PubSub[1] recv: {"type":"PONG"}`}; !reflect.DeepEqual(got, want) {
			t.Errorf("deep debug lines = %q, want %q", got, want)
		}
	})
}
