package update

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"a plain version", "1.6.0", "1.6.0"},
		{"a leading v", "v1.6.0", "1.6.0"},
		{"spaces around it", " v1.6.0 ", "1.6.0"},
		{"commits since the tag", "v1.6.0-2-gabc1234", "1.6.0"},
		{"a pre-release", "1.6.0-rc1", "1.6.0"},
		{"a word after a space", "1.6.0 beta", "1.6.0"},
		{"build metadata", "1.6.0+build5", "1.6.0"},
		{"a pre-release with build metadata", "1.6.0-rc1+build5", "1.6.0"},
		{"a leading dash", "-1.6.0", "-1.6.0"},
		{"a leading plus", "+1.6.0", "+1.6.0"},
		{"two leading vs", "vv1.6.0", "v1.6.0"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeVersion(tt.raw); got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"the same version", "1.5.0", "1.5.0", 0},
		{"a newer major", "2.0.0", "1.9.9", 1},
		{"an older major", "1.9.9", "2.0.0", -1},
		{"a newer minor", "1.6.0", "1.5.0", 1},
		{"an older minor", "1.5.0", "1.6.0", -1},
		{"a newer patch", "1.5.2", "1.5.1", 1},
		{"an older patch", "1.5.1", "1.5.2", -1},
		{"two digits against one", "1.5.10", "1.5.9", 1},
		{"a missing patch against zero", "1.5", "1.5.0", 0},
		{"a missing patch against one", "1.5", "1.5.1", -1},
		{"a longer version against a shorter older one", "1.6.0", "1.5", 1},
		{"a part that is not a number", "1.x.0", "1.0.0", 0},
		{"parts in spaces", "1. 5 .0", "1.5.0", 0},
		{"a negative part", "1.-1.0", "1.0.0", -1},
		{"empty against zero", "", "0", 0},
		{"both empty", "", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareVersions(tt.a, tt.b); got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
