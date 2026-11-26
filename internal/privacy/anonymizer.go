package privacy

import (
	"strconv"
	"strings"
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
)

type Anonymizer struct {
	enabled bool

	mu sync.Mutex

	nextStreamerIndex int
	streamerAliases   map[string]string
	points            map[string]*pointsState

	initialPoints func() int
}

type pointsState struct {
	initialized bool
	lastReal    int
	pseudo      int
}

func New(enabled bool) *Anonymizer {
	return newAnonymizer(enabled, func() int { return random.Int(100, 1000) })
}

// newAnonymizer takes initialPoints, the made-up balance a streamer starts from.
func newAnonymizer(enabled bool, initialPoints func() int) *Anonymizer {
	return &Anonymizer{
		enabled:           enabled,
		nextStreamerIndex: 1,
		streamerAliases:   make(map[string]string),
		points:            make(map[string]*pointsState),
		initialPoints:     initialPoints,
	}
}

func (a *Anonymizer) Enabled() bool {
	return a != nil && a.enabled
}

func (a *Anonymizer) StreamerName(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	if name := strings.TrimSpace(streamer.Username); name != "" {
		return a.Name(name)
	}
	if id := strings.TrimSpace(streamer.ChannelID); id != "" {
		return a.Name("id:" + id)
	}
	return ""
}

// Shown is the name a streamer goes by in logs: the stand-in while logs are
// anonymised, and otherwise the login with a capital.
func (a *Anonymizer) Shown(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	if a.Enabled() {
		return a.StreamerName(streamer)
	}
	return streamer.DisplayName()
}

func (a *Anonymizer) Name(raw string) string {
	if !a.Enabled() {
		return raw
	}
	key := domain.LoginKey(raw)
	if key == "" {
		return ""
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if alias, ok := a.streamerAliases[key]; ok {
		return alias
	}
	alias := "Streamer" + strconv.Itoa(a.nextStreamerIndex)
	a.nextStreamerIndex++
	a.streamerAliases[key] = alias
	return alias
}

func (a *Anonymizer) PseudoChannelPoints(streamer *domain.Streamer) int {
	if streamer == nil {
		return 0
	}
	if !a.Enabled() {
		return streamer.ChannelPoints
	}

	key := streamerKey(streamer)
	if key == "" {
		return streamer.ChannelPoints
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	state := a.points[key]
	if state == nil {
		state = &pointsState{}
		a.points[key] = state
	}

	if !state.initialized {
		state.pseudo = a.initialPoints()
		state.lastReal = streamer.ChannelPoints
		state.initialized = true
		return state.pseudo
	}

	delta := streamer.ChannelPoints - state.lastReal
	state.pseudo += delta
	state.lastReal = streamer.ChannelPoints
	return state.pseudo
}
