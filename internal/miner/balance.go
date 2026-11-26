package miner

// balanceAfterGain is the local balance after a PubSub points message: prev is the
// balance before it, earned the change it reports and balance the total it reports.
func balanceAfterGain(prev, earned, balance int) int {
	expected := prev + earned

	// Prefer applying the delta (`earned`) over trusting the absolute balance from PubSub
	// PubSub messages can arrive out of order, and may contain a stale pre-spend balance
	// (e.g. after placing a prediction bet), which would otherwise incorrectly inflate the local state
	newBalance := expected
	if earned == 0 && balance != 0 {
		newBalance = balance
	}

	if newBalance < 0 {
		newBalance = 0
	}
	// For positive earn events, keep the balance monotonic to avoid logging negative deltas
	// when older balances arrive after a streak of gains.
	if earned >= 0 && newBalance < prev {
		newBalance = prev
	}
	return newBalance
}
