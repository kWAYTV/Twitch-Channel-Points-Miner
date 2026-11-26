package domain

import (
	"reflect"
	"testing"
)

func TestCommunityGoal_AmountLeft(t *testing.T) {
	for _, row := range []struct {
		name string
		goal CommunityGoal
		want int
	}{
		{"nothing contributed", CommunityGoal{AmountNeeded: 1000}, 1000},
		{"partly contributed", CommunityGoal{AmountNeeded: 1000, PointsContributed: 400}, 600},
		{"fully contributed", CommunityGoal{AmountNeeded: 1000, PointsContributed: 1000}, 0},
		{"over-contributed", CommunityGoal{AmountNeeded: 1000, PointsContributed: 1200}, -200},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := row.goal.AmountLeft(); got != row.want {
				t.Errorf("AmountLeft() = %d, want %d", got, row.want)
			}
		})
	}
}

func TestNewCommunityGoalFromGQL(t *testing.T) {
	t.Run("a goal with every field", func(t *testing.T) {
		raw := map[string]interface{}{
			"id":                               "goal-1",
			"title":                            "New emotes",
			"isInStock":                        true,
			"pointsContributed":                float64(400),
			"amountNeeded":                     float64(1000),
			"perStreamUserMaximumContribution": float64(250),
			"status":                           "STARTED",
		}
		want := &CommunityGoal{
			ID:                               "goal-1",
			Title:                            "New emotes",
			IsInStock:                        true,
			PointsContributed:                400,
			AmountNeeded:                     1000,
			PerStreamUserMaximumContribution: 250,
			Status:                           "STARTED",
		}
		if got := NewCommunityGoalFromGQL(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromGQL() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty goal", func(t *testing.T) {
		raw := map[string]interface{}{}
		want := &CommunityGoal{}
		if got := NewCommunityGoalFromGQL(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromGQL() = %+v, want %+v", got, want)
		}
	})

	t.Run("a goal with the PubSub field names", func(t *testing.T) {
		raw := map[string]interface{}{
			"id":                                   "goal-1",
			"is_in_stock":                          true,
			"points_contributed":                   float64(400),
			"goal_amount":                          float64(1000),
			"per_stream_maximum_user_contribution": float64(250),
		}
		want := &CommunityGoal{ID: "goal-1"}
		if got := NewCommunityGoalFromGQL(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromGQL() = %+v, want %+v", got, want)
		}
	})

	t.Run("no goal", func(t *testing.T) {
		if got := NewCommunityGoalFromGQL(nil); got != nil {
			t.Errorf("NewCommunityGoalFromGQL(nil) = %+v, want nil", got)
		}
	})
}

func TestNewCommunityGoalFromPubSub(t *testing.T) {
	t.Run("a goal with every field", func(t *testing.T) {
		raw := map[string]interface{}{
			"id":                                   "goal-1",
			"title":                                "New emotes",
			"is_in_stock":                          true,
			"points_contributed":                   float64(400),
			"goal_amount":                          float64(1000),
			"per_stream_maximum_user_contribution": float64(250),
			"status":                               "STARTED",
		}
		want := &CommunityGoal{
			ID:                               "goal-1",
			Title:                            "New emotes",
			IsInStock:                        true,
			PointsContributed:                400,
			AmountNeeded:                     1000,
			PerStreamUserMaximumContribution: 250,
			Status:                           "STARTED",
		}
		if got := NewCommunityGoalFromPubSub(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromPubSub() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty goal", func(t *testing.T) {
		raw := map[string]interface{}{}
		want := &CommunityGoal{}
		if got := NewCommunityGoalFromPubSub(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromPubSub() = %+v, want %+v", got, want)
		}
	})

	t.Run("a goal with the GQL field names", func(t *testing.T) {
		raw := map[string]interface{}{
			"id":                               "goal-1",
			"isInStock":                        true,
			"pointsContributed":                float64(400),
			"amountNeeded":                     float64(1000),
			"perStreamUserMaximumContribution": float64(250),
		}
		want := &CommunityGoal{ID: "goal-1"}
		if got := NewCommunityGoalFromPubSub(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewCommunityGoalFromPubSub() = %+v, want %+v", got, want)
		}
	})

	t.Run("no goal", func(t *testing.T) {
		if got := NewCommunityGoalFromPubSub(nil); got != nil {
			t.Errorf("NewCommunityGoalFromPubSub(nil) = %+v, want nil", got)
		}
	})
}
