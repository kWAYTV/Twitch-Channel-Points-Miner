package twitch

import "testing"

type object = map[string]interface{}
type list = []interface{}

func TestCampaignNameFromInventory(t *testing.T) {
	tests := []struct {
		name     string
		campaign object
		want     string
	}{
		{"a campaign with every name", object{"name": "Summer", "displayName": "Summer Event", "gameDisplayName": "Tanks", "game": object{"displayName": "Tanks!", "name": "tanks"}}, "Summer"},
		{"a display name", object{"displayName": "Summer Event"}, "Summer Event"},
		{"an empty name before a display name", object{"name": "", "displayName": "Summer Event"}, "Summer Event"},
		{"a game display name on the campaign", object{"gameDisplayName": "Tanks"}, "Tanks"},
		{"a game with a display name", object{"game": object{"displayName": "Tanks!", "name": "tanks"}}, "Tanks!"},
		{"a game with only a name", object{"game": object{"name": "tanks"}}, "tanks"},
		{"a name that is a number", object{"name": 7}, ""},
		{"an empty campaign", object{}, ""},
		{"no campaign", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := campaignNameFromInventory(tt.campaign); got != tt.want {
				t.Errorf("campaignNameFromInventory(%v) = %q, want %q", tt.campaign, got, tt.want)
			}
		})
	}
}

func TestRewardNameFromInventory(t *testing.T) {
	tests := []struct {
		name string
		drop object
		want string
	}{
		{"a drop with every name", object{"benefit": object{"name": "Token", "displayName": "Gold Token"}, "name": "Drop 1", "displayName": "First drop"}, "Token"},
		{"a benefit with a display name", object{"benefit": object{"displayName": "Gold Token"}, "name": "Drop 1"}, "Gold Token"},
		{"a benefit with no names", object{"benefit": object{}, "name": "Drop 1"}, "Drop 1"},
		{"a benefit that is not an object", object{"benefit": "Token", "name": "Drop 1"}, "Drop 1"},
		{"a drop name", object{"name": "Drop 1", "displayName": "First drop"}, "Drop 1"},
		{"a drop display name", object{"displayName": "First drop"}, "First drop"},
		{"benefit edges keyed by position", object{"benefit": object{"edges": object{"0": object{"node": object{"name": "Token"}}}}}, "Token"},
		{"an empty drop", object{}, ""},
		{"no drop", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewardNameFromInventory(tt.drop); got != tt.want {
				t.Errorf("rewardNameFromInventory(%v) = %q, want %q", tt.drop, got, tt.want)
			}
		})
	}

	t.Run("benefit edges as a list", func(t *testing.T) {
		drop := object{"benefit": object{"edges": list{object{"node": object{"name": "Token"}}}}}
		if got := rewardNameFromInventory(drop); got != "Token" {
			t.Errorf("rewardNameFromInventory(%v) = %q, want %q", drop, got, "Token")
		}
	})
}

func TestGameNameFromInventory(t *testing.T) {
	tests := []struct {
		name     string
		campaign object
		want     string
	}{
		{"a campaign with every name", object{"game": object{"displayName": "Tanks!", "name": "tanks"}, "gameDisplayName": "Tanks", "gameName": "tnk"}, "Tanks!"},
		{"a game with only a name", object{"game": object{"name": "tanks"}, "gameDisplayName": "Tanks"}, "tanks"},
		{"a game with no names", object{"game": object{}, "gameDisplayName": "Tanks"}, "Tanks"},
		{"a game that is not an object", object{"game": "tanks", "gameDisplayName": "Tanks"}, "Tanks"},
		{"a game display name on the campaign", object{"gameDisplayName": "Tanks", "gameName": "tnk"}, "Tanks"},
		{"a game name on the campaign", object{"gameName": "tnk"}, "tnk"},
		{"an empty campaign", object{}, ""},
		{"no campaign", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gameNameFromInventory(tt.campaign); got != tt.want {
				t.Errorf("gameNameFromInventory(%v) = %q, want %q", tt.campaign, got, tt.want)
			}
		})
	}
}

func TestChannelNameFromInventory(t *testing.T) {
	tests := []struct {
		name     string
		campaign object
		want     string
	}{
		{
			"a campaign with every name",
			object{"channelLogin": "streamer", "channelName": "Streamer", "channelDisplayName": "The Streamer", "channel": object{"login": "other"}, "channels": list{object{"login": "listed"}}},
			"streamer",
		},
		{"a channel name", object{"channelName": "Streamer", "channelDisplayName": "The Streamer"}, "Streamer"},
		{"a channel display name", object{"channelDisplayName": "The Streamer"}, "The Streamer"},
		{"a channel with every name", object{"channel": object{"login": "streamer", "name": "Streamer", "displayName": "The Streamer"}}, "streamer"},
		{"a channel with a padded login", object{"channel": object{"login": " streamer "}}, "streamer"},
		{"a channel with a blank login and a name", object{"channel": object{"login": "  ", "name": "Streamer"}}, "Streamer"},
		{"a channel with only a display name", object{"channel": object{"displayName": "The Streamer"}}, "The Streamer"},
		{"a list of channels", object{"channels": list{object{"login": "streamer"}}, "eligibleChannels": list{object{"login": "eligible"}}}, "streamer"},
		{"a list of eligible channels", object{"eligibleChannels": list{object{"login": "streamer"}}, "allowChannels": list{object{"login": "allow"}}}, "streamer"},
		{"a list of allow channels", object{"allowChannels": list{object{"login": "streamer"}}, "allowedChannels": list{object{"login": "allowed"}}}, "streamer"},
		{"a list of allowed channels", object{"allowedChannels": list{object{"login": "streamer"}}}, "streamer"},
		{"an empty list of channels before a full one", object{"channels": list{}, "allowedChannels": list{object{"login": "streamer"}}}, "streamer"},
		{"an empty campaign", object{}, ""},
		{"no campaign", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := channelNameFromInventory(tt.campaign); got != tt.want {
				t.Errorf("channelNameFromInventory(%v) = %q, want %q", tt.campaign, got, tt.want)
			}
		})
	}
}

func TestChannelNameFromList(t *testing.T) {
	tests := []struct {
		name string
		raw  interface{}
		want string
	}{
		{"a channel with every name", list{object{"login": "streamer", "name": "Streamer", "displayName": "The Streamer"}}, "streamer"},
		{"a channel with a name", list{object{"name": "Streamer", "displayName": "The Streamer"}}, "Streamer"},
		{"a channel with a display name", list{object{"displayName": "The Streamer"}}, "The Streamer"},
		{"a nested channel with every name", list{object{"channel": object{"login": "streamer", "name": "Streamer", "displayName": "The Streamer"}}}, "streamer"},
		{"a nested channel with a padded name", list{object{"channel": object{"name": " Streamer "}}}, "Streamer"},
		{"a nested channel with a display name", list{object{"channel": object{"displayName": "The Streamer"}}}, "The Streamer"},
		{"two channels", list{object{"login": "first"}, object{"login": "second"}}, "first"},
		{"a channel with no names before one with a login", list{object{}, object{"login": "second"}}, "second"},
		{"a string before a channel", list{"noise", object{"login": "second"}}, "second"},
		{"an empty list", list{}, ""},
		{"an object where a list belongs", object{"login": "streamer"}, ""},
		{"nothing", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := channelNameFromList(tt.raw); got != tt.want {
				t.Errorf("channelNameFromList(%v) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDropClaimable(t *testing.T) {
	tests := []struct {
		name     string
		drop     object
		self     object
		current  int
		required int
		claimed  bool
		want     bool
	}{
		{"progress one short of required", object{}, object{}, 179, 180, false, false},
		{"progress at required", object{}, object{}, 180, 180, false, true},
		{"progress past required", object{}, object{}, 181, 180, false, true},
		{"nothing required", object{}, object{}, 0, 0, false, false},
		{"nothing required, some progress", object{}, object{}, 5, 0, false, false},
		{"a claimed drop at required", object{}, object{}, 180, 180, true, false},
		{"a claimed drop flagged claimable", object{}, object{"isClaimable": true}, 180, 180, true, false},
		{"flagged claimable on the viewer's progress, short of required", object{}, object{"isClaimable": true}, 10, 180, false, true},
		{"flagged not claimable on the viewer's progress, at required", object{}, object{"isClaimable": false}, 180, 180, false, false},
		{"the claimable flag", object{}, object{"claimable": true}, 10, 180, false, true},
		{"the canClaim flag", object{}, object{"canClaim": true}, 10, 180, false, true},
		{"the isClaimAvailable flag", object{}, object{"isClaimAvailable": true}, 10, 180, false, true},
		{"the isClaimReady flag", object{}, object{"isClaimReady": true}, 10, 180, false, true},
		{"flagged on the drop", object{"isClaimable": true}, object{}, 10, 180, false, true},
		{"flagged no on the viewer's progress and yes on the drop", object{"isClaimable": true}, object{"isClaimReady": false}, 10, 180, false, false},
		{"a flag that is not a boolean", object{}, object{"isClaimable": "yes"}, 10, 180, false, false},
		{"no drop and no progress, at required", nil, nil, 180, 180, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropClaimable(tt.drop, tt.self, tt.current, tt.required, tt.claimed); got != tt.want {
				t.Errorf("dropClaimable(%v, %v, %d, %d, %t) = %t, want %t", tt.drop, tt.self, tt.current, tt.required, tt.claimed, got, tt.want)
			}
		})
	}
}

func TestDropProgress(t *testing.T) {
	tests := []struct {
		name         string
		drop         object
		self         object
		wantCurrent  int
		wantRequired int
	}{
		{"minutes watched", object{"requiredMinutesWatched": float64(180)}, object{"currentMinutesWatched": float64(63)}, 63, 180},
		{"seconds watched", object{"requiredSecondsWatched": float64(600)}, object{"currentSecondsWatched": float64(30)}, 30, 600},
		{"progress", object{"requiredProgress": float64(5)}, object{"currentProgress": float64(2)}, 2, 5},
		{"amounts", object{"requiredAmount": float64(9)}, object{"currentAmount": float64(4)}, 4, 9},
		{"minutes of zero before seconds", object{"requiredMinutesWatched": float64(0), "requiredSecondsWatched": float64(600)}, object{"currentMinutesWatched": float64(0), "currentSecondsWatched": float64(30)}, 30, 600},
		{"required minutes on the viewer's progress", object{}, object{"currentMinutesWatched": float64(63), "requiredMinutesWatched": float64(180)}, 63, 180},
		{"required seconds on the viewer's progress", object{}, object{"requiredSecondsWatched": float64(600)}, 0, 600},
		{"required on both", object{"requiredMinutesWatched": float64(180)}, object{"requiredMinutesWatched": float64(90)}, 0, 180},
		{"a fraction of a minute", object{"requiredMinutesWatched": 180.9}, object{"currentMinutesWatched": 63.9}, 63, 180},
		{"a number written as text", object{"requiredMinutesWatched": "180"}, object{"currentMinutesWatched": "63"}, 0, 0},
		{"nothing known", object{}, object{}, 0, 0},
		{"no drop and no progress", nil, nil, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, required := dropProgress(tt.drop, tt.self)
			if current != tt.wantCurrent || required != tt.wantRequired {
				t.Errorf("dropProgress(%v, %v) = %d, %d, want %d, %d", tt.drop, tt.self, current, required, tt.wantCurrent, tt.wantRequired)
			}
		})
	}
}

func TestDropStatusesFromInventory(t *testing.T) {
	tests := []struct {
		name string
		inv  object
		want []DropStatus
	}{
		{
			"a drop in progress",
			object{"dropCampaignsInProgress": list{object{
				"name":     "World of Tanks",
				"game":     object{"displayName": "World of Tanks"},
				"channels": list{object{"login": "example_streamer"}},
				"timeBasedDrops": list{object{
					"name":                   "Summer Token Store #12",
					"requiredMinutesWatched": float64(180),
					"self":                   object{"dropInstanceID": "drop-1", "currentMinutesWatched": float64(63), "isClaimed": false},
				}},
			}}},
			[]DropStatus{{DropInstanceID: "drop-1", RewardName: "Summer Token Store #12", CampaignName: "World of Tanks", GameName: "World of Tanks", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180}},
		},
		{
			"a drop watched in full",
			object{"dropCampaignsInProgress": list{object{
				"name": "Summer",
				"timeBasedDrops": list{object{
					"name":                   "Token",
					"requiredMinutesWatched": float64(180),
					"self":                   object{"dropInstanceID": "drop-1", "currentMinutesWatched": float64(180), "isClaimed": false},
				}},
			}}},
			[]DropStatus{{DropInstanceID: "drop-1", RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180, Claimable: true}},
		},
		{
			"a claimed drop",
			object{"dropCampaignsInProgress": list{object{
				"name": "Summer",
				"timeBasedDrops": list{object{
					"name":                   "Token",
					"requiredMinutesWatched": float64(180),
					"self":                   object{"dropInstanceID": "drop-1", "currentMinutesWatched": float64(180), "isClaimed": true},
				}},
			}}},
			[]DropStatus{{DropInstanceID: "drop-1", RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180, Claimed: true}},
		},
		{
			"a drop with nothing but the viewer's empty progress",
			object{"dropCampaignsInProgress": list{object{"timeBasedDrops": list{object{"self": object{}}}}}},
			[]DropStatus{{}},
		},
		{
			"two campaigns with a drop each",
			object{"dropCampaignsInProgress": list{
				object{"name": "Summer", "timeBasedDrops": list{object{"name": "Token", "self": object{"dropInstanceID": "drop-1"}}}},
				object{"name": "Winter", "timeBasedDrops": list{object{"name": "Crate", "self": object{"dropInstanceID": "drop-2"}}}},
			}},
			[]DropStatus{
				{DropInstanceID: "drop-1", RewardName: "Token", CampaignName: "Summer"},
				{DropInstanceID: "drop-2", RewardName: "Crate", CampaignName: "Winter"},
			},
		},
		{
			"a drop the viewer has no progress on, before one with progress",
			object{"dropCampaignsInProgress": list{object{"name": "Summer", "timeBasedDrops": list{
				object{"name": "Token"},
				"noise",
				object{"name": "Crate", "self": object{"dropInstanceID": "drop-2"}},
			}}}},
			[]DropStatus{{DropInstanceID: "drop-2", RewardName: "Crate", CampaignName: "Summer"}},
		},
		{
			"a string before a campaign",
			object{"dropCampaignsInProgress": list{
				"noise",
				object{"name": "Winter", "timeBasedDrops": list{object{"name": "Crate", "self": object{"dropInstanceID": "drop-2"}}}},
			}},
			[]DropStatus{{DropInstanceID: "drop-2", RewardName: "Crate", CampaignName: "Winter"}},
		},
		{"a campaign with no drops", object{"dropCampaignsInProgress": list{object{"name": "Summer"}}}, []DropStatus{}},
		{"no campaigns in progress", object{"dropCampaignsInProgress": nil}, []DropStatus{}},
		{"an empty inventory", object{}, []DropStatus{}},
		{"no inventory", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropStatusesFromInventory(tt.inv); !equal(got, tt.want) {
				t.Errorf("dropStatusesFromInventory(%v) = %#v, want %#v", tt.inv, got, tt.want)
			}
		})
	}
}
