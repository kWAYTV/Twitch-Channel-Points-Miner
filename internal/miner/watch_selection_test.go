package miner

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// streakStreamer is a fresh online streamer with a pending watch streak, watched for minutes so far.
func streakStreamer(login, channelID, game string, minutes float64) *domain.Streamer {
	streamer := freshStreamer(login, channelID, game)
	streamer.Settings.WatchStreak = true
	streamer.Stream.MinuteWatched = minutes
	return streamer
}

func assertPicked(t *testing.T, got []*domain.Streamer, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if names := logins(got); !reflect.DeepEqual(names, want) {
		t.Errorf("picked %q, want %q", names, want)
	}
}

func TestMiner_gamePreference(t *testing.T) {
	playing := func(game string) *domain.Streamer {
		return &domain.Streamer{Stream: &domain.Stream{Game: map[string]interface{}{"displayName": game}}}
	}
	lists := Options{GamePriority: []string{"valorant", "wow"}, GameExclude: []string{"banned"}}
	tests := []struct {
		name         string
		options      Options
		streamer     *domain.Streamer
		wantRank     int
		wantExcluded bool
	}{
		{"no streamer", lists, nil, 3, false},
		{"a streamer with no stream", lists, &domain.Streamer{}, 3, false},
		{"a stream with no game", lists, playing("  "), 3, false},
		{"the first priority game, in capitals with spaces", lists, playing(" Valorant "), 0, false},
		{"the second priority game", lists, playing("WoW"), 1, false},
		{"an excluded game", lists, playing("Banned"), 0, true},
		{"a game on neither list", lists, playing("Chess"), 3, false},
		{"a game with no lists set", Options{}, playing("Chess"), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)

			rank, excluded := h.miner.gamePreference(tt.streamer)

			if rank != tt.wantRank || excluded != tt.wantExcluded {
				t.Errorf("gamePreference() = %d, %t, want %d, %t", rank, excluded, tt.wantRank, tt.wantExcluded)
			}
		})
	}
}

func TestMiner_pickByPriority(t *testing.T) {
	t.Run("ORDER", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("ORDER with the first two on one game: list order wins", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Poker")
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("STREAK", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"STREAK"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := streakStreamer("carol", "3", "Go", 0)

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("STREAK with two pending streaks on one game", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"STREAK"}})
		alice := streakStreamer("alice", "1", "Poker", 0)
		bob := streakStreamer("bob", "2", "Poker", 0)
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("DROPS with two streams that carry drops", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := dropsStreamer("bob", "2", "Chess")
		carol := dropsStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		// One slot for drops, the other back to the top of the list.
		assertPicked(t, got, "bob", "alice")
	})

	t.Run("DROPS with drop claims on but no drops on the stream", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS", "ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := freshStreamer("carol", "3", "Go")
		carol.Settings.ClaimDrops = true

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("DROPS with drops on the stream but drop claims off", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS", "ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := dropsStreamer("carol", "3", "Go")
		carol.Settings.ClaimDrops = false

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("DROPS with drops known by campaign only", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS", "ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := freshStreamer("carol", "3", "Go")
		carol.Settings.ClaimDrops = true
		carol.Stream.CampaignIDs = []string{"campaign-1"}

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("DROPS with drops on a priority game further down the list", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS", "ORDER"}, GamePriority: []string{"go", "chess"}})
		alice := dropsStreamer("alice", "1", "Poker")
		bob := dropsStreamer("bob", "2", "Chess")
		carol := dropsStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		// Go is first in game_priority, so its drops win the drops slot.
		assertPicked(t, got, "carol", "alice")
	})

	t.Run("DROPS with drops only on a game outside game_priority", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS", "ORDER"}, GamePriority: []string{"go"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := dropsStreamer("carol", "3", "Wardogs")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("DROPS when a drops stream already holds a slot", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"ORDER", "DROPS"}})
		alice := dropsStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := dropsStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("SUBSCRIBED", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"SUBSCRIBED"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		bob.ActiveMultipliers = []domain.ActiveMultiplier{{Factor: 1.2}}
		carol := freshStreamer("carol", "3", "Go")
		carol.ActiveMultipliers = []domain.ActiveMultiplier{{Factor: 2}}

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "bob")
	})

	t.Run("SUBSCRIBED with nobody subscribed, then POINTS_ASC", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"SUBSCRIBED", "POINTS_ASC"}})
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 1_000
		bob := freshStreamer("bob", "2", "Chess")
		bob.ChannelPoints = 10
		carol := freshStreamer("carol", "3", "Go")
		carol.ChannelPoints = 20

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "carol")
	})

	t.Run("POINTS_ASC", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"POINTS_ASC"}})
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 900
		bob := freshStreamer("bob", "2", "Chess")
		bob.ChannelPoints = 1_000
		carol := freshStreamer("carol", "3", "Go")
		carol.ChannelPoints = 10

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("POINTS_DESC", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"POINTS_DESC"}})
		alice := freshStreamer("alice", "1", "Poker")
		alice.ChannelPoints = 900
		bob := freshStreamer("bob", "2", "Chess")
		bob.ChannelPoints = 1_000
		carol := freshStreamer("carol", "3", "Go")
		carol.ChannelPoints = 10

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "alice")
	})

	t.Run("STREAK, then ORDER once the first is full", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"STREAK", "ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := streakStreamer("bob", "2", "Chess", 0)
		carol := streakStreamer("carol", "3", "Go", 0)

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "carol")
	})
}

func TestMiner_pickStreamersToWatch(t *testing.T) {
	t.Run("no streamers", func(t *testing.T) {
		h := setup(t, Options{})

		got := h.miner.pickStreamersToWatch(nil)

		assertPicked(t, got)
	})

	t.Run("nobody online", func(t *testing.T) {
		h := setup(t, Options{})
		alice := &domain.Streamer{Username: "alice", ChannelID: "1"}

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		assertPicked(t, got)
	})

	t.Run("one streamer online", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		assertPicked(t, got, "alice")
	})

	t.Run("a pending streak last in the list", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := streakStreamer("carol", "3", "Go", 3)

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("a pending streak the priorities never ask for", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"ORDER"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := streakStreamer("carol", "3", "Go", 0)

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "carol")
	})

	t.Run("a pending streak on the game already picked, with one slot free", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"SUBSCRIBED"}})
		alice := freshStreamer("alice", "1", "Just Chatting")
		alice.ActiveMultipliers = []domain.ActiveMultiplier{{Factor: 1.2}}
		bob := streakStreamer("bob", "2", "Just Chatting", 0)
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("a pending streak on a game off the priority list", func(t *testing.T) {
		h := setup(t, Options{GamePriority: []string{"chess"}})
		alice := streakStreamer("alice", "1", "Poker", 0)
		bob := freshStreamer("bob", "2", "Chess")
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "alice")
	})

	t.Run("a pending streak on a priority game", func(t *testing.T) {
		h := setup(t, Options{GamePriority: []string{"go"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := streakStreamer("carol", "3", "Go", 0)

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("a streamer backed off after failures", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"STREAK", "ORDER"}})
		alice := streakStreamer("alice", "1", "Poker", 0)
		alice.WatchBackoffUntil = testNow.Add(10 * time.Minute)
		bob := streakStreamer("bob", "2", "Poker", 0)
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "carol")
	})

	t.Run("nothing asked for by the priorities", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"DROPS"}, GamePriority: []string{"go"}})
		alice := freshStreamer("alice", "1", "Poker")
		bob := freshStreamer("bob", "2", "Chess")
		carol := freshStreamer("carol", "3", "Go")

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})
}

func TestMiner_keepActiveStreaks(t *testing.T) {
	options := Options{GamePriority: []string{"priority game"}}

	t.Run("two streaks in progress when a priority-game streak appears", func(t *testing.T) {
		h := setup(t, options)
		alice := streakStreamer("alice", "1", "Poker", 1.5)
		bob := streakStreamer("bob", "2", "Chess", 2)
		carol := streakStreamer("carol", "3", "Priority Game", 0)
		h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob})

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "alice", "bob")
	})

	t.Run("two streaks not started when a priority-game streak appears", func(t *testing.T) {
		h := setup(t, options)
		alice := streakStreamer("alice", "1", "Poker", 0)
		bob := streakStreamer("bob", "2", "Chess", 0)
		carol := streakStreamer("carol", "3", "Priority Game", 0)
		h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob})

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "carol", "alice")
	})

	t.Run("a streak in progress that has since resolved", func(t *testing.T) {
		h := setup(t, options)
		alice := streakStreamer("alice", "1", "Poker", 1.5)
		bob := streakStreamer("bob", "2", "Chess", 2)
		carol := streakStreamer("carol", "3", "Priority Game", 0)
		h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob})
		alice.Stream.WatchStreakMissing = false

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "carol")
	})

	t.Run("a streak in progress whose broadcast has since changed", func(t *testing.T) {
		h := setup(t, options)
		alice := streakStreamer("alice", "1", "Poker", 1.5)
		bob := streakStreamer("bob", "2", "Chess", 2)
		carol := streakStreamer("carol", "3", "Priority Game", 0)
		h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob})
		alice.Stream.BroadcastID = "alice-broadcast-2"

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{alice, bob, carol})

		assertPicked(t, got, "bob", "carol")
	})

	t.Run("a streak in progress on the game of a streak not started", func(t *testing.T) {
		h := setup(t, Options{WatchPriority: []string{"STREAK"}})
		alice := streakStreamer("alice", "1", "Just Chatting", 2)
		bob := streakStreamer("bob", "2", "Just Chatting", 0)
		carol := freshStreamer("carol", "3", "Go")
		h.miner.pickStreamersToWatch([]*domain.Streamer{alice})

		got := h.miner.pickStreamersToWatch([]*domain.Streamer{carol, bob, alice})

		assertPicked(t, got, "alice", "bob")
	})
}
