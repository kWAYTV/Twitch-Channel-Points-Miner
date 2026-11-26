package pubsub

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/humanize"
)

// midway is a Timing whose every draw is the middle of its range: a reaction
// takes 5 seconds, and a bet stays at its set moment.
func midway() *humanize.Timing {
	return humanize.NewFrom(func() float64 { return 0.5 })
}

const (
	bonusAvailable = `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`
	raidStarting   = `{"type":"raid_update_v2","raid":{"id":"raid-1","target_login":"target"}}`
	momentActive   = `{"type":"active","data":{"moment_id":"moment-1"}}`
)

func TestClient_react(t *testing.T) {
	t.Run("a bonus, with human timings", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, human: midway()})

		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		if got, want := h.effects(), (effects{Timers: []time.Duration{5 * time.Second}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want only the pause: %+v", got, want)
		}
	})

	t.Run("a bonus, once the pause is over", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, human: midway()})
		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		h.timers[0].fire()

		if got, want := h.account.BonusClaims, []claimed{{"tester", "claim-1"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("bonus claims = %+v, want %+v", got, want)
		}
	})

	t.Run("the same bonus announced again during the pause", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, human: midway()})
		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		if got := len(h.timers); got != 1 {
			t.Errorf("%d answers pending, want 1", got)
		}
	})

	t.Run("the same bonus announced again after it was claimed", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, human: midway()})
		h.receive(t, "community-points-user-v1.999", bonusAvailable)
		h.timers[0].fire()

		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		if got := len(h.timers); got != 2 {
			t.Errorf("%d answers in all, want 2: nothing is pending any more", got)
		}
	})

	t.Run("a raid, with human timings", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{raider()}, human: midway()})
		h.receive(t, "raid.123", raidStarting)
		h.receive(t, "raid.123", raidStarting)

		h.timers[0].fire()

		want := effects{
			Timers:      []time.Duration{5 * time.Second},
			RaidsJoined: []claimed{{"tester", "raid-1"}},
			Lines:       []string{":performing_arts: JOIN_RAID: Joining raid from Tester to target"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a moment, with human timings", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}, human: midway()})
		h.receive(t, "community-moments-channel-v1.123", momentActive)

		h.timers[0].fire()

		if got, want := h.account.MomentClaims, []claimed{{"tester", "moment-1"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("moment claims = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus, with exact timings", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "community-points-user-v1.999", bonusAvailable)

		if got, want := h.effects(), (effects{BonusClaims: []claimed{{"tester", "claim-1"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want it claimed at once: %+v", got, want)
		}
	})
}

func TestClient_scheduleBet_humanTimings(t *testing.T) {
	// The prediction runs for 120 seconds and bettor() bets at its end.
	t.Run("the earliest draw", func(t *testing.T) {
		early := humanize.NewFrom(func() float64 { return 0 })
		h := setup(world{streamers: []*domain.Streamer{bettor()}, human: early})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got, want := h.timers[0].After, 108*time.Second; got != want {
			t.Errorf("bet after %s, want %s: a tenth of the window early", got, want)
		}
	})

	t.Run("the latest draw", func(t *testing.T) {
		late := humanize.NewFrom(func() float64 { return 1 })
		h := setup(world{streamers: []*domain.Streamer{bettor()}, human: late})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got, want := h.timers[0].After, 120*time.Second; got != want {
			t.Errorf("bet after %s, want %s: never later than it was set", got, want)
		}
	})
}
