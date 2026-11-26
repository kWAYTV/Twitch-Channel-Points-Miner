package twitch

import (
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// dropClaims is what the client remembers about drops between inventory
// checks: which ones Twitch has already answered for, and which games the
// user wants no drops from.
type dropClaims struct {
	mu            sync.Mutex
	settled       map[string]struct{}
	excludedGames map[string]struct{}
	// unconnectedGames are the games whose drops the last inventory check
	// showed cannot be handed over.
	unconnectedGames map[string]struct{}
}

// DropsUnavailableFor reports whether drops of this game cannot be received
// because the account is not connected to it, as far as the last inventory
// check showed. Watching for them would earn nothing.
func (t *Client) DropsUnavailableFor(game string) bool {
	t.dropClaims.mu.Lock()
	defer t.dropClaims.mu.Unlock()
	_, unconnected := t.dropClaims.unconnectedGames[domain.GameKey(game)]
	return unconnected
}

// rememberUnconnected replaces what is known with what this inventory says,
// so a game stops counting as soon as its account is connected.
func (c *dropClaims) rememberUnconnected(statuses []DropStatus) {
	games := make(map[string]struct{})
	for _, status := range statuses {
		if key := domain.GameKey(status.GameName); status.AccountNotConnected && key != "" {
			games[key] = struct{}{}
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unconnectedGames = games
}

// ExcludeGamesFromDropClaims stops the drops of these games from being
// claimed. The names are lower-case, as the game_exclude setting is kept.
func (t *Client) ExcludeGamesFromDropClaims(games map[string]struct{}) {
	t.dropClaims.mu.Lock()
	defer t.dropClaims.mu.Unlock()
	t.dropClaims.excludedGames = games
}

// skips reports whether a drop is left alone: its game is excluded, or Twitch
// has already said it is claimed. Without the second check a drop Twitch keeps
// listing as unclaimed, as it does for a game the account is not connected to,
// would be claimed and announced again on every check.
func (c *dropClaims) skips(status DropStatus) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, settled := c.settled[status.DropInstanceID]; settled {
		return true
	}
	_, excluded := c.excludedGames[domain.GameKey(status.GameName)]
	return excluded
}

func (c *dropClaims) settle(dropInstanceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.settled == nil {
		c.settled = make(map[string]struct{})
	}
	c.settled[dropInstanceID] = struct{}{}
}
