package logging

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

var aliceOnline = notify.Message{Headline: "Alice is online", Details: []string{"Playing Valorant"}}

func TestSubscribe(t *testing.T) {
	t.Run("one subscriber", func(t *testing.T) {
		subscriber := subscribeFake(t)

		publish(events.EventStreamerOnline, aliceOnline)

		want := []publication{{events.EventStreamerOnline, aliceOnline}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("two subscribers", func(t *testing.T) {
		first := subscribeFake(t)
		second := &fakeSubscriber{}
		Subscribe(t.Name()+"/second", second.receive)
		t.Cleanup(func() { Subscribe(t.Name()+"/second", nil) })

		publish(events.EventStreamerOnline, aliceOnline)

		want := []publication{{events.EventStreamerOnline, aliceOnline}}
		if !reflect.DeepEqual(first.Received, want) {
			t.Errorf("the first subscriber received %+v, want %+v", first.Received, want)
		}
		if !reflect.DeepEqual(second.Received, want) {
			t.Errorf("the second subscriber received %+v, want %+v", second.Received, want)
		}
	})

	t.Run("two events", func(t *testing.T) {
		subscriber := subscribeFake(t)
		publish(events.EventStreamerOnline, aliceOnline)

		publish(events.EventShutdown, notify.Message{Headline: "Session ended"})

		want := []publication{
			{events.EventStreamerOnline, aliceOnline},
			{events.EventShutdown, notify.Message{Headline: "Session ended"}},
		}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})

	t.Run("a name subscribed twice", func(t *testing.T) {
		replaced := subscribeFake(t)
		replacement := subscribeFake(t)

		publish(events.EventStreamerOnline, aliceOnline)

		if len(replaced.Received) != 0 {
			t.Errorf("the replaced subscriber received %+v, want nothing", replaced.Received)
		}
		want := []publication{{events.EventStreamerOnline, aliceOnline}}
		if !reflect.DeepEqual(replacement.Received, want) {
			t.Errorf("the replacement received %+v, want %+v", replacement.Received, want)
		}
	})

	t.Run("a subscriber removed", func(t *testing.T) {
		subscriber := subscribeFake(t)
		Subscribe(t.Name(), nil)

		publish(events.EventStreamerOnline, aliceOnline)

		if len(subscriber.Received) != 0 {
			t.Errorf("subscriber received %+v, want nothing", subscriber.Received)
		}
	})

	t.Run("a name removed that was never subscribed", func(t *testing.T) {
		subscriber := subscribeFake(t)
		Subscribe(t.Name()+"/unknown", nil)

		publish(events.EventStreamerOnline, aliceOnline)

		want := []publication{{events.EventStreamerOnline, aliceOnline}}
		if !reflect.DeepEqual(subscriber.Received, want) {
			t.Errorf("subscriber received %+v, want %+v", subscriber.Received, want)
		}
	})
}
