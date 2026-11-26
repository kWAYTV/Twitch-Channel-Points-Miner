package dashboard

import "time"

var ranges = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"all": 0,
}

// window is the stretch of history a request asks for.
type window struct {
	name  string
	since int64
	now   int64
}

// windowFor resolves a range name such as "7d" at the time now (unix ms). An
// empty name means the last day.
func windowFor(rangeName string, now int64) (window, bool) {
	if rangeName == "" {
		rangeName = "24h"
	}
	span, ok := ranges[rangeName]
	if !ok {
		return window{}, false
	}
	w := window{name: rangeName, now: now}
	if span > 0 {
		w.since = w.now - span.Milliseconds()
	}
	return w, true
}
