package dashboard

import (
	"fmt"
	"sort"

	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

// Gains shown as chart markers, derived from recorded point reasons.
var markerReasons = map[string]bool{"WATCH_STREAK": true, "CLAIM": true, "RAID": true}

func pointsSentence(delta int) string {
	n := delta
	if n < 0 {
		n = -n
	}
	switch {
	case delta < 0:
		return fmt.Sprintf("Lost %d points.", n)
	case delta == 0:
		return "Points did not change."
	default:
		return fmt.Sprintf("Gained %d points.", n)
	}
}

// withMarkers returns the stored events plus one for every gain the chart
// marks, in time order.
func withMarkers(events []Event, points []Point) []Event {
	// Lines stored before events were recorded as sentences are rewritten here.
	for i := range events {
		events[i].Msg = notify.Sentence(events[i].Msg)
	}
	for _, p := range points {
		if markerReasons[p.R] {
			events = append(events, Event{T: p.T, Type: p.R, Msg: pointsSentence(p.D)})
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].T < events[j].T })
	return events
}
