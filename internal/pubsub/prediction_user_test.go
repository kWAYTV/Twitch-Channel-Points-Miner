package pubsub

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_processPredictionUser(t *testing.T) {
	t.Run("a result for a placed bet", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, betWonLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a lost bet", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"LOSE","points_won":0}}}}`)

		lost := ":bar_chart: BET_LOSE: EventPrediction: tester - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorRed + "LOSE, Lost: -50" + ansi.ColorReset
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, lost}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a result of a type it doesn't know", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"VOID"}}}}`)

		voided := ":bar_chart: info: EventPrediction: tester - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorRed + "VOID, Gained: -50" + ansi.ColorReset
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, voided}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a result, logs anonymised", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}, anonymise: true})

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		want := []string{
			":alarm_clock: BET_START: Place bet after 2m0s for Streamer1",
			":four_leaf_clover: BET_GENERAL: Place bet on: Yes (BLUE), Points: 100, Users: 10 (66.67%), Odds: 1.5 (66.67%) for Streamer1",
			":bar_chart: BET_WIN: EventPrediction: Streamer1 - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorGreen + "WIN" + ansi.ColorReset,
		}
		if got := h.logger.Lines; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("the same result twice", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, betWonLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a result after the channel resolved the event", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`)

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, betWonLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a result with nothing in it", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1"}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("a bet confirmed", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-user-v1.999", `{"type":"prediction-made","data":{"prediction":{"event_id":"ev-1"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{
			Timers: []time.Duration{2 * time.Minute},
			Bets:   []bet{{"ev-1", "yes", 50}},
			Gains:  []gain{{"tester", -50, "PREDICTION", 0}},
			Lines:  []string{betStartLine, betPlacedLine},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a result for an event it isn't waiting on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a message with no prediction", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result","data":{}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a message with no data", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-user-v1.999", `{"type":"prediction-result"}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}
