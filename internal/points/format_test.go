package points

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		want   string
	}{
		{"zero", 0, "0"},
		{"one digit", 5, "5"},
		{"under a thousand", 950, "950"},
		{"one under a thousand", 999, "999"},
		{"exactly a thousand", 1_000, "1k"},
		{"a thousand and a half", 1_500, "1.5k"},
		{"thousands with two decimals", 1_250, "1.25k"},
		{"thousands rounded to two decimals", 1_234, "1.23k"},
		{"one over a thousand", 1_001, "1k"},
		{"ten thousand", 10_000, "10k"},
		{"the last amount shown in thousands", 999_994, "999.99k"},
		{"the first amount shown in millions", 999_995, "1M"},
		{"one under a million", 999_999, "1M"},
		{"exactly a million", 1_000_000, "1M"},
		{"a million and a half", 1_500_000, "1.5M"},
		{"millions with two decimals", 1_250_000, "1.25M"},
		{"millions rounded to two decimals", 2_340_000, "2.34M"},
		{"millions with a decimal", 2_500_000, "2.5M"},
		{"a negative digit", -5, "-5"},
		{"a negative amount", -950, "-950"},
		{"negative thousands", -1_500, "-1.5k"},
		{"negative millions", -2_500_000, "-2.5M"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Format(tt.amount); got != tt.want {
				t.Errorf("Format(%d) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}
