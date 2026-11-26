package twitch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/auth"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

var (
	ErrStreamerOffline = errors.New("streamer offline")
	ErrChannelNotFound = errors.New("channel not found")
)

type debugLogger interface {
	Debugf(format string, args ...interface{})
	DebugEnabled() bool
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorEventf(event events.Event, format string, args ...interface{})
	EmojiPrintf(emoji, format string, args ...interface{})
}

type deepDebugLogger interface {
	DeepDebugf(format string, args ...interface{})
	DeepDebugEnabled() bool
}

// session is what the client needs from a logged-in account.
type session interface {
	Login(cookiesPath string) error
	AuthToken() string
	UserID() string
}

type Client struct {
	username       string
	userAgent      string
	deviceID       string
	clientSession  string
	clientVersion  string
	versionMu      sync.Mutex
	versionFetched time.Time
	versionTTL     time.Duration
	twitchLogin    session
	client         *http.Client
	twilightRegexp *regexp.Regexp
	settingsRegex  *regexp.Regexp
	spadeRegex     *regexp.Regexp
	logger         debugLogger
	anonymizer     *privacy.Anonymizer
	onGameChange   func(streamer *domain.Streamer, previous, current string)
	dropClaims     dropClaims
	// playlistFailures holds the streamers whose playlist check is failing,
	// so the failure is logged once and not on every minute watched.
	playlistFailures sync.Map
	// sleep replaces time.Sleep when set, so waits can be skipped under test.
	sleep func(time.Duration)
	// now replaces time.Now when set.
	now func() time.Time
	// lookupHost resolves a host name; nil skips the check that needs it.
	lookupHost       func(ctx context.Context, host string) ([]string, error)
	watchReportCheck sync.Once
	// loginExpiredReported is set while an expired login has been reported
	// and no request has worked since.
	loginExpiredReported atomic.Bool
}

// clientParts is everything a Client is built from: who it is, and how it
// reaches the account, Twitch, the log and the clock.
type clientParts struct {
	username      string
	userAgent     string
	deviceID      string
	clientSession string
	session       session
	http          *http.Client
	logger        debugLogger
	anonymizer    *privacy.Anonymizer
	sleep         func(time.Duration)
	now           func() time.Time
	lookupHost    func(ctx context.Context, host string) ([]string, error)
}

func NewClient(username, userAgent string, logger debugLogger, anonymizer *privacy.Anonymizer) (*Client, error) {
	deviceID := random.String(32)
	login, err := auth.NewLogin(twitchapi.ClientID, deviceID, username, userAgent)
	if err != nil {
		return nil, err
	}

	return newClient(clientParts{
		username:      username,
		userAgent:     userAgent,
		deviceID:      deviceID,
		clientSession: random.Hex(8),
		session:       login,
		http:          login.Client(),
		logger:        logger,
		anonymizer:    anonymizer,
		lookupHost:    net.DefaultResolver.LookupHost,
	}), nil
}

func newClient(parts clientParts) *Client {
	return &Client{
		username:       parts.username,
		userAgent:      parts.userAgent,
		deviceID:       parts.deviceID,
		clientSession:  parts.clientSession,
		clientVersion:  twitchapi.ClientVersion,
		versionTTL:     10 * time.Hour,
		twitchLogin:    parts.session,
		client:         parts.http,
		twilightRegexp: regexp.MustCompile(`window\.__twilightBuildID\s*=\s*"([0-9a-fA-F\-]{36})"`),
		settingsRegex:  regexp.MustCompile(`(https://static\.twitchcdn\.net/config/settings.*?\.js|https://assets\.twitch\.tv/config/settings.*?\.js)`),
		spadeRegex:     regexp.MustCompile(`"spade_url":"(.*?)"`),
		logger:         parts.logger,
		anonymizer:     parts.anonymizer,
		sleep:          parts.sleep,
		now:            parts.now,
		lookupHost:     parts.lookupHost,
	}
}

// SetGameChangeHandler registers a callback fired whenever stream game metadata changes.
func (t *Client) SetGameChangeHandler(handler func(streamer *domain.Streamer, previous, current string)) {
	t.onGameChange = handler
}

func (t *Client) Login(username string) error {
	cookiesPath := filepath.Join("cookies", fmt.Sprintf("%s.json", username))
	if err := t.twitchLogin.Login(cookiesPath); err != nil {
		return err
	}
	return nil
}

func (t *Client) ChatToken() string {
	if t == nil || t.twitchLogin == nil {
		return ""
	}
	return t.twitchLogin.AuthToken()
}

func (t *Client) debugf(format string, args ...interface{}) {
	if t.logger != nil && t.logger.DebugEnabled() {
		t.logger.Debugf(format, args...)
	}
}

// deepDebugEnabled reports whether whole requests and answers are logged,
// which they never are while logs are anonymised.
func (t *Client) deepDebugEnabled() bool {
	if t == nil || t.anonymizer.Enabled() {
		return false
	}
	deep, ok := t.logger.(deepDebugLogger)
	return ok && deep.DeepDebugEnabled()
}

func (t *Client) deepDebugf(format string, args ...interface{}) {
	if t.deepDebugEnabled() {
		t.logger.(deepDebugLogger).DeepDebugf(format, args...)
	}
}

// UpdateClientVersion refreshes the Twitch build id used for GQL calls.
// clientVersionRetry is how long after a failed fetch the version is asked
// for again, where a fetch that worked is kept for the whole TTL.
const clientVersionRetry = 5 * time.Minute

func (t *Client) retryClientVersionSoon(ttl time.Duration) {
	t.versionMu.Lock()
	defer t.versionMu.Unlock()
	t.versionFetched = t.timeNow().Add(clientVersionRetry - ttl)
}

func (t *Client) UpdateClientVersion() string {
	if t == nil {
		return twitchapi.ClientVersion
	}

	t.versionMu.Lock()
	ttl := t.versionTTL
	if ttl <= 0 {
		ttl = 10 * time.Hour
		t.versionTTL = ttl
	}
	if !t.versionFetched.IsZero() && t.timeNow().Sub(t.versionFetched) < ttl {
		version := t.clientVersion
		t.versionMu.Unlock()
		return version
	}
	t.versionFetched = t.timeNow()
	version := t.clientVersion
	client := t.client
	t.versionMu.Unlock()

	if client == nil {
		return version
	}

	resp, err := client.Get(twitchapi.URL)
	if err != nil {
		t.retryClientVersionSoon(ttl)
		return version
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.debugf("UpdateClientVersion request failed with status %d", resp.StatusCode)
		t.retryClientVersionSoon(ttl)
		return version
	}
	m := t.twilightRegexp.FindStringSubmatch(string(body))
	if len(m) > 1 {
		t.versionMu.Lock()
		t.clientVersion = m[1]
		version = t.clientVersion
		t.versionMu.Unlock()
		t.debugf("Client version updated to %s", version)
		return version
	} else {
		t.debugf("UpdateClientVersion: unable to extract build id")
	}
	t.versionMu.Lock()
	version = t.clientVersion
	t.versionMu.Unlock()
	return version
}

// AuthToken is the OAuth token of the logged-in account.
func (t *Client) AuthToken() string {
	return t.twitchLogin.AuthToken()
}

// UserID is the Twitch user id of the logged-in account.
func (t *Client) UserID() string {
	return t.twitchLogin.UserID()
}

// LoginName is the login the session was opened with.
func (t *Client) LoginName() string {
	return t.username
}

// timeNow is the current time.
func (t *Client) timeNow() time.Time {
	if t.now != nil {
		return t.now()
	}
	return time.Now()
}

// pause waits for d between the steps of a slow exchange with Twitch.
func (t *Client) pause(d time.Duration) {
	if t.sleep != nil {
		t.sleep(d)
		return
	}
	time.Sleep(d)
}
