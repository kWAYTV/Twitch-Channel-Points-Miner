package streak

import "time"

// orNow is t, or the present moment when t was never set.
func orNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
