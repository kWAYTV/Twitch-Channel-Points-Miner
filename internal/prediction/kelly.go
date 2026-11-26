package prediction

import (
	"math"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// A Twitch prediction is a parimutuel pool: the winners share everything that
// was bet, in proportion to their stakes, and nothing is taken off the top. A
// bet is worth making only when the outcome is likelier than its share of the
// pool says, and the Kelly criterion says how much of the balance to put on
// it: the stake that makes the balance grow fastest over many bets.
//
// The pool gives the odds, but not the chance of winning. That is estimated
// here from the outcome's share of the viewers who voted, one head one vote,
// which a few large stakes cannot move the way they move the pool.
//
// A share of six voters says little and a share of six hundred says a lot.
// So the estimate is not the share itself but the lower end of its Wilson
// score interval: the lowest share the voters as a whole can be trusted to
// hold, given how few or many of them there are. It is the standard way to
// compare proportions from samples of different sizes. A small prediction
// then offers no edge and is left alone, and a large one is judged by its
// voters, without a separate rule for each size.

// wilsonZ is the normal quantile of a 95% interval, the usual confidence
// level for the Wilson score.
const wilsonZ = 1.96

// wilsonLowerBound returns the lower end of the Wilson score interval for a
// share of successes out of n. It is 0 when n is 0.
func wilsonLowerBound(successes, n int) float64 {
	if n <= 0 {
		return 0
	}
	count := float64(n)
	share := float64(successes) / count
	z2 := wilsonZ * wilsonZ
	centre := share + z2/(2*count)
	margin := wilsonZ * math.Sqrt((share*(1-share)+z2/(4*count))/count)
	return max(0, (centre-margin)/(1+z2/count))
}

// kellyFraction is the part of the full Kelly stake that is bet. The full
// stake is right only when the chance of winning is known exactly; with an
// estimate, a fraction of it gives up little growth and loses far less when
// the estimate is wrong.
const kellyFraction = 0.25

// kellyStake returns the stake on one outcome that makes the expected
// logarithm of the balance largest, given the chance p that the outcome wins.
// outcomePool is what is already bet on the outcome and totalPool what is bet
// on all of them. The stake counts against itself: it joins the pool, so the
// more is bet, the less each point pays. It is 0 when the outcome is not
// worth a bet.
func kellyStake(balance, p, outcomePool, totalPool float64) float64 {
	if balance <= 0 || p <= 0 || outcomePool < 0 || totalPool <= outcomePool {
		return 0
	}
	if p >= 1 {
		p = 1
	}
	growth := func(stake float64) float64 {
		won := balance - stake + stake*(totalPool+stake)/(outcomePool+stake)
		lost := balance - stake
		if lost <= 0 {
			return math.Inf(-1)
		}
		return p*math.Log(won) + (1-p)*math.Log(lost)
	}
	// The growth is concave in the stake, so its peak is found by narrowing
	// the range from both ends.
	low, high := 0.0, balance
	for i := 0; i < 100; i++ {
		third := (high - low) / 3
		left, right := low+third, high-third
		if growth(left) < growth(right) {
			low = left
		} else {
			high = right
		}
	}
	best := (low + high) / 2
	if best < 0.5 || growth(best) <= growth(0) {
		return 0
	}
	return best
}

// kellyChoice picks the outcome a Kelly bettor would bet on, and the full
// Kelly stake for it. The index is -1 when no outcome is worth a bet: none is
// likelier, by its estimated chance, than its share of the pool.
func kellyChoice(outcomes []Outcome, balance int) (index int, stake float64) {
	totalUsers, totalPool := 0, 0
	for _, o := range outcomes {
		totalUsers += o.TotalUsers
		totalPool += o.TotalPoints
	}
	if totalUsers == 0 || totalPool == 0 {
		return -1, 0
	}
	index = -1
	for i, o := range outcomes {
		chance := wilsonLowerBound(o.TotalUsers, totalUsers)
		if s := kellyStake(float64(balance), chance, float64(o.TotalPoints), float64(totalPool)); s > stake {
			index, stake = i, s
		}
	}
	return index, stake
}

// kellyAmount turns the full Kelly stake into the points to bet: the fraction
// of it that is bet, within the limits the settings put on every stake. When
// that fraction is under Twitch's minimum it is raised to the minimum, but
// only while the minimum is itself no more than the full Kelly stake: past
// that point a bet would be larger than the edge is worth, so none is placed.
func kellyAmount(full float64, balance int, settings domain.BetSettings) int {
	if full < MinimumStake {
		return 0
	}
	limit := stakeLimit(balance, settings)
	amount := int(math.Round(full * kellyFraction))
	if amount > limit {
		amount = limit
	}
	if amount < MinimumStake && limit >= MinimumStake {
		amount = MinimumStake
	}
	return amount
}

// stakeLimit is the most the settings allow on one bet: percentage percent of
// the balance, at most max_points, and never more than the balance.
func stakeLimit(balance int, settings domain.BetSettings) int {
	percentage := domain.DefaultBetPercentage
	if settings.Percentage != nil {
		percentage = *settings.Percentage
	}
	limit := int(float64(balance) * (float64(percentage) / 100))
	if settings.MaxPoints != nil && limit > *settings.MaxPoints {
		limit = *settings.MaxPoints
	}
	if limit > balance {
		limit = balance
	}
	return limit
}
