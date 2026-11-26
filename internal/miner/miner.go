package miner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/chat"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

const (
	colorGreen       = ansi.ColorGreen
	colorRed         = ansi.ColorRed
	colorCyan        = ansi.ColorCyan
	colorGameLabel   = ansi.ColorPurple
	colorDropsAccent = ansi.ColorYellow
	colorReset       = ansi.ColorReset
)

type Miner struct {
	Username            string
	StreamerSettings    domain.StreamerSettings
	StreamerOverrides   map[string]domain.StreamerSettings
	logger              *logging.Logger
	startedAt           time.Time
	twitch              twitchAPI
	clock               func() time.Time
	streamers           []*domain.Streamer
	initialPoints       map[string]int
	initialPointsMu     sync.Mutex
	startingUp          atomic.Bool
	stop                chan struct{}
	watchPriorities     []watchPriority
	streamerExclusions  map[string]struct{}
	gamePriority        []string
	gamePriorityIndex   map[string]int
	gameExclusions      map[string]struct{}
	chat                chatRooms
	disableAtInNickname bool
	showGameInfo        bool
	showDropsProgress   bool
	logWatchQueue       bool
	anonymizer          *privacy.Anonymizer
	warmStartCache      *streak.WarmStartCache
	warmStartCachePath  string
	activeStreakWatches map[string]streak.ActiveWatch
	activeStreakMu      sync.Mutex
}

// Options is everything a Miner is configured with.
type Options struct {
	Username            string
	Logger              logging.LoggerSettings
	StreamerSettings    domain.StreamerSettings
	StreamerOverrides   map[string]domain.StreamerSettings
	WatchPriority       []string
	StreamersExclude    []string
	GamePriority        []string
	GameExclude         []string
	DisableAtInNickname bool
	ShowGame            bool
	WatchQueueLogging   bool
	ShowDropsProgress   bool
}

// New returns a Miner that is ready to Mine. It does not contact Twitch yet.
func New(options Options) *Miner {
	options.StreamerSettings.Default()
	gamePriority := normalizeNames(options.GamePriority)
	m := &Miner{
		Username:            options.Username,
		StreamerSettings:    options.StreamerSettings,
		StreamerOverrides:   options.StreamerOverrides,
		logger:              logging.NewLogger(options.Logger, options.Username),
		initialPoints:       make(map[string]int),
		watchPriorities:     parseWatchPriorities(options.WatchPriority),
		streamerExclusions:  nameSet(normalizeNames(options.StreamersExclude)),
		gamePriority:        gamePriority,
		gamePriorityIndex:   positions(gamePriority),
		gameExclusions:      nameSet(normalizeNames(options.GameExclude)),
		disableAtInNickname: options.DisableAtInNickname,
		showGameInfo:        options.ShowGame,
		showDropsProgress:   options.ShowDropsProgress,
		logWatchQueue:       options.WatchQueueLogging,
		anonymizer:          privacy.New(options.Logger.AnonymizeLogs),
		warmStartCachePath:  warmStartCachePathFor(options.Username),
		activeStreakWatches: make(map[string]streak.ActiveWatch),
	}
	m.chat = chat.NewPresence(options.Username, m.chatToken, m.logger, options.DisableAtInNickname, m.anonymizer, m.styledStreamerName)
	return m
}

// now is the miner's clock: the wall clock unless another was given.
func (m *Miner) now() time.Time {
	if m == nil || m.clock == nil {
		return time.Now()
	}
	return m.clock()
}

// chatToken is empty until the miner has logged in.
func (m *Miner) chatToken() string {
	if m.twitch == nil {
		return ""
	}
	return m.twitch.ChatToken()
}

// Mine runs the miner for an explicit list of streamers.
func (m *Miner) Mine(streamers []string) {
	m.run(streamers, false, domain.FollowersOrderASC)
}

// MineFollowers runs the miner using the follower list.
func (m *Miner) MineFollowers(order domain.FollowersOrder) {
	m.run(nil, true, order)
}

func (m *Miner) filterExcludedTargets(targets []string) ([]string, int) {
	if len(targets) == 0 || len(m.streamerExclusions) == 0 {
		return targets, 0
	}
	filtered := make([]string, 0, len(targets))
	excluded := 0
	for _, name := range targets {
		key := domain.LoginKey(name)
		if key == "" {
			continue
		}
		if _, ok := m.streamerExclusions[key]; ok {
			excluded++
			continue
		}
		filtered = append(filtered, name)
	}
	return filtered, excluded
}
