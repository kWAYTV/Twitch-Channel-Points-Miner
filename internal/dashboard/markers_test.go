package dashboard

import (
	"reflect"
	"testing"
)

func TestPointsSentence(t *testing.T) {
	tests := []struct {
		name  string
		delta int
		want  string
	}{
		{"a gain", 450, "Gained 450 points."},
		{"a loss", -200, "Lost 200 points."},
		{"no change", 0, "Points did not change."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pointsSentence(tt.delta); got != tt.want {
				t.Errorf("pointsSentence(%d) = %q, want %q", tt.delta, got, tt.want)
			}
		})
	}

	t.Run("a gain of one", func(t *testing.T) {
		if got, want := pointsSentence(1), "Gained 1 point."; got != want {
			t.Errorf("pointsSentence(1) = %q, want %q", got, want)
		}
	})

	t.Run("a loss of one", func(t *testing.T) {
		if got, want := pointsSentence(-1), "Lost 1 point."; got != want {
			t.Errorf("pointsSentence(-1) = %q, want %q", got, want)
		}
	})
}

func TestWithMarkers(t *testing.T) {
	tests := []struct {
		name   string
		events []Event
		points []Point
		want   []Event
	}{
		{
			name: "no events and no points",
			want: nil,
		},
		{
			name:   "a watch streak",
			points: []Point{{T: 2, P: 1450, D: 450, R: "WATCH_STREAK"}},
			want:   []Event{{T: 2, Type: "WATCH_STREAK", Msg: "Gained 450 points."}},
		},
		{
			name:   "a claim",
			points: []Point{{T: 2, P: 1050, D: 50, R: "CLAIM"}},
			want:   []Event{{T: 2, Type: "CLAIM", Msg: "Gained 50 points."}},
		},
		{
			name:   "a raid",
			points: []Point{{T: 2, P: 1250, D: 250, R: "RAID"}},
			want:   []Event{{T: 2, Type: "RAID", Msg: "Gained 250 points."}},
		},
		{
			name:   "a plain watch gain",
			points: []Point{{T: 2, P: 1010, D: 10, R: "WATCH"}},
			want:   nil,
		},
		{
			name:   "a balance with no reason",
			points: []Point{{T: 2, P: 1000}},
			want:   nil,
		},
		{
			name:   "a stored sentence",
			events: []Event{{T: 1, Type: "BET_WIN", Msg: "Won a bet on alice."}},
			want:   []Event{{T: 1, Type: "BET_WIN", Msg: "Won a bet on alice."}},
		},
		{
			name:   "a console line stored before events were sentences",
			events: []Event{{T: 1, Type: "STREAMER_ONLINE", Msg: "🥳 Alice (1.2k points) is Online!"}},
			want:   []Event{{T: 1, Type: "STREAMER_ONLINE", Msg: "Alice is online. Balance 1.2k points."}},
		},
		{
			name:   "a streak before a stored event",
			events: []Event{{T: 3, Type: "BET_WIN", Msg: "Won a bet on alice."}},
			points: []Point{{T: 2, P: 1450, D: 450, R: "WATCH_STREAK"}},
			want: []Event{
				{T: 2, Type: "WATCH_STREAK", Msg: "Gained 450 points."},
				{T: 3, Type: "BET_WIN", Msg: "Won a bet on alice."},
			},
		},
		{
			name:   "a streak in the same millisecond as a stored event",
			events: []Event{{T: 2, Type: "BET_WIN", Msg: "Won a bet on alice."}},
			points: []Point{{T: 2, P: 1450, D: 450, R: "WATCH_STREAK"}},
			want: []Event{
				{T: 2, Type: "BET_WIN", Msg: "Won a bet on alice."},
				{T: 2, Type: "WATCH_STREAK", Msg: "Gained 450 points."},
			},
		},
		{
			name: "stored events out of order",
			events: []Event{
				{T: 5, Type: "BET_WIN", Msg: "Won a bet on alice."},
				{T: 1, Type: "DROP_CLAIM", Msg: "Claimed a drop."},
			},
			want: []Event{
				{T: 1, Type: "DROP_CLAIM", Msg: "Claimed a drop."},
				{T: 5, Type: "BET_WIN", Msg: "Won a bet on alice."},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withMarkers(tt.events, tt.points); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("withMarkers = %+v, want %+v", got, tt.want)
			}
		})
	}
}
