package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestChannelIDFromPayload(t *testing.T) {
	object := func(pairs ...interface{}) map[string]interface{} {
		made := map[string]interface{}{}
		for i := 0; i < len(pairs); i += 2 {
			made[pairs[i].(string)] = pairs[i+1]
		}
		return made
	}
	channel := func(id string) map[string]interface{} { return object("channel_id", id) }

	for _, row := range []struct {
		name    string
		payload map[string]interface{}
		topic   string
		want    string
	}{
		{
			"a channel in every place",
			object("data", object("prediction", channel("1"), "claim", channel("2"), "channel_id", "3", "balance", channel("4"))),
			"raid.5",
			"1",
		},
		{
			"no prediction",
			object("data", object("claim", channel("2"), "channel_id", "3", "balance", channel("4"))),
			"raid.5",
			"2",
		},
		{
			"no prediction or claim",
			object("data", object("channel_id", "3", "balance", channel("4"))),
			"raid.5",
			"3",
		},
		{
			"only a balance and the topic",
			object("data", object("balance", channel("4"))),
			"raid.5",
			"4",
		},
		{
			"only the topic",
			object("data", object()),
			"raid.5",
			"5",
		},
		{
			"a blank channel in the prediction",
			object("data", object("prediction", channel(" "), "channel_id", "3")),
			"raid.5",
			"3",
		},
		{"no data", object(), "raid.5", "5"},
		{"no payload", nil, "raid.5", "5"},
		{"a topic with spaces around its ID", object(), "raid. 5 ", "5"},
		{"a topic that ends in a dot", object(), "raid.", ""},
		{"a topic without a dot", object(), "raid", ""},
		{"a topic with two dots", object(), "raid.5.6", ""},
		{"no topic", object(), "", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := channelIDFromPayload(row.payload, row.topic); got != row.want {
				t.Errorf("channelIDFromPayload(%v, %q) = %q, want %q", row.payload, row.topic, got, row.want)
			}
		})
	}
}

func viewed() *domain.Streamer {
	return &domain.Streamer{Username: "tester", ChannelID: "123"}
}

func TestClient_handleMessage(t *testing.T) {
	t.Run("a message for a streamer", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})

		err := h.client.handleMessage([]byte(`{"type":"MESSAGE","data":{"topic":"video-playback-by-id.123","message":"{\"type\":\"stream-down\"}"}}`), nil)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.effects(), (effects{Presence: []presence{{"tester", false, "stream-down"}}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a PONG", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})
		pongs := 0

		err := h.client.handleMessage([]byte(`{"type":"PONG"}`), func() { pongs++ })

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if pongs != 1 {
			t.Errorf("told of %d PONGs, want 1", pongs)
		}
	})

	t.Run("a PONG with nobody to tell", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})

		err := h.client.handleMessage([]byte(`{"type":"PONG"}`), nil)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
	})

	for _, row := range []struct {
		name string
		raw  string
	}{
		{"a RESPONSE", `{"type":"RESPONSE","error":"ERR_BADAUTH","nonce":"abc"}`},
		{"a RECONNECT", `{"type":"RECONNECT"}`},
		{"a type it doesn't know", `{"type":"AUTH_REVOKED"}`},
		{"no type", `{}`},
		{"a MESSAGE without data", `{"type":"MESSAGE"}`},
		{"a MESSAGE with nothing in it", `{"type":"MESSAGE","data":{"topic":"video-playback-by-id.123","message":""}}`},
		{"a topic it doesn't know", `{"type":"MESSAGE","data":{"topic":"whispers.123","message":"{\"type\":\"stream-down\"}"}}`},
		{"no topic", `{"type":"MESSAGE","data":{"message":"{\"type\":\"stream-down\"}"}}`},
	} {
		t.Run(row.name, func(t *testing.T) {
			h := setup(world{streamers: []*domain.Streamer{viewed()}})

			err := h.client.handleMessage([]byte(row.raw), func() { t.Errorf("told of a PONG") })

			if err != nil {
				t.Errorf("handleMessage = %v, want nil", err)
			}
			if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
				t.Errorf("effects = %+v, want none", got)
			}
		})
	}

	t.Run("not JSON", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})

		err := h.client.handleMessage([]byte(`stream-down`), nil)

		if got, want := errorText(err), "invalid character 's' looking for beginning of value"; got != want {
			t.Errorf("handleMessage = %q, want %q", got, want)
		}
	})

	t.Run("a MESSAGE whose contents are not JSON", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})

		err := h.receive(t, "video-playback-by-id.123", "stream-down")

		if got, want := errorText(err), "invalid character 's' looking for beginning of value"; got != want {
			t.Errorf("handleMessage = %q, want %q", got, want)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("debug on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})
		h.logger.DebugOn = true

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-down"}`)

		if got, want := h.logger.Debug, []string{"PubSub topic video-playback-by-id type=stream-down"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("debug off", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}})

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-down"}`)

		if got := h.logger.Debug; got != nil {
			t.Errorf("debug lines = %q, want none", got)
		}
	})

	t.Run("deep debug on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}, deepDebug: true})
		h.deep.DeepOn = true

		h.receive(t, "video-playback-by-id.123", ` {"type":"stream-down"} `)

		if got, want := h.deep.Deep, []string{`PubSub topic video-playback-by-id.123 payload {"type":"stream-down"}`}; !reflect.DeepEqual(got, want) {
			t.Errorf("deep debug lines = %q, want %q", got, want)
		}
	})

	t.Run("deep debug off", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}, deepDebug: true})

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-down"}`)

		if got := h.deep.Deep; got != nil {
			t.Errorf("deep debug lines = %q, want none", got)
		}
	})

	t.Run("deep debug on, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{viewed()}, deepDebug: true, anonymise: true})
		h.deep.DeepOn = true

		h.receive(t, "video-playback-by-id.123", `{"type":"stream-down"}`)

		if got := h.deep.Deep; got != nil {
			t.Errorf("deep debug lines = %q, want none", got)
		}
	})
}

func errorText(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func TestPayloadType(t *testing.T) {
	for _, row := range []struct {
		name    string
		payload map[string]interface{}
		want    string
	}{
		{"a type in lower case", map[string]interface{}{"type": "stream-up"}, "stream-up"},
		{"a type in upper case", map[string]interface{}{"type": "EVENT-CREATED"}, "event-created"},
		{"no type", map[string]interface{}{}, "<nil>"},
		{"no payload", nil, "<nil>"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := payloadType(row.payload); got != row.want {
				t.Errorf("payloadType(%v) = %q, want %q", row.payload, got, row.want)
			}
		})
	}
}
