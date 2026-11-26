package logging

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

type publication struct {
	Event   events.Event
	Message notify.Message
}

// fakeSubscriber keeps everything published to it.
type fakeSubscriber struct {
	Received []publication
}

func (s *fakeSubscriber) receive(event events.Event, message notify.Message) {
	s.Received = append(s.Received, publication{event, message})
}

// subscribeFake subscribes a fake under the test's own name until the test ends.
func subscribeFake(t *testing.T) *fakeSubscriber {
	t.Helper()
	subscriber := &fakeSubscriber{}
	Subscribe(t.Name(), subscriber.receive)
	t.Cleanup(func() { Subscribe(t.Name(), nil) })
	return subscriber
}
