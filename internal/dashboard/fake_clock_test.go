package dashboard

import (
	"sync/atomic"
	"time"
)

// Times are unix milliseconds, as the store keeps them.
const (
	t0     int64 = 1_700_000_000_000
	minute int64 = 60_000
	hour   int64 = 3_600_000
	day    int64 = 86_400_000
)

// fakeClock is the time a store reads. It stands still until a test sets it.
type fakeClock struct{ ms atomic.Int64 }

func newFakeClock(ms int64) *fakeClock {
	c := &fakeClock{}
	c.Set(ms)
	return c
}

func (c *fakeClock) Now() time.Time { return time.UnixMilli(c.ms.Load()) }

func (c *fakeClock) Set(ms int64) { c.ms.Store(ms) }
