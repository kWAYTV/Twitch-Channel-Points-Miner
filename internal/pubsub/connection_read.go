package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// readMessages reads from conn until it fails or the server asks for a reconnect,
// which it reports on readErr. PONGs refresh lastPong; everything else is queued
// on msgCh for the workers.
func (p *Client) readMessages(ctx context.Context, conn *websocket.Conn, connIndex int, msgCh chan<- []byte, readErr chan<- error, lastPong *atomic.Pointer[time.Time]) {
	for {
		_, message, err := conn.Read(ctx)
		if err != nil {
			select {
			case readErr <- err:
			default:
			}
			return
		}
		msgType, envelope := messageType(message)
		if msgType != "" {
			p.debugf("PubSub[%d] recv type=%s (%d bytes)", connIndex, msgType, len(message))
		} else {
			p.debugf("PubSub[%d] recv (%d bytes)", connIndex, len(message))
		}
		p.deepDebugf("PubSub[%d] recv: %s", connIndex, strings.TrimSpace(string(message)))
		if msgType == "PONG" {
			now := p.env.now()
			lastPong.Store(&now)
			continue
		}
		if msgType == "RECONNECT" {
			select {
			case readErr <- fmt.Errorf("PubSub[%d] server requested reconnect: %w", connIndex, ErrReconnectRequested):
			default:
			}
			return
		}
		if msgType == "RESPONSE" && envelope != nil {
			p.logResponseError(connIndex, envelope)
		}

		select {
		case msgCh <- message:
		default:
			p.logger.Errorf("PubSub[%d] dropping message: worker queue full", connIndex)
		}
	}
}

// messageType returns the upper-case "type" of a PubSub message with the decoded
// message, or "" when it has none.
func messageType(message []byte) (string, map[string]interface{}) {
	var envelope map[string]interface{}
	if err := json.Unmarshal(message, &envelope); err != nil {
		return "", envelope
	}
	if t, ok := envelope["type"].(string); ok {
		return strings.ToUpper(strings.TrimSpace(t)), envelope
	}
	return "", envelope
}
