package twitch

import (
	"regexp"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func newBet(amount int) *prediction.Event {
	return &prediction.Event{
		Streamer: newStreamer(),
		EventID:  "event-1",
		Decision: prediction.Decision{OutcomeID: "outcome-1", Amount: amount},
	}
}

func TestClient_MakePrediction(t *testing.T) {
	t.Run("a bet of 10 points", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("MakePrediction", "", ok(`{"data":{"makePrediction":{"error":null}}}`))

		err := client.MakePrediction(newBet(10))

		if err != nil {
			t.Fatalf("MakePrediction error = %v, want none", err)
		}
		bet := regexp.MustCompile(`^\{"input":\{"eventID":"event-1","outcomeID":"outcome-1","points":10,"transactionID":"[0-9a-f]{32}"\}\}$`)
		if got := fake.twitch.sent("MakePrediction"); !matchAll(got, bet) {
			t.Errorf("bets sent = %v, want one matching %v", got, bet)
		}
	})

	t.Run("a bet of 9 points", func(t *testing.T) {
		client, fake := setup(t)

		err := client.MakePrediction(newBet(9))

		if want := "invalid prediction decision"; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("a bet on no outcome", func(t *testing.T) {
		client, _ := setup(t)
		bet := newBet(10)
		bet.Decision.OutcomeID = ""

		err := client.MakePrediction(bet)

		if want := "invalid prediction decision"; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
	})

	t.Run("a bet on an event with no id", func(t *testing.T) {
		client, _ := setup(t)
		bet := newBet(10)
		bet.EventID = ""

		err := client.MakePrediction(bet)

		if want := "invalid prediction decision"; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
	})

	t.Run("a bet on an event with no streamer", func(t *testing.T) {
		client, _ := setup(t)
		bet := newBet(10)
		bet.Streamer = nil

		err := client.MakePrediction(bet)

		if want := "nil prediction event"; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
	})

	t.Run("no event", func(t *testing.T) {
		client, _ := setup(t)

		err := client.MakePrediction(nil)

		if want := "nil prediction event"; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("MakePrediction", "", down("connection refused"))

		err := client.MakePrediction(newBet(10))

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("MakePrediction error = %v, want %s", err, want)
		}
	})
}
