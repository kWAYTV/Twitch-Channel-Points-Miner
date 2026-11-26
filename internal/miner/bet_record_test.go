package miner

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func TestRecordSettledBet(t *testing.T) {
	bet := prediction.SettledBet{
		Streamer: "alice", Title: "Who wins?", Strategy: "KELLY", Outcome: "Yes",
		Stake: 480, Voters: 400, TotalVoters: 600, Pool: 5_000, TotalPool: 10_000, Chance: 0.628,
		Result: "WIN", Gained: 470,
	}

	t.Run("a sink installed", func(t *testing.T) {
		sink := installAnalyticsSink(t)

		recordSettledBet(bet)

		if want := []prediction.SettledBet{bet}; !reflect.DeepEqual(sink.Bets, want) {
			t.Errorf("recorded bets = %+v, want %+v", sink.Bets, want)
		}
	})

	t.Run("no sink installed", func(t *testing.T) {
		SetAnalyticsSink(nil)

		recordSettledBet(bet)
	})
}
