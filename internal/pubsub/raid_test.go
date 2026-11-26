package pubsub

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func raider() *domain.Streamer {
	return &domain.Streamer{
		Username:  "tester",
		ChannelID: "123",
		Settings:  domain.StreamerSettings{FollowRaid: true},
	}
}

func TestClient_processRaidMessage(t *testing.T) {
	t.Run("a raid", func(t *testing.T) {
		streamer := raider()
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		err := h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			RaidsJoined: []claimed{{"tester", "raid-1"}},
			Lines:       []string{":performing_arts: JOIN_RAID: Joining raid from Tester to target"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
		if got, want := streamer.LastRaidID, "raid-1"; got != want {
			t.Errorf("last raid = %q, want %q", got, want)
		}
	})

	t.Run("a raid with no target named", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1"}}`)

		want := effects{
			RaidsJoined: []claimed{{"tester", "raid-1"}},
			Lines:       []string{":performing_arts: JOIN_RAID: Joining raid from Tester to raid target"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a raid, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}, anonymise: true})

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		if got, want := h.logger.Lines, []string{":performing_arts: JOIN_RAID: Joining raid from Streamer1 to Streamer2"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("the same raid twice", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})
		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		want := effects{
			RaidsJoined: []claimed{{"tester", "raid-1"}},
			Lines:       []string{":performing_arts: JOIN_RAID: Joining raid from Tester to target"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a second raid", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})
		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-2","target_login":"target"}}`)

		if got, want := h.account.RaidsJoined, []claimed{{"tester", "raid-1"}, {"tester", "raid-2"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("raids joined = %+v, want %+v", got, want)
		}
	})

	t.Run("joining fails", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})
		h.account.RaidErr = errors.New("503")

		err := h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			RaidsJoined: []claimed{{"tester", "raid-1"}},
			Lines:       []string{"error: join raid Tester->target: 503"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("the same raid again after joining failed", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})
		h.account.RaidErr = errors.New("503")
		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)
		h.account.RaidErr = nil

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		want := effects{
			RaidsJoined: []claimed{{"tester", "raid-1"}, {"tester", "raid-1"}},
			Lines: []string{
				"error: join raid Tester->target: 503",
				":performing_arts: JOIN_RAID: Joining raid from Tester to target",
			},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a raid without an ID", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"target_login":"target"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a raid without an ID after another raid", func(t *testing.T) {
		streamer := raider()
		streamer.LastRaidID = "raid-1"
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"target_login":"target"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
		if got, want := streamer.LastRaidID, "raid-1"; got != want {
			t.Errorf("last raid = %q, want %q", got, want)
		}
	})

	t.Run("follow_raid off", func(t *testing.T) {
		streamer := raider()
		streamer.Settings.FollowRaid = false
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "raid.123", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}})

		h.receive(t, "raid.456", `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}
