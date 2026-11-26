// Package drops claims finished Twitch drops and reports their progress.
package drops

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

// claimInterval is how often the inventory is checked for claimable drops.
const claimInterval = 30 * time.Minute

// Inventory is the part of the Twitch client that lists and claims drops.
type Inventory interface {
	ClaimAllDropsFromInventoryWithStatuses() ([]twitch.ClaimedDrop, []twitch.DropStatus, error)
}

// Logger is where claims and progress are reported.
type Logger interface {
	Printf(format string, args ...interface{})
	Eventf(event events.Event, format string, args ...interface{})
	EmojiEventf(emoji string, event events.Event, format string, args ...interface{})
}

// silent is the Logger of a claimer that was given none.
type silent struct{}

func (silent) Printf(string, ...interface{})                            {}
func (silent) Eventf(events.Event, string, ...interface{})              {}
func (silent) EmojiEventf(string, events.Event, string, ...interface{}) {}

// Claimer claims drops from the inventory, and logs their progress when asked to.
type Claimer struct {
	inventory    Inventory
	logger       Logger
	showProgress bool
	// warnedGames are the unconnected games already reported, so each is
	// reported once and again only after it was connected in between.
	warnedGames map[string]struct{}
}

// NewClaimer claims from inventory. Without a logger it claims in silence.
func NewClaimer(inventory Inventory, logger Logger, showProgress bool) *Claimer {
	if logger == nil {
		logger = silent{}
	}
	return &Claimer{inventory: inventory, logger: logger, showProgress: showProgress}
}

// Run claims on every interval until stop is closed.
func (c *Claimer) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(claimInterval)
	defer ticker.Stop()
	c.run(stop, ticker.C)
}

func (c *Claimer) run(stop <-chan struct{}, ticks <-chan time.Time) {
	for {
		select {
		case <-ticks:
			c.Claim("drop claim failed")
		case <-stop:
			return
		}
	}
}

// Claim claims whatever is claimable now. A failure is logged with errorPrefix.
func (c *Claimer) Claim(errorPrefix string) {
	if c == nil || c.inventory == nil {
		return
	}
	drops, statuses, err := c.inventory.ClaimAllDropsFromInventoryWithStatuses()
	if err != nil {
		c.logger.Printf("%s: %v", errorPrefix, err)
		return
	}
	c.logStatuses(statuses)
	c.warnUnconnected(statuses)
	c.logClaimed(drops)
}

// warnUnconnected says which games' drops are not being collected because the
// account is not connected to them.
func (c *Claimer) warnUnconnected(statuses []twitch.DropStatus) {
	unconnected := make(map[string]struct{})
	for _, status := range statuses {
		if !status.AccountNotConnected {
			continue
		}
		game := gameLabel(status.GameName)
		if _, listed := unconnected[game]; listed {
			continue
		}
		unconnected[game] = struct{}{}
		if _, warned := c.warnedGames[game]; !warned {
			c.logger.EmojiEventf(":electric_plug:", events.EventDropNotConnected, "%s", formatUnconnectedLine(status))
		}
	}
	c.warnedGames = unconnected
}

func (c *Claimer) logClaimed(drops []twitch.ClaimedDrop) {
	for _, drop := range drops {
		if drop.AccountNotConnected {
			c.logger.EmojiEventf(":electric_plug:", events.EventDropNotConnected, "%s", formatClaimLine(drop))
			continue
		}
		c.logger.EmojiEventf(":package:", events.EventDropClaim, "%s", formatClaimLine(drop))
	}
}

func (c *Claimer) logStatuses(statuses []twitch.DropStatus) {
	if c == nil || c.logger == nil || !c.showProgress {
		return
	}
	for _, status := range statuses {
		line := formatDropStatusLine(status)
		if line == "" {
			continue
		}
		c.logger.Eventf(events.EventDropStatus, "%s", line)
	}
}
