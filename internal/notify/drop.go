package notify

import (
	"regexp"
	"strings"
)

var reDropBar = regexp.MustCompile(`(?i)^DROP_PROGRESS\s+\[[#\-]*\]\s*(\d+%)\s*(?:\|\s*(.*))?$`)

// describeDropClaim covers "Claim Cool Hat (Summer Drops) 60/60 (100%)".
func describeDropClaim(head string) (Message, bool) {
	const prefix = "claim "
	if !strings.HasPrefix(strings.ToLower(head), prefix) {
		return Message{}, false
	}
	rest := strings.TrimSpace(head[len(prefix):])
	open := strings.Index(rest, " (")
	if open <= 0 {
		return Message{}, false
	}
	closeRel := strings.Index(rest[open:], ") ")
	if closeRel < 0 {
		return Message{}, false
	}
	reward := strings.TrimSpace(rest[:open])
	campaign := strings.TrimSpace(rest[open+2 : open+closeRel])
	progress := strings.TrimSpace(rest[open+closeRel+2:])
	if reward == "" || campaign == "" || progress == "" {
		return Message{}, false
	}
	return newMessage("Claimed "+reward, campaign+" campaign", "Progress "+progress), true
}

// describeDropProgress covers "DROP_PROGRESS [####----] 40% | Game / Campaign / channel / 12m of 30m / ...".
func describeDropProgress(head string, other []string) (Message, bool) {
	m := reDropBar.FindStringSubmatch(head)
	if m == nil {
		return Message{}, false
	}
	desc := strings.TrimSpace(m[2])
	if desc == "" {
		desc = strings.Join(other, " | ")
	}
	if i := strings.Index(strings.ToLower(desc), " / remaining "); i >= 0 {
		desc = desc[:i]
	}
	desc = strings.TrimSpace(desc)
	headline := "Drop " + m[1] + " complete"
	if i := strings.LastIndex(desc, " / "); i > 0 {
		tail := strings.TrimSpace(desc[i+3:])
		if watched, total, ok := strings.Cut(tail, " of "); ok && strings.Contains(tail, "m") {
			return newMessage(headline, strings.TrimSpace(desc[:i]), strings.TrimSpace(watched)+" watched of "+strings.TrimSpace(total)), true
		}
	}
	return newMessage(headline, desc), true
}
