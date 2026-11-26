package dashboard

// maxEvents caps the markers sent per request (the most recent are kept).
const maxEvents = 2000

type seriesStats struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Gain  int `json:"gain"`
	Min   int `json:"min"`
	Max   int `json:"max"`
}

type streamerResponse struct {
	Name            string      `json:"name"`
	Range           string      `json:"range"`
	From            int64       `json:"from"`
	To              int64       `json:"to"`
	Series          [][2]int64  `json:"series"`
	Events          []Event     `json:"events"`
	EventsTruncated bool        `json:"events_truncated"`
	Stats           seriesStats `json:"stats"`
}

// streamerHistory builds the response for one streamer and window. previous is
// the last balance before the window, which the series starts from.
func streamerHistory(name string, span window, points []Point, previous *Point, events []Event) streamerResponse {
	resp := streamerResponse{Name: name, Range: span.name, From: span.since, To: span.now}
	resp.Events, resp.EventsTruncated = latestEvents(events)
	if previous != nil {
		points = append([]Point{{T: span.since, P: previous.P}}, points...)
	}
	switch {
	case len(points) > 0:
		if span.since == 0 {
			resp.From = points[0].T
		}
		resp.Stats = statsOf(points)
	case span.since == 0:
		resp.From = span.now
	}
	resp.Series = seriesOf(Downsample(points, MaxSeriesPoints))
	return resp
}

// latestEvents keeps the most recent maxEvents and says whether older ones were dropped.
func latestEvents(events []Event) ([]Event, bool) {
	if len(events) > maxEvents {
		return events[len(events)-maxEvents:], true
	}
	if events == nil {
		return []Event{}, false
	}
	return events, false
}

func statsOf(points []Point) seriesStats {
	stats := seriesStats{Start: points[0].P, End: points[len(points)-1].P, Min: points[0].P, Max: points[0].P}
	for _, p := range points {
		if p.P < stats.Min {
			stats.Min = p.P
		}
		if p.P > stats.Max {
			stats.Max = p.P
		}
	}
	stats.Gain = stats.End - stats.Start
	return stats
}

func seriesOf(points []Point) [][2]int64 {
	series := make([][2]int64, len(points))
	for i, p := range points {
		series[i] = [2]int64{p.T, int64(p.P)}
	}
	return series
}
