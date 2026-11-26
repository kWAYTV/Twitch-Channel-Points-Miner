package random

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// repeated is a random source that yields the same byte forever.
func repeated(b byte) *bytes.Reader {
	return bytes.NewReader(bytes.Repeat([]byte{b}, 256))
}

func TestString(t *testing.T) {
	t.Run("ten characters", func(t *testing.T) {
		shape := regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)
		if got := String(10); !shape.MatchString(got) {
			t.Errorf("String(10) = %q, want it to match %s", got, shape)
		}
	})

	t.Run("no characters", func(t *testing.T) {
		if got := String(0); got != "" {
			t.Errorf("String(0) = %q, want %q", got, "")
		}
	})

	t.Run("a negative length", func(t *testing.T) {
		if got := String(-1); got != "" {
			t.Errorf("String(-1) = %q, want %q", got, "")
		}
	})
}

func TestStringFrom(t *testing.T) {
	tests := []struct {
		name   string
		source byte
		length int
		want   string
	}{
		{"the lowest draw", 0, 3, "aaa"},
		{"a draw in the capitals", 26, 3, "AAA"},
		{"the highest draw", 61, 3, "999"},
		{"one character", 1, 1, "b"},
		{"no characters", 1, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringFrom(repeated(tt.source), tt.length); got != tt.want {
				t.Errorf("stringFrom(%d..., %d) = %q, want %q", tt.source, tt.length, got, tt.want)
			}
		})
	}
}

func TestHex(t *testing.T) {
	t.Run("sixteen bytes", func(t *testing.T) {
		shape := regexp.MustCompile(`^[0-9a-f]{32}$`)
		if got := Hex(16); !shape.MatchString(got) {
			t.Errorf("Hex(16) = %q, want it to match %s", got, shape)
		}
	})
}

func TestHexFrom(t *testing.T) {
	tests := []struct {
		name   string
		source byte
		length int
		want   string
	}{
		{"two bytes", 0xab, 2, "abab"},
		{"one byte", 0x0f, 1, "0f"},
		{"no bytes", 0xab, 0, ""},
		{"a negative length", 0xab, -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hexFrom(repeated(tt.source), tt.length); got != tt.want {
				t.Errorf("hexFrom(%#x..., %d) = %q, want %q", tt.source, tt.length, got, tt.want)
			}
		})
	}

	t.Run("a source that runs dry", func(t *testing.T) {
		shape := regexp.MustCompile(`^[a-zA-Z0-9]{4}$`)
		if got := hexFrom(strings.NewReader(""), 4); !shape.MatchString(got) {
			t.Errorf("hexFrom(empty, 4) = %q, want it to match %s", got, shape)
		}
	})
}

func TestInt(t *testing.T) {
	t.Run("a range of one number", func(t *testing.T) {
		if got := Int(7, 7); got != 7 {
			t.Errorf("Int(7, 7) = %d, want 7", got)
		}
	})

	t.Run("a range of two numbers", func(t *testing.T) {
		if got := Int(7, 8); got != 7 && got != 8 {
			t.Errorf("Int(7, 8) = %d, want 7 or 8", got)
		}
	})
}

func TestIntFrom(t *testing.T) {
	tests := []struct {
		name     string
		source   byte
		min, max int
		want     int
	}{
		{"the lowest draw", 0x00, 1, 4, 1},
		{"a draw in the middle", 0x02, 1, 4, 3},
		{"the highest draw", 0xff, 1, 4, 4},
		{"a range of one number", 0xff, 3, 3, 3},
		{"max below min", 0xff, 3, 2, 3},
		{"a range of two numbers", 0xff, 3, 4, 4},
		{"a range below zero", 0x00, -4, -1, -4},
		{"a range across zero", 0xff, -2, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intFrom(repeated(tt.source), tt.min, tt.max); got != tt.want {
				t.Errorf("intFrom(%#x..., %d, %d) = %d, want %d", tt.source, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
