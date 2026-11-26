package twitch

import (
	"encoding/json"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestGameToInterfaceMap(t *testing.T) {
	tests := []struct {
		name string
		game *gqlGame
		want map[string]interface{}
	}{
		{
			"a game with every field",
			&gqlGame{ID: "509658", Name: "Just Chatting", DisplayName: "Just Chatting!"},
			map[string]interface{}{"id": "509658", "name": "Just Chatting", "displayName": "Just Chatting!"},
		},
		{"a game with only a name", &gqlGame{Name: "Just Chatting"}, map[string]interface{}{"name": "Just Chatting"}},
		{"a game with every field empty", &gqlGame{}, map[string]interface{}{}},
		{"no game", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gameToInterfaceMap(tt.game); !equal(got, tt.want) {
				t.Errorf("gameToInterfaceMap(%+v) = %#v, want %#v", tt.game, got, tt.want)
			}
		})
	}
}

func TestTagsToInterfaceMaps(t *testing.T) {
	tests := []struct {
		name string
		tags []gqlID
		want []map[string]interface{}
	}{
		{"two tags", []gqlID{{ID: "tag-1"}, {ID: "tag-2"}}, []map[string]interface{}{{"id": "tag-1"}, {"id": "tag-2"}}},
		{"a tag with no id among others", []gqlID{{ID: "tag-1"}, {}, {ID: "tag-2"}}, []map[string]interface{}{{"id": "tag-1"}, {"id": "tag-2"}}},
		{"only a tag with no id", []gqlID{{}}, []map[string]interface{}{}},
		{"an empty list", []gqlID{}, nil},
		{"no list", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagsToInterfaceMaps(tt.tags); !equal(got, tt.want) {
				t.Errorf("tagsToInterfaceMaps(%+v) = %#v, want %#v", tt.tags, got, tt.want)
			}
		})
	}
}

func TestParseCommunityGoalsJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]*domain.CommunityGoal
	}{
		{
			"a goal with every field",
			`[{"id":"goal-1","title":"New chair","isInStock":true,"pointsContributed":400,"amountNeeded":1000,"perStreamUserMaximumContribution":500,"status":"STARTED"}]`,
			map[string]*domain.CommunityGoal{
				"goal-1": {ID: "goal-1", Title: "New chair", IsInStock: true, PointsContributed: 400, AmountNeeded: 1000, PerStreamUserMaximumContribution: 500, Status: "STARTED"},
			},
		},
		{
			"a goal with only an id",
			`[{"id":"goal-1"}]`,
			map[string]*domain.CommunityGoal{"goal-1": {ID: "goal-1"}},
		},
		{
			"two goals",
			`[{"id":"goal-1","status":"STARTED"},{"id":"goal-2","status":"ENDED"}]`,
			map[string]*domain.CommunityGoal{
				"goal-1": {ID: "goal-1", Status: "STARTED"},
				"goal-2": {ID: "goal-2", Status: "ENDED"},
			},
		},
		{"a goal with no id", `[{"title":"New chair"}]`, map[string]*domain.CommunityGoal{}},
		{"a null goal", `[null]`, map[string]*domain.CommunityGoal{}},
		{"an empty list", `[]`, map[string]*domain.CommunityGoal{}},
		{"null", `null`, nil},
		{"nothing", ``, nil},
		{"an object where a list belongs", `{"id":"goal-1"}`, nil},
		{"broken JSON", `[{"id":`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseCommunityGoalsJSON(json.RawMessage(tt.raw)); !equal(got, tt.want) {
				t.Errorf("parseCommunityGoalsJSON(%s) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}
