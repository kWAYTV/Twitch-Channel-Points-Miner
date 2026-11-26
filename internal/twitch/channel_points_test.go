package twitch

import (
	"errors"
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const (
	pointsVariables = `{"channelLogin":"streamer"}`
	claimVariables  = `{"input":{"channelID":"123456","claimID":"claim-1"}}`

	pointsNoBonus   = `{"data":{"community":{"channel":{"self":{"communityPoints":{"balance":1500,"activeMultipliers":[],"availableClaim":null}},"communityPointsSettings":{"goals":[{"id":"goal-1","title":"New chair","isInStock":true,"pointsContributed":400,"amountNeeded":1000,"perStreamUserMaximumContribution":500,"status":"STARTED"}]}}}}}`
	pointsWithBonus = `{"data":{"community":{"channel":{"self":{"communityPoints":{"balance":1500,"activeMultipliers":[],"availableClaim":{"id":"claim-1"}}},"communityPointsSettings":{"goals":[]}}}}}`
	pointsBonusNoID = `{"data":{"community":{"channel":{"self":{"communityPoints":{"balance":1500,"activeMultipliers":[],"availableClaim":{"id":""}}},"communityPointsSettings":{"goals":[]}}}}}`
	claimSucceeded  = `{"data":{"claimCommunityPoints":{"status":"SUCCESS"}}}`
)

func TestClient_LoadChannelPointsContext(t *testing.T) {
	t.Run("a channel with points", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsNoBonus))
		streamer := newStreamer()
		streamer.ActiveMultipliers = []domain.ActiveMultiplier{{Factor: 2}}

		balance, err := client.LoadChannelPointsContext(streamer)

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if balance != 1500 {
			t.Errorf("balance = %d, want 1500", balance)
		}
		if streamer.ChannelPoints != 1500 {
			t.Errorf("streamer's points = %d, want 1500", streamer.ChannelPoints)
		}
		if streamer.ActiveMultipliers != nil {
			t.Errorf("active multipliers = %v, want none", streamer.ActiveMultipliers)
		}
		if streamer.CommunityGoals != nil {
			t.Errorf("community goals = %v, want none: the setting is off", streamer.CommunityGoals)
		}
		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("a channel with an active multiplier", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(`{"data":{"community":{"channel":{"self":{"communityPoints":{"balance":1500,"activeMultipliers":[{"factor":0.2}],"availableClaim":null}}}}}}`))
		streamer := newStreamer()

		_, err := client.LoadChannelPointsContext(streamer)

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if want := []domain.ActiveMultiplier{{Factor: 0.2}}; !equal(streamer.ActiveMultipliers, want) {
			t.Errorf("active multipliers = %v, want %v", streamer.ActiveMultipliers, want)
		}
	})

	t.Run("community goals on", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsNoBonus))
		fake.twitch.answer("UserPointsContribution", pointsVariables, ok(`{"data":{"user":{"channel":{"self":{"communityPoints":{"goalContributions":[{"goal":{"id":"goal-1"},"userPointsContributedThisStream":100}]}}}}}}`))
		fake.twitch.answer("ContributeCommunityPointsCommunityGoal", "", ok(`{"data":{"contributeCommunityPointsCommunityGoal":{"error":null}}}`))
		streamer := newStreamer()
		streamer.Settings.CommunityGoals = true

		balance, err := client.LoadChannelPointsContext(streamer)

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if balance != 1500 {
			t.Errorf("balance = %d, want 1500, the balance before contributing", balance)
		}
		wantGoals := map[string]*domain.CommunityGoal{
			"goal-1": {ID: "goal-1", Title: "New chair", IsInStock: true, PointsContributed: 400, AmountNeeded: 1000, PerStreamUserMaximumContribution: 500, Status: "STARTED"},
		}
		if !equal(streamer.CommunityGoals, wantGoals) {
			t.Errorf("community goals = %v, want %v", streamer.CommunityGoals, wantGoals)
		}
		if streamer.ChannelPoints != 1100 {
			t.Errorf("streamer's points = %d, want 1100 after contributing 400", streamer.ChannelPoints)
		}
		contribution := regexp.MustCompile(`^\{"input":\{"amount":400,"channelID":"123456","goalID":"goal-1","transactionID":"[0-9a-f]{32}"\}\}$`)
		if got := fake.twitch.sent("ContributeCommunityPointsCommunityGoal"); !matchAll(got, contribution) {
			t.Errorf("contributions sent = %v, want one matching %v", got, contribution)
		}
	})

	t.Run("a missing channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(`{"data":{"community":{"channel":null}}}`))

		balance, err := client.LoadChannelPointsContext(newStreamer())

		if !errors.Is(err, ErrChannelNotFound) {
			t.Errorf("LoadChannelPointsContext error = %v, want it to wrap ErrChannelNotFound", err)
		}
		if want := "channel missing for streamer: channel not found"; errorText(err) != want {
			t.Errorf("LoadChannelPointsContext error = %v, want %s", err, want)
		}
		if balance != 0 {
			t.Errorf("balance = %d, want 0", balance)
		}
	})

	t.Run("a missing channel, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(`{"data":{"community":{"channel":null}}}`))

		_, err := client.LoadChannelPointsContext(newStreamer())

		if want := "channel missing for Streamer1: channel not found"; errorText(err) != want {
			t.Errorf("LoadChannelPointsContext error = %v, want %s", err, want)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, down("connection refused"))
		streamer := newStreamer()
		streamer.ChannelPoints = 700

		balance, err := client.LoadChannelPointsContext(streamer)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("LoadChannelPointsContext error = %v, want %s", err, want)
		}
		if balance != 0 {
			t.Errorf("balance = %d, want 0", balance)
		}
		if streamer.ChannelPoints != 700 {
			t.Errorf("streamer's points = %d, want 700 as before", streamer.ChannelPoints)
		}
	})

	t.Run("a bonus to claim", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsWithBonus))
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		balance, err := client.LoadChannelPointsContext(newStreamer())

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if balance != 1500 {
			t.Errorf("balance = %d, want 1500", balance)
		}
		if want := []string{claimVariables}; !equal(fake.twitch.sent("ClaimCommunityPoints"), want) {
			t.Errorf("claims sent = %v, want %v", fake.twitch.sent("ClaimCommunityPoints"), want)
		}
		wantLines := []string{
			"emoji :gift: Pending bonus detected for streamer (claim claim-1, channel 123456)",
			"print: Claim bonus success for streamer (claim claim-1)",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a bonus to claim, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsWithBonus))
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		_, err := client.LoadChannelPointsContext(newStreamer())

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		wantLines := []string{
			"emoji :gift: Pending bonus detected for Streamer1",
			"print: Claim bonus success for Streamer1",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a bonus whose claim fails", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsWithBonus))
		fake.twitch.answer("ClaimCommunityPoints", "", code(500, "oops"))

		balance, err := client.LoadChannelPointsContext(newStreamer())

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if balance != 1500 {
			t.Errorf("balance = %d, want 1500", balance)
		}
		wantLines := []string{
			"emoji :gift: Pending bonus detected for streamer (claim claim-1, channel 123456)",
			"error: claim bonus on context load for streamer failed: claim bonus status 500: oops",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a bonus whose claim fails, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsWithBonus))
		fake.twitch.answer("ClaimCommunityPoints", "", code(500, "oops"))

		_, err := client.LoadChannelPointsContext(newStreamer())

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		wantLines := []string{
			"emoji :gift: Pending bonus detected for Streamer1",
			"error: claim bonus on context load for Streamer1 failed: claim bonus status 500",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a bonus with no id", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsBonusNoID))
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		balance, err := client.LoadChannelPointsContext(newStreamer())

		if err != nil {
			t.Fatalf("LoadChannelPointsContext error = %v, want none", err)
		}
		if balance != 1500 {
			t.Errorf("balance = %d, want 1500", balance)
		}
		if got := fake.twitch.sent("ClaimCommunityPoints"); len(got) != 0 {
			t.Errorf("claims sent = %v, want none", got)
		}
		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("a bonus with no id, debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsBonusNoID))

		_, _ = client.LoadChannelPointsContext(newStreamer())

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL ChannelPointsContext (points/bonus) | Status 200",
			"debug: availableClaim present but missing id for streamer",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a bonus with no id, debug logging on, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.logger.Debug = true
		fake.twitch.answer("ChannelPointsContext", pointsVariables, ok(pointsBonusNoID))

		_, _ = client.LoadChannelPointsContext(newStreamer())

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL ChannelPointsContext (points/bonus) | Status 200",
			"debug: availableClaim present but missing id for Streamer1",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})
}
