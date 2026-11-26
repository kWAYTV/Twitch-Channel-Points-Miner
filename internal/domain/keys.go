package domain

import "strings"

// LoginKey is the form a Twitch login is compared and looked up in: without
// surrounding spaces, in lower case.
func LoginKey(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

// GameKey is the form a game's name is compared and looked up in: without
// surrounding spaces, in lower case.
func GameKey(game string) string {
	return strings.ToLower(strings.TrimSpace(game))
}
