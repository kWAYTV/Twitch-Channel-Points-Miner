package dashboard

import (
	"encoding/json"
	"path/filepath"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// betsFile holds every settled bet, one per line, for all streamers.
const betsFile = "bets.jsonl"

// Bet is one settled bet as it is stored: when it settled (unix ms), what was
// bet, the figures it was decided from, and how it came out.
type Bet struct {
	T        int64  `json:"t"`
	Streamer string `json:"streamer"`
	Title    string `json:"title"`
	Strategy string `json:"strategy"`
	Outcome  string `json:"outcome"`
	Stake    int    `json:"stake"`

	Voters      int `json:"voters"`
	TotalVoters int `json:"total_voters"`
	Pool        int `json:"pool"`
	TotalPool   int `json:"total_pool"`
	// Chance is the strategy's estimate; left out when it made none.
	Chance float64 `json:"chance,omitempty"`

	// Result is WIN, LOSE or REFUND; Gained is the net change in points.
	Result string `json:"result"`
	Gained int    `json:"gained"`
}

// RecordBet stores a bet that has just settled.
func (s *Store) RecordBet(bet prediction.SettledBet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appendLine(betsFile, Bet{
		T:           s.now().UnixMilli(),
		Streamer:    bet.Streamer,
		Title:       bet.Title,
		Strategy:    bet.Strategy,
		Outcome:     bet.Outcome,
		Stake:       bet.Stake,
		Voters:      bet.Voters,
		TotalVoters: bet.TotalVoters,
		Pool:        bet.Pool,
		TotalPool:   bet.TotalPool,
		Chance:      bet.Chance,
		Result:      bet.Result,
		Gained:      bet.Gained,
	})
}

// Bets loads the bets settled at or after since, oldest first.
func (s *Store) Bets(since int64) ([]Bet, error) {
	var bets []Bet
	err := readLines(filepath.Join(s.dir, betsFile), func(line []byte) {
		if bet, ok := parseBet(line); ok && bet.T >= since {
			bets = append(bets, bet)
		}
	})
	if err != nil {
		return nil, err
	}
	return bets, nil
}

// parseBet reads one line of the bets file. The file is append-only, so a
// torn last line simply fails to parse.
func parseBet(line []byte) (Bet, bool) {
	var bet Bet
	if json.Unmarshal(line, &bet) != nil || bet.T == 0 || bet.Result == "" {
		return Bet{}, false
	}
	return bet, true
}
