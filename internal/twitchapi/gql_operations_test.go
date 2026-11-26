package twitchapi

import (
	"reflect"
	"regexp"
	"testing"
)

func TestClonePersistedOperation(t *testing.T) {
	tests := []struct {
		name string
		op   GQLPersistedOperation
	}{
		{"an operation with no variables", GQLPersistedOperation{OperationName: "JoinRaid"}},
		{"an operation with empty variables", GQLPersistedOperation{OperationName: "JoinRaid", Variables: map[string]interface{}{}}},
		{
			"an operation with a hash and plain variables",
			GQLPersistedOperation{
				OperationName: "ChannelFollows",
				Variables:     map[string]interface{}{"limit": 100, "order": "ASC", "live": true, "cursor": nil},
				Extensions:    GQLPersistedExtensions{PersistedQuery: GQLPersistedQuery{Version: 1, Sha256Hash: "eecf"}},
			},
		},
		{
			"an operation with nested variables",
			GQLPersistedOperation{
				OperationName: "PersonalSections",
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"sectionInputs":         []string{"FOLLOWED_SECTION"},
						"recommendationContext": map[string]interface{}{"platform": "web"},
						"pages":                 []interface{}{map[string]interface{}{"cursor": "c1"}, "c2"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClonePersistedOperation(tt.op); !reflect.DeepEqual(got, tt.op) {
				t.Errorf("ClonePersistedOperation(%+v) = %+v, want an equal operation", tt.op, got)
			}
		})
	}
}

// A copy shares nothing with the operation it was made from.
func TestClonePersistedOperation_Independence(t *testing.T) {
	t.Run("a variable set on the copy", func(t *testing.T) {
		original := GQLPersistedOperation{Variables: map[string]interface{}{"limit": 100}}

		ClonePersistedOperation(original).Variables["limit"] = 5

		if want := (map[string]interface{}{"limit": 100}); !reflect.DeepEqual(original.Variables, want) {
			t.Errorf("original variables = %v, want %v", original.Variables, want)
		}
	})

	t.Run("a nested variable set on the copy", func(t *testing.T) {
		original := GQLPersistedOperation{Variables: map[string]interface{}{"criteria": map[string]interface{}{"filter": "ALL_TIME"}}}

		ClonePersistedOperation(original).Variables["criteria"].(map[string]interface{})["filter"] = "LAST_DAY"

		if want := (map[string]interface{}{"criteria": map[string]interface{}{"filter": "ALL_TIME"}}); !reflect.DeepEqual(original.Variables, want) {
			t.Errorf("original variables = %v, want %v", original.Variables, want)
		}
	})

	t.Run("an entry of a list of strings set on the copy", func(t *testing.T) {
		original := GQLPersistedOperation{Variables: map[string]interface{}{"sectionInputs": []string{"FOLLOWED_SECTION"}}}

		ClonePersistedOperation(original).Variables["sectionInputs"].([]string)[0] = "RECOMMENDED_SECTION"

		if want := (map[string]interface{}{"sectionInputs": []string{"FOLLOWED_SECTION"}}); !reflect.DeepEqual(original.Variables, want) {
			t.Errorf("original variables = %v, want %v", original.Variables, want)
		}
	})

	t.Run("an object in a list set on the copy", func(t *testing.T) {
		original := GQLPersistedOperation{Variables: map[string]interface{}{"pages": []interface{}{map[string]interface{}{"cursor": "c1"}}}}

		ClonePersistedOperation(original).Variables["pages"].([]interface{})[0].(map[string]interface{})["cursor"] = "c2"

		if want := (map[string]interface{}{"pages": []interface{}{map[string]interface{}{"cursor": "c1"}}}); !reflect.DeepEqual(original.Variables, want) {
			t.Errorf("original variables = %v, want %v", original.Variables, want)
		}
	})

	t.Run("an entry of a list of numbers set on the copy (today: shared)", func(t *testing.T) {
		original := GQLPersistedOperation{Variables: map[string]interface{}{"ids": []int{1}}}

		ClonePersistedOperation(original).Variables["ids"].([]int)[0] = 2

		if want := (map[string]interface{}{"ids": []int{2}}); !reflect.DeepEqual(original.Variables, want) {
			t.Errorf("original variables = %v, want %v", original.Variables, want)
		}
	})
}

func TestGQLOperations(t *testing.T) {
	sha256Hash := regexp.MustCompile(`^[0-9a-f]{64}$`)
	operations := map[string]GQLPersistedOperation{
		"WithIsStreamLiveQuery":                  GQLOperations.WithIsStreamLiveQuery,
		"PlaybackAccessToken":                    GQLOperations.PlaybackAccessToken,
		"VideoPlayerStreamInfoOverlay":           GQLOperations.VideoPlayerStreamInfoOverlay,
		"RewardList":                             GQLOperations.RewardList,
		"ClaimCommunityPoints":                   GQLOperations.ClaimCommunityPoints,
		"CommunityMomentCalloutClaim":            GQLOperations.CommunityMomentCalloutClaim,
		"DropsPageClaimDropRewards":              GQLOperations.DropsPageClaimDropRewards,
		"ChannelPointsContext":                   GQLOperations.ChannelPointsContext,
		"JoinRaid":                               GQLOperations.JoinRaid,
		"ModViewChannelQuery":                    GQLOperations.ModViewChannelQuery,
		"Inventory":                              GQLOperations.Inventory,
		"MakePrediction":                         GQLOperations.MakePrediction,
		"ViewerDropsDashboard":                   GQLOperations.ViewerDropsDashboard,
		"DropCampaignDetails":                    GQLOperations.DropCampaignDetails,
		"DropsHighlightServiceAvailable":         GQLOperations.DropsHighlightServiceAvailable,
		"GetIDFromLogin":                         GQLOperations.GetIDFromLogin,
		"PersonalSections":                       GQLOperations.PersonalSections[0],
		"ChannelFollows":                         GQLOperations.ChannelFollows,
		"UserPointsContribution":                 GQLOperations.UserPointsContribution,
		"ContributeCommunityPointsCommunityGoal": GQLOperations.ContributeCommunityPointsCommunityGoal,
		"ClipsCardsUser":                         GQLOperations.ClipsCardsUser,
		"FilterableVideoTowerVideos":             GQLOperations.FilterableVideoTowerVideos,
	}
	for name, op := range operations {
		t.Run(name, func(t *testing.T) {
			if op.OperationName == "" {
				t.Errorf("operation name is empty")
			}
			if op.Extensions.PersistedQuery.Version != 1 {
				t.Errorf("persisted query version = %d, want 1", op.Extensions.PersistedQuery.Version)
			}
			if hash := op.Extensions.PersistedQuery.Sha256Hash; !sha256Hash.MatchString(hash) {
				t.Errorf("persisted query hash = %q, want 64 hex digits", hash)
			}
		})
	}
}
