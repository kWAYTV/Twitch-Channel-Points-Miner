package pubsub

import (
	"reflect"
	"regexp"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_Start(t *testing.T) {
	t.Run("one streamer", func(t *testing.T) {
		received := make(chan sent, 2)
		endpoint := fakePubSub(t, func(conn *websocket.Conn) { forward(conn, received) })
		h := setup(world{endpoint: endpoint, streamers: []*domain.Streamer{{Username: "tester", ChannelID: "123"}}})
		h.account.ContextLoaded = make(chan string, 1)
		stop := make(chan struct{})
		defer close(stop)

		h.client.Start(stop)

		want := []sent{
			{"type": "LISTEN", "data": sent{"topics": []interface{}{"community-points-user-v1.999"}, "auth_token": "token-999"}},
			{"type": "LISTEN", "data": sent{"topics": []interface{}{"video-playback-by-id.123"}}},
		}
		got := []sent{withoutNonce(t, within(t, received)), withoutNonce(t, within(t, received))}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("sent %v, want %v", got, want)
		}
		if got, want := within(t, h.account.ContextLoaded), "tester"; got != want {
			t.Errorf("checked for a pending bonus of %q, want %q", got, want)
		}
	})

	t.Run("more topics than one connection takes", func(t *testing.T) {
		var connections atomic.Int32
		listens := make(chan int32, 61)
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			connection := connections.Add(1)
			for {
				if _, ok := receive(conn); !ok {
					return
				}
				listens <- connection
			}
		})
		var streamers []*domain.Streamer
		for _, id := range []string{
			"101", "102", "103", "104", "105", "106", "107", "108", "109", "110",
			"111", "112", "113", "114", "115", "116", "117", "118", "119", "120",
			"121", "122", "123", "124", "125", "126", "127", "128", "129", "130",
		} {
			streamers = append(streamers, &domain.Streamer{ChannelID: id, Settings: domain.StreamerSettings{FollowRaid: true}})
		}
		h := setup(world{endpoint: endpoint, streamers: streamers})
		stop := make(chan struct{})
		defer close(stop)

		h.client.Start(stop)

		perConnection := map[int32]int{}
		for i := 0; i < 61; i++ {
			perConnection[within(t, listens)]++
		}
		got := []int{perConnection[1], perConnection[2]}
		sort.Ints(got)
		if want := []int{11, 50}; !reflect.DeepEqual(got, want) {
			t.Errorf("topics per connection = %v, want %v", got, want)
		}
	})

	t.Run("no user ID", func(t *testing.T) {
		h := setup(world{endpoint: notPubSub(t)})
		h.account.ID = ""

		h.client.Start(make(chan struct{}))

		if got, want := h.effects(), (effects{Lines: []string{"error: PubSub topic error: no user id for pubsub"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("the server asks to reconnect", func(t *testing.T) {
		var connections atomic.Int32
		reconnected := make(chan sent, 1)
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			if connections.Add(1) == 1 {
				receive(conn)
				send(conn, `{"type":"RECONNECT"}`)
				untilClosed(conn)
				return
			}
			forward(conn, reconnected)
		})
		waiting := make(chan time.Duration)
		resume := make(chan struct{})
		h := setup(world{endpoint: endpoint, sleep: func(d time.Duration) {
			waiting <- d
			<-resume
		}})
		stop := make(chan struct{})
		defer close(stop)

		h.client.Start(stop)

		if got, want := within(t, waiting), 60*time.Second; got != want {
			t.Errorf("waited %v, want %v", got, want)
		}
		want := []string{
			"info: Connected to Twitch PubSub (conn #1) with 1 topic(s)",
			"info: PubSub[1] reconnect requested; waiting ~60 seconds",
		}
		if got := h.logger.Shown(); !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
		close(resume)
		wantListen := sent{"type": "LISTEN", "data": sent{"topics": []interface{}{"community-points-user-v1.999"}, "auth_token": "token-999"}}
		if got := withoutNonce(t, within(t, reconnected)); !reflect.DeepEqual(got, wantListen) {
			t.Errorf("sent %v on the second connection, want %v", got, wantListen)
		}
	})

	t.Run("the server hangs up", func(t *testing.T) {
		var connections atomic.Int32
		reconnected := make(chan sent, 1)
		endpoint := fakePubSub(t, func(conn *websocket.Conn) {
			if connections.Add(1) == 1 {
				receive(conn)
				hangUp(conn)
				return
			}
			forward(conn, reconnected)
		})
		waiting := make(chan time.Duration)
		resume := make(chan struct{})
		h := setup(world{endpoint: endpoint, sleep: func(d time.Duration) {
			waiting <- d
			<-resume
		}})
		stop := make(chan struct{})
		defer close(stop)

		h.client.Start(stop)

		if got, want := within(t, waiting), 10*time.Second; got != want {
			t.Errorf("waited %v, want %v", got, want)
		}
		got := h.logger.Shown()
		if want := "info: Connected to Twitch PubSub (conn #1) with 1 topic(s)"; len(got) != 2 || got[0] != want {
			t.Errorf("lines = %q, want %q and one more", got, want)
		}
		if shape := regexp.MustCompile(`^error: PubSub\[1\] connection error: .*status = StatusNormalClosure and reason = "bye"$`); len(got) == 2 && !shape.MatchString(got[1]) {
			t.Errorf("second line = %q, want it to match %s", got[1], shape)
		}
		close(resume)
		wantListen := sent{"type": "LISTEN", "data": sent{"topics": []interface{}{"community-points-user-v1.999"}, "auth_token": "token-999"}}
		if got := withoutNonce(t, within(t, reconnected)); !reflect.DeepEqual(got, wantListen) {
			t.Errorf("sent %v on the second connection, want %v", got, wantListen)
		}
	})
}
