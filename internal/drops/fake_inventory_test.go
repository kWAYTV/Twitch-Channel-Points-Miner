package drops

import "github.com/kWAYTV/twitch-points-miner/internal/twitch"

// fakeInventory hands back the drops and statuses it holds, and keeps the
// claims asked of it.
type fakeInventory struct {
	Claimed  []twitch.ClaimedDrop
	Statuses []twitch.DropStatus
	Err      error
	Calls    []string
}

func (i *fakeInventory) ClaimAllDropsFromInventory() ([]twitch.ClaimedDrop, error) {
	i.Calls = append(i.Calls, "claim")
	if i.Err != nil {
		return nil, i.Err
	}
	return i.Claimed, nil
}

func (i *fakeInventory) ClaimAllDropsFromInventoryWithStatuses() ([]twitch.ClaimedDrop, []twitch.DropStatus, error) {
	i.Calls = append(i.Calls, "claim with statuses")
	if i.Err != nil {
		return nil, nil, i.Err
	}
	return i.Claimed, i.Statuses, nil
}
