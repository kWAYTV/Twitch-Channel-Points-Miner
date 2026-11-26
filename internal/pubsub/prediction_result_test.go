package pubsub

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/ansi"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

func TestResultColor(t *testing.T) {
	for _, row := range []struct {
		name   string
		gained int
		want   string
	}{
		{"a gain of one point", 1, ansi.ColorGreen},
		{"nothing gained", 0, ansi.ColorReset},
		{"a loss of one point", -1, ansi.ColorRed},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := resultColor(row.gained); got != row.want {
				t.Errorf("resultColor(%d) = %q, want %q", row.gained, got, row.want)
			}
		})
	}
}

func TestAnonymisedLabel(t *testing.T) {
	for _, row := range []struct {
		name     string
		streamer string
		title    string
		want     string
	}{
		{"an alias and a title", "Streamer1", "who wins", "EventPrediction: Streamer1 - who wins"},
		{"only an alias", "Streamer1", "", "EventPrediction: Streamer1"},
		{"only a title", "", "who wins", "EventPrediction: who wins"},
		{"neither", "", "", "EventPrediction"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := anonymisedLabel(row.streamer, row.title); got != row.want {
				t.Errorf("anonymisedLabel(%q, %q) = %q, want %q", row.streamer, row.title, got, row.want)
			}
		})
	}
}

func TestSettledResult(t *testing.T) {
	stake := prediction.Decision{Choice: 0, OutcomeID: "yes", Amount: 50}
	outcomes := []prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 50}}

	for _, row := range []struct {
		name        string
		status      string
		winningID   string
		wantType    string
		wantPoints  int
		wantSettled bool
	}{
		{"resolved for the outcome bet on", "RESOLVED", "yes", "WIN", 75, true},
		{"resolved for the other outcome", "RESOLVED", "no", "LOSE", 0, true},
		{"resolved with no winner named", "RESOLVED", "", "", 0, false},
		{"canceled", "CANCELED", "", "REFUND", 0, true},
		{"cancelled, British spelling", "CANCELLED", "", "REFUND", 0, true},
		{"canceled with a winner named", "CANCELED", "yes", "REFUND", 0, true},
		{"still open", "ACTIVE", "", "", 0, false},
		{"locked", "LOCKED", "yes", "", 0, false},
		{"no status", "", "yes", "", 0, false},
		{"resolved in lower case", "resolved", "yes", "", 0, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			resultType, points, settled := settledResult(row.status, row.winningID, stake, outcomes)
			if resultType != row.wantType || points != row.wantPoints || settled != row.wantSettled {
				t.Errorf("settledResult(%q, %q, ...) = %q, %d, %t, want %q, %d, %t",
					row.status, row.winningID, resultType, points, settled, row.wantType, row.wantPoints, row.wantSettled)
			}
		})
	}
}

func TestWinningOutcomeID(t *testing.T) {
	type object = map[string]interface{}
	type list = []interface{}

	for _, row := range []struct {
		name  string
		event object
		want  string
	}{
		{
			"the winner named on the event",
			object{"winning_outcome_id": "yes", "outcomes": list{object{"id": "no", "is_winning_outcome": true}}},
			"yes",
		},
		{
			"the winner flagged on its outcome",
			object{"outcomes": list{object{"id": "no"}, object{"id": "yes", "is_winning_outcome": true}}},
			"yes",
		},
		{
			"an outcome flagged as not winning",
			object{"outcomes": list{object{"id": "no", "is_winning_outcome": false}}},
			"",
		},
		{
			"an outcome in state RESOLVED",
			object{"outcomes": list{object{"id": "no", "state": "LOST"}, object{"id": "yes", "state": "RESOLVED"}}},
			"yes",
		},
		{
			"an outcome in state winner, lower case",
			object{"outcomes": list{object{"id": "yes", "state": "winner"}}},
			"yes",
		},
		{
			"an outcome in state WIN",
			object{"outcomes": list{object{"id": "yes", "state": "WIN"}}},
			"yes",
		},
		{
			"a flagged outcome without an ID, then one in state WIN",
			object{"outcomes": list{object{"is_winning_outcome": true}, object{"id": "yes", "state": "WIN"}}},
			"yes",
		},
		{
			"an outcome in state WIN without an ID",
			object{"outcomes": list{object{"state": "WIN"}}},
			"",
		},
		{
			"an outcome that is not an object",
			object{"outcomes": list{"yes", object{"id": "no", "is_winning_outcome": true}}},
			"no",
		},
		{"no outcomes", object{"outcomes": list{}}, ""},
		{"an empty event", object{}, ""},
		{"no event", nil, ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := winningOutcomeID(row.event); got != row.want {
				t.Errorf("winningOutcomeID(%v) = %q, want %q", row.event, got, row.want)
			}
		})
	}
}

func TestPayoutForOutcome(t *testing.T) {
	for _, row := range []struct {
		name      string
		decision  prediction.Decision
		outcomes  []prediction.Outcome
		winningID string
		want      int
	}{
		{
			"a win at odds of 1.5",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 50}},
			"yes",
			75,
		},
		{
			"a payout that rounds up from a half",
			prediction.Decision{OutcomeID: "yes", Amount: 1},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 2}, {ID: "no", TotalPoints: 1}},
			"yes",
			2,
		},
		{
			"a payout that rounds down",
			prediction.Decision{OutcomeID: "yes", Amount: 10},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 3}, {ID: "no", TotalPoints: 1}},
			"yes",
			13,
		},
		{
			"every point on the winner",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 0}},
			"yes",
			50,
		},
		{
			"a bet on the other outcome",
			prediction.Decision{OutcomeID: "no", Amount: 50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 50}},
			"yes",
			0,
		},
		{
			"a stake of zero",
			prediction.Decision{OutcomeID: "yes", Amount: 0},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 50}},
			"yes",
			0,
		},
		{
			"a negative stake",
			prediction.Decision{OutcomeID: "yes", Amount: -50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 100}, {ID: "no", TotalPoints: 50}},
			"yes",
			0,
		},
		{
			"no points on the winner",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 0}, {ID: "no", TotalPoints: 50}},
			"yes",
			50,
		},
		{
			"no points on any outcome",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			[]prediction.Outcome{{ID: "yes", TotalPoints: 0}, {ID: "no", TotalPoints: 0}},
			"yes",
			50,
		},
		{
			"a winner that is not among the outcomes",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			[]prediction.Outcome{{ID: "no", TotalPoints: 50}},
			"yes",
			50,
		},
		{
			"no outcomes",
			prediction.Decision{OutcomeID: "yes", Amount: 50},
			nil,
			"yes",
			50,
		},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := payoutForOutcome(row.decision, row.outcomes, row.winningID); got != row.want {
				t.Errorf("payoutForOutcome(%+v, %+v, %q) = %d, want %d", row.decision, row.outcomes, row.winningID, got, row.want)
			}
		})
	}
}
