package prediction

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/points"
)

type Outcome struct {
	ID              string
	Title           string
	Color           string
	TotalUsers      int
	TotalPoints     int
	TopPoints       int
	PercentageUsers float64
	Odds            float64
	OddsPercentage  float64
}

// parseOutcomes reads the outcomes of a prediction as Twitch sends them, and
// works out each one's share of the users and its odds.
func parseOutcomes(outcomes []interface{}) []Outcome {
	parsed := make([]Outcome, 0, len(outcomes))
	totalUsers := 0
	totalPoints := 0
	for _, raw := range outcomes {
		oc, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		outcome := Outcome{
			ID:          jsonvalue.String(oc["id"]),
			Title:       jsonvalue.String(oc["title"]),
			Color:       jsonvalue.String(oc["color"]),
			TotalUsers:  jsonvalue.Int(oc["total_users"]),
			TotalPoints: jsonvalue.Int(oc["total_points"]),
		}
		if topPredictors, ok := oc["top_predictors"].([]interface{}); ok && len(topPredictors) > 0 {
			if first, ok := topPredictors[0].(map[string]interface{}); ok {
				outcome.TopPoints = jsonvalue.Int(first["points"])
			}
		}
		parsed = append(parsed, outcome)
		totalUsers += outcome.TotalUsers
		totalPoints += outcome.TotalPoints
	}
	for i := range parsed {
		if totalUsers > 0 {
			parsed[i].PercentageUsers = (float64(parsed[i].TotalUsers) * 100) / float64(totalUsers)
		}
		if parsed[i].TotalPoints > 0 {
			parsed[i].Odds = float64(totalPoints) / float64(parsed[i].TotalPoints)
		}
		if parsed[i].Odds > 0 {
			parsed[i].OddsPercentage = 100 / parsed[i].Odds
		}
	}
	return parsed
}

func (o Outcome) String() string {
	return fmt.Sprintf(
		"%s (%s), Points: %s, Users: %s (%.2f%%), Odds: %s (%s%%)",
		strings.TrimSpace(o.Title),
		strings.ToUpper(o.Color),
		points.Format(o.TotalPoints),
		points.Format(o.TotalUsers),
		o.PercentageUsers,
		FormatFloat(o.Odds),
		FormatFloat(o.OddsPercentage),
	)
}
