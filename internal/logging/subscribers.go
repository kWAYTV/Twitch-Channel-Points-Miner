package logging

import (
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

// A Subscriber receives every logged event, described without colour codes.
type Subscriber func(event events.Event, message notify.Message)

var subscribers sync.Map

// Subscribe installs fn under name, replacing an earlier subscriber of that name.
// A nil fn removes it. Subscribers must be safe for concurrent use and must not block for long.
func Subscribe(name string, fn Subscriber) {
	if fn == nil {
		subscribers.Delete(name)
		return
	}
	subscribers.Store(name, fn)
}

func publish(event events.Event, message notify.Message) {
	subscribers.Range(func(_, fn any) bool {
		fn.(Subscriber)(event, message)
		return true
	})
}
