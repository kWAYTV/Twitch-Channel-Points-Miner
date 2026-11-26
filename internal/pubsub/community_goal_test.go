package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func goalContributor() *domain.Streamer {
	return &domain.Streamer{
		Username:  "tester",
		ChannelID: "123",
		Settings:  domain.StreamerSettings{CommunityGoals: true},
	}
}

func TestClient_processCommunityPointChannel(t *testing.T) {
	t.Run("a goal created", func(t *testing.T) {
		streamer := goalContributor()
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		err := h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-created","data":{"community_goal":{"id":"goal-1","title":"New emote","is_in_stock":true,"points_contributed":100,"goal_amount":1000,"per_stream_maximum_user_contribution":500,"status":"STARTED"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		wantGoals := map[string]*domain.CommunityGoal{
			"goal-1": {
				ID:                               "goal-1",
				Title:                            "New emote",
				IsInStock:                        true,
				PointsContributed:                100,
				AmountNeeded:                     1000,
				PerStreamUserMaximumContribution: 500,
				Status:                           "STARTED",
			},
		}
		if !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got, want := h.effects(), (effects{GoalContributions: []string{"tester"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a goal updated", func(t *testing.T) {
		streamer := goalContributor()
		streamer.CommunityGoals = map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New emote", PointsContributed: 100, AmountNeeded: 1000, Status: "STARTED"},
			"goal-2": {ID: "goal-2", Title: "Movie night", Status: "STARTED"},
		}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-updated","data":{"community_goal":{"id":"goal-1","title":"New emote","points_contributed":1000,"goal_amount":1000,"status":"ENDED"}}}`)

		wantGoals := map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New emote", PointsContributed: 1000, AmountNeeded: 1000, Status: "ENDED"},
			"goal-2": {ID: "goal-2", Title: "Movie night", Status: "STARTED"},
		}
		if !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got, want := h.account.GoalContributions, []string{"tester"}; !reflect.DeepEqual(got, want) {
			t.Errorf("goal contributions = %q, want %q", got, want)
		}
	})

	t.Run("a goal created without an ID", func(t *testing.T) {
		streamer := goalContributor()
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-created","data":{"community_goal":{"title":"New emote"}}}`)

		if wantGoals := map[string]*domain.CommunityGoal{}; !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got, want := h.account.GoalContributions, []string{"tester"}; !reflect.DeepEqual(got, want) {
			t.Errorf("goal contributions = %q, want %q", got, want)
		}
	})

	t.Run("a goal created with nothing about it", func(t *testing.T) {
		streamer := goalContributor()
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-created"}`)

		if wantGoals := map[string]*domain.CommunityGoal{}; !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got, want := h.account.GoalContributions, []string{"tester"}; !reflect.DeepEqual(got, want) {
			t.Errorf("goal contributions = %q, want %q", got, want)
		}
	})

	t.Run("a goal deleted", func(t *testing.T) {
		streamer := goalContributor()
		streamer.CommunityGoals = map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New emote", Status: "STARTED"},
			"goal-2": {ID: "goal-2", Title: "Movie night", Status: "STARTED"},
		}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-deleted","data":{"community_goal":{"id":"goal-1"}}}`)

		wantGoals := map[string]*domain.CommunityGoal{
			"goal-2": {ID: "goal-2", Title: "Movie night", Status: "STARTED"},
		}
		if !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a goal deleted without an ID", func(t *testing.T) {
		streamer := goalContributor()
		streamer.CommunityGoals = map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New emote", Status: "STARTED"},
		}
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-deleted","data":{"community_goal":{}}}`)

		wantGoals := map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New emote", Status: "STARTED"},
		}
		if !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
	})

	t.Run("a type it doesn't know", func(t *testing.T) {
		streamer := goalContributor()
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"custom-reward-created"}`)

		if wantGoals := map[string]*domain.CommunityGoal{}; !reflect.DeepEqual(streamer.CommunityGoals, wantGoals) {
			t.Errorf("goals = %+v, want %+v", streamer.CommunityGoals, wantGoals)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("community_goals off", func(t *testing.T) {
		streamer := goalContributor()
		streamer.Settings.CommunityGoals = false
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "community-points-channel-v1.123", `{"type":"community-goal-created","data":{"community_goal":{"id":"goal-1","title":"New emote"}}}`)

		if streamer.CommunityGoals != nil {
			t.Errorf("goals = %+v, want none", streamer.CommunityGoals)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{goalContributor()}})

		h.receive(t, "community-points-channel-v1.456", `{"type":"community-goal-created","data":{"community_goal":{"id":"goal-1","title":"New emote"}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}
