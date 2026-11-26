package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestPointGainOf(t *testing.T) {
	type object = map[string]interface{}

	for _, row := range []struct {
		name   string
		data   object
		want   pointGain
		wantOK bool
	}{
		{
			"a gain with the balance after it",
			object{"point_gain": object{"reason_code": "WATCH", "total_points": 10.0}, "balance": object{"balance": 1010.0}},
			pointGain{"WATCH", 10, 1010},
			true,
		},
		{
			"a gain without a balance",
			object{"point_gain": object{"reason_code": "WATCH", "total_points": 10.0}},
			pointGain{"WATCH", 10, 1000},
			true,
		},
		{
			"a balance of zero",
			object{"point_gain": object{"reason_code": "WATCH", "total_points": 10.0}, "balance": object{"balance": 0.0}},
			pointGain{"WATCH", 10, 0},
			true,
		},
		{
			"a reason in lower case",
			object{"point_gain": object{"reason_code": "watch_streak", "total_points": 450.0}},
			pointGain{"WATCH_STREAK", 450, 1000},
			true,
		},
		{
			"points taken away",
			object{"point_gain": object{"reason_code": "PREDICTION", "total_points": -50.0}},
			pointGain{"PREDICTION", -50, 1000},
			true,
		},
		{
			"no points",
			object{"point_gain": object{"reason_code": "WATCH"}},
			pointGain{"WATCH", 0, 1000},
			true,
		},
		{
			"points written as text",
			object{"point_gain": object{"reason_code": "WATCH", "total_points": "10"}},
			pointGain{"WATCH", 0, 1000},
			true,
		},
		{"no gain", object{"balance": object{"balance": 1010.0}}, pointGain{}, false},
		{"a gain that is not an object", object{"point_gain": "WATCH"}, pointGain{}, false},
		{"no data", nil, pointGain{}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, ok := pointGainOf(row.data, 1000)
			if got != row.want || ok != row.wantOK {
				t.Errorf("pointGainOf(%v, 1000) = %+v, %t, want %+v, %t", row.data, got, ok, row.want, row.wantOK)
			}
		})
	}

	t.Run("no reason", func(t *testing.T) {
		got, ok := pointGainOf(object{"point_gain": object{"total_points": 10.0}}, 1000)
		if want := (pointGain{"", 10, 1000}); got != want || !ok {
			t.Errorf("pointGainOf = %+v, %t, want %+v, true", got, ok, want)
		}
	})
}

func TestClient_processPointsEarned(t *testing.T) {
	t.Run("a gain on a followed channel", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", ChannelPoints: 1000}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		err := h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"123","point_gain":{"reason_code":"WATCH","total_points":10},"balance":{"channel_id":"123","balance":1010}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.effects(), (effects{Gains: []gain{{"tester", 10, "WATCH", 1010}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
		if got, want := streamer.ChannelPoints, 1000; got != want {
			t.Errorf("balance = %d, want %d", got, want)
		}
	})

	t.Run("the same gain twice", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123", ChannelPoints: 1000}}})
		h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"123","point_gain":{"reason_code":"WATCH","total_points":10},"balance":{"channel_id":"123","balance":1010}}}`)

		h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"123","point_gain":{"reason_code":"WATCH","total_points":10},"balance":{"channel_id":"123","balance":1010}}}`)

		if got, want := h.gains, []gain{{"tester", 10, "WATCH", 1010}, {"tester", 10, "WATCH", 1010}}; !reflect.DeepEqual(got, want) {
			t.Errorf("gains = %+v, want %+v", got, want)
		}
	})

	t.Run("a gain on a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123", ChannelPoints: 1000}}})

		h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"456","point_gain":{"reason_code":"WATCH","total_points":10},"balance":{"channel_id":"456","balance":1010}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a message with no gain in it", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123", ChannelPoints: 1000}}})

		h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"123","balance":{"channel_id":"123","balance":1010}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a message with no data", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123", ChannelPoints: 1000}}})

		err := h.receive(t, "community-points-user-v1.999", `{"type":"points-earned"}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a gain with no gain handler", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester", ChannelID: "123", ChannelPoints: 1000}
		h := setup(world{streamers: []*domain.Streamer{streamer}, noHandlers: true})

		err := h.receive(t, "community-points-user-v1.999", `{"type":"points-earned","data":{"channel_id":"123","point_gain":{"reason_code":"WATCH","total_points":10},"balance":{"channel_id":"123","balance":1010}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := streamer.ChannelPoints, 1000; got != want {
			t.Errorf("balance = %d, want %d", got, want)
		}
	})
}
