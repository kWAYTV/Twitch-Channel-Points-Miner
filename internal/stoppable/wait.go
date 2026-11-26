// Package stoppable holds waits that end early when the miner is stopping.
package stoppable

import "time"

// Wait waits for d, or until stop is closed, and reports whether it was
// stopped. A wait of no time returns at once without looking at stop.
func Wait(d time.Duration, stop <-chan struct{}) (stopped bool) {
	if d <= 0 {
		return false
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-stop:
		return true
	case <-timer.C:
		return false
	}
}
