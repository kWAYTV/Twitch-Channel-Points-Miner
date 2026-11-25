package pubsub

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// fakeAccount records the bets it is asked to make.
type fakeAccount struct {
	Account
	bets   []*prediction.Event
	betErr error
}

func (a *fakeAccount) MakePrediction(event *prediction.Event) error {
	a.bets = append(a.bets, event)
	return a.betErr
}

// lineLogger keeps every logged line with its event.
type lineLogger struct {
	stubLogger
	lines []string
}

func (l *lineLogger) Eventf(event events.Event, format string, args ...interface{}) {
	l.lines = append(l.lines, string(event)+": "+fmt.Sprintf(format, args...))
}

func (l *lineLogger) EmojiEventf(_ string, event events.Event, format string, args ...interface{}) {
	l.lines = append(l.lines, string(event)+": "+fmt.Sprintf(format, args...))
}

func (l *lineLogger) ErrorEventf(event events.Event, format string, args ...interface{}) {
	l.lines = append(l.lines, string(event)+": "+fmt.Sprintf(format, args...))
}

type gain struct {
	earned int
	reason string
}

type betFixture struct {
	client   *Client
	account  *fakeAccount
	logger   *lineLogger
	streamer *domain.Streamer
	event    *prediction.Event
	gains    []gain
}

func newBetFixture(t *testing.T, points int, adjust func(*domain.BetSettings)) *betFixture {
	t.Helper()
	settings := domain.BetSettings{Strategy: domain.StrategyMostVoted}
	settings.Default()
	if adjust != nil {
		adjust(&settings)
	}
	f := &betFixture{account: &fakeAccount{}, logger: &lineLogger{}}
	f.streamer = &domain.Streamer{Username: "tester", ChannelPoints: points, Settings: domain.StreamerSettings{Bet: settings}}
	f.event = prediction.NewEvent(f.streamer, map[string]interface{}{
		"id":     "ev-1",
		"status": "ACTIVE",
		"title":  "who wins",
		"outcomes": []interface{}{
			map[string]interface{}{"id": "a", "title": "A", "color": "blue", "total_users": 10, "total_points": 100},
			map[string]interface{}{"id": "b", "title": "B", "color": "pink", "total_users": 5, "total_points": 50},
		},
	})
	if f.event == nil {
		t.Fatalf("expected event")
	}
	f.client = &Client{
		account:     f.account,
		logger:      f.logger,
		predictions: map[string]*prediction.Event{f.event.EventID: f.event},
		onGain: func(_ *domain.Streamer, earned int, reason string, _ int) {
			f.gains = append(f.gains, gain{earned, reason})
		},
	}
	return f
}

func (f *betFixture) tracked() bool {
	_, ok := f.client.predictions[f.event.EventID]
	return ok
}

func (f *betFixture) expectSkipped(t *testing.T, wantLine string) {
	t.Helper()
	if f.tracked() || f.event.ResultType != "SKIPPED" {
		t.Fatalf("a skipped bet should stop being tracked and be marked SKIPPED, got tracked=%t result=%q", f.tracked(), f.event.ResultType)
	}
	if len(f.account.bets) != 0 || len(f.gains) != 0 || f.event.BetPlaced {
		t.Fatalf("a skipped bet must not be placed: bets=%d gains=%v placed=%t", len(f.account.bets), f.gains, f.event.BetPlaced)
	}
	if len(f.logger.lines) != 1 || f.logger.lines[0] != wantLine {
		t.Fatalf("logged %q, want %q", f.logger.lines, wantLine)
	}
}

func TestPlacePredictionIgnoresUnknownEvent(t *testing.T) {
	f := newBetFixture(t, 1000, nil)
	f.client.placePrediction("some-other-event")
	if !f.tracked() || len(f.account.bets) != 0 || len(f.logger.lines) != 0 {
		t.Fatalf("an unknown event should change nothing")
	}
}

func TestPlacePredictionSkipsEventThatIsNotActive(t *testing.T) {
	f := newBetFixture(t, 1000, nil)
	f.event.Status = "LOCKED"
	f.client.placePrediction(f.event.EventID)
	f.expectSkipped(t, "BET_GENERAL: Skip bet for tester: event status is LOCKED")
}

func TestPlacePredictionSkipsBelowMinimumPoints(t *testing.T) {
	f := newBetFixture(t, 1000, func(b *domain.BetSettings) { minimum := 5000; b.MinimumPoints = &minimum })
	f.client.placePrediction(f.event.EventID)
	f.expectSkipped(t, "BET_FILTERS: Skip bet for tester: balance 1000 <= minimum_points 5000")
}

func TestPlacePredictionSkipsStakeBelowTwitchMinimum(t *testing.T) {
	poor := newBetFixture(t, 5, nil)
	poor.client.placePrediction(poor.event.EventID)
	poor.expectSkipped(t, "BET_FILTERS: Skip bet for tester: balance 5 below Twitch minimum 10")

	capped := newBetFixture(t, 1000, func(b *domain.BetSettings) { limit := 5; b.MaxPoints = &limit })
	capped.client.placePrediction(capped.event.EventID)
	capped.expectSkipped(t, "BET_FILTERS: Skip bet for tester: max_points 5 below Twitch minimum 10")

}

func TestPlacePredictionRaisesSmallStakeToTwitchMinimum(t *testing.T) {
	f := newBetFixture(t, 100, func(b *domain.BetSettings) { percentage := 1; b.Percentage = &percentage })
	f.client.placePrediction(f.event.EventID)
	if len(f.account.bets) != 1 || f.event.Decision.Amount != 10 {
		t.Fatalf("1%% of 100 points should be raised to the minimum stake of 10, got %d bets with stake %d", len(f.account.bets), f.event.Decision.Amount)
	}
}

func TestPlacePredictionReportsFailedBet(t *testing.T) {
	f := newBetFixture(t, 1000, nil)
	f.account.betErr = errors.New("twitch said no")
	f.client.placePrediction(f.event.EventID)
	if f.tracked() || f.event.ResultType != "ERROR" || f.event.BetPlaced || len(f.gains) != 0 {
		t.Fatalf("a failed bet should stop being tracked and be marked ERROR: tracked=%t result=%q placed=%t gains=%v", f.tracked(), f.event.ResultType, f.event.BetPlaced, f.gains)
	}
	if want := "BET_FAILED: prediction tester: twitch said no"; len(f.logger.lines) != 1 || f.logger.lines[0] != want {
		t.Fatalf("logged %q, want %q", f.logger.lines, want)
	}
}

func TestPlacePredictionPlacesBetAndDeductsStake(t *testing.T) {
	f := newBetFixture(t, 1000, nil)
	f.client.placePrediction(f.event.EventID)

	if len(f.account.bets) != 1 || f.account.bets[0] != f.event {
		t.Fatalf("the bet should be made once, got %d", len(f.account.bets))
	}
	stake := f.event.Decision.Amount
	if stake < 10 {
		t.Fatalf("stake should be at least Twitch's minimum, got %d", stake)
	}
	if !f.tracked() || !f.event.BetPlaced || !f.event.BetConfirmed || f.event.ResultType != "" {
		t.Fatalf("a placed bet stays tracked until its result: tracked=%t placed=%t confirmed=%t result=%q", f.tracked(), f.event.BetPlaced, f.event.BetConfirmed, f.event.ResultType)
	}
	if len(f.gains) != 1 || f.gains[0] != (gain{-stake, "PREDICTION"}) {
		t.Fatalf("the stake should be deducted once, got %v", f.gains)
	}
	if len(f.logger.lines) != 1 || !strings.HasPrefix(f.logger.lines[0], fmt.Sprintf("BET_GENERAL: Place %d points on: ", stake)) || !strings.HasSuffix(f.logger.lines[0], " for tester") {
		t.Fatalf("unexpected log %q", f.logger.lines)
	}
}

func TestPlacePredictionStakeDeductionIsOptional(t *testing.T) {
	kept := newBetFixture(t, 1000, func(b *domain.BetSettings) { deduct := false; b.DeductStakeOnPlace = &deduct })
	kept.client.placePrediction(kept.event.EventID)
	if len(kept.gains) != 0 || kept.streamer.ChannelPoints != 1000 || !kept.event.BetPlaced {
		t.Fatalf("with deduct_stake_on_place off the balance is left alone: gains=%v points=%d", kept.gains, kept.streamer.ChannelPoints)
	}

	direct := newBetFixture(t, 1000, nil)
	direct.client.onGain = nil
	direct.client.placePrediction(direct.event.EventID)
	if want := 1000 - direct.event.Decision.Amount; direct.streamer.ChannelPoints != want {
		t.Fatalf("without a gain handler the stake comes straight off the balance: got %d, want %d", direct.streamer.ChannelPoints, want)
	}
}
