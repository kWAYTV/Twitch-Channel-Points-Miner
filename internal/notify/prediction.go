package notify

import (
	"strings"
)

// describeBetResult covers "EventPrediction: bob - title - Decision: Yes - Result: WIN, Gained: +500".
func describeBetResult(head string) (Message, bool) {
	const prefix = "eventprediction:"
	if !strings.HasPrefix(strings.ToLower(head), prefix) {
		return Message{}, false
	}
	rest := strings.TrimSpace(head[len(prefix):])
	var name, title, choice, result string
	for _, part := range strings.Split(rest, " - ") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		switch {
		case strings.HasPrefix(lower, "decision:"):
			choice = shortOutcome(strings.TrimSpace(part[len("decision:"):]))
		case strings.HasPrefix(lower, "result:"):
			result = strings.TrimSpace(part[len("result:"):])
			result = replaceFold(result, ", gained: ", " ")
			result = replaceFold(result, ", lost: ", " ")
			result = replaceFold(result, ", refunded: ", " ")
			result = collapseSpace(result)
		case name == "":
			name = part
		case title == "":
			title = part
		default:
			title = strings.TrimSpace(title + " " + part)
		}
	}
	if name == "" {
		return Message{}, false
	}

	kind, pts := splitResult(result)
	var headline string
	var details []string
	switch {
	case kind == "WIN" && pts != "":
		headline = "Won " + pts + " points on " + name
	case kind == "WIN":
		headline = "Won the bet on " + name
	case kind == "LOSE" && pts != "":
		headline = "Lost " + pts + " points on " + name
	case kind == "LOSE":
		headline = "Lost the bet on " + name
	case kind == "REFUND" && pts != "":
		headline = "Refunded " + pts + " points on " + name
	case kind == "REFUND":
		headline = "Bet refunded on " + name
	default:
		headline = "Bet finished on " + name
		details = append(details, result)
	}
	if title != "" {
		details = append(details, `"`+title+`"`)
	}
	if choice != "" {
		details = append(details, "Chose "+choice)
	}
	return newMessage(headline, details...), true
}

func splitResult(result string) (kind, pts string) {
	fields := strings.Fields(result)
	if len(fields) == 0 {
		return "", ""
	}
	kind = strings.ToUpper(fields[0])
	switch kind {
	case "WIN", "LOSE", "REFUND":
	default:
		return "", result
	}
	if len(fields) >= 2 {
		pts = strings.TrimLeft(fields[len(fields)-1], "+-−")
	}
	return kind, pts
}

// describeBetPlaced covers "Place 1,234 points on: Yes ... for bob".
func describeBetPlaced(head string) (Message, bool) {
	lower := strings.ToLower(head)
	amount := ""
	var rest string
	switch {
	case strings.HasPrefix(lower, "place bet on:"):
		rest = strings.TrimSpace(head[len("place bet on:"):])
	case strings.HasPrefix(lower, "place ") && strings.Contains(lower, " points on:"):
		i := strings.Index(lower, " points on:")
		amount = strings.TrimSpace(head[len("place "):i])
		rest = strings.TrimSpace(head[i+len(" points on:"):])
	default:
		return Message{}, false
	}
	outcome, who, ok := splitLastFold(rest, " for ")
	if !ok {
		return Message{}, false
	}
	outcome = shortOutcome(outcome)
	detail := ""
	switch {
	case amount != "" && outcome != "":
		detail = amount + " points on " + outcome
	case outcome != "":
		detail = "On " + outcome
	case amount != "":
		detail = amount + " points"
	}
	return newMessage("Bet placed on "+who, detail), true
}

// describeBetNotice covers a bet that is scheduled, skipped or failed.
func describeBetNotice(head string) (Message, bool) {
	lower := strings.ToLower(head)
	if strings.HasPrefix(lower, "place bet after ") {
		rest := strings.TrimSpace(head[len("place bet after "):])
		if when, who, ok := splitLastFold(rest, " for "); ok {
			return newMessage("Bet on " + who + " in " + when), true
		}
	}
	if who, reason, ok := splitOnceFold(head, "skip bet for ", ": "); ok {
		return newMessage("Bet skipped on "+who, skipReason(reason)), true
	}
	if name, detail, ok := splitOnceFold(head, "prediction ", ": "); ok {
		return newMessage("Failed prediction on "+name, strings.TrimRight(strings.TrimSpace(detail), ".")), true
	}
	return Message{}, false
}

func skipReason(reason string) string {
	reason = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(reason), "_", " "))
	switch {
	case reason == "":
		return ""
	case strings.Contains(reason, "no outcome"):
		return "No outcome was selected"
	case strings.Contains(reason, "twitch"):
		return "The stake is below Twitch's minimum"
	case strings.Contains(reason, "minimum"):
		return "The balance is below the minimum"
	case strings.Contains(reason, "filter"):
		return "The bet filter did not match"
	case strings.Contains(reason, "status"):
		return "The prediction is no longer active"
	default:
		return strings.ToUpper(reason[:1]) + reason[1:]
	}
}
