package miner

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/chat"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/privacy"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

const (
	colorGreen       = ansi.ColorGreen
	colorRed         = ansi.ColorRed
	colorCyan        = ansi.ColorCyan
	colorGameLabel   = ansi.ColorPurple
	colorDropsAccent = ansi.ColorYellow
	colorReset       = ansi.ColorReset
)

func normalizeGameList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, raw := range values {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

func normalizeStreamerList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, raw := range values {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

type Miner struct {
	Username                   string
	Password                   string
	ClaimDropsStartup          bool
	DisableSSLCertVerification bool
	StreamerSettings           domain.StreamerSettings
	StreamerOverrides          map[string]domain.StreamerSettings
	WatchStreakWarmStartCache  bool
	logger                     *logging.Logger
	startedAt                  time.Time
	twitch                     *twitch.Client
	streamers                  []*domain.Streamer
	initialPoints              map[string]int
	initialPointsMu            sync.Mutex
	startingUp                 atomic.Bool
	stop                       chan struct{}
	watchPriorities            []watchPriority
	streamerExclusions         map[string]struct{}
	gamePriority               []string
	gamePriorityIndex          map[string]int
	gameExclusions             map[string]struct{}
	chat                       *chat.Presence
	disableAtInNickname        bool
	showGameInfo               bool
	showDropsProgress          bool
	logWatchQueue              bool
	anonymizer                 *privacy.Anonymizer
	warmStartCache             *streak.WarmStartCache
	warmStartCachePath         string
	activeStreakWatches        map[string]streak.ActiveWatch
	activeStreakMu             sync.Mutex
}

// Options is everything a Miner is configured with.
type Options struct {
	Username                   string
	Password                   string
	ClaimDropsStartup          bool
	DisableSSLCertVerification bool
	Logger                     logging.LoggerSettings
	StreamerSettings           domain.StreamerSettings
	StreamerOverrides          map[string]domain.StreamerSettings
	WatchPriority              []string
	StreamersExclude           []string
	GamePriority               []string
	GameExclude                []string
	DisableAtInNickname        bool
	ShowGame                   bool
	WatchQueueLogging          bool
	WatchStreakWarmStartCache  bool
	ShowDropsProgress          bool
}

// New returns a Miner that is ready to Mine. It does not contact Twitch yet.
func New(options Options) *Miner {
	options.StreamerSettings.Default()
	gamePriority := normalizeGameList(options.GamePriority)
	m := &Miner{
		Username:                   options.Username,
		Password:                   options.Password,
		ClaimDropsStartup:          options.ClaimDropsStartup,
		DisableSSLCertVerification: options.DisableSSLCertVerification,
		StreamerSettings:           options.StreamerSettings,
		StreamerOverrides:          options.StreamerOverrides,
		WatchStreakWarmStartCache:  options.WatchStreakWarmStartCache,
		logger:                     logging.NewLogger(options.Logger, options.Username),
		watchPriorities:            parseWatchPriorities(options.WatchPriority),
		streamerExclusions:         nameSet(normalizeStreamerList(options.StreamersExclude)),
		gamePriority:               gamePriority,
		gamePriorityIndex:          positions(gamePriority),
		gameExclusions:             nameSet(normalizeGameList(options.GameExclude)),
		disableAtInNickname:        options.DisableAtInNickname,
		showGameInfo:               options.ShowGame,
		showDropsProgress:          options.ShowDropsProgress,
		logWatchQueue:              options.WatchQueueLogging,
		anonymizer:                 privacy.New(options.Logger.AnonymizeLogs),
		warmStartCachePath:         warmStartCachePathFor(options.Username),
		activeStreakWatches:        make(map[string]streak.ActiveWatch),
	}
	m.chat = chat.NewPresence(options.Username, m.chatToken, m.logger, options.DisableAtInNickname, m.anonymizer, m.styledStreamerName)
	return m
}

func nameSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

// positions maps each name to its place in the list.
func positions(names []string) map[string]int {
	index := make(map[string]int, len(names))
	for i, name := range names {
		index[name] = i
	}
	return index
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
		key := strings.ToLower(strings.TrimSpace(name))
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
