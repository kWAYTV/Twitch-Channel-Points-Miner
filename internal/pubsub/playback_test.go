package pubsub

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestJustWentOffline(t *testing.T) {
	for _, row := range []struct {
		name      string
		offlineAt time.Time
		want      bool
	}{
		{"never offline", time.Time{}, false},
		{"offline this instant", noon, true},
		{"offline just under a minute ago", noon.Add(-time.Minute + time.Nanosecond), true},
		{"offline exactly a minute ago", noon.Add(-time.Minute), false},
		{"offline just over a minute ago", noon.Add(-time.Minute - time.Nanosecond), false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := justWentOffline(row.offlineAt, noon); got != row.want {
				t.Errorf("justWentOffline(%v, noon) = %t, want %t", row.offlineAt, got, row.want)
			}
		})
	}
}

func TestClient_processPlaybackMessage(t *testing.T) {
	t.Run("stream-up for a streamer not seen live yet", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123"}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		err := h.receive(t, "video-playback-by-id.123", `{"type":"stream-up"}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := domain.NewStream()
		want.StreamUpAt = noon
		if !reflect.DeepEqual(streamer.Stream, want) {
			t.Errorf("stream = %+v, want %+v", streamer.Stream, want)
		}
		if got, want := h.effects(), (effects{Presence: []presence{{"tester", true, "stream-up"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("stream-up for a streamer watched before", func(t *testing.T) {
		streamer := &domain.Streamer{
			Username:  "tester",
			ChannelID: "123",
			Stream: &domain.Stream{
				BroadcastID:   "broadcast-1",
				MinuteWatched: 12.5,
				WatchCount:    3,
				CreatedAt:     noon.Add(-time.Hour),
				StreamUpAt:    noon.Add(-time.Hour),
			},
		}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-up"}`)

		want := &domain.Stream{BroadcastID: "broadcast-1", WatchStreakMissing: true, StreamUpAt: noon}
		if !reflect.DeepEqual(streamer.Stream, want) {
			t.Errorf("stream = %+v, want %+v", streamer.Stream, want)
		}
		if got, want := h.presence, []presence{{"tester", true, "stream-up"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("presence = %+v, want %+v", got, want)
		}
	})

	t.Run("stream-down", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-down"}`)

		if got, want := h.effects(), (effects{Presence: []presence{{"tester", false, "stream-down"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("viewcount while live", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.account.Live = true

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got, want := h.effects(), (effects{Presence: []presence{{"tester", true, "viewcount"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("viewcount while not live", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.account.Live = false

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("viewcount, the live check fails", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.account.Live = true
		h.account.LiveErr = errors.New("503")

		err := h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.effects(), (effects{Lines: []string{"error: live check Tester: 503"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("viewcount long after stream-up", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", Stream: &domain.Stream{StreamUpAt: noon}}
		h := setup(world{streamers: []*domain.Streamer{streamer}})
		h.account.Live = true

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got, want := h.presence, []presence{{"tester", true, "viewcount"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("presence = %+v, want %+v", got, want)
		}
	})

	t.Run("viewcount right after stream-up", func(t *testing.T) {
		// domain.Stream measures the time since stream-up on the real clock.
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", Stream: &domain.Stream{StreamUpAt: time.Now()}}
		h := setup(world{streamers: []*domain.Streamer{streamer}})
		h.account.Live = true

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("viewcount half a minute after going offline", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", OfflineAt: noon.Add(-30 * time.Second)}
		h := setup(world{streamers: []*domain.Streamer{streamer}})
		h.account.Live = true

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("viewcount two minutes after going offline", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", OfflineAt: noon.Add(-2 * time.Minute)}
		h := setup(world{streamers: []*domain.Streamer{streamer}})
		h.account.Live = true

		h.receive(t, "video-playback-by-id.123", `{"type":"viewcount","viewers":42}`)

		if got, want := h.presence, []presence{{"tester", true, "viewcount"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("presence = %+v, want %+v", got, want)
		}
	})

	t.Run("a type it doesn't know", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "video-playback-by-id.123", `{"type":"commercial","length":30}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "video-playback-by-id.456", `{"type":"stream-down"}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("no channel in the topic", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "video-playback-by-id.", `{"type":"stream-down"}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("stream-up with no presence handler", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123"}
		h := setup(world{streamers: []*domain.Streamer{streamer}, noHandlers: true})

		err := h.receive(t, "video-playback-by-id.123", `{"type":"stream-up"}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if streamer.Stream != nil {
			t.Errorf("stream = %+v, want none", streamer.Stream)
		}
	})
}
