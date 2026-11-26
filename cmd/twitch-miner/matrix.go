package main

import (
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

type matrixConfig struct {
	Homeserver  string   `json:"homeserver"`
	AccessToken string   `json:"access_token"`
	RoomID      string   `json:"room_id"`
	Events      []string `json:"events"`
}

func defaultMatrixConfig() map[string]interface{} {
	return map[string]interface{}{
		"homeserver":   "",
		"access_token": "",
		"room_id":      "",
		"events":       []interface{}{},
	}
}

func (c matrixConfig) settings() notify.MatrixSettings {
	return notify.MatrixSettings{
		Homeserver:  c.Homeserver,
		AccessToken: c.AccessToken,
		RoomID:      c.RoomID,
		Events:      c.Events,
	}
}

// startMatrix sends logged events to a Matrix room when a homeserver, an access
// token and a room are configured.
func startMatrix(cfg matrixConfig) {
	notifier := notify.NewMatrixNotifier(cfg.settings())
	if notifier == nil {
		return
	}
	logging.Subscribe("matrix", func(event events.Event, message notify.Message) {
		notifier.Send(message, event)
	})
}
