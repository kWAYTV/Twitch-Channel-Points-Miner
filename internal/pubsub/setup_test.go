package pubsub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
)

var noon = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

type gain struct {
	Streamer string
	Earned   int
	Reason   string
	Balance  int
}

type presence struct {
	Streamer string
	Online   bool
	Reason   string
}

// timer is work the client asked to have done later.
type timer struct {
	After time.Duration
	fire  func()
}

// world is what a client under test is built with. The zero value is a client
// with no streamers at noon, with plain logs.
type world struct {
	streamers []*domain.Streamer
	// now is the time the client reads; noon when left zero.
	now       time.Time
	endpoint  string
	insecure  bool
	anonymise bool
	// deepDebug gives the client a logger that offers deep debug output.
	deepDebug bool
	// noHandlers leaves out the gain and presence handlers.
	noHandlers bool
	// sleep stands in for waiting between connections; it returns at once when left nil.
	sleep func(time.Duration)
}

type harness struct {
	client  *Client
	account *fakeAccount
	logger  *fakeLogger
	deep    *fakeDeepLogger

	gains    []gain
	presence []presence
	timers   []timer
}

func setup(w world) *harness {
	h := &harness{
		account: &fakeAccount{ID: "999", Token: "token-999", Login: "miner"},
		logger:  &fakeLogger{},
	}
	var logger Logger = h.logger
	if w.deepDebug {
		h.deep = &fakeDeepLogger{fakeLogger: h.logger}
		logger = h.deep
	}
	if w.now.IsZero() {
		w.now = noon
	}
	if w.sleep == nil {
		w.sleep = func(time.Duration) {}
	}
	env := environment{
		endpoint: w.endpoint,
		now:      func() time.Time { return w.now },
		after:    func(d time.Duration, f func()) { h.timers = append(h.timers, timer{d, f}) },
		sleep:    w.sleep,
	}
	onGain := func(s *domain.Streamer, earned int, reason string, balance int) {
		h.gains = append(h.gains, gain{s.Username, earned, reason, balance})
	}
	onPresence := func(s *domain.Streamer, online bool, reason string) {
		h.presence = append(h.presence, presence{s.Username, online, reason})
	}
	if w.noHandlers {
		onGain, onPresence = nil, nil
	}
	h.client = newClient(env, h.account, logger, privacy.New(w.anonymise), w.streamers, w.insecure, onGain, onPresence)
	return h
}

// receive hands the client a message published on topic.
func (h *harness) receive(t *testing.T, topic, message string) error {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"type": "MESSAGE",
		"data": map[string]string{"topic": topic, "message": message},
	})
	if err != nil {
		t.Fatalf("building the message: %v", err)
	}
	return h.client.handleMessage(raw, nil)
}

// effects is everything a client can do that a test can see.
type effects struct {
	Lines             []string
	Gains             []gain
	Presence          []presence
	Timers            []time.Duration
	ContextLoads      []string
	BonusClaims       []claimed
	MomentClaims      []claimed
	RaidsJoined       []claimed
	Bets              []bet
	GoalContributions []string
}

func (h *harness) effects() effects {
	var timers []time.Duration
	for _, timer := range h.timers {
		timers = append(timers, timer.After)
	}
	return effects{
		Lines:             h.logger.Lines,
		Gains:             h.gains,
		Presence:          h.presence,
		Timers:            timers,
		ContextLoads:      h.account.ContextLoads,
		BonusClaims:       h.account.BonusClaims,
		MomentClaims:      h.account.MomentClaims,
		RaidsJoined:       h.account.RaidsJoined,
		Bets:              h.account.Bets,
		GoalContributions: h.account.GoalContributions,
	}
}
