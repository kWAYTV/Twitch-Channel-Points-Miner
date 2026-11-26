package pubsub

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestClient_processPredictionChannel(t *testing.T) {
	t.Run("an event created", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		want := effects{Timers: []time.Duration{2 * time.Minute}, Lines: []string{betStartLine}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("an event created half a minute ago", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}, now: noon.Add(30 * time.Second)})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		want := effects{
			Timers: []time.Duration{90 * time.Second},
			Lines:  []string{":alarm_clock: BET_START: Place bet after 1m30s for Tester"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("an event created, bets set to six seconds before the end", func(t *testing.T) {
		streamer := bettor()
		delay := 6.0
		streamer.Settings.Bet.DelayMode = domain.DelayModeFromEnd
		streamer.Settings.Bet.Delay = &delay
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		want := effects{
			Timers: []time.Duration{114 * time.Second},
			Lines:  []string{":alarm_clock: BET_START: Place bet after 1m54s for Tester"},
		}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("an event created, logs anonymised", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}, anonymise: true})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got, want := h.logger.Lines, []string{":alarm_clock: BET_START: Place bet after 2m0s for Streamer1"}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an event created with a balance equal to minimum_points", func(t *testing.T) {
		streamer := bettor()
		minimum := 1000
		streamer.Settings.Bet.MinimumPoints = &minimum
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an event created with a balance one over minimum_points", func(t *testing.T) {
		streamer := bettor()
		minimum := 999
		streamer.Settings.Bet.MinimumPoints = &minimum
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		want := effects{Timers: []time.Duration{2 * time.Minute}, Lines: []string{betStartLine}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("an event created already locked", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-created","data":{"event":{"id":"ev-1","status":"LOCKED","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[]}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an event created without an ID", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-created","data":{"event":{"status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[]}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("a message with no event", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-channel-v1.123", `{"type":"event-created","data":{}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an event created, make_predictions off", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.MakePredictions = false
		h := setup(world{streamers: []*domain.Streamer{streamer}})

		h.receive(t, "predictions-channel-v1.123", whoWinsCreated)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an event created on a channel it doesn't follow", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.456", whoWinsCreated)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an update that resolves a placed bet as won", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes","outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, betWonLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an update that brings more points before it resolves", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes","outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":20,"total_points":300}]}}}`)

		wonMore := ":bar_chart: BET_WIN: EventPrediction: tester - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorGreen + "WIN, Gained: +150" + ansi.ColorReset
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, wonMore}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an update that cancels a placed bet", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"CANCELED"}}}`)

		refunded := ":bar_chart: BET_REFUND: EventPrediction: tester - who wins - Decision: A: Yes (BLUE) - Result: " + ansi.ColorReset + "REFUND, Refunded: +0" + ansi.ColorReset
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, refunded}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("the same resolving update twice", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`)

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`)

		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine, betWonLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an update while a placed bet is still open", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"LOCKED"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got, want := h.logger.Lines, []string{betStartLine, betPlacedLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an update that resolves an event before the bet", func(t *testing.T) {
		h := betDue(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`)

		if got, want := h.logger.Lines, []string{betStartLine}; !reflect.DeepEqual(got, want) {
			t.Errorf("lines = %q, want %q", got, want)
		}
	})

	t.Run("an update for an event it isn't waiting on", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		err := h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`)

		if err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an update for an open event it never saw created", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		want := effects{Timers: []time.Duration{2 * time.Minute}, Lines: []string{betStartLine}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want %+v", got, want)
		}
	})

	t.Run("a second update for an open event it never saw created", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		want := effects{Timers: []time.Duration{2 * time.Minute}, Lines: []string{betStartLine}}
		if got := h.effects(); !reflect.DeepEqual(got, want) {
			t.Errorf("effects = %+v, want one bet scheduled: %+v", got, want)
		}
	})

	t.Run("an update for an open event that does not say when it opened", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","prediction_window_seconds":120}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an update for an open event that does not say how long it runs", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","created_at":"2026-10-30T12:00:00Z"}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an update for a locked event it never saw created", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"LOCKED","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})

	t.Run("an update for an open event after its bet was skipped", func(t *testing.T) {
		streamer := bettor()
		streamer.Settings.Bet.FilterCondition = crowdedFilter()
		h := betDue(t, world{streamers: []*domain.Streamer{streamer}})
		h.timers[0].fire()

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		if got := len(h.timers); got != 1 {
			t.Errorf("%d bets scheduled, want the first one only", got)
		}
	})

	t.Run("an update for an open event the account bet on from elsewhere", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-user-v1.999", `{"type":"prediction-made","data":{"prediction":{"event_id":"ev-1"}}}`)

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"ACTIVE","title":"who wins","created_at":"2026-10-30T12:00:00Z","prediction_window_seconds":120,"outcomes":[{"id":"yes","title":"Yes","color":"BLUE","total_users":10,"total_points":100},{"id":"no","title":"No","color":"PINK","total_users":5,"total_points":50}]}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none: the bet is already made", got)
		}
	})

	t.Run("a type it doesn't know", func(t *testing.T) {
		h := setup(world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", `{"type":"event-archived","data":{"event":{"id":"ev-1","status":"ACTIVE"}}}`)

		if got := h.effects(); !reflect.DeepEqual(got, effects{}) {
			t.Errorf("effects = %+v, want none", got)
		}
	})
}
