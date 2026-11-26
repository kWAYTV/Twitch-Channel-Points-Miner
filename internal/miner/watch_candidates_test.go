package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

func TestStreakReadyCandidates(t *testing.T) {
	plain := watchCandidate{idx: 0}
	ready := watchCandidate{idx: 1, isStreakReady: true}
	readyToo := watchCandidate{idx: 2, isStreakReady: true}
	tests := []struct {
		name       string
		candidates []watchCandidate
		want       []watchCandidate
	}{
		{"no candidates", nil, []watchCandidate{}},
		{"none with a pending streak", []watchCandidate{plain}, []watchCandidate{}},
		{"some with a pending streak", []watchCandidate{ready, plain, readyToo}, []watchCandidate{ready, readyToo}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streakReadyCandidates(tt.candidates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("streakReadyCandidates() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAnyPriorityGame(t *testing.T) {
	offList := watchCandidate{idx: 0}
	onList := watchCandidate{idx: 1, priorityGame: true}
	tests := []struct {
		name       string
		candidates []watchCandidate
		want       bool
	}{
		{"no candidates", nil, false},
		{"none on a priority game", []watchCandidate{offList}, false},
		{"the last one on a priority game", []watchCandidate{offList, onList}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := anyPriorityGame(tt.candidates); got != tt.want {
				t.Errorf("anyPriorityGame() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSortWatchCandidates(t *testing.T) {
	a := watchCandidate{idx: 10, position: 0, rank: 2}
	b := watchCandidate{idx: 11, position: 1, rank: 0}
	c := watchCandidate{idx: 12, position: 2, rank: 2}
	highestIdxFirst := func(x, y watchCandidate) bool { return x.idx > y.idx }
	sameForAll := func(x, y watchCandidate) bool { return false }
	tests := []struct {
		name            string
		list            []watchCandidate
		less            func(a, b watchCandidate) bool
		includeGameRank bool
		want            []watchCandidate
	}{
		{"an empty list", nil, nil, true, nil},
		{"by position", []watchCandidate{c, a, b}, nil, false, []watchCandidate{a, b, c}},
		{"by game rank, then position", []watchCandidate{c, a, b}, nil, true, []watchCandidate{b, a, c}},
		{"by the given order", []watchCandidate{a, b, c}, highestIdxFirst, true, []watchCandidate{c, b, a}},
		{"a given order that ties, with game rank", []watchCandidate{c, a, b}, sameForAll, true, []watchCandidate{b, a, c}},
		{"a given order that ties, without game rank", []watchCandidate{c, a, b}, sameForAll, false, []watchCandidate{a, b, c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortWatchCandidates(tt.list, tt.less, tt.includeGameRank); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortWatchCandidates() = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("the list passed in", func(t *testing.T) {
		list := []watchCandidate{c, a, b}

		sortWatchCandidates(list, nil, false)

		if want := []watchCandidate{c, a, b}; !reflect.DeepEqual(list, want) {
			t.Errorf("list after sorting = %+v, want %+v", list, want)
		}
	})
}

func TestMiner_watchCandidates(t *testing.T) {
	options := Options{
		Username:     "account",
		GamePriority: []string{"chess"},
		GameExclude:  []string{"slots"},
	}

	t.Run("no streamers", func(t *testing.T) {
		h := setup(t, options)

		got := h.miner.watchCandidates(nil, h.now)

		if want := []watchCandidate{}; !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer on a game off the list", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		want := []watchCandidate{{idx: 0, rank: 2, game: "poker", position: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer on a priority game with a pending streak", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Chess")
		alice.Settings.WatchStreak = true

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		want := []watchCandidate{{idx: 0, rank: 0, game: "chess", position: 0, priorityGame: true, isStreakReady: true}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a missing streamer and an offline one before an online one", func(t *testing.T) {
		h := setup(t, options)
		offline := &domain.Streamer{Username: "bob", ChannelID: "2"}
		alice := h.live("alice", "1", "Poker")

		got := h.miner.watchCandidates([]*domain.Streamer{nil, offline, alice}, h.now)

		want := []watchCandidate{{idx: 2, rank: 2, game: "poker", position: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer backed off after failures", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.WatchBackoffUntil = testNow.Add(time.Minute)

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		if want := []watchCandidate{}; !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("online for 29 seconds", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.OnlineAt = testNow.Add(-29 * time.Second)

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		if want := []watchCandidate{}; !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("online for exactly 30 seconds", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.OnlineAt = testNow.Add(-30 * time.Second)

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		want := []watchCandidate{{idx: 0, rank: 2, game: "poker", position: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("online since an unknown time", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.OnlineAt = time.Time{}

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		want := []watchCandidate{{idx: 0, rank: 2, game: "poker", position: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer Twitch now reports offline", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Online = false

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		if want := []watchCandidate{}; !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
	})

	t.Run("a streamer on an excluded game", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Slots")

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		if want := []watchCandidate{}; !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
	})

	t.Run("a pending streak watched for the whole priority time", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.Settings.WatchStreak = true
		alice.Stream.MinuteWatched = 7

		got := h.miner.watchCandidates([]*domain.Streamer{alice}, h.now)

		want := []watchCandidate{{idx: 0, rank: 2, game: "poker", position: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("watchCandidates() = %+v, want %+v", got, want)
		}
		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		entry, _ := h.cache.Get("alice")
		wantEntry := streak.WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "alice",
			ChannelID:       "1",
			BroadcastID:     "alice-broadcast",
			StreamCreatedAt: testNow.Add(-5 * time.Minute),
			CheckedAt:       testNow,
			IsOnline:        true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})
}

func TestMiner_refreshStreamForPreference(t *testing.T) {
	options := Options{Username: "account"}

	t.Run("before login", func(t *testing.T) {
		m := New(Options{})
		alice := onlineStreamer("alice", "1", "")

		m.refreshStreamForPreference(alice)

		if got := alice.Stream.GameName(); got != "" {
			t.Errorf("game = %q, want none", got)
		}
	})

	t.Run("an offline streamer", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.IsOnline = false
		alice.Stream.Game = map[string]interface{}{}

		h.miner.refreshStreamForPreference(alice)

		if got := alice.Stream.GameName(); got != "" {
			t.Errorf("game = %q, want none", got)
		}
	})

	t.Run("a stream with no game yet", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.Stream.Game = map[string]interface{}{}

		h.miner.refreshStreamForPreference(alice)

		if got := alice.Stream.GameName(); got != "Poker" {
			t.Errorf("game = %q, want %q", got, "Poker")
		}
	})

	t.Run("a stream with a game that is due a refresh", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Game = "Chess"

		h.miner.refreshStreamForPreference(alice)

		if got := alice.Stream.GameName(); got != "Chess" {
			t.Errorf("game = %q, want %q", got, "Chess")
		}
	})

	t.Run("a stream with a game, refreshed a moment ago", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Online: true, BroadcastID: "alice-broadcast", Game: "Chess"}

		h.miner.refreshStreamForPreference(alice)

		if got := alice.Stream.GameName(); got != "Poker" {
			t.Errorf("game = %q, want %q", got, "Poker")
		}
	})

	t.Run("a streamer Twitch now reports offline", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Online = false

		h.miner.refreshStreamForPreference(alice)

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
		if !alice.OfflineAt.Equal(testNow) {
			t.Errorf("OfflineAt = %s, want %s", alice.OfflineAt, testNow)
		}
		wantLog := []string{"[INFO] Alice (" + colorCyan + "0" + colorReset + " points) is " + colorRed + "Offline" + colorReset + "!"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		debug := options
		debug.Logger = logging.LoggerSettings{Debug: true}
		h := setup(t, debug)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Fail["UpdateStream"] = errors.New("503")

		h.miner.refreshStreamForPreference(alice)

		if !alice.IsOnline {
			t.Errorf("IsOnline = false, want true")
		}
		wantLog := []string{"[DEBUG] stream info Alice: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch reporting the streak complete", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].StreakComplete = true

		h.miner.refreshStreamForPreference(alice)

		entry, _ := h.cache.Get("alice")
		wantEntry := streak.WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "alice",
			ChannelID:       "1",
			BroadcastID:     "alice-broadcast",
			StreamCreatedAt: testNow.Add(-5 * time.Minute),
			CheckedAt:       testNow,
			IsOnline:        true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})

	t.Run("a new broadcast after a completed streak", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		alice.CompletedWatchStreak = true
		h.twitch.Channels["alice"].BroadcastID = "alice-broadcast-2"

		h.miner.refreshStreamForPreference(alice)

		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		if want := testNow.Add(30 * time.Minute); !alice.ResolvedStreakCarryoverUntil.Equal(want) {
			t.Errorf("ResolvedStreakCarryoverUntil = %s, want %s", alice.ResolvedStreakCarryoverUntil, want)
		}
		entry, _ := h.cache.Get("alice")
		wantEntry := streak.WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "alice",
			ChannelID:       "1",
			BroadcastID:     "alice-broadcast-2",
			StreamCreatedAt: testNow.Add(-5 * time.Minute),
			CheckedAt:       testNow,
			IsOnline:        true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})

	t.Run("nothing new from Twitch", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "")

		h.miner.refreshStreamForPreference(alice)

		if entry, ok := h.cache.Get("alice"); ok {
			t.Errorf("cached entry = %+v, want none", entry)
		}
	})
}
