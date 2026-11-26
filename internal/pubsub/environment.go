package pubsub

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// environment is where a Client finds PubSub and how it reads and spends time.
type environment struct {
	endpoint string
	now      func() time.Time
	// after runs f once d has passed, without blocking.
	after func(d time.Duration, f func())
	sleep func(d time.Duration)
}

// twitchEnvironment is Twitch's PubSub on the real clock.
func twitchEnvironment() environment {
	return environment{
		endpoint: twitchapi.WebsocketURL,
		now:      time.Now,
		after:    func(d time.Duration, f func()) { time.AfterFunc(d, f) },
		sleep:    time.Sleep,
	}
}
