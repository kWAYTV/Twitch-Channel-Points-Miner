package prediction

import (
	"math"
	"sort"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func (p *Event) Decide(balance int) Decision {
	decision := Decision{}
	if p.Streamer == nil || len(p.Outcomes) == 0 {
		return decision
	}
	settings := p.Streamer.Settings.Bet

	if settings.Strategy == "" {
		settings.Strategy = domain.DefaultBetStrategy
	}
	var choice, amount int
	var chance float64
	if settings.Strategy == domain.StrategyKelly {
		choice, amount, chance = kellyDecision(p.Outcomes, balance, settings)
	} else {
		choice = selectOutcome(p.Outcomes, settings)
		if choice >= 0 && choice < len(p.Outcomes) {
			amount = stake(balance, settings, p.Outcomes[choice].TopPoints)
		}
	}
	if choice < 0 || choice >= len(p.Outcomes) {
		return decision
	}
	p.Basis = basisOf(p.Outcomes, choice, chance)

	decision = Decision{
		Choice:    choice,
		OutcomeID: p.Outcomes[choice].ID,
		Amount:    amount,
	}
	p.Decision = decision
	return decision
}

// kellyDecision is the outcome and stake of the KELLY strategy. When no
// outcome is worth a bet it still names the outcome most viewers chose, with
// a stake of 0, so the log can say what was passed over. chance is the
// estimate the choice rests on.
func kellyDecision(outcomes []Outcome, balance int, settings domain.BetSettings) (choice, amount int, chance float64) {
	choice, full := kellyChoice(outcomes, balance)
	if choice < 0 {
		choice = maxIndex(outcomes, func(o Outcome) float64 { return float64(o.TotalUsers) })
	} else {
		amount = underTopBet(kellyAmount(full, balance, settings), settings, outcomes[choice].TopPoints)
	}
	if choice < 0 {
		return choice, 0, 0
	}
	totalVoters := 0
	for _, outcome := range outcomes {
		totalVoters += outcome.TotalUsers
	}
	return choice, amount, wilsonLowerBound(outcomes[choice].TotalUsers, totalVoters)
}

// underTopBet lowers a stake to one point under the largest bet on the
// outcome when stealth mode is on, so it is never the top bet.
func underTopBet(amount int, settings domain.BetSettings, topPoints int) int {
	if settings.StealthMode == nil || !*settings.StealthMode || topPoints <= 0 || amount < topPoints {
		return amount
	}
	return max(topPoints-1, 1)
}

// MinimumStake is the smallest bet Twitch accepts.
const MinimumStake = 10

// stake is how many points to bet from balance. topPoints is the largest bet
// already placed on the chosen outcome, which stealth mode stays below.
func stake(balance int, settings domain.BetSettings, topPoints int) int {
	amount := underTopBet(stakeLimit(balance, settings), settings, topPoints)
	if amount < MinimumStake {
		if settings.MaxPoints != nil && *settings.MaxPoints < MinimumStake {
			amount = min(*settings.MaxPoints, balance)
		} else if balance >= MinimumStake {
			amount = MinimumStake
		}
	}
	return amount
}

func selectOutcome(outcomes []Outcome, settings domain.BetSettings) int {
	if len(outcomes) == 0 {
		return -1
	}
	switch settings.Strategy {
	case domain.StrategyMostVoted:
		return maxIndex(outcomes, func(o Outcome) float64 { return float64(o.TotalUsers) })
	case domain.StrategyHighOdds:
		return maxIndex(outcomes, func(o Outcome) float64 { return o.Odds })
	case domain.StrategyPercentage:
		return maxIndex(outcomes, func(o Outcome) float64 { return o.OddsPercentage })
	case domain.StrategySmartMoney:
		return maxIndex(outcomes, func(o Outcome) float64 { return float64(o.TopPoints) })
	case domain.StrategyNumber1:
		if len(outcomes) > 0 {
			return 0
		}
	case domain.StrategyNumber2:
		if len(outcomes) > 1 {
			return 1
		}
	case domain.StrategyNumber3:
		if len(outcomes) > 2 {
			return 2
		}
	case domain.StrategyNumber4:
		if len(outcomes) > 3 {
			return 3
		}
	case domain.StrategyNumber5:
		if len(outcomes) > 4 {
			return 4
		}
	case domain.StrategyNumber6:
		if len(outcomes) > 5 {
			return 5
		}
	case domain.StrategyNumber7:
		if len(outcomes) > 6 {
			return 6
		}
	case domain.StrategyNumber8:
		if len(outcomes) > 7 {
			return 7
		}
	case domain.StrategySmart:
		gap := domain.DefaultBetPercentageGap
		if settings.PercentageGap != nil {
			gap = *settings.PercentageGap
		}
		percents := append([]Outcome(nil), outcomes...)
		sort.SliceStable(percents, func(i, j int) bool {
			return percents[i].PercentageUsers > percents[j].PercentageUsers
		})
		if len(percents) >= 2 {
			if math.Abs(percents[0].PercentageUsers-percents[1].PercentageUsers) < float64(gap) {
				return maxIndex(outcomes, func(o Outcome) float64 { return o.Odds })
			}
		}
		return maxIndex(outcomes, func(o Outcome) float64 { return float64(o.TotalUsers) })
	}

	return maxIndex(outcomes, func(o Outcome) float64 { return o.Odds })
}

func maxIndex(outcomes []Outcome, value func(Outcome) float64) int {
	if len(outcomes) == 0 {
		return -1
	}
	best := 0
	bestVal := value(outcomes[0])
	for i := 1; i < len(outcomes); i++ {
		if v := value(outcomes[i]); v > bestVal {
			best = i
			bestVal = v
		}
	}
	return best
}
