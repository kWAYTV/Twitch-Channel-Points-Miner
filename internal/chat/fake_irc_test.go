package chat

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// patience bounds every wait on the client, so a broken test fails instead of
// hanging. No passing test waits this long.
const patience = 5 * time.Second

// fakeIRC stands in for Twitch chat: each dial opens an in-memory connection
// whose server end the test accepts, and each wait before a retry parks the
// client until the test wakes it.
type fakeIRC struct {
	conns chan *ircConn
	slept chan time.Duration
	wake  chan struct{}

	mu      sync.Mutex
	refusal error
	once    bool
}

func newFakeIRC() *fakeIRC {
	return &fakeIRC{
		conns: make(chan *ircConn, 16),
		slept: make(chan time.Duration),
		wake:  make(chan struct{}),
	}
}

func (f *fakeIRC) link() link {
	return link{dial: f.dial, wait: f.wait}
}

// Refuse makes every dial from now on fail with err.
func (f *fakeIRC) Refuse(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refusal, f.once = err, false
}

// RefuseOnce makes only the next dial fail with err.
func (f *fakeIRC) RefuseOnce(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refusal, f.once = err, true
}

func (f *fakeIRC) dial() (net.Conn, error) {
	f.mu.Lock()
	err := f.refusal
	if f.once {
		f.refusal, f.once = nil, false
	}
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	client, server := net.Pipe()
	_ = server.SetDeadline(time.Now().Add(patience))
	f.conns <- &ircConn{conn: server, reader: bufio.NewReader(server)}
	return client, nil
}

func (f *fakeIRC) wait(d time.Duration, stop <-chan struct{}) {
	select {
	case f.slept <- d:
	case <-stop:
		return
	}
	select {
	case <-f.wake:
	case <-stop:
	}
}

// Accept returns the server end of the next connection a client opens.
func (f *fakeIRC) Accept(t *testing.T) *ircConn {
	t.Helper()
	select {
	case conn := <-f.conns:
		return conn
	case <-time.After(patience):
		t.Fatal("no connection was opened")
		return nil
	}
}

// Unaccepted counts the connections opened that no test accepted.
func (f *fakeIRC) Unaccepted() int {
	return len(f.conns)
}

// Slept returns how long the client now waits before its next try. The client
// stays parked until Wake.
func (f *fakeIRC) Slept(t *testing.T) time.Duration {
	t.Helper()
	select {
	case d := <-f.slept:
		return d
	case <-time.After(patience):
		t.Fatal("the client never waited to retry")
		return 0
	}
}

// Wake ends the wait a client is parked in.
func (f *fakeIRC) Wake() {
	f.wake <- struct{}{}
}

// StopParked stops a client that is parked in a wait, refusing it any further
// connection on its way out.
func (f *fakeIRC) StopParked(stop func()) {
	f.Refuse(net.ErrClosed)
	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()
	for {
		select {
		case f.wake <- struct{}{}:
		case <-f.slept:
		case <-done:
			return
		}
	}
}

// ircConn is the server end of one chat connection.
type ircConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Line reads the next line the client sent, without its line ending.
func (c *ircConn) Line(t *testing.T) string {
	t.Helper()
	line, err := c.reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read a line from the client: %v", err)
	}
	return strings.TrimSuffix(line, "\r\n")
}

// Lines reads the next n lines the client sent.
func (c *ircConn) Lines(t *testing.T, n int) []string {
	t.Helper()
	lines := make([]string, n)
	for i := range lines {
		lines[i] = c.Line(t)
	}
	return lines
}

// Send delivers a line to the client.
func (c *ircConn) Send(t *testing.T, line string) {
	t.Helper()
	if _, err := c.conn.Write([]byte(line + "\r\n")); err != nil {
		t.Fatalf("send %q to the client: %v", line, err)
	}
}

// Settle returns once the client has handled every line sent so far, by
// trading a PING for its PONG.
func (c *ircConn) Settle(t *testing.T) {
	t.Helper()
	c.Send(t, "PING :settle")
	if got := c.Line(t); got != "PONG :settle" {
		t.Fatalf("client answered %q, want %q", got, "PONG :settle")
	}
}

// Close hangs up on the client.
func (c *ircConn) Close() {
	c.conn.Close()
}

// ClosedByClient reports whether the client has hung up.
func (c *ircConn) ClosedByClient() bool {
	_, err := c.reader.ReadString('\n')
	return err != nil && !isTimeout(err)
}

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}
