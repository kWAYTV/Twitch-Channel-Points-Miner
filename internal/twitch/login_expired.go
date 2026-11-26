package twitch

import (
	"net/http"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// noteLogin reports an expired login from the status of a GQL answer. Twitch
// answers 401 once the saved login is no longer accepted, and nothing the
// miner does works after that until someone logs in again. It is reported
// once, as an event so that it reaches the notifications, and again only
// after a request has worked in between.
func (t *Client) noteLogin(status int) {
	switch {
	case status >= 200 && status <= 299:
		t.loginExpiredReported.Store(false)
	case status == http.StatusUnauthorized:
		if t.loginExpiredReported.Swap(true) || t.logger == nil {
			return
		}
		t.logger.ErrorEventf(events.EventLoginExpired, "Twitch login expired: restart the miner and enter the new code at twitch.tv/activate")
	}
}
