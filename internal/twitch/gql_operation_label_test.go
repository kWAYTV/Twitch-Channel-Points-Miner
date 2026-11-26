package twitch

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func TestOperationLabel(t *testing.T) {
	tests := []struct {
		name        string
		payload     interface{}
		includeNote bool
		want        string
	}{
		{"an operation with a note", twitchapi.GQLPersistedOperation{OperationName: "JoinRaid"}, true, "JoinRaid (join raid)"},
		{"an operation with a note, notes off", twitchapi.GQLPersistedOperation{OperationName: "JoinRaid"}, false, "JoinRaid"},
		{"an operation with no note", twitchapi.GQLPersistedOperation{OperationName: "Unheard"}, true, "Unheard"},
		{"an operation with no name", twitchapi.GQLPersistedOperation{}, true, "unknown"},
		{"no payload", nil, true, "unknown"},
		{
			"a batch of two",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "Inventory"}},
			true,
			"batch[JoinRaid, Inventory]",
		},
		{
			"a batch of three",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "Inventory"}, {OperationName: "RewardList"}},
			true,
			"batch[JoinRaid, Inventory, RewardList]",
		},
		{
			"a batch of four",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "Inventory"}, {OperationName: "RewardList"}, {OperationName: "MakePrediction"}},
			true,
			"batch[JoinRaid, Inventory, RewardList +1]",
		},
		{
			"a batch of the same operation twice",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "JoinRaid"}},
			true,
			"JoinRaid (join raid)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := operationLabel(tt.payload, tt.includeNote); got != tt.want {
				t.Errorf("operationLabel(%v, %t) = %q, want %q", tt.payload, tt.includeNote, got, tt.want)
			}
		})
	}
}

type namedOperation struct {
	OperationName string
}

type unnamedOperation struct {
	Query string
}

type numberedOperation struct {
	OperationName int
}

func TestOperationNames(t *testing.T) {
	var noOperation *twitchapi.GQLPersistedOperation
	var noNamedOperation *namedOperation

	tests := []struct {
		name    string
		payload interface{}
		want    []string
	}{
		{"an operation", twitchapi.GQLPersistedOperation{OperationName: "JoinRaid"}, []string{"JoinRaid"}},
		{"an operation with a padded name", twitchapi.GQLPersistedOperation{OperationName: " JoinRaid "}, []string{"JoinRaid"}},
		{"an operation with a blank name", twitchapi.GQLPersistedOperation{OperationName: "  "}, nil},
		{"a pointer to an operation", &twitchapi.GQLPersistedOperation{OperationName: "JoinRaid"}, []string{"JoinRaid"}},
		{"a nil pointer to an operation", noOperation, nil},
		{
			"a list of operations",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "Inventory"}},
			[]string{"JoinRaid", "Inventory"},
		},
		{
			"a list with the same operation twice",
			[]twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, {OperationName: "Inventory"}, {OperationName: "JoinRaid"}},
			[]string{"JoinRaid", "Inventory"},
		},
		{"an empty list of operations", []twitchapi.GQLPersistedOperation{}, nil},
		{
			"a list of pointers with a nil one",
			[]*twitchapi.GQLPersistedOperation{{OperationName: "JoinRaid"}, nil, {OperationName: "Inventory"}},
			[]string{"JoinRaid", "Inventory"},
		},
		{"a decoded object", map[string]interface{}{"operationName": "JoinRaid"}, []string{"JoinRaid"}},
		{"a decoded object with no name", map[string]interface{}{"query": "{}"}, nil},
		{"a decoded object whose name is a number", map[string]interface{}{"operationName": 7}, nil},
		{
			"a decoded list",
			[]interface{}{map[string]interface{}{"operationName": "JoinRaid"}, "noise", map[string]interface{}{"query": "{}"}, map[string]interface{}{"operationName": "Inventory"}},
			[]string{"JoinRaid", "Inventory"},
		},
		{"another struct with an operation name", namedOperation{OperationName: "Custom"}, []string{"Custom"}},
		{"a pointer to another struct", &namedOperation{OperationName: "Custom"}, []string{"Custom"}},
		{"a nil pointer to another struct", noNamedOperation, nil},
		{"a struct with no operation name", unnamedOperation{Query: "{}"}, nil},
		{"a struct whose operation name is a number", numberedOperation{OperationName: 7}, nil},
		{
			"a list of other structs",
			[]namedOperation{{OperationName: "Custom"}, {OperationName: "Other"}},
			[]string{"Custom", "Other"},
		},
		{
			"a list of pointers to other structs with a nil one",
			[]*namedOperation{{OperationName: "Custom"}, nil, {OperationName: "Other"}},
			[]string{"Custom", "Other"},
		},
		{"a list of structs with no operation name", []unnamedOperation{{Query: "{}"}}, nil},
		{"a list of strings", []string{"JoinRaid"}, nil},
		{"a string", "JoinRaid", nil},
		{"no payload", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := operationNames(tt.payload); !equal(got, tt.want) {
				t.Errorf("operationNames(%v) = %q, want %q", tt.payload, got, tt.want)
			}
		})
	}
}

func TestOperationNote(t *testing.T) {
	tests := []struct {
		operation string
		want      string
	}{
		{"WithIsStreamLiveQuery", "check live"},
		{"VideoPlayerStreamInfoOverlayChannel", "stream info"},
		{"RewardList", "reward list"},
		{"PlaybackAccessToken", "playback token"},
		{"ChannelPointsContext", "points/bonus"},
		{"ClaimCommunityPoints", "claim bonus"},
		{"CommunityMomentCallout_Claim", "claim moment"},
		{"JoinRaid", "join raid"},
		{"DropsPage_ClaimDropRewards", "claim drop"},
		{"ViewerDropsDashboard", "drops dashboard"},
		{"DropCampaignDetails", "drop details"},
		{"DropsHighlightService_AvailableDrops", "available drops"},
		{"Inventory", "inventory"},
		{"GetIDFromLogin", "resolve id"},
		{"ChannelFollows", "followers"},
		{"PersonalSections", "followed streams"},
		{"UserPointsContribution", "community goal"},
		{"ContributeCommunityPointsCommunityGoal", "contribute goal"},
		{"MakePrediction", "make prediction"},
		{"ModViewChannelQuery", "mod view"},
		{"ClipsCards__User", "clips"},
		{"FilterableVideoTower_Videos", "vods"},
		{"Unheard", ""},
		{"joinraid", ""},
	}
	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			if got := operationNote(tt.operation); got != tt.want {
				t.Errorf("operationNote(%q) = %q, want %q", tt.operation, got, tt.want)
			}
		})
	}
}
