package auth

import "time"

// fakeClock is a clock that moves only when something sleeps on it.
type fakeClock struct {
	Time  time.Time
	Slept []time.Duration
}

func (c *fakeClock) Now() time.Time {
	return c.Time
}

func (c *fakeClock) Sleep(d time.Duration) {
	c.Slept = append(c.Slept, d)
	c.Time = c.Time.Add(d)
}
