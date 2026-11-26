package twitch

import (
	"net/http"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
)

var now = time.Date(2026, time.November, 26, 12, 0, 0, 0, time.UTC)

const (
	buildID  = "0a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
	homePage = `<script>window.__twilightBuildID="0a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d";</script>`
)

// fakeClock is the client's clock: the time it reads and the waits it asked for.
type fakeClock struct {
	Now   time.Time
	Slept []time.Duration
}

type fakes struct {
	twitch  *fakeTwitch
	session *fakeSession
	logger  *fakeLogger
	clock   *fakeClock
}

// setup builds a client logged in as user-id, on a Twitch that serves its home page.
func setup(t *testing.T) (*Client, fakes) {
	t.Helper()
	return setupWith(t, nil)
}

// setupAnonymized is setup with streamer names hidden: "streamer" is logged as "Streamer1".
func setupAnonymized(t *testing.T) (*Client, fakes) {
	t.Helper()
	return setupWith(t, privacy.New(true))
}

func setupWith(t *testing.T, anonymizer *privacy.Anonymizer) (*Client, fakes) {
	t.Helper()
	fake := fakes{
		twitch:  newFakeTwitch("token"),
		session: &fakeSession{Token: "token", ID: "user-id"},
		logger:  &fakeLogger{},
		clock:   &fakeClock{Now: now},
	}
	fake.twitch.serve("GET https://www.twitch.tv", ok(homePage))
	client := newClient(clientParts{
		username:      "miner",
		userAgent:     "agent",
		deviceID:      "device",
		clientSession: "session",
		session:       fake.session,
		http:          &http.Client{Transport: fake.twitch},
		logger:        fake.logger,
		anonymizer:    anonymizer,
		sleep:         func(d time.Duration) { fake.clock.Slept = append(fake.clock.Slept, d) },
		now:           func() time.Time { return fake.clock.Now },
	})
	return client, fake
}

// newStreamer is the channel "streamer", id 123456, with nothing known about its stream.
func newStreamer() *domain.Streamer {
	return &domain.Streamer{
		Username:  "streamer",
		ChannelID: "123456",
		Stream:    domain.NewStream(),
	}
}

// gqlHeaders is what every GQL request of the client built by setup carries.
func gqlHeaders() http.Header {
	return http.Header{
		"Authorization":     {"OAuth token"},
		"Client-Id":         {"ue6666qo983tsx6so1t0vnawi233wa"},
		"Client-Session-Id": {"session"},
		"Client-Version":    {buildID},
		"User-Agent":        {"agent"},
		"X-Device-Id":       {"device"},
		"Content-Type":      {"application/json"},
	}
}

func errorText(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func equal(got, want interface{}) bool {
	return reflect.DeepEqual(got, want)
}

// matchAll reports whether got has one value per pattern, each matching its pattern.
func matchAll(got []string, patterns ...*regexp.Regexp) bool {
	if len(got) != len(patterns) {
		return false
	}
	for i, pattern := range patterns {
		if !pattern.MatchString(got[i]) {
			return false
		}
	}
	return true
}
