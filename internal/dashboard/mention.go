package dashboard

import "strings"

// mentionedStreamer returns the known streamer whose name appears first in the
// message as a whole word (longest name wins a tie), or "".
func mentionedStreamer(message string, known map[string]*streamerState) string {
	msg := strings.ToLower(message)
	best, bestAt := "", -1
	for name := range known {
		at := indexWord(msg, name)
		if at < 0 {
			continue
		}
		if bestAt < 0 || at < bestAt || (at == bestAt && len(name) > len(best)) {
			best, bestAt = name, at
		}
	}
	return best
}

func indexWord(msg, word string) int {
	for from := 0; from < len(msg); {
		i := strings.Index(msg[from:], word)
		if i < 0 {
			return -1
		}
		i += from
		end := i + len(word)
		if (i == 0 || !isNameByte(msg[i-1])) && (end == len(msg) || !isNameByte(msg[end])) {
			return i
		}
		from = i + 1
	}
	return -1
}
