package dashboard

import "testing"

func TestWindowFor(t *testing.T) {
	tests := []struct {
		name      string
		rangeName string
		now       int64
		want      window
		wantOK    bool
	}{
		{"no range", "", 1_700_000_000_000, window{name: "24h", since: 1_699_913_600_000, now: 1_700_000_000_000}, true},
		{"a day", "24h", 1_700_000_000_000, window{name: "24h", since: 1_699_913_600_000, now: 1_700_000_000_000}, true},
		{"a week", "7d", 1_700_000_000_000, window{name: "7d", since: 1_699_395_200_000, now: 1_700_000_000_000}, true},
		{"thirty days", "30d", 1_700_000_000_000, window{name: "30d", since: 1_697_408_000_000, now: 1_700_000_000_000}, true},
		{"everything", "all", 1_700_000_000_000, window{name: "all", since: 0, now: 1_700_000_000_000}, true},
		{"a range it does not know", "1y", 1_700_000_000_000, window{}, false},
		{"a known range in capitals", "24H", 1_700_000_000_000, window{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := windowFor(tt.rangeName, tt.now)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("windowFor(%q, %d) = %+v, %v, want %+v, %v", tt.rangeName, tt.now, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
