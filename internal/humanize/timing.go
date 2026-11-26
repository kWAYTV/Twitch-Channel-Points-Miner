// Package humanize varies the miner's timings the way a person's vary, so
// that it does not act at the same instant every time.
package humanize

import (
	"math/rand/v2"
	"time"
)

// Timing draws the variations. A nil Timing is switched off: every method
// then returns the exact time it was given, or no wait at all.
type Timing struct {
	// uniform returns a number from 0 up to, but not including, 1.
	uniform func() float64
}

// New returns a Timing that varies, or nil when enabled is false.
func New(enabled bool) *Timing {
	if !enabled {
		return nil
	}
	return &Timing{uniform: rand.Float64}
}

// NewFrom returns a Timing that draws from uniform, for a caller that needs
// the same variation every time.
func NewFrom(uniform func() float64) *Timing {
	return &Timing{uniform: uniform}
}

// Between is a wait somewhere from low to high: how long a person takes to
// notice something and react. It is no wait at all when switched off.
func (t *Timing) Between(low, high time.Duration) time.Duration {
	if t == nil || high <= low {
		return 0
	}
	return low + time.Duration(t.uniform()*float64(high-low))
}

// Around is d made up to spread shorter or longer: 0.2 gives anything from
// 80% to 120% of d. It is d itself when switched off.
func (t *Timing) Around(d time.Duration, spread float64) time.Duration {
	if t == nil || d <= 0 || spread <= 0 {
		return d
	}
	factor := 1 + spread*(2*t.uniform()-1)
	return time.Duration(float64(d) * factor)
}

// betSpread is how far a bet may move from its set moment, as a share of the
// prediction's window.
const betSpread = 0.10

// BetMoment moves the moment a bet is placed, in seconds after the prediction
// opened, by up to a tenth of the window either way. It never moves the bet
// into the last tenth of the window unless it was set there, because a bet
// placed later than asked is the one that can be missed. It is moment itself
// when switched off.
func (t *Timing) BetMoment(moment, window float64) float64 {
	if t == nil || window <= 0 {
		return moment
	}
	shifted := moment + window*betSpread*(2*t.uniform()-1)
	latest := max(moment, window*(1-betSpread))
	// Nor earlier than a second after the prediction opened, unless it was
	// set earlier than that.
	earliest := min(1, moment)
	return min(max(shifted, earliest), latest)
}

// Every sends the time on the returned channel once per interval until stop
// is closed, each interval made up to spread shorter or longer. Switched off,
// the intervals are exact. A tick nobody is waiting for is dropped.
func (t *Timing) Every(interval time.Duration, spread float64, stop <-chan struct{}) <-chan time.Time {
	ticks := make(chan time.Time)
	go func() {
		for {
			timer := time.NewTimer(t.Around(interval, spread))
			select {
			case <-stop:
				timer.Stop()
				return
			case at := <-timer.C:
				select {
				case ticks <- at:
				case <-stop:
					return
				default:
				}
			}
		}
	}()
	return ticks
}

// Stake rounds a stake down to the kind of number a person types: two
// leading digits and zeros, so 487 becomes 480 and 12,345 becomes 12,000.
// Rounding down keeps it within every limit it was already within. It is the
// stake itself when switched off, and a stake under 20 is left alone.
func (t *Timing) Stake(points int) int {
	if t == nil || points < 20 {
		return points
	}
	step := 10
	for points/step >= 100 {
		step *= 10
	}
	return points / step * step
}
