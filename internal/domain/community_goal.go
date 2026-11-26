package domain

import (
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
)

type CommunityGoal struct {
	ID                               string
	Title                            string
	IsInStock                        bool
	PointsContributed                int
	AmountNeeded                     int
	PerStreamUserMaximumContribution int
	Status                           string
}

func (c *CommunityGoal) AmountLeft() int {
	return c.AmountNeeded - c.PointsContributed
}

func NewCommunityGoalFromGQL(goal map[string]interface{}) *CommunityGoal {
	if goal == nil {
		return nil
	}
	return &CommunityGoal{
		ID:                               jsonvalue.String(goal["id"]),
		Title:                            jsonvalue.String(goal["title"]),
		IsInStock:                        jsonvalue.Bool(goal["isInStock"]),
		PointsContributed:                jsonvalue.Int(goal["pointsContributed"]),
		AmountNeeded:                     jsonvalue.Int(goal["amountNeeded"]),
		PerStreamUserMaximumContribution: jsonvalue.Int(goal["perStreamUserMaximumContribution"]),
		Status:                           jsonvalue.String(goal["status"]),
	}
}

func NewCommunityGoalFromPubSub(goal map[string]interface{}) *CommunityGoal {
	if goal == nil {
		return nil
	}
	return &CommunityGoal{
		ID:                               jsonvalue.String(goal["id"]),
		Title:                            jsonvalue.String(goal["title"]),
		IsInStock:                        jsonvalue.Bool(goal["is_in_stock"]),
		PointsContributed:                jsonvalue.Int(goal["points_contributed"]),
		AmountNeeded:                     jsonvalue.Int(goal["goal_amount"]),
		PerStreamUserMaximumContribution: jsonvalue.Int(goal["per_stream_maximum_user_contribution"]),
		Status:                           jsonvalue.String(goal["status"]),
	}
}
