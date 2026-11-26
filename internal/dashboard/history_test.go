package dashboard

import (
	"reflect"
	"testing"
)

// numberedEvents returns n events at the times 1 to n.
func numberedEvents(n int) []Event {
	events := make([]Event, n)
	for i := range events {
		events[i] = Event{T: int64(i + 1), Type: "BET_WIN", Msg: "Won a bet on alice."}
	}
	return events
}

func TestLatestEvents(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		got, truncated := latestEvents(nil)
		if want := []Event{}; !reflect.DeepEqual(got, want) || truncated {
			t.Errorf("latestEvents(nil) = %#v, %v, want %#v, false", got, truncated, want)
		}
	})

	t.Run("one event", func(t *testing.T) {
		events := []Event{{T: 1, Type: "BET_WIN", Msg: "Won a bet on alice."}}
		got, truncated := latestEvents(events)
		if want := []Event{{T: 1, Type: "BET_WIN", Msg: "Won a bet on alice."}}; !reflect.DeepEqual(got, want) || truncated {
			t.Errorf("latestEvents = %+v, %v, want %+v, false", got, truncated, want)
		}
	})

	t.Run("1999 events", func(t *testing.T) {
		got, truncated := latestEvents(numberedEvents(1999))
		if want := numberedEvents(1999); !reflect.DeepEqual(got, want) || truncated {
			t.Errorf("latestEvents = %d events from t=%d, %v, want all 1999, false", len(got), got[0].T, truncated)
		}
	})

	t.Run("2000 events", func(t *testing.T) {
		got, truncated := latestEvents(numberedEvents(2000))
		if want := numberedEvents(2000); !reflect.DeepEqual(got, want) || truncated {
			t.Errorf("latestEvents = %d events from t=%d, %v, want all 2000, false", len(got), got[0].T, truncated)
		}
	})

	t.Run("2001 events", func(t *testing.T) {
		got, truncated := latestEvents(numberedEvents(2001))
		if want := numberedEvents(2001)[1:]; !reflect.DeepEqual(got, want) || !truncated {
			t.Errorf("latestEvents = %d events from t=%d, %v, want 2000 from t=2, true", len(got), got[0].T, truncated)
		}
	})
}

func TestStatsOf(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
		want   seriesStats
	}{
		{"one point", []Point{{T: 1, P: 1000}}, seriesStats{Start: 1000, End: 1000, Gain: 0, Min: 1000, Max: 1000}},
		{"a rise", []Point{{T: 1, P: 1000}, {T: 2, P: 1010}, {T: 3, P: 1460}}, seriesStats{Start: 1000, End: 1460, Gain: 460, Min: 1000, Max: 1460}},
		{"a fall", []Point{{T: 1, P: 1000}, {T: 2, P: 750}, {T: 3, P: 300}}, seriesStats{Start: 1000, End: 300, Gain: -700, Min: 300, Max: 1000}},
		{"a dip and a spike in between", []Point{{T: 1, P: 1000}, {T: 2, P: 200}, {T: 3, P: 5000}, {T: 4, P: 1100}}, seriesStats{Start: 1000, End: 1100, Gain: 100, Min: 200, Max: 5000}},
		{"a flat line at zero", []Point{{T: 1, P: 0}, {T: 2, P: 0}}, seriesStats{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statsOf(tt.points); got != tt.want {
				t.Errorf("statsOf(%+v) = %+v, want %+v", tt.points, got, tt.want)
			}
		})
	}
}

func TestSeriesOf(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
		want   [][2]int64
	}{
		{"no points", nil, [][2]int64{}},
		{"one point", []Point{{T: 1700000000000, P: 1000}}, [][2]int64{{1700000000000, 1000}}},
		{"a gain with its reason", []Point{{T: 1, P: 1000}, {T: 2, P: 1010, D: 10, R: "WATCH"}}, [][2]int64{{1, 1000}, {2, 1010}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := seriesOf(tt.points); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("seriesOf(%+v) = %#v, want %#v", tt.points, got, tt.want)
			}
		})
	}
}

func TestStreamerHistory(t *testing.T) {
	lastDay := window{name: "24h", since: 5000, now: 9000}
	everything := window{name: "all", since: 0, now: 9000}

	t.Run("no points, over the last day", func(t *testing.T) {
		got := streamerHistory("alice", lastDay, nil, nil, nil)

		want := streamerResponse{
			Name: "alice", Range: "24h", From: 5000, To: 9000,
			Series: [][2]int64{}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("no points, over everything", func(t *testing.T) {
		got := streamerHistory("alice", everything, nil, nil, nil)

		want := streamerResponse{
			Name: "alice", Range: "all", From: 9000, To: 9000,
			Series: [][2]int64{}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("two points, over the last day", func(t *testing.T) {
		points := []Point{{T: 6000, P: 1000}, {T: 7000, P: 1010, D: 10, R: "WATCH"}}

		got := streamerHistory("alice", lastDay, points, nil, nil)

		want := streamerResponse{
			Name: "alice", Range: "24h", From: 5000, To: 9000,
			Series: [][2]int64{{6000, 1000}, {7000, 1010}}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{Start: 1000, End: 1010, Gain: 10, Min: 1000, Max: 1010},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("two points, over everything", func(t *testing.T) {
		points := []Point{{T: 6000, P: 1000}, {T: 7000, P: 1010, D: 10, R: "WATCH"}}

		got := streamerHistory("alice", everything, points, nil, nil)

		want := streamerResponse{
			Name: "alice", Range: "all", From: 6000, To: 9000,
			Series: [][2]int64{{6000, 1000}, {7000, 1010}}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{Start: 1000, End: 1010, Gain: 10, Min: 1000, Max: 1010},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("a balance from before the window, then a point in it", func(t *testing.T) {
		points := []Point{{T: 7000, P: 1010, D: 10, R: "WATCH"}}
		previous := &Point{T: 4000, P: 1000}

		got := streamerHistory("alice", lastDay, points, previous, nil)

		want := streamerResponse{
			Name: "alice", Range: "24h", From: 5000, To: 9000,
			Series: [][2]int64{{5000, 1000}, {7000, 1010}}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{Start: 1000, End: 1010, Gain: 10, Min: 1000, Max: 1010},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("only a balance from before the window", func(t *testing.T) {
		previous := &Point{T: 4000, P: 1000}

		got := streamerHistory("alice", lastDay, nil, previous, nil)

		want := streamerResponse{
			Name: "alice", Range: "24h", From: 5000, To: 9000,
			Series: [][2]int64{{5000, 1000}}, Events: []Event{}, EventsTruncated: false,
			Stats: seriesStats{Start: 1000, End: 1000, Gain: 0, Min: 1000, Max: 1000},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("one event", func(t *testing.T) {
		events := []Event{{T: 6000, Type: "BET_WIN", Msg: "Won a bet on alice."}}

		got := streamerHistory("alice", lastDay, nil, nil, events)

		want := streamerResponse{
			Name: "alice", Range: "24h", From: 5000, To: 9000,
			Series: [][2]int64{}, Events: []Event{{T: 6000, Type: "BET_WIN", Msg: "Won a bet on alice."}}, EventsTruncated: false,
			Stats: seriesStats{},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory = %+v, want %+v", got, want)
		}
	})

	t.Run("2001 events", func(t *testing.T) {
		got := streamerHistory("alice", everything, nil, nil, numberedEvents(2001))

		want := streamerResponse{
			Name: "alice", Range: "all", From: 9000, To: 9000,
			Series: [][2]int64{}, Events: numberedEvents(2001)[1:], EventsTruncated: true,
			Stats: seriesStats{},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("streamerHistory: %d events from t=%d, truncated %v, want 2000 from t=2, truncated true",
				len(got.Events), got.Events[0].T, got.EventsTruncated)
		}
	})

	t.Run("801 points, one more than a chart holds", func(t *testing.T) {
		points := make([]Point, 801)
		for i := range points {
			points[i] = Point{T: int64(6000 + i), P: 1000 + i}
		}

		got := streamerHistory("alice", everything, points, nil, nil)

		if len(got.Series) != 800 {
			t.Errorf("series holds %d points, want 800", len(got.Series))
		}
		if first, want := got.Series[0], [2]int64{6000, 1000}; first != want {
			t.Errorf("series starts at %v, want %v", first, want)
		}
		if last, want := got.Series[len(got.Series)-1], [2]int64{6800, 1800}; last != want {
			t.Errorf("series ends at %v, want %v", last, want)
		}
		if want := (seriesStats{Start: 1000, End: 1800, Gain: 800, Min: 1000, Max: 1800}); got.Stats != want {
			t.Errorf("stats = %+v, want %+v", got.Stats, want)
		}
	})
}
