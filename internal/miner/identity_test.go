package miner

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

const channelMissing = "channel missing for oldlogin: channel not found"

func TestStreamerKey(t *testing.T) {
	tests := []struct {
		name     string
		streamer *domain.Streamer
		want     string
	}{
		{"no streamer", nil, ""},
		{"a channel id and a login", &domain.Streamer{ChannelID: "123456", Username: "Alice"}, "123456"},
		{"a channel id with spaces around", &domain.Streamer{ChannelID: " 123456 ", Username: "Alice"}, "123456"},
		{"a blank channel id", &domain.Streamer{ChannelID: "  ", Username: " Alice "}, "alice"},
		{"no channel id and no login", &domain.Streamer{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streamerKey(tt.streamer); got != tt.want {
				t.Errorf("streamerKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldRepairStreamerIdentity(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"no error", nil, false},
		{"channel not found", twitch.ErrChannelNotFound, true},
		{"channel not found, wrapped", fmt.Errorf("context: %w", twitch.ErrChannelNotFound), true},
		{"streamer offline", twitch.ErrStreamerOffline, false},
		{"a network error", errors.New("503"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRepairStreamerIdentity(tt.err); got != tt.want {
				t.Errorf("shouldRepairStreamerIdentity(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

// renamedStreamer is held under its old login while Twitch knows the channel as "newlogin".
func renamedStreamer(h *harness) *domain.Streamer {
	streamer := onlineStreamer("oldlogin", "123", "Poker")
	streamer.StreamerURL = "https://www.twitch.tv/oldlogin"
	h.twitch.Channels["newlogin"] = &fakeChannel{
		ID:          "123",
		Points:      250,
		Online:      true,
		BroadcastID: streamer.Stream.BroadcastID,
		Game:        "Poker",
		CreatedAt:   streamer.Stream.CreatedAt,
	}
	return streamer
}

func TestMiner_applyStreamerLoginChange(t *testing.T) {
	options := Options{Username: "account"}

	t.Run("a new login", func(t *testing.T) {
		h := setup(t, options)
		streamer := renamedStreamer(h)
		streamer.Stream.SpadeURL = "https://spade.example/old"

		got := h.miner.applyStreamerLoginChange(streamer, "newlogin")

		if !got {
			t.Errorf("applyStreamerLoginChange() = false, want true")
		}
		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
		if want := "https://www.twitch.tv/newlogin"; streamer.StreamerURL != want {
			t.Errorf("StreamerURL = %q, want %q", streamer.StreamerURL, want)
		}
		if streamer.Stream.SpadeURL != "" {
			t.Errorf("SpadeURL = %q, want none", streamer.Stream.SpadeURL)
		}
		if want := []string{"oldlogin"}; !reflect.DeepEqual(h.chat.Left, want) {
			t.Errorf("chats left = %q, want %q", h.chat.Left, want)
		}
		if want := []chatUpdate{{Login: "newlogin", Online: true}}; !reflect.DeepEqual(h.chat.Updates, want) {
			t.Errorf("chat updates = %+v, want %+v", h.chat.Updates, want)
		}
		wantLog := []string{"[INFO] identity refresh oldlogin -> newlogin for channel 123"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
		entry, _ := h.cache.Get("newlogin")
		wantEntry := streak.WarmStartEntry{
			AccountName:        "account",
			StreamerLogin:      "newlogin",
			ChannelID:          "123",
			BroadcastID:        "oldlogin-broadcast",
			StreamCreatedAt:    streamer.Stream.CreatedAt,
			WatchStreakMissing: true,
			CheckedAt:          testNow,
			IsOnline:           true,
		}
		if entry != wantEntry {
			t.Errorf("cached entry = %+v, want %+v", entry, wantEntry)
		}
	})

	t.Run("a new login in capitals with spaces around", func(t *testing.T) {
		h := setup(t, options)
		streamer := renamedStreamer(h)

		h.miner.applyStreamerLoginChange(streamer, "  NewLogin ")

		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
	})

	t.Run("a streamer with no stream", func(t *testing.T) {
		h := setup(t, options)
		streamer := &domain.Streamer{Username: "oldlogin", ChannelID: "123"}

		got := h.miner.applyStreamerLoginChange(streamer, "newlogin")

		if !got {
			t.Errorf("applyStreamerLoginChange() = false, want true")
		}
		if want := []chatUpdate{{Login: "newlogin", Online: false}}; !reflect.DeepEqual(h.chat.Updates, want) {
			t.Errorf("chat updates = %+v, want %+v", h.chat.Updates, want)
		}
	})

	t.Run("anonymized logs", func(t *testing.T) {
		anonymized := options
		anonymized.Logger = logging.LoggerSettings{AnonymizeLogs: true}
		h := setup(t, anonymized)
		streamer := renamedStreamer(h)

		h.miner.applyStreamerLoginChange(streamer, "newlogin")

		wantLog := []string{"[INFO] identity refresh Streamer1"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	unchanged := []struct {
		name     string
		streamer *domain.Streamer
		login    string
	}{
		{"the same login in another case", &domain.Streamer{Username: "oldlogin", ChannelID: "123"}, "OldLogin"},
		{"a blank login", &domain.Streamer{Username: "oldlogin", ChannelID: "123"}, "   "},
		{"no streamer", nil, "newlogin"},
	}
	for _, tt := range unchanged {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, options)

			got := h.miner.applyStreamerLoginChange(tt.streamer, tt.login)

			if got {
				t.Errorf("applyStreamerLoginChange() = true, want false")
			}
			if h.chat.Left != nil {
				t.Errorf("chats left = %q, want none", h.chat.Left)
			}
			if got := h.logLines(t); got != nil {
				t.Errorf("log = %q, want nothing", got)
			}
		})
	}
}

func TestMiner_repairStreamerIdentity(t *testing.T) {
	notFound := fmt.Errorf("channel missing for oldlogin: %w", twitch.ErrChannelNotFound)

	t.Run("a channel renamed on Twitch", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)

		got := h.miner.repairStreamerIdentity(streamer, notFound)

		if !got {
			t.Errorf("repairStreamerIdentity() = false, want true")
		}
		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
	})

	t.Run("an error other than channel not found", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)

		got := h.miner.repairStreamerIdentity(streamer, errors.New("503"))

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
		if streamer.Username != "oldlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "oldlogin")
		}
	})

	t.Run("the lookup failing", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		h.twitch.Fail["GetUserByID"] = errors.New("503")

		got := h.miner.repairStreamerIdentity(streamer, notFound)

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
		wantLog := []string{"[INFO] identity refresh Oldlogin: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		got := h.miner.repairStreamerIdentity(streamer, notFound)

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
	})

	t.Run("a channel with no login", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")
		h.twitch.Channels[""] = &fakeChannel{ID: "123"}

		got := h.miner.repairStreamerIdentity(streamer, notFound)

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
	})

	t.Run("a channel still under the same login", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")
		h.twitch.Channels["oldlogin"] = &fakeChannel{ID: "123"}

		got := h.miner.repairStreamerIdentity(streamer, notFound)

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
	})

	t.Run("before login", func(t *testing.T) {
		m := New(Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		got := m.repairStreamerIdentity(streamer, notFound)

		if got {
			t.Errorf("repairStreamerIdentity() = true, want false")
		}
	})
}

func TestWithIdentityRepair(t *testing.T) {
	notFound := fmt.Errorf("channel missing: %w", twitch.ErrChannelNotFound)

	t.Run("an answer the first time", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		calls := 0

		got, err := withIdentityRepair(h.miner, streamer, func() (string, error) {
			calls++
			return "answer", nil
		})

		if got != "answer" || err != nil || calls != 1 {
			t.Errorf("withIdentityRepair() = %q, %v after %d calls, want %q, nil after 1", got, err, calls, "answer")
		}
		if streamer.Username != "oldlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "oldlogin")
		}
	})

	t.Run("a failure that is not about the channel", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		failure := errors.New("503")
		calls := 0

		got, err := withIdentityRepair(h.miner, streamer, func() (string, error) {
			calls++
			return "partial", failure
		})

		if got != "partial" || err != failure || calls != 1 {
			t.Errorf("withIdentityRepair() = %q, %v after %d calls, want %q, %v after 1", got, err, calls, "partial", failure)
		}
	})

	t.Run("a channel renamed on Twitch is asked about again", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		calls := 0

		got, err := withIdentityRepair(h.miner, streamer, func() (string, error) {
			calls++
			if calls == 1 {
				return "", notFound
			}
			return streamer.Username, nil
		})

		if got != "newlogin" || err != nil || calls != 2 {
			t.Errorf("withIdentityRepair() = %q, %v after %d calls, want %q, nil after 2", got, err, calls, "newlogin")
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")
		calls := 0

		got, err := withIdentityRepair(h.miner, streamer, func() (string, error) {
			calls++
			return "first", notFound
		})

		if got != "first" || err != notFound || calls != 1 {
			t.Errorf("withIdentityRepair() = %q, %v after %d calls, want %q, %v after 1", got, err, calls, "first", notFound)
		}
	})

	t.Run("a second failure after the repair", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		failure := errors.New("503")
		calls := 0

		got, err := withIdentityRepair(h.miner, streamer, func() (string, error) {
			calls++
			if calls == 1 {
				return "first", notFound
			}
			return "second", failure
		})

		if got != "second" || err != failure || calls != 2 {
			t.Errorf("withIdentityRepair() = %q, %v after %d calls, want %q, %v after 2", got, err, calls, "second", failure)
		}
	})
}

func TestMiner_loadChannelPointsContextWithIdentityRepair(t *testing.T) {
	t.Run("Twitch answering", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Points = 250

		got, err := h.miner.loadChannelPointsContextWithIdentityRepair(alice)

		if got != 250 || err != nil {
			t.Errorf("loadChannelPointsContextWithIdentityRepair() = %d, %v, want 250, nil", got, err)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		failure := errors.New("503")
		h.twitch.Fail["LoadChannelPointsContext"] = failure

		got, err := h.miner.loadChannelPointsContextWithIdentityRepair(alice)

		if got != 0 || err != failure {
			t.Errorf("loadChannelPointsContextWithIdentityRepair() = %d, %v, want 0, %v", got, err, failure)
		}
	})

	t.Run("a channel renamed on Twitch", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)

		got, err := h.miner.loadChannelPointsContextWithIdentityRepair(streamer)

		if got != 250 || err != nil {
			t.Errorf("loadChannelPointsContextWithIdentityRepair() = %d, %v, want 250, nil", got, err)
		}
		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		got, err := h.miner.loadChannelPointsContextWithIdentityRepair(streamer)

		if got != 0 || err == nil || err.Error() != channelMissing {
			t.Errorf("loadChannelPointsContextWithIdentityRepair() = %d, %v, want 0, %s", got, err, channelMissing)
		}
	})
}

func TestMiner_checkStreamerOnlineWithIdentityRepair(t *testing.T) {
	t.Run("Twitch answering", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")

		got, err := h.miner.checkStreamerOnlineWithIdentityRepair(alice)

		if !got || err != nil {
			t.Errorf("checkStreamerOnlineWithIdentityRepair() = %t, %v, want true, nil", got, err)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		failure := errors.New("503")
		h.twitch.Fail["CheckStreamerOnline"] = failure

		got, err := h.miner.checkStreamerOnlineWithIdentityRepair(alice)

		if !got || err != failure {
			t.Errorf("checkStreamerOnlineWithIdentityRepair() = %t, %v, want true, %v", got, err, failure)
		}
	})

	t.Run("a channel renamed on Twitch", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)
		streamer.IsOnline = false

		got, err := h.miner.checkStreamerOnlineWithIdentityRepair(streamer)

		if !got || err != nil {
			t.Errorf("checkStreamerOnlineWithIdentityRepair() = %t, %v, want true, nil", got, err)
		}
		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		got, err := h.miner.checkStreamerOnlineWithIdentityRepair(streamer)

		if !got || err == nil || err.Error() != channelMissing {
			t.Errorf("checkStreamerOnlineWithIdentityRepair() = %t, %v, want true, %s", got, err, channelMissing)
		}
	})
}

func TestMiner_updateStreamWithIdentityRepair(t *testing.T) {
	t.Run("Twitch answering", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		h.twitch.Channels["alice"].Game = "Chess"

		err := h.miner.updateStreamWithIdentityRepair(alice)

		if err != nil {
			t.Errorf("updateStreamWithIdentityRepair() = %v, want nil", err)
		}
		if got := alice.Stream.GameName(); got != "Chess" {
			t.Errorf("game = %q, want %q", got, "Chess")
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		failure := errors.New("503")
		h.twitch.Fail["UpdateStream"] = failure

		err := h.miner.updateStreamWithIdentityRepair(alice)

		if err != failure {
			t.Errorf("updateStreamWithIdentityRepair() = %v, want %v", err, failure)
		}
	})

	t.Run("a channel renamed on Twitch", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)

		err := h.miner.updateStreamWithIdentityRepair(streamer)

		if err != nil {
			t.Errorf("updateStreamWithIdentityRepair() = %v, want nil", err)
		}
		if streamer.Username != "newlogin" {
			t.Errorf("Username = %q, want %q", streamer.Username, "newlogin")
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		err := h.miner.updateStreamWithIdentityRepair(streamer)

		if err == nil || err.Error() != channelMissing {
			t.Errorf("updateStreamWithIdentityRepair() = %v, want %s", err, channelMissing)
		}
	})
}

func TestMiner_sendMinuteWatchedWithIdentityRepair(t *testing.T) {
	t.Run("Twitch answering", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")

		err := h.miner.sendMinuteWatchedWithIdentityRepair(alice)

		if err != nil {
			t.Errorf("sendMinuteWatchedWithIdentityRepair() = %v, want nil", err)
		}
		if want := []string{"alice"}; !reflect.DeepEqual(h.twitch.MinutesWatched, want) {
			t.Errorf("minutes watched = %q, want %q", h.twitch.MinutesWatched, want)
		}
	})

	t.Run("Twitch failing", func(t *testing.T) {
		h := setup(t, Options{})
		alice := h.live("alice", "1", "Poker")
		failure := errors.New("503")
		h.twitch.Fail["SendMinuteWatched"] = failure

		err := h.miner.sendMinuteWatchedWithIdentityRepair(alice)

		if err != failure {
			t.Errorf("sendMinuteWatchedWithIdentityRepair() = %v, want %v", err, failure)
		}
		if h.twitch.MinutesWatched != nil {
			t.Errorf("minutes watched = %q, want none", h.twitch.MinutesWatched)
		}
	})

	t.Run("a channel renamed on Twitch", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := renamedStreamer(h)

		err := h.miner.sendMinuteWatchedWithIdentityRepair(streamer)

		if err != nil {
			t.Errorf("sendMinuteWatchedWithIdentityRepair() = %v, want nil", err)
		}
		if want := []string{"newlogin"}; !reflect.DeepEqual(h.twitch.MinutesWatched, want) {
			t.Errorf("minutes watched = %q, want %q", h.twitch.MinutesWatched, want)
		}
	})

	t.Run("a channel Twitch no longer knows", func(t *testing.T) {
		h := setup(t, Options{})
		streamer := onlineStreamer("oldlogin", "123", "Poker")

		err := h.miner.sendMinuteWatchedWithIdentityRepair(streamer)

		if err == nil || err.Error() != channelMissing {
			t.Errorf("sendMinuteWatchedWithIdentityRepair() = %v, want %s", err, channelMissing)
		}
	})
}
