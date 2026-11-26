// Package textcase changes the case of text a letter at a time.
package textcase

import (
	"unicode"
	"unicode/utf8"
)

// UpperFirst returns s with its first letter in upper case. The letter may be
// more than one byte long.
func UpperFirst(s string) string {
	first, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	return string(unicode.ToUpper(first)) + s[size:]
}
