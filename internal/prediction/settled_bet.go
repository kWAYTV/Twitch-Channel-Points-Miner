package prediction

// SettledBet is one bet with how it came out, and the figures it was decided
// from. It is what gets recorded, so that a strategy can be judged by its
// results and not by its reasoning.
type SettledBet struct {
	// Streamer is the channel's login.
	Streamer string
	// Title is the prediction's question.
	Title    string
	Strategy string
	// Outcome is the title of the outcome that was bet on.
	Outcome string
	Stake   int

	Voters      int
	TotalVoters int
	Pool        int
	TotalPool   int
	// Chance is the strategy's estimate, 0 when it made none.
	Chance float64

	// Result is WIN, LOSE or REFUND.
	Result string
	// Gained is the net change in points: positive for a win, negative for a
	// loss, 0 for a refund.
	Gained int
}

// Settled describes the event's bet once its result is known.
func (p *Event) Settled(result string, gained int) SettledBet {
	bet := SettledBet{
		Title:       p.Title,
		Outcome:     p.Decision.OutcomeID,
		Stake:       p.Decision.Amount,
		Voters:      p.Basis.Voters,
		TotalVoters: p.Basis.TotalVoters,
		Pool:        p.Basis.Pool,
		TotalPool:   p.Basis.TotalPool,
		Chance:      p.Basis.Chance,
		Result:      result,
		Gained:      gained,
	}
	if outcome := p.DecisionOutcome(); outcome != nil {
		bet.Outcome = outcome.Title
	}
	if p.Streamer != nil {
		bet.Streamer = p.Streamer.Username
		bet.Strategy = string(p.Streamer.Settings.Bet.Strategy)
	}
	return bet
}
