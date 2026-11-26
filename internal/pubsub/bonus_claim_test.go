package pubsub

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_processClaimAvailable(t *testing.T) {
	t.Run("a bonus on a followed channel", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		err := h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.effects(), (effects{BonusClaims: []claimed{{"tester", "claim-1"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("the same bonus twice", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`)

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`)

		if got, want := h.account.BonusClaims, []claimed{{"tester", "claim-1"}, {"tester", "claim-1"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("bonus claims = %+v, want %+v", got, want)
		}
	})

	t.Run("claiming fails", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.account.BonusErr = errors.New("503")

		err := h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			BonusClaims: []claimed{{"tester", "claim-1"}},
			Lines:       []string{"error: claim bonus Tester (channel 123): 503"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("claiming fails, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, anonymise: true})
		h.account.BonusErr = errors.New("503")

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"123"}}}`)

		if got, want := h.logger.Lines, []string{"error: claim bonus Streamer1: 503"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a bonus on a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1","channel_id":"456"}}}`)

		want := effects{}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus without an ID", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"channel_id":"123"}}}`)

		want := effects{Lines: []string{"error: claim-available ignored: channel=123 claim= streamer=true"}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus without an ID, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, anonymise: true})

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"channel_id":"123"}}}`)

		want := effects{Lines: []string{"error: claim-available ignored: [redacted]"}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus without an ID, deep debug on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, deepDebug: true})
		h.deep.DeepOn = true

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"channel_id":"123"}}}`)

		want := []string{
			`PubSub topic community-points-user-v1.999 payload {"type":"claim-available","data":{"claim":{"channel_id":"123"}}}`,
			"claim-available payload: map[data:map[claim:map[channel_id:123]] type:claim-available]",
		}
		if got := h.deep.Deep; !reflect.DeepEqual(got, want) {
			t.Errorf("deep debug lines = %q, want %q", got, want)
		}
	})

	t.Run("a bonus naming no channel, on no topic, one streamer followed", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "", `{"type":"claim-available","data":{"claim":{"id":"claim-1"}}}`)

		if got, want := h.effects(), (effects{BonusClaims: []claimed{{"tester", "claim-1"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus naming no channel, on no topic, two streamers followed", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{
			{Username: "tester", ChannelID: "123"},
			{Username: "other", ChannelID: "456"},
		}})

		h.receive(t, "", `{"type":"claim-available","data":{"claim":{"id":"claim-1"}}}`)

		want := effects{}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a bonus naming no channel, one streamer followed", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		h.receive(t, "community-points-user-v1.999", `{"type":"claim-available","data":{"claim":{"id":"claim-1"}}}`)

		if got, want := h.effects(), (effects{BonusClaims: []claimed{{"tester", "claim-1"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a message with no data", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})

		err := h.receive(t, "community-points-user-v1.999", `{"type":"claim-available"}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}

func TestClient_checkPendingClaims(t *testing.T) {
	t.Run("two streamers", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{
			{Username: "tester", ChannelID: "123"},
			{Username: "other", ChannelID: "456"},
		}})

		h.client.checkPendingClaims()

		if got, want := h.effects(), (effects{ContextLoads: []string{"tester", "other"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer without a channel ID", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{
			{Username: "tester"},
			{Username: "other", ChannelID: "456"},
		}})

		h.client.checkPendingClaims()

		if got, want := h.effects(), (effects{ContextLoads: []string{"other"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer without a name", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{
			{ChannelID: "123"},
			{Username: "other", ChannelID: "456"},
		}})

		h.client.checkPendingClaims()

		if got, want := h.effects(), (effects{ContextLoads: []string{"other"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("no streamers", func(t *testing.T) {
		h := setup(world{})

		h.client.checkPendingClaims()

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("checking fails", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{
			{Username: "tester", ChannelID: "123"},
			{Username: "other", ChannelID: "456"},
		}})
		h.account.ContextErr = errors.New("503")

		h.client.checkPendingClaims()

		want := effects{
			ContextLoads: []string{"tester", "other"},
			Lines:        []string{"error: pending bonus check Tester: 503", "error: pending bonus check Other: 503"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("debug on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.logger.DebugOn = true

		h.client.checkPendingClaims()

		if got, want := h.logger.Debug, []string{"Checking pending bonus for Tester (123)"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("debug on, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}, anonymise: true})
		h.logger.DebugOn = true

		h.client.checkPendingClaims()

		if got, want := h.logger.Debug, []string{"Checking pending bonus for Streamer1"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})
}
