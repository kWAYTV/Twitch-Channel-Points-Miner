package points

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		want   string
	}{
		{"zero", 0, "0"},
		{"under a thousand", 950, "950"},
		{"one under a thousand", 999, "999"},
		{"exactly a thousand", 1_000, "1k"},
		{"a thousand and a half", 1_500, "1.5k"},
		{"thousands with two decimals", 1_250, "1.25k"},
		{"thousands rounded to two decimals", 1_234, "1.23k"},
		{"the last amount shown in thousands", 999_994, "999.99k"},
		{"the first amount shown in millions", 999_995, "1M"},
		{"exactly a million", 1_000_000, "1M"},
		{"millions with a decimal", 2_500_000, "2.5M"},
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
