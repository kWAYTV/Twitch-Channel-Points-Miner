package pubsub

import (
	"errors"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/humanize"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
)

type Logger interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	EmojiPrintf(emoji, format string, args ...interface{})
	Eventf(event events.Event, format string, args ...interface{})
	EmojiEventf(emoji string, event events.Event, format string, args ...interface{})
	ErrorEventf(event events.Event, format string, args ...interface{})
	Debugf(format string, args ...interface{})
	DebugEnabled() bool
}

var ErrReconnectRequested = errors.New("pubsub reconnect requested")

// Account is what the PubSub client does on Twitch for the logged-in account.
type Account interface {
	UserID() string
	AuthToken() string
	LoginName() string
	IsStreamLive(channelID string) (bool, error)
	LoadChannelPointsContext(streamer *domain.Streamer) (int, error)
	ClaimBonus(streamer *domain.Streamer, claimID string) error
	ClaimMoment(streamer *domain.Streamer, momentID string) error
	JoinRaid(streamer *domain.Streamer, raidID string) error
	MakePrediction(event *prediction.Event) error
	ContributeToCommunityGoals(streamer *domain.Streamer)
}

type Client struct {
	account     Account
	logger      Logger
	anonymizer  *privacy.Anonymizer
	env         environment
	streamers   []*domain.Streamer
	streamerMap map[string]*domain.Streamer
	// reacting holds what is being answered after a human pause, so a
	// message Twitch repeats meanwhile is not answered twice.
	reacting   map[string]struct{}
	reactingMu sync.Mutex

	predictions map[string]*prediction.Event
	// seenPredictions holds every event that was scheduled, or that the
	// account bet on from elsewhere, so none is scheduled a second time.
	seenPredictions map[string]struct{}
	predMu          sync.Mutex
	onGain          func(streamer *domain.Streamer, earned int, reason string, balance int)
	onPresence      func(streamer *domain.Streamer, online bool, reason string)
	onBetSettled    func(bet prediction.SettledBet)
}

// Handlers are what the client tells its owner about. Any of them may be nil.
type Handlers struct {
	// Gain is called when points are earned or spent on a channel.
	Gain func(streamer *domain.Streamer, earned int, reason string, balance int)
	// Presence is called when a channel goes online or offline.
	Presence func(streamer *domain.Streamer, online bool, reason string)
	// BetSettled is called once for each bet when its result is known.
	BetSettled func(bet prediction.SettledBet)
}

func (p *Client) debugf(format string, args ...interface{}) {
	if p.logger != nil && p.logger.DebugEnabled() {
		p.logger.Debugf(format, args...)
	}
}

// deepDebugLogger is the optional extra a Logger may offer.
type deepDebugLogger interface {
	DeepDebugf(format string, args ...interface{})
	DeepDebugEnabled() bool
}

func (p *Client) deepDebugf(format string, args ...interface{}) {
	if p == nil || p.anonymizer.Enabled() {
		return
	}
	if p.logger == nil {
		return
	}
	deep, ok := p.logger.(deepDebugLogger)
	if !ok || !deep.DeepDebugEnabled() {
		return
	}
	deep.DeepDebugf(format, args...)
}

func NewClient(
	account Account,
	logger Logger,
	anonymizer *privacy.Anonymizer,
	streamers []*domain.Streamer,
	handlers Handlers,
	human *humanize.Timing,
) *Client {
	env := twitchEnvironment()
	env.human = human
	return newClient(env, account, logger, anonymizer, streamers, handlers)
}

func newClient(
	env environment,
	account Account,
	logger Logger,
	anonymizer *privacy.Anonymizer,
	streamers []*domain.Streamer,
	handlers Handlers,
) *Client {
	streamerMap := make(map[string]*domain.Streamer)
	for _, s := range streamers {
		if s.ChannelID != "" {
			streamerMap[s.ChannelID] = s
		}
	}
	return &Client{
		account:         account,
		logger:          logger,
		anonymizer:      anonymizer,
		env:             env,
		streamers:       streamers,
		streamerMap:     streamerMap,
		reacting:        make(map[string]struct{}),
		predictions:     make(map[string]*prediction.Event),
		seenPredictions: make(map[string]struct{}),
		onGain:          handlers.Gain,
		onPresence:      handlers.Presence,
		onBetSettled:    handlers.BetSettled,
	}
}

func (p *Client) Start(stop <-chan struct{}) {
	topics, err := p.buildTopics()
	if err != nil {
		p.logger.Errorf("PubSub topic error: %v", err)
		return
	}
	go p.pollPendingClaims(stop)
	batches := chunkTopics(topics, 50)
	for i, batch := range batches {
		idx := i + 1
		go p.run(idx, batch, stop)
	}
}

func (p *Client) run(connIndex int, topics []string, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		if err := p.connectAndListen(connIndex, topics, stop); err != nil {
			if errors.Is(err, ErrReconnectRequested) {
				p.logger.Printf("PubSub[%d] reconnect requested; waiting ~60 seconds", connIndex)
				p.env.sleep(60 * time.Second)
				continue
			}
			p.logger.Errorf("PubSub[%d] connection error: %v", connIndex, err)
			p.env.sleep(10 * time.Second)
		}
	}
}
