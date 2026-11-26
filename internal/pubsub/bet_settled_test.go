package pubsub

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

const (
	wonOnUserTopic      = `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"WIN","points_won":75}}}}`
	lostOnUserTopic     = `{"type":"prediction-result","data":{"prediction":{"event_id":"ev-1","result":{"type":"LOSE","points_won":0}}}}`
	resolvedOnChannel   = `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"RESOLVED","winning_outcome_id":"yes"}}}`
	cancelledOnChannel  = `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"CANCELED"}}}`
	stillOpenOnChannel  = `{"type":"event-updated","data":{"event":{"id":"ev-1","status":"LOCKED"}}}`
	userPredictionTopic = "predictions-user-v1.999"
)

// A bet of 50 on Yes in whoWinsCreated: ten of fifteen voters and 100 of the
// 150 points were on Yes when it was placed.
func settledWith(result string, gained int) prediction.SettledBet {
	return prediction.SettledBet{
		Streamer: "tester", Title: "who wins", Strategy: "MOST_VOTED", Outcome: "Yes",
		Stake: 50, Voters: 10, TotalVoters: 15, Pool: 100, TotalPool: 150,
		Result: result, Gained: gained,
	}
}

func TestClient_betSettled(t *testing.T) {
	t.Run("a win reported on the user topic", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, userPredictionTopic, wonOnUserTopic)

		if want := []prediction.SettledBet{settledWith("WIN", 25)}; !reflect.DeepEqual(h.settled, want) {
			t.Errorf("settled bets = %+v, want %+v", h.settled, want)
		}
	})

	t.Run("a loss reported on the user topic", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, userPredictionTopic, lostOnUserTopic)

		if want := []prediction.SettledBet{settledWith("LOSE", -50)}; !reflect.DeepEqual(h.settled, want) {
			t.Errorf("settled bets = %+v, want %+v", h.settled, want)
		}
	})

	t.Run("a win seen on the channel topic", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", resolvedOnChannel)

		if want := []prediction.SettledBet{settledWith("WIN", 25)}; !reflect.DeepEqual(h.settled, want) {
			t.Errorf("settled bets = %+v, want %+v", h.settled, want)
		}
	})

	t.Run("a prediction cancelled", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", cancelledOnChannel)

		if want := []prediction.SettledBet{settledWith("REFUND", 0)}; !reflect.DeepEqual(h.settled, want) {
			t.Errorf("settled bets = %+v, want %+v", h.settled, want)
		}
	})

	t.Run("a result on both topics", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})
		h.receive(t, "predictions-channel-v1.123", resolvedOnChannel)

		h.receive(t, userPredictionTopic, wonOnUserTopic)

		if got := len(h.settled); got != 1 {
			t.Errorf("%d settled bets, want 1", got)
		}
	})

	t.Run("a prediction that is locked but not decided", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}})

		h.receive(t, "predictions-channel-v1.123", stillOpenOnChannel)

		if len(h.settled) != 0 {
			t.Errorf("settled bets = %+v, want none yet", h.settled)
		}
	})

	t.Run("no handler for settled bets", func(t *testing.T) {
		h := betPlaced(t, world{streamers: []*domain.Streamer{bettor()}, noHandlers: true})

		if err := h.receive(t, userPredictionTopic, wonOnUserTopic); err != nil {
			t.Errorf("handleMessage = %v, want nil", err)
		}
	})
}
