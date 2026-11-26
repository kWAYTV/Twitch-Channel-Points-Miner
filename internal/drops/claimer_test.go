package drops

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

var (
	now = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	tokenDrop = twitch.ClaimedDrop{RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180}
	crateDrop = twitch.ClaimedDrop{RewardName: "Crate", CampaignName: "Winter", CurrentValue: 60, RequiredValue: 60}

	tokenStatus = twitch.DropStatus{
		RewardName:    "Token",
		CampaignName:  "Summer",
		GameName:      "Tanks",
		ChannelName:   "example_streamer",
		CurrentValue:  63,
		RequiredValue: 180,
	}
)

const (
	tokenClaimLine  = "Claim Token (Summer) 180/180 (100%)"
	crateClaimLine  = "Claim Crate (Winter) 60/60 (100%)"
	tokenStatusLine = "DROP_PROGRESS [#######-------------] 35% | Tanks / Summer / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS"
)

func setup(showProgress bool) (*Claimer, *fakeInventory, *fakeLogger) {
	inventory := &fakeInventory{}
	logger := &fakeLogger{}
	return NewClaimer(inventory, logger, showProgress), inventory, logger
}

func checkEntries(t *testing.T, logger *fakeLogger, want []logEntry) {
	t.Helper()
	if !reflect.DeepEqual(logger.Entries, want) {
		t.Errorf("logged %+v, want %+v", logger.Entries, want)
	}
}

func checkCalls(t *testing.T, inventory *fakeInventory, want []string) {
	t.Helper()
	if !reflect.DeepEqual(inventory.Calls, want) {
		t.Errorf("inventory was asked to %q, want %q", inventory.Calls, want)
	}
}

func TestClaimer_Claim(t *testing.T) {
	t.Run("one finished drop", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Claimed = []twitch.ClaimedDrop{tokenDrop}

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{{Emoji: ":package:", Event: "DROP_CLAIM", Text: tokenClaimLine}})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("two finished drops", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Claimed = []twitch.ClaimedDrop{tokenDrop, crateDrop}

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{
			{Emoji: ":package:", Event: "DROP_CLAIM", Text: tokenClaimLine},
			{Emoji: ":package:", Event: "DROP_CLAIM", Text: crateClaimLine},
		})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("nothing to claim", func(t *testing.T) {
		claimer, inventory, logger := setup(false)

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, nil)
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("a drop in progress, with progress hidden", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Statuses = []twitch.DropStatus{tokenStatus}

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, nil)
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("inventory down", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Err = errors.New("503")

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{{Text: "startup claim failed: 503"}})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("a drop in progress, with progress shown", func(t *testing.T) {
		claimer, inventory, logger := setup(true)
		inventory.Statuses = []twitch.DropStatus{tokenStatus}

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{{Event: "DROP_STATUS", Text: tokenStatusLine}})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("a drop in progress and a finished one, with progress shown", func(t *testing.T) {
		claimer, inventory, logger := setup(true)
		inventory.Statuses = []twitch.DropStatus{tokenStatus}
		inventory.Claimed = []twitch.ClaimedDrop{crateDrop}

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{
			{Event: "DROP_STATUS", Text: tokenStatusLine},
			{Emoji: ":package:", Event: "DROP_CLAIM", Text: crateClaimLine},
		})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("nothing to claim, with progress shown", func(t *testing.T) {
		claimer, inventory, logger := setup(true)

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, nil)
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("inventory down, with progress shown", func(t *testing.T) {
		claimer, inventory, logger := setup(true)
		inventory.Err = errors.New("503")

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{{Text: "startup claim failed: 503"}})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("the same claim twice", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Claimed = []twitch.ClaimedDrop{tokenDrop}
		claimer.Claim("startup claim failed")

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, []logEntry{
			{Emoji: ":package:", Event: "DROP_CLAIM", Text: tokenClaimLine},
			{Emoji: ":package:", Event: "DROP_CLAIM", Text: tokenClaimLine},
		})
		checkCalls(t, inventory, []string{"claim", "claim"})
	})

	t.Run("no logger", func(t *testing.T) {
		inventory := &fakeInventory{
			Claimed:  []twitch.ClaimedDrop{{RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180}},
			Statuses: []twitch.DropStatus{{RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180}},
		}
		claimer := NewClaimer(inventory, nil, true)

		claimer.Claim("startup claim failed")

		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("no logger, the inventory unreadable", func(t *testing.T) {
		inventory := &fakeInventory{Err: errors.New("503")}
		claimer := NewClaimer(inventory, nil, false)

		claimer.Claim("startup claim failed")

		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("no inventory", func(t *testing.T) {
		logger := &fakeLogger{}
		claimer := NewClaimer(nil, logger, true)

		claimer.Claim("startup claim failed")

		checkEntries(t, logger, nil)
	})

	t.Run("no claimer", func(t *testing.T) {
		var claimer *Claimer

		claimer.Claim("startup claim failed")
	})
}

func TestClaimer_Run(t *testing.T) {
	t.Run("stopped before the first interval", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Claimed = []twitch.ClaimedDrop{tokenDrop}
		stop := make(chan struct{})
		close(stop)

		claimer.Run(stop)

		checkEntries(t, logger, nil)
		checkCalls(t, inventory, nil)
	})
}

// runThrough runs the claimer through the given number of intervals, then stops it.
func runThrough(claimer *Claimer, intervals int) {
	stop := make(chan struct{})
	ticks := make(chan time.Time)
	done := make(chan struct{})
	go func() {
		claimer.run(stop, ticks)
		close(done)
	}()
	for i := 0; i < intervals; i++ {
		ticks <- now.Add(time.Duration(i+1) * claimInterval)
	}
	close(stop)
	<-done
}

func TestClaimer_run(t *testing.T) {
	t.Run("one interval", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Claimed = []twitch.ClaimedDrop{tokenDrop}

		runThrough(claimer, 1)

		checkEntries(t, logger, []logEntry{{Emoji: ":package:", Event: "DROP_CLAIM", Text: tokenClaimLine}})
		checkCalls(t, inventory, []string{"claim"})
	})

	t.Run("two intervals", func(t *testing.T) {
		claimer, inventory, _ := setup(false)

		runThrough(claimer, 2)

		checkCalls(t, inventory, []string{"claim", "claim"})
	})

	t.Run("one interval, inventory down", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Err = errors.New("503")

		runThrough(claimer, 1)

		checkEntries(t, logger, []logEntry{{Text: "drop claim failed: 503"}})
	})

	t.Run("no intervals", func(t *testing.T) {
		claimer, inventory, _ := setup(false)

		runThrough(claimer, 0)

		checkCalls(t, inventory, nil)
	})
}

func TestClaimer_Claim_unconnectedGames(t *testing.T) {
	const tanksLine = "Drops for Tanks are not collected: your Twitch account is not connected to the game. Connect it at https://www.twitch.tv/drops/inventory"
	unconnected := []twitch.DropStatus{
		{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", AccountNotConnected: true},
		{RewardName: "Crate", CampaignName: "Summer", GameName: "Tanks", AccountNotConnected: true},
	}

	t.Run("two drops of one unconnected game", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Statuses = unconnected

		claimer.Claim("drop claim failed")

		checkEntries(t, logger, []logEntry{{Text: tanksLine}})
	})

	t.Run("the same game at the next check", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Statuses = unconnected
		claimer.Claim("drop claim failed")

		claimer.Claim("drop claim failed")

		checkEntries(t, logger, []logEntry{{Text: tanksLine}})
	})

	t.Run("a game connected, then disconnected again", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Statuses = unconnected
		claimer.Claim("drop claim failed")
		inventory.Statuses = []twitch.DropStatus{{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks"}}
		claimer.Claim("drop claim failed")
		inventory.Statuses = unconnected

		claimer.Claim("drop claim failed")

		checkEntries(t, logger, []logEntry{{Text: tanksLine}, {Text: tanksLine}})
	})

	t.Run("a connected game", func(t *testing.T) {
		claimer, inventory, logger := setup(false)
		inventory.Statuses = []twitch.DropStatus{{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks"}}

		claimer.Claim("drop claim failed")

		checkEntries(t, logger, nil)
	})
}
