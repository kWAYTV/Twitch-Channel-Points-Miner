package chat

import (
	"fmt"
	"strings"
)

// isAuthFailure reports whether a line from Twitch says the login was refused.
func isAuthFailure(line string) bool {
	return strings.Contains(strings.ToLower(line), "authentication failed")
}

// parsePrivmsg splits a chat message line into who wrote it and what they
// wrote. ok is false for a line that carries no chat message.
func parsePrivmsg(line string) (nick, msg string, ok bool) {
	if !strings.Contains(line, "PRIVMSG") {
		return "", "", false
	}
	parts := strings.SplitN(line, " :", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	msg = parts[1]
	prefix := parts[0]
	if strings.HasPrefix(prefix, ":") {
		prefix = strings.TrimPrefix(prefix, ":")
		if idx := strings.Index(prefix, "!"); idx > 0 {
			nick = prefix[:idx]
		}
	}
	return nick, msg, true
}

// mentions reports whether msg names username, which is in lower case: as
// @username, or, when bareName is set, as the name alone.
func mentions(msg, username string, bareName bool) bool {
	msgLower := strings.ToLower(msg)
	target := "@" + username
	mentioned := strings.Contains(msgLower, target)
	if bareName && !mentioned {
		mentioned = strings.Contains(msgLower, username)
	}
	return mentioned
}

// mentionLine is what the log says about a mention.
func mentionLine(nick, channel, msg string) string {
	displayNick := nick
	if displayNick == "" {
		displayNick = "unknown"
	}
	return fmt.Sprintf("%s at #%s wrote: %s", displayNick, channel, msg)
}
