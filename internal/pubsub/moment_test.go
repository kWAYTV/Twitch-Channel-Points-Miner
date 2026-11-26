package pubsub

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func momentClaimer() *domain.Streamer {
	return &domain.Streamer{
		Username:  "tester",
		ChannelID: "123",
		Settings:  domain.StreamerSettings{ClaimMoments: true},
	}
}

func TestClient_processMomentMessage(t *testing.T) {
	t.Run("an active moment", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}})

		err := h.receive(t, "community-moments-channel-v1.123", `{"type":"active","data":{"moment_id":"moment-1"}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			MomentClaims: []claimed{{"tester", "moment-1"}},
			Lines:        []string{":video_camera: MOMENT_CLAIM: tester Claimed Moment"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("an active moment, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}, anonymise: true})

		h.receive(t, "community-moments-channel-v1.123", `{"type":"active","data":{"moment_id":"moment-1"}}`)

		if got, want := h.logger.Lines, []string{":video_camera: MOMENT_CLAIM: Streamer1 Claimed Moment"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("claiming fails", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}})
		h.account.MomentErr = errors.New("503")

		err := h.receive(t, "community-moments-channel-v1.123", `{"type":"active","data":{"moment_id":"moment-1"}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			MomentClaims: []claimed{{"tester", "moment-1"}},
			Lines:        []string{"error: claim moment tester: 503"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a moment that is over", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}})

		h.receive(t, "community-moments-channel-v1.123", `{"type":"expired","data":{"moment_id":"moment-1"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an active moment without an ID", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}})

		h.receive(t, "community-moments-channel-v1.123", `{"type":"active","data":{}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("claim_moments off", func(t *testing.T) {
		streamer := momentClaimer()
		streamer.Settings.ClaimMoments = false
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-moments-channel-v1.123", `{"type":"active","data":{"moment_id":"moment-1"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{momentClaimer()}})

		h.receive(t, "community-moments-channel-v1.456", `{"type":"active","data":{"moment_id":"moment-1"}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}
