package notify

import (
	"testing"
)

type split struct {
	before, after string
	ok            bool
}

func TestStripLeadingEmoji(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"a line with a glyph in front", "🥳 Alice is Online!", "Alice is Online!"},
		{"a glyph with a variation selector", "▶️ Start session", "Start session"},
		{"two glyphs joined", "👨‍💻 Alice", "Alice"},
		{"a glyph with no space after it", "🥳Alice", "Alice"},
		{"a line without a glyph", "Alice is Online!", "Alice is Online!"},
		{"a glyph in the middle", "Alice 🥳 is Online!", "Alice 🥳 is Online!"},
		{"a plus sign in front", "+50 → Alice", "+50 → Alice"},
		{"a space before the glyph", " 🥳 Alice", " 🥳 Alice"},
		{"only a glyph", "🥳", ""},
		{"an empty line", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripLeadingEmoji(tc.in); got != tc.want {
				t.Errorf("stripLeadingEmoji(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCollapseSpace(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"single spaces", "Alice is online", "Alice is online"},
		{"a run of spaces", "Alice   is online", "Alice is online"},
		{"a tab and a line break", "Alice\tis\nonline", "Alice is online"},
		{"spaces at both ends", "  Alice is online  ", "Alice is online"},
		{"only spaces", "   ", ""},
		{"an empty string", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := collapseSpace(tc.in); got != tc.want {
				t.Errorf("collapseSpace(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIndexFold(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		sub  string
		want int
	}{
		{"the same case", "nick wrote: hi", " wrote: ", 4},
		{"another case", "nick WROTE: hi", " wrote: ", 4},
		{"found twice", "a wrote: b wrote: c", " wrote: ", 1},
		{"after a letter outside ASCII", "é wrote: hi", " wrote: ", 2},
		{"not found", "nick said hi", " wrote: ", -1},
		{"an empty needle", "nick", "", 0},
		{"an empty string", "", " wrote: ", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := indexFold(tc.s, tc.sub); got != tc.want {
				t.Errorf("indexFold(%q, %q) = %d, want %d", tc.s, tc.sub, got, tc.want)
			}
		})
	}

	t.Run("after a letter whose lower case is shorter", func(t *testing.T) {
		if got, want := indexFold("İ wrote: hi", " wrote: "), 2; got != want {
			t.Errorf(`indexFold("İ wrote: hi", " wrote: ") = %d, want %d`, got, want)
		}
	})
}

func TestSplitLastFold(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		sep  string
		want split
	}{
		{"one separator", "Yes for bob", " for ", split{"Yes", "bob", true}},
		{"a separator in another case", "Yes FOR bob", " for ", split{"Yes", "bob", true}},
		{"two separators", "Yes for sure for bob", " for ", split{"Yes for sure", "bob", true}},
		{"spaces around both halves", "Yes   for   bob", " for ", split{"Yes", "bob", true}},
		{"nothing before the separator", " for bob", " for ", split{"", "bob", true}},
		{"nothing after the separator", "Yes for ", " for ", split{"Yes", "", true}},
		{"no separator", "Yes bob", " for ", split{"", "", false}},
		{"an empty string", "", " for ", split{"", "", false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, after, ok := splitLastFold(tc.s, tc.sep)
			if got := (split{before, after, ok}); got != tc.want {
				t.Errorf("splitLastFold(%q, %q) = %+v, want %+v", tc.s, tc.sep, got, tc.want)
			}
		})
	}

	t.Run("after a letter whose lower case is shorter", func(t *testing.T) {
		before, after, ok := splitLastFold("İyi (BLUE) for bob", " for ")
		if got, want := (split{before, after, ok}), (split{"İyi (BLUE)", "bob", true}); got != want {
			t.Errorf(`splitLastFold("İyi (BLUE) for bob", " for ") = %+v, want %+v`, got, want)
		}
	})
}

func TestSplitOnceFold(t *testing.T) {
	for _, tc := range []struct {
		name   string
		s      string
		prefix string
		sep    string
		want   split
	}{
		{"a prefix and a separator", "Skip bet for Alice: no outcome", "skip bet for ", ": ", split{"Alice", "no outcome", true}},
		{"a prefix in another case", "SKIP BET FOR Alice: no outcome", "skip bet for ", ": ", split{"Alice", "no outcome", true}},
		{"a separator in another case", "Joining raid from Alice TO Bob", "joining raid from ", " to ", split{"Alice", "Bob", true}},
		{"two separators", "Skip bet for Alice: status: locked", "skip bet for ", ": ", split{"Alice", "status: locked", true}},
		{"only spaces after the separator", "Skip bet for Alice: ", "skip bet for ", ": ", split{"", "", false}},
		{"nothing between the prefix and the separator", "Skip bet for : no outcome", "skip bet for ", ": ", split{"", "", false}},
		{"no separator", "Skip bet for Alice", "skip bet for ", ": ", split{"", "", false}},
		{"another prefix", "Place bet for Alice: no outcome", "skip bet for ", ": ", split{"", "", false}},
		{"shorter than the prefix", "Skip", "skip bet for ", ": ", split{"", "", false}},
		{"an empty string", "", "skip bet for ", ": ", split{"", "", false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, after, ok := splitOnceFold(tc.s, tc.prefix, tc.sep)
			if got := (split{before, after, ok}); got != tc.want {
				t.Errorf("splitOnceFold(%q, %q, %q) = %+v, want %+v", tc.s, tc.prefix, tc.sep, got, tc.want)
			}
		})
	}
}

func TestCutSuffixFold(t *testing.T) {
	type cut struct {
		before string
		ok     bool
	}
	for _, tc := range []struct {
		name   string
		s      string
		suffix string
		want   cut
	}{
		{"the suffix as written", "Alice claimed moment", " claimed moment", cut{"Alice", true}},
		{"the suffix in another case", "Alice Claimed Moment", " claimed moment", cut{"Alice", true}},
		{"a name outside ASCII", "İnci Claimed Moment", " claimed moment", cut{"İnci", true}},
		{"only the suffix", " Claimed Moment", " claimed moment", cut{"", true}},
		{"the suffix in the middle", "Alice claimed moment today", " claimed moment", cut{"", false}},
		{"no suffix", "Alice is online", " claimed moment", cut{"", false}},
		{"an empty string", "", " claimed moment", cut{"", false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, ok := cutSuffixFold(tc.s, tc.suffix)
			if got := (cut{before, ok}); got != tc.want {
				t.Errorf("cutSuffixFold(%q, %q) = %+v, want %+v", tc.s, tc.suffix, got, tc.want)
			}
		})
	}
}

func TestReplaceFold(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		old  string
		new  string
		want string
	}{
		{"one match", "WIN, gained: +500", ", gained: ", " ", "WIN +500"},
		{"a match in another case", "WIN, Gained: +500", ", gained: ", " ", "WIN +500"},
		{"two matches", "WIN, Gained: +500, GAINED: +1", ", gained: ", " ", "WIN +500 +1"},
		{"a match at the end", "WIN, Gained: ", ", gained: ", " ", "WIN "},
		{"a letter outside ASCII before the match", "é, Gained: +500", ", gained: ", " ", "é +500"},
		{"no match", "WIN, Lost: -500", ", gained: ", " ", "WIN, Lost: -500"},
		{"an empty string", "", ", gained: ", " ", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := replaceFold(tc.s, tc.old, tc.new); got != tc.want {
				t.Errorf("replaceFold(%q, %q, %q) = %q, want %q", tc.s, tc.old, tc.new, got, tc.want)
			}
		})
	}

	t.Run("after a letter whose lower case is shorter", func(t *testing.T) {
		if got, want := replaceFold("İ, Gained: +500", ", gained: ", " "), "İ +500"; got != want {
			t.Errorf(`replaceFold("İ, Gained: +500", ", gained: ", " ") = %q, want %q`, got, want)
		}
	})
}

func TestLimitRunes(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"one rune under the limit", "abcd", 5, "abcd"},
		{"exactly the limit", "abcde", 5, "abcde"},
		{"one rune over the limit", "abcdef", 5, "abcd…"},
		{"letters outside ASCII over the limit", "éèêëàâ", 5, "éèêë…"},
		{"a space where the cut falls", "abc def", 5, "abc…"},
		{"a limit of one", "abc", 1, "…"},
		{"one rune and a limit of one", "a", 1, "a"},
		{"a limit of zero", "abc", 0, "abc"},
		{"a negative limit", "abc", -1, "abc"},
		{"an empty string", "", 5, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := limitRunes(tc.s, tc.n); got != tc.want {
				t.Errorf("limitRunes(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
			}
		})
	}
}
