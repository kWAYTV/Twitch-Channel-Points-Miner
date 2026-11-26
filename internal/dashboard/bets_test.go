package dashboard

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

var (
	wonBet = prediction.SettledBet{
		Streamer: "alice", Title: "Who wins?", Strategy: "KELLY", Outcome: "Yes",
		Stake: 480, Voters: 400, TotalVoters: 600, Pool: 5_000, TotalPool: 10_000, Chance: 0.628,
		Result: "WIN", Gained: 470,
	}
	lostBet = prediction.SettledBet{
		Streamer: "bob", Title: "Map two?", Strategy: "SMART", Outcome: "No",
		Stake: 50, Voters: 10, TotalVoters: 15, Pool: 100, TotalPool: 150,
		Result: "LOSE", Gained: -50,
	}
)

func TestStore_betRoundTrip(t *testing.T) {
	t.Run("a bet with every figure", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordBet(wonBet)

		got, err := store.Bets(0)

		want := []Bet{{
			T: t0, Streamer: "alice", Title: "Who wins?", Strategy: "KELLY", Outcome: "Yes",
			Stake: 480, Voters: 400, TotalVoters: 600, Pool: 5_000, TotalPool: 10_000, Chance: 0.628,
			Result: "WIN", Gained: 470,
		}}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("Bets(0) = %+v, %v, want %+v", got, err, want)
		}
	})

	t.Run("a bet by a strategy that makes no estimate", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordBet(lostBet)

		got, err := store.Bets(0)

		want := []Bet{{
			T: t0, Streamer: "bob", Title: "Map two?", Strategy: "SMART", Outcome: "No",
			Stake: 50, Voters: 10, TotalVoters: 15, Pool: 100, TotalPool: 150,
			Result: "LOSE", Gained: -50,
		}}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("Bets(0) = %+v, %v, want %+v", got, err, want)
		}
	})

	t.Run("two bets, in the order they settled", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordBet(wonBet)
		clock.Set(t0 + minute)
		store.RecordBet(lostBet)

		got, _ := store.Bets(0)

		if len(got) != 2 || got[0].Streamer != "alice" || got[1].Streamer != "bob" || got[1].T != t0+minute {
			t.Errorf("Bets(0) = %+v, want alice's bet then bob's a minute later", got)
		}
	})

	t.Run("no bets yet", func(t *testing.T) {
		store, _, _ := setup(t)

		got, err := store.Bets(0)

		if err != nil || len(got) != 0 {
			t.Errorf("Bets(0) = %+v, %v, want none", got, err)
		}
	})

	t.Run("bets read back after the store is opened again", func(t *testing.T) {
		store, clock, dir := setup(t)
		store.RecordBet(wonBet)

		got, _ := openStore(t, dir, clock).Bets(0)

		if len(got) != 1 || got[0].Streamer != "alice" {
			t.Errorf("Bets(0) = %+v, want alice's bet", got)
		}
	})
}

func TestStore_Bets(t *testing.T) {
	t.Run("a bet settled exactly at since", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordBet(wonBet)

		if got, _ := store.Bets(t0); len(got) != 1 {
			t.Errorf("Bets(t0) = %+v, want the bet settled at t0", got)
		}
	})

	t.Run("a bet settled a millisecond before since", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordBet(wonBet)

		if got, _ := store.Bets(t0 + 1); len(got) != 0 {
			t.Errorf("Bets(t0+1) = %+v, want none", got)
		}
	})

	t.Run("a last line cut short", func(t *testing.T) {
		store, _, dir := setup(t)
		store.RecordBet(wonBet)
		appendToFile(t, dir, "bets.jsonl", `{"t":1700000060000,"streamer":"bo`)

		got, err := store.Bets(0)

		if err != nil || len(got) != 1 || got[0].Streamer != "alice" {
			t.Errorf("Bets(0) = %+v, %v, want alice's bet alone", got, err)
		}
	})

	t.Run("a line that is not a bet", func(t *testing.T) {
		store, _, dir := setup(t)
		writeFile(t, dir, "bets.jsonl", "{\"note\":\"hello\"}\n")

		got, err := store.Bets(0)

		if err != nil || len(got) != 0 {
			t.Errorf("Bets(0) = %+v, %v, want none", got, err)
		}
	})

	t.Run("a bets file that cannot be read", func(t *testing.T) {
		store, _, dir := setup(t)
		replaceWithFolder(t, dir, "bets.jsonl")

		if _, err := store.Bets(0); err == nil {
			t.Errorf("Bets(0) error = nil, want an error")
		}
	})
}

func TestStrategyTotals(t *testing.T) {
	bet := func(strategy, result string, stake, gained int) Bet {
		return Bet{Strategy: strategy, Result: result, Stake: stake, Gained: gained}
	}
	tests := []struct {
		name string
		bets []Bet
		want []StrategyTotals
	}{
		{"no bets", nil, []StrategyTotals{}},
		{"one win", []Bet{bet("KELLY", "WIN", 100, 90)},
			[]StrategyTotals{{Strategy: "KELLY", Bets: 1, Won: 1, Staked: 100, Net: 90}}},
		{"one loss", []Bet{bet("KELLY", "LOSE", 100, -100)},
			[]StrategyTotals{{Strategy: "KELLY", Bets: 1, Lost: 1, Staked: 100, Net: -100}}},
		{"one refund", []Bet{bet("KELLY", "REFUND", 100, 0)},
			[]StrategyTotals{{Strategy: "KELLY", Bets: 1, Refunded: 1, Staked: 100, Net: 0}}},
		{"a win, a loss and a refund", []Bet{
			bet("KELLY", "WIN", 100, 90), bet("KELLY", "LOSE", 50, -50), bet("KELLY", "REFUND", 20, 0),
		}, []StrategyTotals{{Strategy: "KELLY", Bets: 3, Won: 1, Lost: 1, Refunded: 1, Staked: 170, Net: 40}}},
		{"two strategies, by name", []Bet{
			bet("SMART", "LOSE", 50, -50), bet("KELLY", "WIN", 100, 90), bet("SMART", "WIN", 50, 25),
		}, []StrategyTotals{
			{Strategy: "KELLY", Bets: 1, Won: 1, Staked: 100, Net: 90},
			{Strategy: "SMART", Bets: 2, Won: 1, Lost: 1, Staked: 100, Net: -25},
		}},
		{"a bet with no strategy recorded", []Bet{bet("", "WIN", 100, 90)},
			[]StrategyTotals{{Strategy: "", Bets: 1, Won: 1, Staked: 100, Net: 90}}},
		{"a result it does not know", []Bet{bet("KELLY", "VOID", 100, 0)},
			[]StrategyTotals{{Strategy: "KELLY", Bets: 1, Staked: 100, Net: 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strategyTotals(tt.bets); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strategyTotals() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLatestBets(t *testing.T) {
	numbered := func(n int) []Bet {
		bets := make([]Bet, n)
		for i := range bets {
			bets[i] = Bet{T: int64(i + 1)}
		}
		return bets
	}

	t.Run("fewer than the limit, newest first", func(t *testing.T) {
		got := latestBets(numbered(3), 50)

		if want := []Bet{{T: 3}, {T: 2}, {T: 1}}; !reflect.DeepEqual(got, want) {
			t.Errorf("latestBets() = %+v, want %+v", got, want)
		}
	})

	t.Run("more than the limit", func(t *testing.T) {
		got := latestBets(numbered(5), 2)

		if want := []Bet{{T: 5}, {T: 4}}; !reflect.DeepEqual(got, want) {
			t.Errorf("latestBets() = %+v, want %+v", got, want)
		}
	})

	t.Run("none", func(t *testing.T) {
		if got := latestBets(nil, 50); len(got) != 0 || got == nil {
			t.Errorf("latestBets(nil) = %#v, want an empty list", got)
		}
	})
}

func TestGET_api_bets(t *testing.T) {
	t.Run("200: no bets", func(t *testing.T) {
		own := startEmptyApp(t)

		req, _ := http.NewRequest(http.MethodGet, own.url+"/api/bets", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if got, want := string(body), `{"bets":[],"strategies":[]}`; got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("200: a win and a loss by two strategies", func(t *testing.T) {
		own := startEmptyApp(t)
		own.store.RecordBet(wonBet)
		own.clock.Set(t0 + minute)
		own.store.RecordBet(lostBet)

		req, _ := http.NewRequest(http.MethodGet, own.url+"/api/bets", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"bets":[` +
			`{"t":1700000060000,"streamer":"bob","title":"Map two?","strategy":"SMART","outcome":"No","stake":50,"voters":10,"total_voters":15,"pool":100,"total_pool":150,"result":"LOSE","gained":-50},` +
			`{"t":1700000000000,"streamer":"alice","title":"Who wins?","strategy":"KELLY","outcome":"Yes","stake":480,"voters":400,"total_voters":600,"pool":5000,"total_pool":10000,"chance":0.628,"result":"WIN","gained":470}],` +
			`"strategies":[` +
			`{"strategy":"KELLY","bets":1,"won":1,"lost":0,"refunded":0,"staked":480,"net":470},` +
			`{"strategy":"SMART","bets":1,"won":0,"lost":1,"refunded":0,"staked":50,"net":-50}]}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("500: a bets file that cannot be read", func(t *testing.T) {
		own := startEmptyApp(t)
		replaceWithFolder(t, own.dir, "bets.jsonl")

		req, _ := http.NewRequest(http.MethodGet, own.url+"/api/bets", nil)
		resp, _ := send(t, req)

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
	})
}

// appendToFile adds text to the end of a file in dir, without a line ending.
func appendToFile(t *testing.T, dir, name, text string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		t.Fatalf("append to %s: %v", name, err)
	}
}
