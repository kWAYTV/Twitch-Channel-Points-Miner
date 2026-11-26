package prediction

import (
	"fmt"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/points"
)

type Decision struct {
	Choice    int
	OutcomeID string
	Amount    int
}

type Event struct {
	Streamer      *domain.Streamer
	EventID       string
	Title         string
	Status        string
	CreatedAt     time.Time
	WindowSeconds float64
	Outcomes      []Outcome
	Decision      Decision
	BetPlaced     bool
	BetConfirmed  bool
	ResultType    string
	ResultString  string
}

func NewEvent(streamer *domain.Streamer, event map[string]interface{}) *Event {
	return newEventAt(streamer, event, time.Now())
}

// newEventAt builds the event; now stands in for a missing or unreadable
// created_at.
func newEventAt(streamer *domain.Streamer, event map[string]interface{}, now time.Time) *Event {
	if streamer == nil || event == nil {
		return nil
	}
	eventID, _ := event["id"].(string)
	title, _ := event["title"].(string)
	status := strings.ToUpper(jsonvalue.String(event["status"]))
	window := float64(jsonvalue.Float(event["prediction_window_seconds"]))
	created := now
	if createdStr, ok := event["created_at"].(string); ok && createdStr != "" {
		if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
			created = t
		}
	}
	pe := &Event{
		Streamer:      streamer,
		EventID:       eventID,
		Title:         strings.TrimSpace(title),
		Status:        status,
		CreatedAt:     created,
		WindowSeconds: window,
		BetPlaced:     false,
	}
	rawOutcomes, _ := event["outcomes"].([]interface{})
	pe.UpdateOutcomes(rawOutcomes)
	return pe
}

func (p *Event) UpdateOutcomes(outcomes []interface{}) {
	p.Outcomes = parseOutcomes(outcomes)
}

func (p *Event) ClosingAfter(now time.Time) time.Duration {
	elapsed := now.Sub(p.CreatedAt).Seconds()
	remaining := p.WindowSeconds - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return time.Duration(remaining * float64(time.Second))
}

func (p *Event) ParseResult(result map[string]interface{}) (gained, placed, won int, resultType, resultString string) {
	resultType = strings.ToUpper(jsonvalue.String(result["type"]))
	placed = p.Decision.Amount
	won = jsonvalue.Int(result["points_won"])
	if resultType == "REFUND" {
		placed = 0
		won = 0
	}
	gained = won - placed
	p.ResultType = resultType
	action := "Gained"
	switch resultType {
	case "LOSE":
		action = "Lost"
	case "REFUND":
		action = "Refunded"
	}
	sign := ""
	if gained >= 0 {
		sign = "+"
	}
	resultString = fmt.Sprintf("%s, %s: %s%s", resultType, action, sign, points.Format(gained))
	p.ResultString = resultString
	return
}

// Label names a prediction by who runs it and its title, leaving out
// whichever is missing.
func Label(name, title string) string {
	switch {
	case name != "" && title != "":
		return fmt.Sprintf("EventPrediction: %s - %s", name, title)
	case name != "":
		return fmt.Sprintf("EventPrediction: %s", name)
	case title != "":
		return fmt.Sprintf("EventPrediction: %s", title)
	default:
		return "EventPrediction"
	}
}

func (p *Event) String() string {
	name := ""
	if p.Streamer != nil {
		name = p.Streamer.Username
	}
	return Label(name, p.Title)
}

func (p *Event) DecisionOutcome() *Outcome {
	choice := p.Decision.Choice
	if choice >= 0 && choice < len(p.Outcomes) {
		return &p.Outcomes[choice]
	}
	for i := range p.Outcomes {
		if p.Outcomes[i].ID == p.Decision.OutcomeID {
			return &p.Outcomes[i]
		}
	}
	return nil
}

func (p *Event) DecisionOutcomeString() string {
	if out := p.DecisionOutcome(); out != nil {
		return out.String()
	}
	return p.Decision.OutcomeID
}

func (p *Event) DecisionLabel() string {
	out := p.DecisionOutcome()
	if out == nil {
		if p.Decision.OutcomeID != "" {
			return fmt.Sprintf("%s: %s", choiceLabel(p.Decision.Choice), p.Decision.OutcomeID)
		}
		return ""
	}
	return fmt.Sprintf("%s: %s (%s)", choiceLabel(p.Decision.Choice), out.Title, strings.ToUpper(out.Color))
}

func choiceLabel(choice int) string {
	if choice >= 0 && choice < 26 {
		return string(rune('A' + choice))
	}
	return fmt.Sprintf("#%d", choice+1)
}
