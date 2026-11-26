package miner

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
)

func TestFormatLoadDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"no time", 0, "0.0 seconds"},
		{"a few seconds", 5200 * time.Millisecond, "5.2 seconds"},
		{"a second under a minute", 59 * time.Second, "59.0 seconds"},
		{"exactly a minute", time.Minute, "1.0 minutes"},
		{"a minute and a half", 90 * time.Second, "1.5 minutes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatLoadDuration(tt.duration); got != tt.want {
				t.Errorf("formatLoadDuration(%s) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func onlineNamed(logins ...string) []*domain.Streamer {
	streamers := make([]*domain.Streamer, 0, len(logins))
	for _, login := range logins {
		streamers = append(streamers, &domain.Streamer{Username: login, IsOnline: true})
	}
	return streamers
}

func TestMiner_loadSummary(t *testing.T) {
	tests := []struct {
		name      string
		options   Options
		streamers []*domain.Streamer
		took      time.Duration
		want      string
	}{
		{"no streamers", Options{}, nil, time.Second, "0 streamers loaded in 1.0 seconds: 0 online, 0 offline"},
		{"one offline streamer", Options{}, []*domain.Streamer{{Username: "bob"}}, time.Second, "1 streamer loaded in 1.0 seconds: 0 online, 1 offline"},
		{"one online streamer", Options{}, onlineNamed("alice"), time.Second, "1 streamer loaded in 1.0 seconds: 1 online (Alice), 0 offline"},
		{
			"two online around one offline",
			Options{},
			[]*domain.Streamer{{Username: "alice", IsOnline: true}, {Username: "bob"}, {Username: "carol", IsOnline: true}},
			5200 * time.Millisecond,
			"3 streamers loaded in 5.2 seconds: 2 online (Alice, Carol), 1 offline",
		},
		{
			"ten online",
			Options{},
			onlineNamed("a", "b", "c", "d", "e", "f", "g", "h", "i", "j"),
			time.Second,
			"10 streamers loaded in 1.0 seconds: 10 online (A, B, C, D, E, F, G, H, I, J), 0 offline",
		},
		{
			"eleven online",
			Options{},
			onlineNamed("a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"),
			time.Second,
			"11 streamers loaded in 1.0 seconds: 11 online (A, B, C, D, E, F, G, H, I, J and 1 more), 0 offline",
		},
		{"anonymizing on", anonymizedOptions, onlineNamed("alice"), time.Second, "1 streamer loaded in 1.0 seconds: 1 online (Streamer1), 0 offline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)

			if got := h.miner.loadSummary(tt.streamers, tt.took); got != tt.want {
				t.Errorf("loadSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMiner_logLoadSummary(t *testing.T) {
	t.Run("no streamers", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.logLoadSummary(nil, time.Second)

		if got := h.logLines(t); got != nil {
			t.Errorf("log = %q, want nothing", got)
		}
	})

	t.Run("one online streamer", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.logLoadSummary(onlineNamed("alice"), time.Second)

		want := []string{"[INFO] 1 streamer loaded in 1.0 seconds: 1 online (Alice), 0 offline"}
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})
}

func TestMiner_loadStreamer(t *testing.T) {
	options := Options{Username: "account", WatchStreakWarmStartCache: true}
	startedAt := testNow.Add(-5 * time.Minute)
	defaults := domain.StreamerSettings{}
	defaults.Default()

	t.Run("a blank name", func(t *testing.T) {
		h := setup(t, options)

		if got := h.miner.loadStreamer(""); got != nil {
			t.Errorf("loadStreamer() = %+v, want nil", got)
		}
	})

	t.Run("an online channel", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Points: 250, Online: true, BroadcastID: "alice-broadcast", Game: "Poker", CreatedAt: startedAt}

		got := h.miner.loadStreamer("alice")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		type loaded struct {
			Username, ChannelID, StreamerURL, Game string
			ChannelPoints                          int
			PointsInit, IsOnline, PresenceKnown    bool
			Settings                               domain.StreamerSettings
		}
		gotLoaded := loaded{got.Username, got.ChannelID, got.StreamerURL, got.Stream.GameName(), got.ChannelPoints, got.PointsInit, got.IsOnline, got.PresenceKnown, got.Settings}
		wantLoaded := loaded{"alice", "1", "https://www.twitch.tv/alice", "Poker", 250, true, true, true, defaults}
		if !reflect.DeepEqual(gotLoaded, wantLoaded) {
			t.Errorf("loaded streamer = %+v, want %+v", gotLoaded, wantLoaded)
		}
		wantLog := []string{"[INFO] Alice (" + colorCyan + "250" + colorReset + " points) is " + colorGreen + "Online" + colorReset + "!"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("an offline channel", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Points: 250}

		got := h.miner.loadStreamer("alice")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		if got.IsOnline || !got.PresenceKnown {
			t.Errorf("IsOnline, PresenceKnown = %t, %t, want false, true", got.IsOnline, got.PresenceKnown)
		}
		wantLog := []string{"[INFO] Alice (" + colorCyan + "250" + colorReset + " points) is " + colorRed + "Offline" + colorReset + "!"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a channel with settings of its own, named in capitals", func(t *testing.T) {
		own := options
		own.StreamerOverrides = map[string]domain.StreamerSettings{"alice": {ClaimDrops: true, IRCMode: domain.IRCModeNever}}
		h := setup(t, own)
		h.twitch.Channels["ALICE"] = &fakeChannel{ID: "1"}

		got := h.miner.loadStreamer("ALICE")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		if want := (domain.StreamerSettings{ClaimDrops: true, IRCMode: domain.IRCModeNever}); !reflect.DeepEqual(got.Settings, want) {
			t.Errorf("Settings = %+v, want %+v", got.Settings, want)
		}
	})

	t.Run("a channel Twitch does not know", func(t *testing.T) {
		h := setup(t, options)

		got := h.miner.loadStreamer("ghost")

		if got != nil {
			t.Errorf("loadStreamer() = %+v, want nil", got)
		}
		wantLog := []string{"[INFO] skip ghost: channel missing for ghost: channel not found"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a channel Twitch does not know, with anonymizing on", func(t *testing.T) {
		anonymized := options
		anonymized.Logger = logging.LoggerSettings{AnonymizeLogs: true}
		h := setup(t, anonymized)

		h.miner.loadStreamer("ghost")

		wantLog := []string{"[INFO] skip Streamer1: channel missing for ghost: channel not found"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a renamed channel the warm-start cache remembers", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.Channels["newlogin"] = &fakeChannel{ID: "123", Points: 250}
		h.cache.UpdateFromStreamer(&domain.Streamer{Username: "oldlogin", ChannelID: "123"}, testNow.Add(-time.Hour))

		got := h.miner.loadStreamer("oldlogin")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		if got.Username != "newlogin" || got.ChannelID != "123" || got.ChannelPoints != 250 {
			t.Errorf("Username, ChannelID, ChannelPoints = %q, %q, %d, want %q, %q, %d", got.Username, got.ChannelID, got.ChannelPoints, "newlogin", "123", 250)
		}
	})

	t.Run("a channel the warm-start cache remembers without an id", func(t *testing.T) {
		h := setup(t, options)
		h.cache.UpdateFromStreamer(&domain.Streamer{Username: "ghost"}, testNow.Add(-time.Hour))

		if got := h.miner.loadStreamer("ghost"); got != nil {
			t.Errorf("loadStreamer() = %+v, want nil", got)
		}
	})

	t.Run("Twitch failing on the balance", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Points: 250}
		h.twitch.Fail["LoadChannelPointsContext"] = errors.New("503")

		got := h.miner.loadStreamer("alice")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		if got.PointsInit {
			t.Errorf("PointsInit = true, want false")
		}
		wantLog := []string{
			"[INFO] context for alice: 503",
			"[INFO] Alice (" + colorCyan + "0" + colorReset + " points) is " + colorRed + "Offline" + colorReset + "!",
		}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("Twitch failing on the online check", func(t *testing.T) {
		h := setup(t, options)
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Points: 250, Online: true}
		h.twitch.Fail["CheckStreamerOnline"] = errors.New("503")

		got := h.miner.loadStreamer("alice")

		if got == nil {
			t.Fatalf("loadStreamer() = nil, want a streamer")
		}
		if got.PresenceKnown {
			t.Errorf("PresenceKnown = true, want false")
		}
		wantLog := []string{"[INFO] online check Alice: 503"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})
}

func TestMiner_loadStreamers(t *testing.T) {
	t.Run("no targets", func(t *testing.T) {
		h := setup(t, Options{})

		got := h.miner.loadStreamers(nil)

		if want := []string{}; !reflect.DeepEqual(logins(got), want) {
			t.Errorf("loaded %q, want %q", logins(got), want)
		}
	})

	t.Run("six targets, one of them unknown", func(t *testing.T) {
		h := setup(t, Options{})
		for i, login := range []string{"alice", "bob", "carol", "dave", "erin"} {
			h.twitch.Channels[login] = &fakeChannel{ID: string(rune('1' + i)), Online: i%2 == 0}
		}

		got := h.miner.loadStreamers([]string{"alice", "bob", "ghost", "carol", "dave", "erin"})

		if want := []string{"alice", "bob", "carol", "dave", "erin"}; !reflect.DeepEqual(logins(got), want) {
			t.Errorf("loaded %q, want %q", logins(got), want)
		}
		wantLog := []string{"[INFO] skip ghost: channel missing for ghost: channel not found"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})

	t.Run("a streamer going offline after the list has loaded", func(t *testing.T) {
		h := setup(t, Options{})
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Online: true}
		alice := h.miner.loadStreamers([]string{"alice"})[0]

		h.miner.setPresence(alice, false, "poll")

		wantLog := []string{"[INFO] Alice (" + colorCyan + "0" + colorReset + " points) is " + colorRed + "Offline" + colorReset + "!"}
		if got := h.logLines(t); !reflect.DeepEqual(got, wantLog) {
			t.Errorf("log = %q, want %q", got, wantLog)
		}
	})
}
