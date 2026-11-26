package project

import "testing"

func TestLabelOf(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"a release version", "1.6.0", "v1.6.0"},
		{"a build with no version stamped in", "dev", "built from source"},
		{"an empty version", "", "v"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := labelOf(tt.version); got != tt.want {
				t.Errorf("labelOf(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
