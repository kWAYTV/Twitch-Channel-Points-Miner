package update

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

func TestSourceBuildIsNeverReplaced(t *testing.T) {
	if project.Version != project.DevVersion {
		t.Skip("this test binary has a release version stamped in")
	}
	updated, err := RunAutoUpdate()
	if updated || err != nil {
		t.Fatalf("a build without a version must be left alone, got updated=%v err=%v", updated, err)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.6.0", "1.5", 1},
		{"1.5.0", "1.5", 0},
		{"1.5.1", "1.5.10", -1},
		{"2.0.0", "1.9.9", 1},
	}
	for _, c := range cases {
		if got := compareVersions(normalizeVersion(c.a), normalizeVersion(c.b)); got != c.want {
			t.Errorf("compare %s with %s = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	if got := normalizeVersion("v1.6.0-2-gabc1234"); got != "1.6.0" {
		t.Errorf("normalize = %q, want 1.6.0", got)
	}
}
