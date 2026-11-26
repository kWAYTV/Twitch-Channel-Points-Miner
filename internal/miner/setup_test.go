package miner

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

var testNow = time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)

type harness struct {
	miner  *Miner
	twitch *fakeTwitchAPI
	chat   *fakeChatRooms
	cache  *streak.WarmStartCache
	log    *bytes.Buffer
	now    time.Time
}

// setup builds a miner from options on fakes: Twitch, the chat rooms, an in-memory
// warm-start cache when the option asks for one, a log buffer, and a clock at testNow.
func setup(t *testing.T, options Options) *harness {
	t.Helper()
	h := &harness{
		twitch: newFakeTwitchAPI(),
		chat:   &fakeChatRooms{},
		log:    &bytes.Buffer{},
		now:    testNow,
	}
	m := New(options)
	m.logger = logging.NewLoggerTo(h.log, options.Logger)
	m.twitch = h.twitch
	m.chat = h.chat
	m.clock = func() time.Time { return h.now }
	if options.WatchStreakWarmStartCache {
		h.cache = streak.NewWarmStartCache(options.Username, "")
		m.warmStartCache = h.cache
	}
	h.miner = m
	return h
}

var logPrefix = regexp.MustCompile(`^\[(\w+)\] \d{2}:\d{2} \d{2}/\d{2}/\d{2}: `)

// logLines returns what was logged, each line as "[LEVEL] message" without its timestamp.
func (h *harness) logLines(t *testing.T) []string {
	t.Helper()
	text := strings.TrimSuffix(h.log.String(), "\n")
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = logPrefix.ReplaceAllString(line, "[$1] ")
	}
	return lines
}

// historyOf copies the streamer's gains per reason, or returns nil when there are none.
func historyOf(streamer *domain.Streamer) map[string]domain.HistoryEntry {
	if streamer.History == nil {
		return nil
	}
	history := make(map[string]domain.HistoryEntry, len(streamer.History))
	for reason, entry := range streamer.History {
		history[reason] = *entry
	}
	return history
}

// logins lists the streamers' logins in order.
func logins(streamers []*domain.Streamer) []string {
	names := make([]string, 0, len(streamers))
	for _, streamer := range streamers {
		names = append(names, streamer.Username)
	}
	return names
}

// live adds an online channel to Twitch and returns its streamer, whose stream is
// due a refresh from Twitch.
func (h *harness) live(login, channelID, game string) *domain.Streamer {
	streamer := onlineStreamer(login, channelID, game)
	h.twitch.Channels[login] = &fakeChannel{
		ID:          channelID,
		Online:      true,
		BroadcastID: streamer.Stream.BroadcastID,
		Game:        game,
		CreatedAt:   streamer.Stream.CreatedAt,
	}
	return streamer
}

// freshStreamer is an online streamer whose stream was refreshed a moment ago,
// so the miner has no reason to ask Twitch about it.
func freshStreamer(login, channelID, game string) *domain.Streamer {
	streamer := onlineStreamer(login, channelID, game)
	streamer.Stream.Update(streamer.Stream.BroadcastID, "", streamer.Stream.Game, nil, 0, streamer.Stream.CreatedAt, "")
	return streamer
}

// onlineStreamer is a streamer who has been online for five minutes with a pending watch streak.
func onlineStreamer(login, channelID, game string) *domain.Streamer {
	startedAt := testNow.Add(-5 * time.Minute)
	stream := domain.NewStream()
	stream.BroadcastID = login + "-broadcast"
	stream.CreatedAt = startedAt
	stream.StreamUpAt = startedAt
	if game != "" {
		stream.Game = map[string]interface{}{"displayName": game}
	}
	return &domain.Streamer{
		Username:      login,
		ChannelID:     channelID,
		IsOnline:      true,
		PresenceKnown: true,
		OnlineAt:      startedAt,
		Stream:        stream,
	}
}
