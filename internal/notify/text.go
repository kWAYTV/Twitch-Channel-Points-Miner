package notify

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ansiCodes matches the colour escapes the console logger adds to event lines.
var ansiCodes = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripLeadingEmoji drops the glyph the console logger puts in front of a line.
// Every glyph it uses is a "symbol, other" rune, optionally followed by a
// variation selector or joined to another one.
func stripLeadingEmoji(s string) string {
	rest := strings.TrimLeftFunc(s, func(r rune) bool {
		return unicode.Is(unicode.So, r) || r == '️' || r == '‍'
	})
	if len(rest) == len(s) {
		return s
	}
	return strings.TrimSpace(rest)
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// The Fold functions ignore case. They measure in the bytes of s itself, never
// of a lower-cased copy, because a letter's lower case can be shorter in bytes
// than the letter.

// foldPrefixLen returns how many bytes at the start of s spell prefix, or -1.
func foldPrefixLen(s, prefix string) int {
	n := 0
	for _, want := range prefix {
		if n >= len(s) {
			return -1
		}
		got, size := utf8.DecodeRuneInString(s[n:])
		if unicode.ToLower(got) != unicode.ToLower(want) {
			return -1
		}
		n += size
	}
	return n
}

// findFold returns where sub first starts in s and how many bytes it takes there.
func findFold(s, sub string) (index, length int) {
	for i := range s {
		if n := foldPrefixLen(s[i:], sub); n >= 0 {
			return i, n
		}
	}
	if sub == "" {
		return len(s), 0
	}
	return -1, 0
}

// findLastFold is findFold for the last place sub starts in s.
func findLastFold(s, sub string) (index, length int) {
	if sub == "" {
		return len(s), 0
	}
	index = -1
	for i := range s {
		if n := foldPrefixLen(s[i:], sub); n >= 0 {
			index, length = i, n
		}
	}
	return index, length
}

func indexFold(s, sub string) int {
	i, _ := findFold(s, sub)
	return i
}

func splitLastFold(s, sep string) (string, string, bool) {
	i, n := findLastFold(s, sep)
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+n:]), true
}

func splitOnceFold(s, prefix, sep string) (string, string, bool) {
	start := foldPrefixLen(s, prefix)
	if start < 0 {
		return "", "", false
	}
	rest := strings.TrimSpace(s[start:])
	i, n := findFold(rest, sep)
	if i <= 0 {
		return "", "", false
	}
	return strings.TrimSpace(rest[:i]), strings.TrimSpace(rest[i+n:]), true
}

func cutSuffixFold(s, suffix string) (string, bool) {
	i, n := findLastFold(s, suffix)
	if i < 0 || i+n != len(s) {
		return "", false
	}
	return s[:i], true
}

func replaceFold(s, old, new string) string {
	if old == "" {
		return s
	}
	var b strings.Builder
	for {
		i, n := findFold(s, old)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		b.WriteString(new)
		s = s[i+n:]
	}
}

// capitalise upper-cases the first letter of s.
func capitalise(s string) string {
	first, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	return string(unicode.ToUpper(first)) + s[size:]
}

func limitRunes(s string, n int) string {
	r := []rune(s)
	if n < 1 || len(r) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
