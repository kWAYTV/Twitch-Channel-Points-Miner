package notify

import (
	"strings"
	"unicode"

	"github.com/kWAYTV/twitch-points-miner/internal/textcase"
)

// describeGain covers "+50 → Alice (12.4k points) - Reason: WATCH": points the
// account gained or lost on a channel.
func describeGain(head, game, streak string) (Message, bool) {
	arrow := strings.Index(head, "→")
	if arrow < 0 {
		arrow = strings.Index(head, "->")
	}
	if arrow <= 0 {
		return Message{}, false
	}
	sep := "→"
	if head[arrow] == '-' {
		sep = "->"
	}
	amount := strings.TrimSpace(head[:arrow])
	if amount == "" || !isSignedAmount(amount) {
		return Message{}, false
	}
	rest := strings.TrimSpace(head[arrow+len(sep):])
	reason := ""
	if i := indexFold(rest, " - reason:"); i >= 0 {
		reason = strings.TrimSpace(rest[i+len(" - reason:"):])
		rest = strings.TrimSpace(rest[:i])
	}
	name, points := splitPoints(rest)
	if name == "" {
		return Message{}, false
	}

	n, lost := unsignedAmount(amount)
	sign := "+"
	if lost {
		sign = "−"
	}
	details := gainDetails(reason, game, streak)
	details = append(details, balance(points))
	return newMessage(sign+n+" points on "+name, details...), true
}

// gainDetails says why the points changed, and what was being played or how
// long the watch streak is when the line carried that.
func gainDetails(reason, game, streak string) []string {
	kind := ""
	if fields := strings.Fields(reason); len(fields) > 0 {
		kind = strings.ToUpper(fields[0])
	}
	playingGame := playing(game)
	streakLength := watchStreak(streak)
	switch kind {
	case "WATCH":
		if game != "" {
			return []string{"Watching " + game, streakLength}
		}
		return []string{"Watching", streakLength}
	case "WATCH_STREAK":
		if streak != "" {
			return []string{streakLength, playingGame}
		}
		return []string{"Watch streak", playingGame}
	case "CLAIM":
		return []string{"Bonus claim", streakLength}
	case "RAID":
		return []string{"Raid", playingGame, streakLength}
	case "":
		return []string{playingGame, streakLength}
	default:
		label := strings.ToLower(strings.ReplaceAll(kind, "_", " "))
		return []string{textcase.UpperFirst(label), playingGame, streakLength}
	}
}

func unsignedAmount(amount string) (string, bool) {
	amount = strings.TrimSpace(amount)
	lost := strings.HasPrefix(amount, "-") || strings.HasPrefix(amount, "−")
	return strings.TrimLeft(amount, "+-−"), lost
}

func isSignedAmount(s string) bool {
	r := []rune(s)
	if len(r) < 2 {
		return false
	}
	if r[0] != '+' && r[0] != '-' && r[0] != '−' {
		return false
	}
	return unicode.IsDigit(r[1])
}
