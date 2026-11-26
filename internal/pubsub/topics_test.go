package pubsub

import (
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/coder/websocket"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestTopicsFor(t *testing.T) {
	for _, row := range []struct {
		name      string
		streamers []*domain.Streamer
		want      []string
	}{
		{"no streamers", nil, []string{"community-points-user-v1.999"}},
		{
			"a streamer with everything off",
			[]*domain.Streamer{{ChannelID: "123"}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123"},
		},
		{
			"a streamer with follow_raid on",
			[]*domain.Streamer{{ChannelID: "123", Settings: domain.StreamerSettings{FollowRaid: true}}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123", "raid.123"},
		},
		{
			"a streamer with make_predictions on",
			[]*domain.Streamer{{ChannelID: "123", Settings: domain.StreamerSettings{MakePredictions: true}}},
			[]string{"community-points-user-v1.999", "predictions-user-v1.999", "video-playback-by-id.123", "predictions-channel-v1.123"},
		},
		{
			"a streamer with claim_moments on",
			[]*domain.Streamer{{ChannelID: "123", Settings: domain.StreamerSettings{ClaimMoments: true}}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123", "community-moments-channel-v1.123"},
		},
		{
			"a streamer with community_goals on",
			[]*domain.Streamer{{ChannelID: "123", Settings: domain.StreamerSettings{CommunityGoals: true}}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123", "community-points-channel-v1.123"},
		},
		{
			"a streamer with everything on",
			[]*domain.Streamer{{ChannelID: "123", Settings: domain.StreamerSettings{
				FollowRaid: true, MakePredictions: true, ClaimMoments: true, CommunityGoals: true,
			}}},
			[]string{
				"community-points-user-v1.999",
				"predictions-user-v1.999",
				"video-playback-by-id.123",
				"raid.123",
				"predictions-channel-v1.123",
				"community-moments-channel-v1.123",
				"community-points-channel-v1.123",
			},
		},
		{
			"two streamers",
			[]*domain.Streamer{{ChannelID: "123"}, {ChannelID: "456"}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123", "video-playback-by-id.456"},
		},
		{
			"the same streamer twice",
			[]*domain.Streamer{{ChannelID: "123"}, {ChannelID: "123"}},
			[]string{"community-points-user-v1.999", "video-playback-by-id.123"},
		},
		{
			"a streamer without a channel ID",
			[]*domain.Streamer{{Username: "tester", Settings: domain.StreamerSettings{FollowRaid: true}}},
			[]string{"community-points-user-v1.999"},
		},
		{
			"a streamer without a channel ID, make_predictions on",
			[]*domain.Streamer{{Username: "tester", Settings: domain.StreamerSettings{MakePredictions: true}}},
			[]string{"community-points-user-v1.999", "predictions-user-v1.999"},
		},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := topicsFor("999", row.streamers); !reflect.DeepEqual(got, row.want) {
				t.Errorf("topicsFor(\"999\", ...) = %q, want %q", got, row.want)
			}
		})
	}
}

func TestTopicNeedsAuth(t *testing.T) {
	for _, row := range []struct {
		name  string
		topic string
		want  bool
	}{
		{"the account's points", "community-points-user-v1.999", true},
		{"the account's predictions", "predictions-user-v1.999", true},
		{"a channel's points", "community-points-channel-v1.123", false},
		{"a channel's predictions", "predictions-channel-v1.123", false},
		{"a channel's playback", "video-playback-by-id.123", false},
		{"no topic", "", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := topicNeedsAuth(row.topic); got != row.want {
				t.Errorf("topicNeedsAuth(%q) = %t, want %t", row.topic, got, row.want)
			}
		})
	}
}

func TestTopicFamily(t *testing.T) {
	for _, row := range []struct {
		name  string
		topic string
		want  string
	}{
		{"a topic with an ID", "video-playback-by-id.123", "video-playback-by-id"},
		{"a topic with two dots", "raid.123.456", "raid"},
		{"a topic without an ID", "raid", "raid"},
		{"a topic that starts with a dot", ".123", ""},
		{"no topic", "", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := topicFamily(row.topic); got != row.want {
				t.Errorf("topicFamily(%q) = %q, want %q", row.topic, got, row.want)
			}
		})
	}
}

func TestChunkTopics(t *testing.T) {
	for _, row := range []struct {
		name   string
		topics []string
		size   int
		want   [][]string
	}{
		{"fewer topics than a chunk", []string{"a", "b"}, 3, [][]string{{"a", "b"}}},
		{"exactly a chunk", []string{"a", "b", "c"}, 3, [][]string{{"a", "b", "c"}}},
		{"one more than a chunk", []string{"a", "b", "c", "d"}, 3, [][]string{{"a", "b", "c"}, {"d"}}},
		{"exactly two chunks", []string{"a", "b", "c", "d"}, 2, [][]string{{"a", "b"}, {"c", "d"}}},
		{"chunks of one", []string{"a", "b"}, 1, [][]string{{"a"}, {"b"}}},
		{"no topics", nil, 3, nil},
		{"a chunk size of zero", []string{"a", "b"}, 0, [][]string{{"a", "b"}}},
		{"a negative chunk size", []string{"a", "b"}, -1, [][]string{{"a", "b"}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := chunkTopics(row.topics, row.size); !reflect.DeepEqual(got, row.want) {
				t.Errorf("chunkTopics(%q, %d) = %q, want %q", row.topics, row.size, got, row.want)
			}
		})
	}
}

func TestClient_listenTopics(t *testing.T) {
	t.Run("a channel's topic", func(t *testing.T) {
		received := make(chan sent, 1)
		conn := connected(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{})

		err := h.client.listenTopics(context.Background(), conn, []string{"video-playback-by-id.123"})

		if err != nil {
			t.Errorf("listenTopics = %v, want nil", err)
		}
		want := sent{"type": "LISTEN", "data": sent{"topics": []interface{}{"video-playback-by-id.123"}}}
		if got := withoutNonce(t, within(t, received)); !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
	})

	t.Run("the account's own topic", func(t *testing.T) {
		received := make(chan sent, 1)
		conn := connected(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{})

		h.client.listenTopics(context.Background(), conn, []string{"community-points-user-v1.999"})

		want := sent{"type": "LISTEN", "data": sent{"topics": []interface{}{"community-points-user-v1.999"}, "auth_token": "token-999"}}
		if got := withoutNonce(t, within(t, received)); !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
	})

	t.Run("two topics", func(t *testing.T) {
		received := make(chan sent, 2)
		conn := connected(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{})

		h.client.listenTopics(context.Background(), conn, []string{"video-playback-by-id.123", "raid.123"})

		want := []sent{
			{"type": "LISTEN", "data": sent{"topics": []interface{}{"video-playback-by-id.123"}}},
			{"type": "LISTEN", "data": sent{"topics": []interface{}{"raid.123"}}},
		}
		got := []sent{withoutNonce(t, within(t, received)), withoutNonce(t, within(t, received))}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
	})

	t.Run("debug on", func(t *testing.T) {
		conn := connected(t, untilClosed)
		h := setup(world{})
		h.logger.DebugOn = true

		h.client.listenTopics(context.Background(), conn, []string{"video-playback-by-id.123"})

		if got, want := h.logger.Debug, []string{"PubSub LISTEN video-playback-by-id.123"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("debug on, logs anonymised", func(t *testing.T) {
		conn := connected(t, untilClosed)
		h := setup(world{anonymise: true})
		h.logger.DebugOn = true

		h.client.listenTopics(context.Background(), conn, []string{"video-playback-by-id.123"})

		if got, want := h.logger.Debug, []string{"PubSub LISTEN video-playback-by-id"}; !reflect.DeepEqual(got, want) {
			t.Errorf("debug lines = %q, want %q", got, want)
		}
	})

	t.Run("a connection already closed", func(t *testing.T) {
		conn := connected(t, untilClosed)
		conn.CloseNow()
		h := setup(world{})

		err := h.client.listenTopics(context.Background(), conn, []string{"video-playback-by-id.123"})

		if shape := regexp.MustCompile(`^failed to write JSON message: `); !shape.MatchString(errorText(err)) {
			t.Errorf("listenTopics = %q, want it to match %s", errorText(err), shape)
		}
	})
}
