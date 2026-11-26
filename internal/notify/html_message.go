package notify

import (
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// htmlMessage renders a message as HTML of at most limit runes: the polarity
// glyph and the headline in bold, then lineBreak and the details on one line.
// The glyph is the one the console uses, and it is sent even when console
// emojis are turned off.
func htmlMessage(event events.Event, message Message, lineBreak string, limit int) string {
	text := eventGlyphPrefix(event) + "<b>" + htmlEscape(headlineText(event, message)) + "</b>"
	details := htmlEscape(detailLine(message))
	budget := limit - len([]rune(text)) - len([]rune(lineBreak))
	if details == "" || budget < 1 {
		return text
	}
	return text + lineBreak + limitEscaped(details, budget)
}

// headlineText is the headline of a message, cut to the length every service accepts.
func headlineText(event events.Event, message Message) string {
	return limitRunes(headlineOf(event, message), headlineLimit)
}

// detailLine is the details of a message on one line.
func detailLine(message Message) string {
	return strings.Join(message.Details, detailSeparator)
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// limitEscaped truncates an HTML-escaped string without splitting an entity.
func limitEscaped(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 1 {
		return ""
	}
	cut := string(r[:n-1])
	if i := strings.LastIndex(cut, "&"); i >= 0 && !strings.Contains(cut[i:], ";") {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "…"
}
