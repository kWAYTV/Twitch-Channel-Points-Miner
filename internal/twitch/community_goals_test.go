package twitch

import (
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const (
	contributionsOf100 = `{"data":{"user":{"channel":{"self":{"communityPoints":{"goalContributions":[{"goal":{"id":"goal-1"},"userPointsContributedThisStream":100}]}}}}}}`
	contributionTaken  = `{"data":{"contributeCommunityPointsCommunityGoal":{"error":null}}}`
)

// newGoalStreamer has 250 points and one started goal: 600 points short, of which
// this viewer may still give 400 after the 100 the fake reports as given.
func newGoalStreamer() *domain.Streamer {
	streamer := newStreamer()
	streamer.Settings.CommunityGoals = true
	streamer.ChannelPoints = 250
	streamer.CommunityGoals = map[string]*domain.CommunityGoal{
		"goal-1": {ID: "goal-1", Title: "New chair", IsInStock: true, PointsContributed: 400, AmountNeeded: 1000, PerStreamUserMaximumContribution: 500, Status: "STARTED"},
	}
	return streamer
}

func contributionOf(amount string) *regexp.Regexp {
	return regexp.MustCompile(`^\{"input":\{"amount":` + amount + `,"channelID":"123456","goalID":"goal-1","transactionID":"[0-9a-f]{32}"\}\}$`)
}

func TestClient_ContributeToCommunityGoals(t *testing.T) {
	t.Run("a started goal", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(contributionsOf100))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); !matchAll(got, contributionOf("250")) {
			t.Errorf("contributions sent = %v, want one matching %v", got, contributionOf("250"))
		}
		if streamer.ChannelPoints != 0 {
			t.Errorf("streamer's points = %d, want 0 after giving all 250", streamer.ChannelPoints)
		}
	})

	t.Run("a goal the viewer has given the most to", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(`{"data":{"user":{"channel":{"self":{"communityPoints":{"goalContributions":[{"goal":{"id":"goal-1"},"userPointsContributedThisStream":500}]}}}}}}`))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
		if streamer.ChannelPoints != 250 {
			t.Errorf("streamer's points = %d, want 250 as before", streamer.ChannelPoints)
		}
	})

	t.Run("a goal that has ended", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(contributionsOf100))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()
		streamer.CommunityGoals["goal-1"].Status = "ENDED"

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("community goals off", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(contributionsOf100))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()
		streamer.Settings.CommunityGoals = false

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("a streamer with no goals", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(contributionsOf100))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()
		streamer.CommunityGoals = nil

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		client, fake := setup(t)

		client.ContributeToCommunityGoals(nil)

		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("a contribution to a goal the streamer does not have", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(`{"data":{"user":{"channel":{"self":{"communityPoints":{"goalContributions":[{"goal":{"id":"goal-9"},"userPointsContributedThisStream":0}]}}}}}}`))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("contributions that name no goal before one that does", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(`{"data":{"user":{"channel":{"self":{"communityPoints":{"goalContributions":["noise",{"userPointsContributedThisStream":0},{"goal":{"id":"goal-1"},"userPointsContributedThisStream":100}]}}}}}}`))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); !matchAll(got, contributionOf("250")) {
			t.Errorf("contributions sent = %v, want one matching %v", got, contributionOf("250"))
		}
	})

	t.Run("an answer with no contributions", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(`{"data":{"user":{"channel":null}}}`))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, down("connection refused"))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); len(got) != 0 {
			t.Errorf("contributions sent = %v, want none", got)
		}
	})

	t.Run("a contribution Twitch refuses", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(contributionsOf100))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(`{"data":{"contributeCommunityPointsCommunityGoal":{"error":"GOAL_ENDED"}}}`))
		streamer := newGoalStreamer()

		client.ContributeToCommunityGoals(streamer)

		if streamer.ChannelPoints != 250 {
			t.Errorf("streamer's points = %d, want 250 as before", streamer.ChannelPoints)
		}
	})
}

func TestClient_ContributeToCommunityGoal(t *testing.T) {
	t.Run("a contribution", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 100)

		if err != nil {
			t.Fatalf("ContributeToCommunityGoal error = %v, want none", err)
		}
		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); !matchAll(got, contributionOf("100")) {
			t.Errorf("contributions sent = %v, want one matching %v", got, contributionOf("100"))
		}
		if streamer.ChannelPoints != 150 {
			t.Errorf("streamer's points = %d, want 150", streamer.ChannelPoints)
		}
	})

	t.Run("a contribution of more points than the streamer shows", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(contributionTaken))
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 300)

		if err != nil {
			t.Fatalf("ContributeToCommunityGoal error = %v, want none", err)
		}
		if streamer.ChannelPoints != 0 {
			t.Errorf("streamer's points = %d, want 0", streamer.ChannelPoints)
		}
	})

	t.Run("a contribution of nothing", func(t *testing.T) {
		client, fake := setup(t)
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 0)

		if err != nil {
			t.Fatalf("ContributeToCommunityGoal error = %v, want none", err)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("a negative contribution", func(t *testing.T) {
		client, fake := setup(t)
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", -1)

		if err != nil {
			t.Fatalf("ContributeToCommunityGoal error = %v, want none", err)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("no goal id", func(t *testing.T) {
		client, fake := setup(t)
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "", "New chair", 100)

		if err != nil {
			t.Fatalf("ContributeToCommunityGoal error = %v, want none", err)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", down("connection refused"))
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 100)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ContributeToCommunityGoal error = %v, want %s", err, want)
		}
		if streamer.ChannelPoints != 250 {
			t.Errorf("streamer's points = %d, want 250 as before", streamer.ChannelPoints)
		}
	})

	t.Run("a contribution Twitch refuses", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(`{"data":{"contributeCommunityPointsCommunityGoal":{"error":"GOAL_ENDED"}}}`))
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 100)

		if want := "unable to contribute to New chair: GOAL_ENDED"; errorText(err) != want {
			t.Errorf("ContributeToCommunityGoal error = %v, want %s", err, want)
		}
		if streamer.ChannelPoints != 250 {
			t.Errorf("streamer's points = %d, want 250 as before", streamer.ChannelPoints)
		}
	})

	t.Run("a refusal sent as an object", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(`{"data":{"contributeCommunityPointsCommunityGoal":{"error":{"code":"GOAL_ENDED"}}}}`))
		streamer := newGoalStreamer()

		err := client.ContributeToCommunityGoal(streamer, "goal-1", "New chair", 100)

		if err == nil {
			t.Errorf("ContributeToCommunityGoal error = none, want the refusal")
		}
		if streamer.ChannelPoints != 250 {
			t.Errorf("streamer's points = %d, want 250 as before", streamer.ChannelPoints)
		}
	})
}

func TestHasActiveCommunityGoal(t *testing.T) {
	tests := []struct {
		name  string
		goals map[string]*domain.CommunityGoal
		want  bool
	}{
		{"a started goal in stock", map[string]*domain.CommunityGoal{"goal-1": {Status: "STARTED", IsInStock: true}}, true},
		{"a started goal out of stock", map[string]*domain.CommunityGoal{"goal-1": {Status: "STARTED", IsInStock: false}}, false},
		{"an ended goal in stock", map[string]*domain.CommunityGoal{"goal-1": {Status: "ENDED", IsInStock: true}}, false},
		{"a goal started in lower case", map[string]*domain.CommunityGoal{"goal-1": {Status: "started", IsInStock: true}}, false},
		{
			"an ended goal and a started one",
			map[string]*domain.CommunityGoal{"goal-1": {Status: "ENDED", IsInStock: true}, "goal-2": {Status: "STARTED", IsInStock: true}},
			true,
		},
		{"a nil goal", map[string]*domain.CommunityGoal{"goal-1": nil}, false},
		{"no goals", map[string]*domain.CommunityGoal{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasActiveCommunityGoal(tt.goals); got != tt.want {
				t.Errorf("hasActiveCommunityGoal(%v) = %t, want %t", tt.goals, got, tt.want)
			}
		})
	}
}
