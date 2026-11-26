package pubsub

import "time"

// How long a viewer takes to notice something on the page and click it.
const (
	reactionSoonest = 2 * time.Second
	reactionLatest  = 8 * time.Second
)

// react answers something Twitch announced: at once when timings are exact,
// and otherwise after the pause a viewer would take. key names what is being
// answered; while its answer is pending, a repeat of the announcement is
// ignored.
func (p *Client) react(key string, act func()) {
	pause := p.env.human.Between(reactionSoonest, reactionLatest)
	if pause <= 0 {
		act()
		return
	}
	p.reactingMu.Lock()
	_, pending := p.reacting[key]
	if !pending {
		p.reacting[key] = struct{}{}
	}
	p.reactingMu.Unlock()
	if pending {
		return
	}
	p.env.after(pause, func() {
		defer func() {
			p.reactingMu.Lock()
			delete(p.reacting, key)
			p.reactingMu.Unlock()
		}()
		act()
	})
}
