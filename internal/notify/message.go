package notify

import (
	"strconv"
	"strings"
)

// sentenceLimit caps a message written out as sentences. The console line stays complete.
const sentenceLimit = 500

const (
	// headlineLimit keeps a headline within what Discord accepts as an embed title.
	headlineLimit = 256
	// detailSeparator joins the details of a message on one line.
	detailSeparator = " · "
)

// Message is one event told from the account's side: a headline that stands on
// its own, and the details behind it. Points are always the account's points on
// a channel, never the streamer's.
type Message struct {
	Headline string
	Details  []string
}

func newMessage(headline string, details ...string) Message {
	m := Message{Headline: strings.TrimSpace(headline)}
	for _, d := range details {
		if d = strings.TrimSpace(d); d != "" {
			m.Details = append(m.Details, d)
		}
	}
	return m
}

// Sentence writes the message as plain sentences: the headline, then each detail.
// The streamer's name stays in the text, which is how the dashboard attributes it.
func (m Message) Sentence() string {
	parts := make([]string, 0, len(m.Details)+1)
	for _, part := range append([]string{m.Headline}, m.Details...) {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, endSentence(part))
		}
	}
	return limitRunes(strings.Join(parts, " "), sentenceLimit)
}

func endSentence(s string) string {
	if strings.ContainsAny(s[len(s)-1:], ".!?") {
		return s
	}
	return s + "."
}

// Sentence is Describe written out as sentences. It is safe to call on its own
// output, so stored sentences can be passed through it again.
func Sentence(line string) string {
	return Describe(line).Sentence()
}

// Describe turns a console event line into a Message. The console output itself
// is not changed.
func Describe(line string) Message {
	s := strings.TrimSpace(ansiCodes.ReplaceAllString(line, ""))
	s = stripLeadingEmoji(s)
	if s == "" {
		return Message{}
	}
	s = collapseSpace(s)
	head, game, streak, other := splitNotifyPipes(s)

	describers := []func() (Message, bool){
		func() (Message, bool) { return describeSession(head) },
		func() (Message, bool) { return describePresence(head, game, streak, "online") },
		func() (Message, bool) { return describePresence(head, game, streak, "offline") },
		func() (Message, bool) { return describeGain(head, game, streak) },
		func() (Message, bool) { return describeBetResult(head) },
		func() (Message, bool) { return describeBetPlaced(head) },
		func() (Message, bool) { return describeBetNotice(head) },
		func() (Message, bool) { return describeRaid(head) },
		func() (Message, bool) { return describeMoment(head) },
		func() (Message, bool) { return describeDropClaim(head) },
		func() (Message, bool) { return describeDropProgress(head, other) },
		func() (Message, bool) { return describeMention(head) },
	}
	for _, describe := range describers {
		if m, ok := describe(); ok {
			return m
		}
	}
	return describeOther(head, game, streak, other)
}

// describeOther keeps a line no rule recognises as it is, with the game and a
// positive watch streak as details.
func describeOther(head, game, streak string, other []string) Message {
	headline := strings.TrimSpace(head)
	if headline == "" {
		headline = "Something changed"
	}
	lower := strings.ToLower(headline)
	var details []string
	if game != "" && !strings.Contains(lower, strings.ToLower(game)) {
		details = append(details, playing(game))
	}
	if streak != "" && !strings.Contains(lower, "watch streak") {
		details = append(details, watchStreak(streak))
	}
	return newMessage(headline, append(details, other...)...)
}

// splitNotifyPipes keeps the main clause plus the game and a positive streak.
// The raw timestamps from LogExtraInfo are not worth sending.
func splitNotifyPipes(s string) (head, game, streak string, other []string) {
	parts := strings.Split(s, " | ")
	head = strings.TrimSpace(parts[0])
	for _, raw := range parts[1:] {
		p := strings.TrimSpace(raw)
		lower := strings.ToLower(p)
		switch {
		case strings.HasPrefix(lower, "playing:"), strings.HasPrefix(lower, "game:"):
			if v := strings.TrimSpace(p[strings.Index(p, ":")+1:]); v != "" && game == "" {
				game = v
			}
		case strings.HasPrefix(lower, "streak length "):
			n := strings.TrimSpace(p[len("streak length "):])
			if v, err := strconv.Atoi(n); err == nil && v > 0 {
				streak = n
			}
		case strings.HasPrefix(lower, "expires at "),
			strings.HasPrefix(lower, "started at "),
			strings.HasPrefix(lower, "streak achievement"):
			continue
		default:
			if p != "" {
				other = append(other, p)
			}
		}
	}
	return head, game, streak, other
}

// splitPoints separates "Alice (1.2k points)" into the name and the balance.
func splitPoints(subject string) (name, points string) {
	open := strings.LastIndex(subject, "(")
	if open <= 0 || !strings.HasSuffix(subject, ")") {
		return subject, ""
	}
	inner := subject[open+1 : len(subject)-1]
	if !strings.Contains(strings.ToLower(inner), "point") {
		return subject, ""
	}
	return strings.TrimSpace(subject[:open]), strings.TrimSpace(inner)
}

// balance words the account's points on a channel.
func balance(points string) string {
	if points == "" {
		return ""
	}
	return "Balance " + points
}

// playing words the game a channel is on.
func playing(game string) string {
	if game == "" {
		return ""
	}
	return "Playing " + game
}

// watchStreak words how long the account's watch streak on a channel is.
func watchStreak(length string) string {
	if length == "" {
		return ""
	}
	return "Watch streak " + length
}
