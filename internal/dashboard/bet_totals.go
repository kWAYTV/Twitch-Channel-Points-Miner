package dashboard

import "sort"

// StrategyTotals is how one bet strategy has done: how many bets it placed,
// how they came out, what it staked and what it won or lost overall.
type StrategyTotals struct {
	Strategy string `json:"strategy"`
	Bets     int    `json:"bets"`
	Won      int    `json:"won"`
	Lost     int    `json:"lost"`
	Refunded int    `json:"refunded"`
	Staked   int    `json:"staked"`
	Net      int    `json:"net"`
}

// strategyTotals adds the bets up per strategy, in the order of the
// strategies' names.
func strategyTotals(bets []Bet) []StrategyTotals {
	byStrategy := make(map[string]*StrategyTotals)
	for _, bet := range bets {
		totals := byStrategy[bet.Strategy]
		if totals == nil {
			totals = &StrategyTotals{Strategy: bet.Strategy}
			byStrategy[bet.Strategy] = totals
		}
		totals.Bets++
		totals.Staked += bet.Stake
		totals.Net += bet.Gained
		switch bet.Result {
		case "WIN":
			totals.Won++
		case "LOSE":
			totals.Lost++
		case "REFUND":
			totals.Refunded++
		}
	}
	out := make([]StrategyTotals, 0, len(byStrategy))
	for _, totals := range byStrategy {
		out = append(out, *totals)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Strategy < out[j].Strategy })
	return out
}

// latestBets returns at most limit bets, the newest first.
func latestBets(bets []Bet, limit int) []Bet {
	out := make([]Bet, 0, min(len(bets), limit))
	for i := len(bets) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, bets[i])
	}
	return out
}
