package pubsub

import (
	"fmt"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// chanceNote says what a bet rests on, for a strategy that estimates the
// chance of winning: the estimate against the share of the pool, which is the
// chance the pool itself puts on the outcome. It goes after the outcome's
// figures on the bet line, and is empty for a strategy that makes no estimate.
func chanceNote(basis prediction.Basis) string {
	if basis.Chance <= 0 {
		return ""
	}
	return fmt.Sprintf(", Chance: %s vs pool %s", wholePercent(basis.Chance), wholePercent(basis.PoolShare()))
}

// passedOver names the outcome a KELLY bet would have gone on, with the
// figures that made it not worth betting.
func passedOver(event *prediction.Event) string {
	outcome := event.DecisionOutcome()
	if outcome == nil {
		return ""
	}
	return fmt.Sprintf(" (best: %s, chance %s vs pool %s)", outcome.Title, wholePercent(event.Basis.Chance), wholePercent(event.Basis.PoolShare()))
}

func wholePercent(share float64) string {
	return fmt.Sprintf("%.0f%%", share*100)
}
