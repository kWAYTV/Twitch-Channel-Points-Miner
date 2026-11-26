package pubsub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/kWAYTV/twitch-points-miner/internal/random"
)

const (
	// messageWorkers handle incoming messages so a slow one does not stall the socket.
	messageWorkers = 4
	// pongTimeout is how long the connection may go without a PONG before it is replaced.
	pongTimeout = 5 * time.Minute
	// handshakeTimeout bounds opening the connection.
	handshakeTimeout = 45 * time.Second
	// maxMessageBytes is the largest message accepted from PubSub.
	maxMessageBytes = 4 << 20
)

// connectAndListen opens one PubSub connection, subscribes it to topics and serves
// it until stop is closed or the connection has to be replaced.
func (p *Client) connectAndListen(connIndex int, topics []string, stop <-chan struct{}) error {
	ctx, cancel := contextUntil(stop)
	defer cancel()

	conn, err := p.dial(ctx)
	if err != nil {
		return err
	}
	conn.SetReadLimit(maxMessageBytes)

	var lastPong atomic.Pointer[time.Time]
	connectedAt := p.env.now()
	lastPong.Store(&connectedAt)
	pingTimer := time.NewTimer(p.randomPingInterval())
	defer pingTimer.Stop()

	msgCh := make(chan []byte, 256)
	workerWG := p.startWorkers(msgCh, stop)

	var readWG sync.WaitGroup
	defer func() {
		_ = conn.CloseNow()
		readWG.Wait()
		close(msgCh)
		workerWG.Wait()
	}()

	if err := p.listenTopics(ctx, conn, topics); err != nil {
		return err
	}

	p.logger.Printf("Connected to Twitch PubSub (conn #%d) with %d topic(s)", connIndex, len(topics))

	readErr := make(chan error, 1)
	readWG.Add(1)
	go func() {
		defer readWG.Done()
		p.readMessages(ctx, conn, connIndex, msgCh, readErr, &lastPong)
	}()

	return p.keepAlive(ctx, conn, pingTimer, readErr, &lastPong)
}

// contextUntil returns a context that is cancelled when stop is closed.
func contextUntil(stop <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-stop:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

func (p *Client) dial(ctx context.Context) (*websocket.Conn, error) {
	handshake, cancel := context.WithTimeout(ctx, handshakeTimeout)
	defer cancel()
	conn, _, err := websocket.Dial(handshake, p.env.endpoint, nil)
	return conn, err
}

// startWorkers handles queued messages until msgCh is closed or stop is.
func (p *Client) startWorkers(msgCh <-chan []byte, stop <-chan struct{}) *sync.WaitGroup {
	var workers sync.WaitGroup
	workers.Add(messageWorkers)
	for i := 0; i < messageWorkers; i++ {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-stop:
					return
				case raw, ok := <-msgCh:
					if !ok {
						return
					}
					if err := p.handleMessage(raw, nil); err != nil {
						p.logger.Errorf("PubSub message error: %v", err)
					}
				}
			}
		}()
	}
	return &workers
}

// keepAlive pings on a jittered interval and returns when the context ends,
// reading failed, or no PONG has arrived for too long.
func (p *Client) keepAlive(ctx context.Context, conn *websocket.Conn, pingTimer *time.Timer, readErr <-chan error, lastPong *atomic.Pointer[time.Time]) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pingTimer.C:
			if err := wsjson.Write(ctx, conn, map[string]string{"type": "PING"}); err != nil {
				return err
			}
			if pongOverdue(*lastPong.Load(), p.env.now()) {
				return fmt.Errorf("last PONG >5m ago, reconnecting")
			}
			pingTimer.Reset(p.randomPingInterval())
		case err := <-readErr:
			if ctx.Err() != nil {
				// Reading stopped because the client is shutting down.
				return nil
			}
			return err
		}
	}
}

// pongOverdue says whether the last PONG is too old for the connection to be kept.
func pongOverdue(lastPong, now time.Time) bool {
	return now.Sub(lastPong) > pongTimeout
}

func (p *Client) randomPingInterval() time.Duration {
	return time.Duration(random.Int(25, 30)) * time.Second
}
