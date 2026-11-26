package domain

import "testing"

func TestStreamer_HasActiveMultipliers(t *testing.T) {
	t.Run("no multipliers", func(t *testing.T) {
		streamer := &Streamer{}
		if got := streamer.HasActiveMultipliers(); got != false {
			t.Errorf("HasActiveMultipliers() = %t, want false", got)
		}
	})

	t.Run("one multiplier", func(t *testing.T) {
		streamer := &Streamer{ActiveMultipliers: []ActiveMultiplier{{Factor: 1.5}}}
		if got := streamer.HasActiveMultipliers(); got != true {
			t.Errorf("HasActiveMultipliers() = %t, want true", got)
		}
	})
}

func TestStreamer_TotalMultiplier(t *testing.T) {
	for _, row := range []struct {
		name        string
		multipliers []ActiveMultiplier
		want        float64
	}{
		{"no multipliers", nil, 0},
		{"one multiplier", []ActiveMultiplier{{Factor: 1.5}}, 1.5},
		{"two multipliers", []ActiveMultiplier{{Factor: 1.5}, {Factor: 2}}, 3.5},
	} {
		t.Run(row.name, func(t *testing.T) {
			streamer := &Streamer{ActiveMultipliers: row.multipliers}
			if got := streamer.TotalMultiplier(); got != row.want {
				t.Errorf("TotalMultiplier() = %v, want %v", got, row.want)
			}
		})
	}
}

func TestStreamer_PredictionWindowSeconds(t *testing.T) {
	for _, row := range []struct {
		name   string
		mode   DelayMode
		delay  *float64
		window float64
		want   float64
	}{
		{"from the start, a delay under the window", DelayModeFromStart, floatPtr(2), 5, 2},
		{"from the start, a delay of exactly the window", DelayModeFromStart, floatPtr(5), 5, 5},
		{"from the start, a delay over the window", DelayModeFromStart, floatPtr(6), 5, 5},
		{"from the start, no delay", DelayModeFromStart, nil, 5, 0.5},

		{"from the end, a delay under the window", DelayModeFromEnd, floatPtr(2), 5, 3},
		{"from the end, a delay of exactly the window", DelayModeFromEnd, floatPtr(5), 5, 0},
		{"from the end, a delay half a second over the window", DelayModeFromEnd, floatPtr(5.5), 5, 0},
		{"from the end, a delay over the window", DelayModeFromEnd, floatPtr(6), 5, 0},
		{"from the end, no delay", DelayModeFromEnd, nil, 5, 4.5},

		{"a share of the window", DelayModePercentage, floatPtr(0.5), 10, 5},
		{"a share of the window, no delay", DelayModePercentage, nil, 10, 5},

		{"no mode", "", floatPtr(2), 5, 5},
		{"an unknown mode", "SOMETIME", floatPtr(2), 5, 5},
		{"a window of zero", DelayModeFromEnd, floatPtr(2), 0, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			streamer := &Streamer{Settings: StreamerSettings{Bet: BetSettings{Delay: row.delay, DelayMode: row.mode}}}
			if got := streamer.PredictionWindowSeconds(row.window); got != row.want {
				t.Errorf("PredictionWindowSeconds(%v) = %v, want %v", row.window, got, row.want)
			}
		})
	}
}

func TestStreamer_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"a login", "alice", "Alice"},
		{"no login", "", ""},
		{"a first letter outside ASCII", "élan", "Élan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamer := &Streamer{Username: tt.username}
			if got := streamer.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStreamer_Spend(t *testing.T) {
	tests := []struct {
		name    string
		balance int
		points  int
		want    int
	}{
		{"part of the balance", 250, 100, 150},
		{"the whole balance", 250, 250, 0},
		{"more than the balance", 250, 300, 0},
		{"nothing", 250, 0, 250},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamer := &Streamer{ChannelPoints: tt.balance}
			streamer.Spend(tt.points)
			if streamer.ChannelPoints != tt.want {
				t.Errorf("balance after spending %d of %d = %d, want %d", tt.points, tt.balance, streamer.ChannelPoints, tt.want)
			}
		})
	}
}
