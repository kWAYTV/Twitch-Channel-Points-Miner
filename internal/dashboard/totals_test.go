package dashboard

import "testing"

func TestTotalsOf(t *testing.T) {
	tests := []struct {
		name string
		list []StreamerInfo
		want totals
	}{
		{"no streamers", nil, totals{}},
		{
			"one streamer, offline",
			[]StreamerInfo{{Name: "alice", Points: 1000, SessionGain: 10}},
			totals{Points: 1000, SessionGain: 10, Online: 0, Count: 1},
		},
		{
			"one streamer, online",
			[]StreamerInfo{{Name: "alice", Points: 1000, SessionGain: 10, Online: true}},
			totals{Points: 1000, SessionGain: 10, Online: 1, Count: 1},
		},
		{
			"three streamers, two online, one with a loss",
			[]StreamerInfo{
				{Name: "alice", Points: 1000, SessionGain: 10, Online: true},
				{Name: "bob", Points: 300, SessionGain: -200},
				{Name: "carol", Points: 0, SessionGain: 0, Online: true},
			},
			totals{Points: 1300, SessionGain: -190, Online: 2, Count: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := totalsOf(tt.list); got != tt.want {
				t.Errorf("totalsOf(%+v) = %+v, want %+v", tt.list, got, tt.want)
			}
		})
	}
}
