package chat

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
	"github.com/kWAYTV/twitch-points-miner/internal/stoppable"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

type Logger interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	EmojiPrintf(emoji, format string, args ...interface{})
	EmojiEventf(emoji string, event events.Event, format string, args ...interface{})
}

type Client struct {
	username            string
	channel             string
	token               string
	logger              Logger
	anonymizer          *privacy.Anonymizer
	disableAtInNickname bool
	stop                chan struct{}
	stopped             chan struct{}
	started             atomic.Bool
	conn                net.Conn
	connMu              sync.Mutex
	link                link
}

// link is how a chat reaches Twitch, and how it waits before trying again.
// wait returns early when stop is closed.
type link struct {
	dial func() (net.Conn, error)
	wait func(d time.Duration, stop <-chan struct{})
}

func twitchLink() link {
	return link{
		dial: dialTwitch,
		wait: func(d time.Duration, stop <-chan struct{}) { stoppable.Wait(d, stop) },
	}
}

func dialTwitch() (net.Conn, error) {
	addr := net.JoinHostPort(twitchapi.IRC, strconv.Itoa(twitchapi.IRCPort))
	return net.DialTimeout("tcp", addr, 10*time.Second)
}

func NewClient(username, token, channel string, logger Logger, disableAtInNickname bool, anonymizer *privacy.Anonymizer) *Client {
	return newClient(username, token, channel, logger, disableAtInNickname, anonymizer, twitchLink())
}

func newClient(username, token, channel string, logger Logger, disableAtInNickname bool, anonymizer *privacy.Anonymizer, link link) *Client {
	return &Client{
		username:            domain.LoginKey(username),
		channel:             domain.LoginKey(channel),
		token:               strings.TrimSpace(token),
		logger:              logger,
		anonymizer:          anonymizer,
		disableAtInNickname: disableAtInNickname,
		stop:                make(chan struct{}),
		stopped:             make(chan struct{}),
		link:                link,
	}
}

// Start joins the chat and keeps it joined until Stop. Without a name, a
// channel and a token there is nothing to try, so it reports that once and
// starts nothing.
func (c *Client) Start() {
	if c.username == "" || c.channel == "" || c.token == "" {
		c.logError(errors.New("missing chat credentials"))
		return
	}
	if c.started.Swap(true) {
		return
	}
	go c.run()
}

func (c *Client) logError(err error) {
	if c.logger == nil {
		return
	}
	if c.anonymizer.Enabled() {
		c.logger.Errorf("chat error: %v", err)
	} else {
		c.logger.Errorf("chat #%s error: %v", c.channel, err)
	}
}

// Stop ends the chat and returns once it has. A client that was never started
// has nothing to wait for.
func (c *Client) Stop() {
	select {
	case <-c.stop:
	default:
		close(c.stop)
		c.closeConn()
	}
	if c.started.Load() {
		<-c.stopped
	}
}

func (c *Client) run() {
	defer close(c.stopped)
	for !c.isStopped() {
		if err := c.connectAndListen(); err != nil && !c.isStopped() {
			c.logError(err)
		}
		c.link.wait(5*time.Second, c.stop)
	}
}

func (c *Client) connectAndListen() error {
	conn, err := c.link.dial()
	if err != nil {
		return err
	}
	c.setConn(conn)
	defer c.closeConn()
	// Stop closes the connection it finds. One opened while it was looking
	// would otherwise be read until its deadline.
	if c.isStopped() {
		return nil
	}

	if err := c.write(fmt.Sprintf("PASS oauth:%s\r\n", c.token)); err != nil {
		return err
	}
	if err := c.write(fmt.Sprintf("NICK %s\r\n", c.username)); err != nil {
		return err
	}
	if err := c.write(fmt.Sprintf("JOIN #%s\r\n", c.channel)); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(90 * time.Second)); err != nil {
			return err
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if c.isStopped() {
					return nil
				}
				continue
			}
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "PING") {
			payload := strings.TrimPrefix(line, "PING")
			_ = c.write("PONG" + payload + "\r\n")
			continue
		}
		c.handleLine(line)
		if c.isStopped() {
			return nil
		}
	}
}

func (c *Client) handleLine(line string) {
	if isAuthFailure(line) {
		if c.logger != nil {
			if c.anonymizer.Enabled() {
				c.logger.Printf("chat authentication failed")
			} else {
				c.logger.Printf("chat #%s authentication failed", c.channel)
			}
		}
		c.closeConn()
		return
	}
	nick, msg, ok := parsePrivmsg(line)
	if !ok {
		return
	}
	if !mentions(msg, c.username, c.disableAtInNickname) {
		return
	}
	if c.logger != nil {
		if c.anonymizer.Enabled() {
			c.logger.EmojiEventf(":speech_balloon:", events.EventChatMention, "Chat mention in %s", c.anonymizer.Name(c.channel))
		} else {
			c.logger.EmojiEventf(":speech_balloon:", events.EventChatMention, "%s", mentionLine(nick, c.channel, msg))
		}
	}
}

func (c *Client) write(msg string) error {
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()
	if conn == nil {
		return fmt.Errorf("no active chat connection")
	}
	_, err := conn.Write([]byte(msg))
	return err
}

func (c *Client) closeConn() {
	c.connMu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()
}

func (c *Client) setConn(conn net.Conn) {
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
}

func (c *Client) isStopped() bool {
	select {
	case <-c.stop:
		return true
	default:
		return false
	}
}
