package prediction

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

var (
	now       = time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	createdAt = time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
)

func TestNewEventAt(t *testing.T) {
	t.Run("an event with every field", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{
			"id":                        "event-1",
			"title":                     "Will it rain?",
			"status":                    "ACTIVE",
			"created_at":                "2026-03-01T10:00:00Z",
			"prediction_window_seconds": float64(120),
			"outcomes": []interface{}{
				map[string]interface{}{"id": "a", "title": "Yes", "color": "BLUE", "total_users": float64(10), "total_points": float64(100)},
			},
		}
		want := &Event{
			Streamer:      streamer,
			EventID:       "event-1",
			Title:         "Will it rain?",
			Status:        "ACTIVE",
			CreatedAt:     createdAt,
			WindowSeconds: 120,
			Outcomes: []Outcome{
				{ID: "a", Title: "Yes", Color: "BLUE", TotalUsers: 10, TotalPoints: 100, PercentageUsers: 100, Odds: 1, OddsPercentage: 100},
			},
		}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty event", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{}
		want := &Event{Streamer: streamer, CreatedAt: now, Outcomes: []Outcome{}}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("a creation time that is not a date", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{"id": "event-1", "created_at": "yesterday"}
		want := &Event{Streamer: streamer, EventID: "event-1", CreatedAt: now, Outcomes: []Outcome{}}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("a creation time that is not a string", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{"id": "event-1", "created_at": float64(1772359200)}
		want := &Event{Streamer: streamer, EventID: "event-1", CreatedAt: now, Outcomes: []Outcome{}}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("a lower-case status", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{"id": "event-1", "status": "locked", "created_at": "2026-03-01T10:00:00Z"}
		want := &Event{Streamer: streamer, EventID: "event-1", Status: "LOCKED", CreatedAt: createdAt, Outcomes: []Outcome{}}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("a title with spaces around it", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{"id": "event-1", "title": "  Will it rain?\n", "created_at": "2026-03-01T10:00:00Z"}
		want := &Event{Streamer: streamer, EventID: "event-1", Title: "Will it rain?", CreatedAt: createdAt, Outcomes: []Outcome{}}
		if got := newEventAt(streamer, raw, now); !reflect.DeepEqual(got, want) {
			t.Errorf("newEventAt() = %+v, want %+v", got, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		raw := map[string]interface{}{"id": "event-1"}
		if got := newEventAt(nil, raw, now); got != nil {
			t.Errorf("newEventAt(nil, event) = %+v, want nil", got)
		}
	})

	t.Run("no event", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		if got := newEventAt(streamer, nil, now); got != nil {
			t.Errorf("newEventAt(streamer, nil) = %+v, want nil", got)
		}
	})
}

func TestNewEvent(t *testing.T) {
	t.Run("an event with its own creation time", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "tester"}
		raw := map[string]interface{}{"id": "event-1", "created_at": "2026-03-01T10:00:00Z"}
		want := &Event{Streamer: streamer, EventID: "event-1", CreatedAt: createdAt, Outcomes: []Outcome{}}
		if got := NewEvent(streamer, raw); !reflect.DeepEqual(got, want) {
			t.Errorf("NewEvent() = %+v, want %+v", got, want)
		}
	})
}

func TestEvent_UpdateOutcomes(t *testing.T) {
	t.Run("new outcomes replace the old ones", func(t *testing.T) {
		event := &Event{Outcomes: []Outcome{{ID: "old"}, {ID: "older"}}}

		event.UpdateOutcomes([]interface{}{
			map[string]interface{}{"id": "a", "total_users": float64(3), "total_points": float64(60)},
		})

		want := []Outcome{{ID: "a", TotalUsers: 3, TotalPoints: 60, PercentageUsers: 100, Odds: 1, OddsPercentage: 100}}
		if !reflect.DeepEqual(event.Outcomes, want) {
			t.Errorf("Outcomes = %+v, want %+v", event.Outcomes, want)
		}
	})
}

func TestEvent_ClosingAfter(t *testing.T) {
	for _, row := range []struct {
		name          string
		windowSeconds float64
		at            time.Time
		want          time.Duration
	}{
		{"as it opens", 30, createdAt, 30 * time.Second},
		{"part of the way through", 30, createdAt.Add(10 * time.Second), 20 * time.Second},
		{"one second before it closes", 30, createdAt.Add(29 * time.Second), time.Second},
		{"exactly as it closes", 30, createdAt.Add(30 * time.Second), 0},
		{"one second after it closed", 30, createdAt.Add(31 * time.Second), 0},
		{"before it was created", 30, createdAt.Add(-5 * time.Second), 35 * time.Second},
		{"a window of half a second", 0.5, createdAt, 500 * time.Millisecond},
		{"no window", 0, createdAt, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			event := &Event{CreatedAt: createdAt, WindowSeconds: row.windowSeconds}
			if got := event.ClosingAfter(row.at); got != row.want {
				t.Errorf("ClosingAfter(%s) = %s, want %s", row.at, got, row.want)
			}
		})
	}
}

type parsedResult struct {
	gained, placed, won      int
	resultType, resultString string
}

func parseResult(event *Event, result map[string]interface{}) parsedResult {
	gained, placed, won, resultType, resultString := event.ParseResult(result)
	return parsedResult{gained, placed, won, resultType, resultString}
}

func TestEvent_ParseResult(t *testing.T) {
	t.Run("a win", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "WIN", "points_won": float64(1200)})

		want := parsedResult{gained: 700, placed: 500, won: 1200, resultType: "WIN", resultString: "WIN, Gained: +700"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
		if event.ResultType != "WIN" {
			t.Errorf("ResultType = %q, want %q", event.ResultType, "WIN")
		}
		if event.ResultString != "WIN, Gained: +700" {
			t.Errorf("ResultString = %q, want %q", event.ResultString, "WIN, Gained: +700")
		}
	})

	t.Run("a win of thousands", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "WIN", "points_won": float64(3000)})

		want := parsedResult{gained: 2500, placed: 500, won: 3000, resultType: "WIN", resultString: "WIN, Gained: +2.50k"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
	})

	t.Run("a win that pays back the stake", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "WIN", "points_won": float64(500)})

		want := parsedResult{gained: 0, placed: 500, won: 500, resultType: "WIN", resultString: "WIN, Gained: +0"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
	})

	t.Run("a loss", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "LOSE", "points_won": float64(0)})

		want := parsedResult{gained: -500, placed: 500, won: 0, resultType: "LOSE", resultString: "LOSE, Lost: -500"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
		if event.ResultType != "LOSE" {
			t.Errorf("ResultType = %q, want %q", event.ResultType, "LOSE")
		}
		if event.ResultString != "LOSE, Lost: -500" {
			t.Errorf("ResultString = %q, want %q", event.ResultString, "LOSE, Lost: -500")
		}
	})

	t.Run("a refund", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "REFUND", "points_won": float64(500)})

		want := parsedResult{gained: 0, placed: 0, won: 0, resultType: "REFUND", resultString: "REFUND, Refunded: +0"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
		if event.ResultType != "REFUND" {
			t.Errorf("ResultType = %q, want %q", event.ResultType, "REFUND")
		}
		if event.ResultString != "REFUND, Refunded: +0" {
			t.Errorf("ResultString = %q, want %q", event.ResultString, "REFUND, Refunded: +0")
		}
	})

	t.Run("a lower-case type", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{"type": "lose"})

		want := parsedResult{gained: -500, placed: 500, won: 0, resultType: "LOSE", resultString: "LOSE, Lost: -500"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty result", func(t *testing.T) {
		event := &Event{Decision: Decision{Amount: 500}}

		got := parseResult(event, map[string]interface{}{})

		want := parsedResult{gained: -500, placed: 500, won: 0, resultType: "", resultString: ", Gained: -500"}
		if got != want {
			t.Errorf("ParseResult() = %+v, want %+v", got, want)
		}
	})
}

func TestEvent_String(t *testing.T) {
	for _, row := range []struct {
		name  string
		event *Event
		want  string
	}{
		{"a streamer and a title", &Event{Streamer: &domain.Streamer{Username: "tester"}, Title: "Will it rain?"}, "EventPrediction: tester - Will it rain?"},
		{"a streamer with no name", &Event{Streamer: &domain.Streamer{}, Title: "Will it rain?"}, "EventPrediction: Will it rain?"},
		{"no streamer", &Event{Title: "Will it rain?"}, "EventPrediction: Will it rain?"},
		{"no streamer and no title", &Event{}, "EventPrediction"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := row.event.String(); got != row.want {
				t.Errorf("String() = %q, want %q", got, row.want)
			}
		})
	}
}

func TestEvent_DecisionOutcome(t *testing.T) {
	t.Run("the choice is one of the outcomes", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes"}, {ID: "b", Title: "No"}},
			Decision: Decision{Choice: 1, OutcomeID: "b"},
		}
		want := Outcome{ID: "b", Title: "No"}
		if got := event.DecisionOutcome(); got == nil || *got != want {
			t.Errorf("DecisionOutcome() = %+v, want %+v", got, want)
		}
	})

	t.Run("the choice and the id disagree", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes"}, {ID: "b", Title: "No"}},
			Decision: Decision{Choice: 0, OutcomeID: "b"},
		}
		want := Outcome{ID: "a", Title: "Yes"}
		if got := event.DecisionOutcome(); got == nil || *got != want {
			t.Errorf("DecisionOutcome() = %+v, want %+v", got, want)
		}
	})

	t.Run("the choice is past the last outcome", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes"}, {ID: "b", Title: "No"}},
			Decision: Decision{Choice: 2, OutcomeID: "b"},
		}
		want := Outcome{ID: "b", Title: "No"}
		if got := event.DecisionOutcome(); got == nil || *got != want {
			t.Errorf("DecisionOutcome() = %+v, want %+v", got, want)
		}
	})

	t.Run("the choice is below zero", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes"}, {ID: "b", Title: "No"}},
			Decision: Decision{Choice: -1, OutcomeID: "a"},
		}
		want := Outcome{ID: "a", Title: "Yes"}
		if got := event.DecisionOutcome(); got == nil || *got != want {
			t.Errorf("DecisionOutcome() = %+v, want %+v", got, want)
		}
	})

	t.Run("neither the choice nor the id is an outcome", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes"}, {ID: "b", Title: "No"}},
			Decision: Decision{Choice: 2, OutcomeID: "c"},
		}
		if got := event.DecisionOutcome(); got != nil {
			t.Errorf("DecisionOutcome() = %+v, want nil", got)
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		event := &Event{Decision: Decision{Choice: 0, OutcomeID: "a"}}
		if got := event.DecisionOutcome(); got != nil {
			t.Errorf("DecisionOutcome() = %+v, want nil", got)
		}
	})
}

func TestEvent_DecisionOutcomeString(t *testing.T) {
	t.Run("a decision on one of the outcomes", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes", Color: "BLUE", TotalUsers: 10, TotalPoints: 100, PercentageUsers: 100, Odds: 1, OddsPercentage: 100}},
			Decision: Decision{Choice: 0, OutcomeID: "a"},
		}
		want := "Yes (BLUE), Points: 100, Users: 10 (100.00%), Odds: 1 (100%)"
		if got := event.DecisionOutcomeString(); got != want {
			t.Errorf("DecisionOutcomeString() = %q, want %q", got, want)
		}
	})

	t.Run("a decision on an outcome the event does not have", func(t *testing.T) {
		event := &Event{Decision: Decision{Choice: 0, OutcomeID: "a"}}
		if got := event.DecisionOutcomeString(); got != "a" {
			t.Errorf("DecisionOutcomeString() = %q, want %q", got, "a")
		}
	})
}

func TestEvent_DecisionLabel(t *testing.T) {
	t.Run("a decision on one of the outcomes", func(t *testing.T) {
		event := &Event{
			Outcomes: []Outcome{{ID: "a", Title: "Yes", Color: "blue"}, {ID: "b", Title: "No", Color: "pink"}},
			Decision: Decision{Choice: 1, OutcomeID: "b"},
		}
		if got := event.DecisionLabel(); got != "B: No (PINK)" {
			t.Errorf("DecisionLabel() = %q, want %q", got, "B: No (PINK)")
		}
	})

	t.Run("a decision on an outcome the event does not have", func(t *testing.T) {
		event := &Event{Decision: Decision{Choice: 1, OutcomeID: "b"}}
		if got := event.DecisionLabel(); got != "B: b" {
			t.Errorf("DecisionLabel() = %q, want %q", got, "B: b")
		}
	})

	t.Run("no decision", func(t *testing.T) {
		event := &Event{}
		if got := event.DecisionLabel(); got != "" {
			t.Errorf("DecisionLabel() = %q, want %q", got, "")
		}
	})
}

func TestChoiceLabel(t *testing.T) {
	for _, row := range []struct {
		name   string
		choice int
		want   string
	}{
		{"the first choice", 0, "A"},
		{"the second choice", 1, "B"},
		{"the last letter", 25, "Z"},
		{"one past the last letter", 26, "#27"},
		{"below zero", -1, "#0"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := choiceLabel(row.choice); got != row.want {
				t.Errorf("choiceLabel(%d) = %q, want %q", row.choice, got, row.want)
			}
		})
	}
}
