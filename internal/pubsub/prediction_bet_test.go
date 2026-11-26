package pubsub

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_placePrediction(t *testing.T) {
	t.Run("a bet that comes due", func(t *testing.T) {
		streamer := bettor()
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}})

		h.timers[0].fire()

		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Bets:   []bet{{"ev-1", "yes", 50}},
			Gains:  []gain{{"tester", -50, "PREDICTION", 0}},
			Lines:  []string{betStartLine, betPlacedLine},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
		if got, want := streamer.ChannelPoints, 1000; got != want {
			t.Errorf("balance = %d, want %d", got, want)
		}
	})

	t.Run("a bet that comes due, logs anonymised", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}, anonymise: true})

		h.timers[0].fire()

		want := []string{
			":alarm_clock: BET_START: Place bet after 2m0s for Streamer1",
			":four_leaf_clover: BET_GENERAL: Place bet on: Yes (BLUE), Points: 100, Users: 10 (66.67%), Odds: 1.5 (66.67%) for Streamer1",
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("no gain handler", func(t *testing.T) {
		streamer := bettor()
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}, noHandlers: true})

		h.timers[0].fire()

		if got, want := h.account.Bets, []bet{{"ev-1", "yes", 50}}; !reflect.DeepEqual(got, want) {
			t.Errorf("bets = %+v, want %+v", got, want)
		}
		if got, want := streamer.ChannelPoints, 950; got != want {
			t.Errorf("balance = %d, want %d", got, want)
		}
	})

	t.Run("the bet fails", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})
		h.account.BetErr = errors.New("twitch said no")

		h.timers[0].fire()

		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Bets:   []bet{{"ev-1", "yes", 50}},
			Lines:  []string{betStartLine, "error BET_FAILED: prediction Tester: twitch said no"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a result after the bet failed", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})
		h.account.BetErr = errors.New("twitch said no")
		h.timers[0].fire()

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, "error BET_FAILED: prediction Tester: twitch said no"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a filter the event doesn't meet", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.Bet.FilterCondition = crowdedFilter()
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}})

		h.timers[0].fire()

		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Lines:  []string{betStartLine, "BET_FILTERS: Skip bet for Tester: filter_condition TOTAL_USERS GT 1000000 not met (current 15)"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a result after the bet was skipped", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.Bet.FilterCondition = crowdedFilter()
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}})
		h.timers[0].fire()

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		want := []string{betStartLine, "BET_FILTERS: Skip bet for Tester: filter_condition TOTAL_USERS GT 1000000 not met (current 15)"}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a balance that fell to minimum_points before the bet", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.Bet.MinimumPoints = pointer(500)
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}})
		streamer.ChannelPoints = 500

		h.timers[0].fire()

		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Lines:  []string{betStartLine, "BET_FILTERS: Skip bet for Tester: balance 500 <= minimum_points 500"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a balance that fell to minimum_points before the bet, logs anonymised", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.Bet.MinimumPoints = pointer(500)
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}, anonymise: true})
		streamer.ChannelPoints = 500

		h.timers[0].fire()

		want := []string{
			":alarm_clock: BET_START: Place bet after 2m0s for Streamer1",
			"BET_FILTERS: Skip bet for Streamer1: balance below minimum_points",
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an event locked before the bet", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"LOCKED"}}}`)

		h.timers[0].fire()

		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Lines:  []string{betStartLine, "BET_GENERAL: Skip bet for Tester: event status is LOCKED"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("the same event created twice, before the bet", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got, want := h.logger.Lines, []string{betStartLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
		if got := len(h.timers); got != 1 {
			t.Errorf("%d bets scheduled, want 1", got)
		}
	})

	t.Run("the same event created twice", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		h.timers[0].fire()

		if got, want := h.account.Bets, []bet{{"ev-1", "yes", 50}}; !reflect.DeepEqual(got, want) {
			t.Errorf("bets = %+v, want %+v", got, want)
		}
	})
}
