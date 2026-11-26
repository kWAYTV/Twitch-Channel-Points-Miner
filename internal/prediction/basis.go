package prediction

// Basis is what a prediction looked like when its bet was decided: how the
// voters and the pool stood on the chosen outcome, and what the strategy
// made of it. It is kept so the bet can be explained and, later, judged.
type Basis struct {
	// Voters chose the outcome, out of TotalVoters on all outcomes.
	Voters      int
	TotalVoters int
	// Pool is the points on the outcome, out of TotalPool on all outcomes.
	Pool      int
	TotalPool int
	// Chance is the strategy's estimate that the outcome wins. It is 0 for a
	// strategy that makes no estimate.
	Chance float64
}

// PoolShare is the outcome's share of the pool, which is the chance the pool
// itself puts on it. It is 0 for an empty pool.
func (b Basis) PoolShare() float64 {
	if b.TotalPool <= 0 {
		return 0
	}
	return float64(b.Pool) / float64(b.TotalPool)
}

// basisOf reads the basis of a bet on outcomes[choice].
func basisOf(outcomes []Outcome, choice int, chance float64) Basis {
	basis := Basis{Chance: chance}
	for _, outcome := range outcomes {
		basis.TotalVoters += outcome.TotalUsers
		basis.TotalPool += outcome.TotalPoints
	}
	basis.Voters = outcomes[choice].TotalUsers
	basis.Pool = outcomes[choice].TotalPoints
	return basis
}
