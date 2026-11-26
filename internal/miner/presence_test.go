package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

const (
	onlineLine  = "[INFO] Alice (" + colorCyan + "0" + colorReset + " points) is " + colorGreen + "Online" + colorReset + "!"
	offlineLine = "[INFO] Alice (" + colorCyan + "0" + colorReset + " points) is " + colorRed + "Offline" + colorReset + "!"
)

// aliceEntry is what the warm-start cache holds for onlineStreamer("alice", "1", ...) at testNow.
func aliceEntry(online, streakMissing bool) streak.WarmStartEntry {
	return streak.WarmStartEntry{
		AccountName:        "account",
		StreamerLogin:      "alice",
		ChannelID:          "1",
		BroadcastID:        "alice-broadcast",
		StreamCreatedAt:    testNow.Add(-5 * time.Minute),
		WatchStreakMissing: streakMissing,
		CheckedAt:          testNow,
		IsOnline:           online,
	}
}

func TestMiner_setPresence(t *testing.T) {
	options := Options{Username: "account"}

	t.Run("first seen, online", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.PresenceKnown = false
		alice.IsOnline = false
		alice.OnlineAt = time.Time{}

		h.miner.setPresence(alice, true, "poll")

		if !alice.PresenceKnown || !alice.IsOnline {
			t.Errorf("PresenceKnown, IsOnline = %t, %t, want true, true", alice.PresenceKnown, alice.IsOnline)
		}
		if !alice.OnlineAt.Equal(testNow) {
			t.Errorf("OnlineAt = %s, want %s", alice.OnlineAt, testNow)
		}
		if want := []chatUpdate{{Login: "alice", Online: true}}; !reflect.DeepEqual(h.chat.Updates, want) {
			t.Errorf("chat updates = %+v, want %+v", h.chat.Updates, want)
		}
		if got, want := h.logLines(t), []string{onlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(true, true) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(true, true))
		}
	})

	t.Run("first seen, offline", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.PresenceKnown = false
		alice.IsOnline = false

		h.miner.setPresence(alice, false, "poll")

		if !alice.PresenceKnown || alice.IsOnline {
			t.Errorf("PresenceKnown, IsOnline = %t, %t, want true, false", alice.PresenceKnown, alice.IsOnline)
		}
		if !alice.OfflineAt.Equal(testNow) {
			t.Errorf("OfflineAt = %s, want %s", alice.OfflineAt, testNow)
		}
		if want := []chatUpdate{{Login: "alice", Online: false}}; !reflect.DeepEqual(h.chat.Updates, want) {
			t.Errorf("chat updates = %+v, want %+v", h.chat.Updates, want)
		}
		if got, want := h.logLines(t), []string{offlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(false, true) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(false, true))
		}
	})

	t.Run("online, then offline after a completed streak", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.CompletedWatchStreak = true
		alice.Stream.WatchStreakMissing = false

		h.miner.setPresence(alice, false, "poll")

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
		if !alice.OfflineAt.Equal(testNow) {
			t.Errorf("OfflineAt = %s, want %s", alice.OfflineAt, testNow)
		}
		if want := testNow.Add(30 * time.Minute); !alice.ResolvedStreakCarryoverUntil.Equal(want) {
			t.Errorf("ResolvedStreakCarryoverUntil = %s, want %s", alice.ResolvedStreakCarryoverUntil, want)
		}
		if got, want := h.logLines(t), []string{offlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(false, false) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(false, false))
		}
	})

	t.Run("online, then offline while a carryover runs", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.ResolvedStreakCarryover = true
		alice.ResolvedStreakCarryoverUntil = testNow.Add(5 * time.Minute)

		h.miner.setPresence(alice, false, "poll")

		if want := testNow.Add(30 * time.Minute); !alice.ResolvedStreakCarryoverUntil.Equal(want) {
			t.Errorf("ResolvedStreakCarryoverUntil = %s, want %s", alice.ResolvedStreakCarryoverUntil, want)
		}
	})

	t.Run("online, then offline with no completed streak", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.setPresence(alice, false, "poll")

		if alice.ResolvedStreakCarryover || !alice.ResolvedStreakCarryoverUntil.IsZero() {
			t.Errorf("carryover = %t until %s, want none", alice.ResolvedStreakCarryover, alice.ResolvedStreakCarryoverUntil)
		}
	})

	t.Run("offline, then online while a carryover runs", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.IsOnline = false
		alice.ResolvedStreakCarryover = true
		alice.ResolvedStreakCarryoverUntil = testNow.Add(10 * time.Minute)

		h.miner.setPresence(alice, true, "poll")

		if !alice.OnlineAt.Equal(testNow) {
			t.Errorf("OnlineAt = %s, want %s", alice.OnlineAt, testNow)
		}
		if !alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = false, want true")
		}
		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
		if got, want := h.logLines(t), []string{onlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("offline, then online after the carryover ran out", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.IsOnline = false
		alice.CompletedWatchStreak = true
		alice.ResolvedStreakCarryover = true
		alice.ResolvedStreakCarryoverUntil = testNow.Add(-time.Second)

		h.miner.setPresence(alice, true, "poll")

		if alice.CompletedWatchStreak {
			t.Errorf("CompletedWatchStreak = true, want false")
		}
		if alice.ResolvedStreakCarryover || !alice.ResolvedStreakCarryoverUntil.IsZero() {
			t.Errorf("carryover = %t until %s, want none", alice.ResolvedStreakCarryover, alice.ResolvedStreakCarryoverUntil)
		}
		if !alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = false, want true")
		}
	})

	t.Run("online again", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.setPresence(alice, true, "poll")

		if want := testNow.Add(-5 * time.Minute); !alice.OnlineAt.Equal(want) {
			t.Errorf("OnlineAt = %s, want %s", alice.OnlineAt, want)
		}
		if want := []chatUpdate{{Login: "alice", Online: true}}; !reflect.DeepEqual(h.chat.Updates, want) {
			t.Errorf("chat updates = %+v, want %+v", h.chat.Updates, want)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(true, true) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(true, true))
		}
	})

	t.Run("offline again", func(t *testing.T) {
		h := setup(t, options)
		alice := freshStreamer("alice", "1", "Poker")
		alice.IsOnline = false
		alice.OfflineAt = testNow.Add(-time.Hour)

		h.miner.setPresence(alice, false, "poll")

		if want := testNow.Add(-time.Hour); !alice.OfflineAt.Equal(want) {
			t.Errorf("OfflineAt = %s, want %s", alice.OfflineAt, want)
		}
		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
		if entry, _ := h.cache.Get("alice"); entry != aliceEntry(false, true) {
			t.Errorf("cached entry = %+v, want %+v", entry, aliceEntry(false, true))
		}
	})

	t.Run("first seen online, with showing games on and streak details from Twitch", func(t *testing.T) {
		showGame := options
		showGame.ShowGame = true
		h := setup(t, showGame)
		h.twitch.ExtraInfo = " | streak length 3"
		alice := h.live("alice", "1", "")
		h.twitch.Channels["alice"].Game = "Chess"
		alice.PresenceKnown = false
		alice.IsOnline = false

		h.miner.setPresence(alice, true, "poll")

		want := []string{onlineLine + " | " + colorGameLabel + "Playing:" + colorReset + " Chess | streak length 3"}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("first seen offline, with streak details from Twitch", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.ExtraInfo = " | streak length 3"
		alice := freshStreamer("alice", "1", "Poker")
		alice.PresenceKnown = false
		alice.IsOnline = false

		h.miner.setPresence(alice, false, "poll")

		want := []string{offlineLine + " | streak length 3"}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("with no warm-start cache", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.setPresence(alice, false, "poll")

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
	})
}

func TestMiner_handlePubSubPresence(t *testing.T) {
	t.Run("a stream going down", func(t *testing.T) {
		h := setup(t, Options{})
		alice := freshStreamer("alice", "1", "Poker")

		h.miner.handlePubSubPresence(alice, false, "stream-down")

		if alice.IsOnline {
			t.Errorf("IsOnline = true, want false")
		}
		if got, want := h.logLines(t), []string{offlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})
}

func TestMiner_updatePresence(t *testing.T) {
	options := Options{Username: "account"}
	unseen := func(h *harness) *domain.Streamer {
		alice := h.live("alice", "1", "Poker")
		alice.PresenceKnown = false
		alice.IsOnline = false
		return alice
	}

	t.Run("online on Twitch", func(t *testing.T) {
		h := setup(t, options)
		alice := unseen(h)

		h.miner.updatePresence(alice)

		if !alice.PresenceKnown || !alice.IsOnline {
			t.Errorf("PresenceKnown, IsOnline = %t, %t, want true, true", alice.PresenceKnown, alice.IsOnline)
		}
		if got, want := h.logLines(t), []string{onlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("offline on Twitch", func(t *testing.T) {
		h := setup(t, options)
		alice := unseen(h)
		h.twitch.Channels["alice"].Online = false

		h.miner.updatePresence(alice)

		if !alice.PresenceKnown || alice.IsOnline {
			t.Errorf("PresenceKnown, IsOnline = %t, %t, want true, false", alice.PresenceKnown, alice.IsOnline)
		}
		if got, want := h.logLines(t), []string{offlineLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, options)
		alice := unseen(h)
		h.twitch.Fail["CheckStreamerOnline"] = errors.New("503")

		h.miner.updatePresence(alice)

		if alice.PresenceKnown {
			t.Errorf("PresenceKnown = true, want false")
		}
		if got, want := h.logLines(t), []string{"[INFO] online check Alice: 503"}; !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("online on Twitch, with this stream's streak resolved in the warm-start cache", func(t *testing.T) {
		h := setup(t, options)
		alice := unseen(h)
		resolved := onlineStreamer("alice", "1", "Poker")
		resolved.Stream.WatchStreakMissing = false
		h.cache.UpdateFromStreamer(resolved, testNow.Add(-time.Hour))

		h.miner.updatePresence(alice)

		if alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = true, want false")
		}
	})

	t.Run("already seen, with this stream's streak resolved in the warm-start cache", func(t *testing.T) {
		h := setup(t, options)
		alice := h.live("alice", "1", "Poker")
		resolved := onlineStreamer("alice", "1", "Poker")
		resolved.Stream.WatchStreakMissing = false
		h.cache.UpdateFromStreamer(resolved, testNow.Add(-time.Hour))

		h.miner.updatePresence(alice)

		if !alice.Stream.WatchStreakMissing {
			t.Errorf("WatchStreakMissing = false, want true")
		}
	})
}
