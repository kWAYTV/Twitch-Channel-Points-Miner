package twitch

import (
	"testing"
	"time"
)

const (
	inventoryVariables = `{"fetchRewardCampaigns":true}`
	firstDropVariables = `{"input":{"dropInstanceID":"drop-1"}}`
	campaignsVariables = `{"channelID":"123456"}`

	inventoryOneDrop  = `{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"name":"Summer","game":{"displayName":"Tanks"},"timeBasedDrops":[{"name":"Token","requiredMinutesWatched":180,"self":{"dropInstanceID":"drop-1","currentMinutesWatched":180,"isClaimed":false}}]}]}}}}`
	inventoryTwoDrops = `{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"name":"Summer","game":{"displayName":"Tanks"},"timeBasedDrops":[{"name":"Token","requiredMinutesWatched":180,"self":{"dropInstanceID":"drop-1","currentMinutesWatched":180,"isClaimed":false}},{"name":"Crate","requiredMinutesWatched":60,"self":{"dropInstanceID":"drop-2","currentMinutesWatched":60,"isClaimed":false}}]}]}}}}`
	dropGranted       = `{"data":{"claimDropRewards":{"status":"ELIGIBLE_FOR_ALL"}}}`
)

var (
	tokenStatus = DropStatus{DropInstanceID: "drop-1", RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 180, RequiredValue: 180, Claimable: true}
	crateStatus = DropStatus{DropInstanceID: "drop-2", RewardName: "Crate", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 60, RequiredValue: 60, Claimable: true}
	tokenDrop   = ClaimedDrop{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 180, RequiredValue: 180}
	crateDrop   = ClaimedDrop{RewardName: "Crate", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 60, RequiredValue: 60}
)

func TestClient_ClaimDrop(t *testing.T) {
	t.Run("a drop Twitch grants", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, err := client.ClaimDrop("drop-1")

		if err != nil {
			t.Fatalf("ClaimDrop error = %v, want none", err)
		}
		if !claimed {
			t.Errorf("ClaimDrop = false, want true")
		}
		if got, want := fake.twitch.sent("DropsPage_ClaimDropRewards"), []string{firstDropVariables}; !equal(got, want) {
			t.Errorf("drop claims sent = %v, want %v", got, want)
		}
	})

	t.Run("a drop claimed before", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"DROP_INSTANCE_ALREADY_CLAIMED"}}}`))

		claimed, err := client.ClaimDrop("drop-1")

		if err != nil {
			t.Fatalf("ClaimDrop error = %v, want none", err)
		}
		if !claimed {
			t.Errorf("ClaimDrop = false, want true")
		}
	})

	t.Run("a drop Twitch does not grant", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"NOT_ELIGIBLE"}}}`))

		claimed, err := client.ClaimDrop("drop-1")

		if err != nil {
			t.Fatalf("ClaimDrop error = %v, want none", err)
		}
		if claimed {
			t.Errorf("ClaimDrop = true, want false")
		}
	})

	t.Run("an answer with no status", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":null}}`))

		claimed, err := client.ClaimDrop("drop-1")

		if err != nil {
			t.Fatalf("ClaimDrop error = %v, want none", err)
		}
		if claimed {
			t.Errorf("ClaimDrop = true, want false")
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", down("connection refused"))

		claimed, err := client.ClaimDrop("drop-1")

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimDrop error = %v, want %s", err, want)
		}
		if claimed {
			t.Errorf("ClaimDrop = true, want false")
		}
	})
}

func TestClient_ClaimAllDropsFromInventory(t *testing.T) {
	t.Run("an inventory with one drop", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if want := []ClaimedDrop{tokenDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
	})

	t.Run("twitch down for the claim", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", down("connection refused"))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimAllDropsFromInventoryWithStatuses error = %v, want %s", err, want)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
	})
}

func TestClient_ClaimAllDropsFromInventoryWithStatuses(t *testing.T) {
	t.Run("an inventory with one drop", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if want := []ClaimedDrop{tokenDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
		if want := []DropStatus{tokenStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if got, want := fake.twitch.sent("DropsPage_ClaimDropRewards"), []string{firstDropVariables}; !equal(got, want) {
			t.Errorf("drop claims sent = %v, want %v", got, want)
		}
		slept := fake.clock.Slept
		if len(slept) != 1 || slept[0] < 5*time.Second || slept[0] > 10*time.Second || slept[0]%time.Second != 0 {
			t.Errorf("waits = %v, want one of 5 to 10 whole seconds", slept)
		}
	})

	t.Run("an inventory with two drops", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryTwoDrops))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if want := []ClaimedDrop{tokenDrop, crateDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
		if want := []DropStatus{tokenStatus, crateStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
	})

	t.Run("a drop already claimed", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(`{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"name":"Summer","game":{"displayName":"Tanks"},"timeBasedDrops":[{"name":"Token","requiredMinutesWatched":180,"self":{"dropInstanceID":"drop-1","currentMinutesWatched":180,"isClaimed":true}}]}]}}}}`))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		want := []DropStatus{{DropInstanceID: "drop-1", RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 180, RequiredValue: 180, Claimed: true}}
		if !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if got := fake.twitch.sent("DropsPage_ClaimDropRewards"); len(got) != 0 {
			t.Errorf("drop claims sent = %v, want none", got)
		}
	})

	t.Run("a drop of a game the account is not connected to", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"ELIGIBLE_FOR_ALL","isUserAccountConnected":false}}}`))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		want := []ClaimedDrop{{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 180, RequiredValue: 180, AccountNotConnected: true}}
		if !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
	})

	t.Run("a drop of a game the account is connected to", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"ELIGIBLE_FOR_ALL","isUserAccountConnected":true}}}`))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if want := []ClaimedDrop{tokenDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
	})

	t.Run("a drop Twitch says was claimed before", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"DROP_INSTANCE_ALREADY_CLAIMED"}}}`))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none: it is not a new claim", claimed)
		}
		if want := []DropStatus{tokenStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if len(fake.clock.Slept) != 0 {
			t.Errorf("waits = %v, want none", fake.clock.Slept)
		}
	})

	t.Run("a second check, the claimed drop still listed as unclaimed", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))
		if _, _, err := client.ClaimAllDropsFromInventoryWithStatuses(); err != nil {
			t.Fatalf("first check: %v", err)
		}

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if want := []DropStatus{tokenStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if got, want := fake.twitch.sent("DropsPage_ClaimDropRewards"), []string{firstDropVariables}; !equal(got, want) {
			t.Errorf("drop claims sent = %v, want the first check's only: %v", got, want)
		}
	})

	t.Run("a second check after a claim Twitch refused", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"ELIGIBLE_FOR_NONE"}}}`))
		if _, _, err := client.ClaimAllDropsFromInventoryWithStatuses(); err != nil {
			t.Fatalf("first check: %v", err)
		}

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if got, want := fake.twitch.sent("DropsPage_ClaimDropRewards"), []string{firstDropVariables, firstDropVariables}; !equal(got, want) {
			t.Errorf("drop claims sent = %v, want it tried again: %v", got, want)
		}
	})

	t.Run("a drop of an excluded game", func(t *testing.T) {
		client, fake := setup(t)
		client.ExcludeGamesFromDropClaims(map[string]struct{}{"tanks": {}})
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if want := []DropStatus{tokenStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if got := fake.twitch.sent("DropsPage_ClaimDropRewards"); len(got) != 0 {
			t.Errorf("drop claims sent = %v, want none", got)
		}
	})

	t.Run("a drop of a game that is not the excluded one", func(t *testing.T) {
		client, fake := setup(t)
		client.ExcludeGamesFromDropClaims(map[string]struct{}{"planes": {}})
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if want := []ClaimedDrop{tokenDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
	})

	t.Run("a drop not started, with no instance id", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(`{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"name":"Summer","game":{"displayName":"Tanks"},"timeBasedDrops":[{"name":"Token","requiredMinutesWatched":180,"self":{"dropInstanceID":null,"currentMinutesWatched":0,"isClaimed":false}}]}]}}}}`))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		want := []DropStatus{{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", RequiredValue: 180}}
		if !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if got := fake.twitch.sent("DropsPage_ClaimDropRewards"); len(got) != 0 {
			t.Errorf("drop claims sent = %v, want none", got)
		}
	})

	t.Run("a drop Twitch does not grant", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryOneDrop))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(`{"data":{"claimDropRewards":{"status":"NOT_ELIGIBLE"}}}`))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if want := []DropStatus{tokenStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
		if len(fake.clock.Slept) != 0 {
			t.Errorf("waits = %v, want none", fake.clock.Slept)
		}
	})

	t.Run("twitch down for the first of two claims", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryTwoDrops))
		fake.twitch.answer("DropsPage_ClaimDropRewards", firstDropVariables, down("connection refused"))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", ok(dropGranted))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimAllDropsFromInventoryWithStatuses error = %v, want %s", err, want)
		}
		if want := []ClaimedDrop{crateDrop}; !equal(claimed, want) {
			t.Errorf("claimed = %+v, want %+v", claimed, want)
		}
		if want := []DropStatus{tokenStatus, crateStatus}; !equal(statuses, want) {
			t.Errorf("statuses = %+v, want %+v", statuses, want)
		}
	})

	t.Run("twitch down for both claims", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(inventoryTwoDrops))
		fake.twitch.answer("DropsPage_ClaimDropRewards", "", down("connection refused"), down("connection reset"))

		_, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimAllDropsFromInventoryWithStatuses error = %v, want the first failure, %s", err, want)
		}
	})

	t.Run("an inventory with no campaigns", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(`{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":null}}}}`))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if want := []DropStatus{}; !equal(statuses, want) {
			t.Errorf("statuses = %#v, want %#v", statuses, want)
		}
	})

	t.Run("an account with no inventory", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(`{"data":{"currentUser":null}}`))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if statuses != nil {
			t.Errorf("statuses = %#v, want none", statuses)
		}
	})

	t.Run("twitch down for the inventory", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, down("connection refused"))

		claimed, statuses, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none: an unreadable inventory is an empty one", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
		if statuses != nil {
			t.Errorf("statuses = %#v, want none", statuses)
		}
	})
}

func TestClient_ClaimAllDropsFromInventory_malformed(t *testing.T) {
	t.Run("an inventory that is not an object", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("Inventory", inventoryVariables, ok(`{"data":{"currentUser":{"inventory":"none"}}}`))

		claimed, _, err := client.ClaimAllDropsFromInventoryWithStatuses()

		if err != nil {
			t.Fatalf("ClaimAllDropsFromInventoryWithStatuses error = %v, want none", err)
		}
		if claimed != nil {
			t.Errorf("claimed = %+v, want none", claimed)
		}
	})
}

func TestClient_CampaignIDsForStreamer(t *testing.T) {
	t.Run("a channel with two campaigns", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":[{"id":"campaign-1"},{"id":"campaign-2"}]}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if want := []string{"campaign-1", "campaign-2"}; !equal(ids, want) {
			t.Errorf("campaign ids = %#v, want %#v", ids, want)
		}
	})

	t.Run("campaigns that are not a list", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":"none"}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if len(ids) != 0 {
			t.Errorf("campaign ids = %#v, want none", ids)
		}
	})

	t.Run("an entry that is not an object among campaigns", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":["broken",{"id":"campaign-2"}]}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if want := []string{"campaign-2"}; !equal(ids, want) {
			t.Errorf("campaign ids = %#v, want %#v", ids, want)
		}
	})

	t.Run("a campaign with no id among others", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":[{"id":"campaign-1"},{"name":"Summer"},{"id":"campaign-2"}]}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if want := []string{"campaign-1", "campaign-2"}; !equal(ids, want) {
			t.Errorf("campaign ids = %#v, want %#v", ids, want)
		}
	})

	t.Run("a channel with no campaigns", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":null}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if want := []string{}; !equal(ids, want) {
			t.Errorf("campaign ids = %#v, want %#v", ids, want)
		}
	})

	t.Run("a channel with an empty list of campaigns", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, ok(`{"data":{"channel":{"viewerDropCampaigns":[]}}}`))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if err != nil {
			t.Fatalf("CampaignIDsForStreamer error = %v, want none", err)
		}
		if ids != nil {
			t.Errorf("campaign ids = %#v, want nil", ids)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("DropsHighlightService_AvailableDrops", campaignsVariables, down("connection refused"))

		ids, err := client.CampaignIDsForStreamer(newStreamer())

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("CampaignIDsForStreamer error = %v, want %s", err, want)
		}
		if ids != nil {
			t.Errorf("campaign ids = %#v, want nil", ids)
		}
	})
}
