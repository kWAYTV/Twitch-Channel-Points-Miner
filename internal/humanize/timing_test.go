package humanize

import (
	"testing"
	"time"
)

// drawing returns a Timing whose every draw is value.
func drawing(value float64) *Timing {
	return NewFrom(func() float64 { return value })
}

func TestNew(t *testing.T) {
	t.Run("switched off", func(t *testing.T) {
		if got := New(false); got != nil {
			t.Errorf("New(false) = %v, want nil", got)
		}
	})

	t.Run("switched on", func(t *testing.T) {
		timing := New(true)
		if timing == nil {
			t.Fatal("New(true) = nil, want a Timing")
		}
		if got := timing.Between(time.Second, 3*time.Second); got < time.Second || got >= 3*time.Second {
			t.Errorf("Between(1s, 3s) = %s, want from 1s up to 3s", got)
		}
	})
}

func TestTiming_Between(t *testing.T) {
	tests := []struct {
		name   string
		timing *Timing
		low    time.Duration
		high   time.Duration
		want   time.Duration
	}{
		{"the lowest draw", drawing(0), 2 * time.Second, 8 * time.Second, 2 * time.Second},
		{"a draw in the middle", drawing(0.5), 2 * time.Second, 8 * time.Second, 5 * time.Second},
		{"a draw near the top", drawing(0.75), 2 * time.Second, 8 * time.Second, 6500 * time.Millisecond},
		{"low and high the same", drawing(0.5), 2 * time.Second, 2 * time.Second, 0},
		{"high under low", drawing(0.5), 8 * time.Second, 2 * time.Second, 0},
		{"switched off", nil, 2 * time.Second, 8 * time.Second, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.timing.Between(tt.low, tt.high); got != tt.want {
				t.Errorf("Between(%s, %s) = %s, want %s", tt.low, tt.high, got, tt.want)
			}
		})
	}
}

func TestTiming_Around(t *testing.T) {
	tests := []struct {
		name   string
		timing *Timing
		d      time.Duration
		spread float64
		want   time.Duration
	}{
		{"the lowest draw", drawing(0), 20 * time.Second, 0.2, 16 * time.Second},
		{"a draw in the middle", drawing(0.5), 20 * time.Second, 0.2, 20 * time.Second},
		{"a draw three quarters up", drawing(0.75), 20 * time.Second, 0.2, 22 * time.Second},
		{"no spread", drawing(0), 20 * time.Second, 0, 20 * time.Second},
		{"no time", drawing(0), 0, 0.2, 0},
		{"switched off", nil, 20 * time.Second, 0.2, 20 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.timing.Around(tt.d, tt.spread); got != tt.want {
				t.Errorf("Around(%s, %v) = %s, want %s", tt.d, tt.spread, got, tt.want)
			}
		})
	}
}

func TestTiming_BetMoment(t *testing.T) {
	tests := []struct {
		name   string
		timing *Timing
		moment float64
		window float64
		want   float64
	}{
		// A bet set three quarters through a 100-second window.
		{"three quarters through, the lowest draw", drawing(0), 75, 100, 65},
		{"three quarters through, a draw in the middle", drawing(0.5), 75, 100, 75},
		{"three quarters through, a draw three quarters up", drawing(0.75), 75, 100, 80},
		{"three quarters through, the highest draw", drawing(1), 75, 100, 85},
		// A bet set six seconds before the end only moves earlier.
		{"six seconds before the end, the lowest draw", drawing(0), 114, 120, 102},
		{"six seconds before the end, the highest draw", drawing(1), 114, 120, 114},
		// A bet set near the end of the first tenth keeps to the window.
		{"two seconds after the start, the lowest draw", drawing(0), 2, 100, 1},
		{"two seconds after the start, the highest draw", drawing(1), 2, 100, 12},
		{"a window of half a second", drawing(0), 0.25, 0.5, 0.25},
		{"no window", drawing(0), 0, 0, 0},
		{"switched off", nil, 75, 100, 75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.timing.BetMoment(tt.moment, tt.window); got != tt.want {
				t.Errorf("BetMoment(%v, %v) = %v, want %v", tt.moment, tt.window, got, tt.want)
			}
		})
	}
}

func TestTiming_Stake(t *testing.T) {
	tests := []struct {
		name   string
		timing *Timing
		points int
		want   int
	}{
		{"three digits", drawing(0), 487, 480},
		{"already round", drawing(0), 500, 500},
		{"two digits", drawing(0), 47, 40},
		{"four digits", drawing(0), 1_234, 1_200},
		{"five digits", drawing(0), 12_345, 12_000},
		{"six digits", drawing(0), 249_999, 240_000},
		{"exactly a hundred", drawing(0), 100, 100},
		{"just under a hundred", drawing(0), 99, 90},
		{"twenty", drawing(0), 20, 20},
		{"under twenty", drawing(0), 19, 19},
		{"Twitch's minimum", drawing(0), 10, 10},
		{"nothing", drawing(0), 0, 0},
		{"switched off", nil, 487, 487},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.timing.Stake(tt.points); got != tt.want {
				t.Errorf("Stake(%d) = %d, want %d", tt.points, got, tt.want)
			}
		})
	}
}

func TestTiming_Every(t *testing.T) {
	t.Run("ticks until stopped", func(t *testing.T) {
		stop := make(chan struct{})
		ticks := drawing(0.5).Every(time.Millisecond, 0.2, stop)

		<-ticks
		<-ticks
		close(stop)
	})

	t.Run("switched off, it still ticks", func(t *testing.T) {
		stop := make(chan struct{})
		defer close(stop)
		var off *Timing

		<-off.Every(time.Millisecond, 0.2, stop)
	})

	t.Run("stopped before the first tick", func(t *testing.T) {
		stop := make(chan struct{})
		close(stop)
		ticks := drawing(0.5).Every(time.Hour, 0.2, stop)

		select {
		case <-ticks:
			t.Errorf("a tick arrived, want none after stop")
		case <-time.After(5 * time.Millisecond):
		}
	})
}
