package prediction

import "testing"

func TestFormatNumber(t *testing.T) {
	for _, row := range []struct {
		name  string
		value int
		want  string
	}{
		{"zero", 0, "0"},
		{"a small number", 5, "5"},
		{"a small negative number", -5, "-5"},
		{"one below a thousand", 999, "999"},
		{"a thousand and a half", 1_500, "1.5k"},
		{"a thousand with cents", 1_234, "1.23k"},
		{"negative thousands", -1_500, "-1.5k"},
		{"one below a million", 999_999, "1M"},
		{"millions", 2_340_000, "2.34M"},
		{"negative millions", -2_500_000, "-2.5M"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := FormatNumber(row.value); got != row.want {
				t.Errorf("FormatNumber(%d) = %q, want %q", row.value, got, row.want)
			}
		})
	}

	t.Run("exactly a thousand", func(t *testing.T) {
		if got := FormatNumber(1_000); got != "1k" {
			t.Errorf("FormatNumber(1000) = %q, want %q", got, "1k")
		}
	})

	t.Run("exactly a million", func(t *testing.T) {
		if got := FormatNumber(1_000_000); got != "1M" {
			t.Errorf("FormatNumber(1000000) = %q, want %q", got, "1M")
		}
	})
}

func TestFormatFloat(t *testing.T) {
	for _, row := range []struct {
		name  string
		value float64
		want  string
	}{
		{"zero", 0, "0"},
		{"a whole number", 25, "25"},
		{"a whole number that ends in zero", 100, "100"},
		{"one decimal", 1.5, "1.5"},
		{"two decimals", 2.25, "2.25"},
		{"a third decimal that rounds down", 2.254, "2.25"},
		{"a third decimal that rounds up", 2.256, "2.26"},
		{"too small to show", 0.001, "0"},
		{"a negative number", -1.5, "-1.5"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := FormatFloat(row.value); got != row.want {
				t.Errorf("FormatFloat(%v) = %q, want %q", row.value, got, row.want)
			}
		})
	}

	t.Run("a negative number too small to show", func(t *testing.T) {
		if got := FormatFloat(-0.001); got != "0" {
			t.Errorf("FormatFloat(-0.001) = %q, want %q", got, "0")
		}
	})
}

func TestTrimZeros(t *testing.T) {
	for _, row := range []struct {
		name string
		val  string
		want string
	}{
		{"no zeros to trim", "1.25", "1.25"},
		{"one trailing zero", "1.50", "1.5"},
		{"only zeros after the point", "3.00", "3"},
		{"nothing but zeros", "0.00", "0"},
		{"an empty string", "", "0"},
		{"a lone minus sign", "-", "0"},
		{"a whole number with no point", "100", "1"},
		{"a suffix after the zeros", "1.00k", "1.00k"},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := trimZeros(row.val); got != row.want {
				t.Errorf("trimZeros(%q) = %q, want %q", row.val, got, row.want)
			}
		})
	}
}
