package dashboard

import "testing"

func TestNextSample(t *testing.T) {
	tests := []struct {
		name    string
		last    Point
		hasLast bool
		now     int64
		points  int
		delta   int
		reason  string
		want    Point
		wantOK  bool
	}{
		{
			name: "a first reading",
			now:  t0, points: 500,
			want: Point{T: t0, P: 500}, wantOK: true,
		},
		{
			name: "a first reading with a gain and its reason",
			now:  t0, points: 510, delta: 10, reason: "WATCH",
			want: Point{T: t0, P: 510, D: 10, R: "WATCH"}, wantOK: true,
		},
		{
			name: "a gain with its reason",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 510, delta: 10, reason: "WATCH",
			want: Point{T: t0 + minute, P: 510, D: 10, R: "WATCH"}, wantOK: true,
		},
		{
			name: "a delta with no reason is kept as given",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 510, delta: 7,
			want: Point{T: t0 + minute, P: 510, D: 7}, wantOK: true,
		},
		{
			name: "a reason with no delta, on an unchanged balance",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 500, reason: "CLAIM",
			want: Point{T: t0 + minute, P: 500, R: "CLAIM"}, wantOK: true,
		},
		{
			name: "a snapshot of a higher balance",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 540,
			want: Point{T: t0 + minute, P: 540, D: 40}, wantOK: true,
		},
		{
			name: "a snapshot of a lower balance",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 300,
			want: Point{T: t0 + minute, P: 300, D: -200}, wantOK: true,
		},
		{
			name: "a snapshot of the same balance in the same millisecond",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0, points: 500,
			want: Point{}, wantOK: false,
		},
		{
			name: "a snapshot of the same balance a millisecond short of an hour later",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + hour - 1, points: 500,
			want: Point{}, wantOK: false,
		},
		{
			name: "a snapshot of the same balance exactly an hour later",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + hour, points: 500,
			want: Point{T: t0 + hour, P: 500}, wantOK: true,
		},
		{
			name: "a snapshot of the same balance an hour and a millisecond later",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + hour + 1, points: 500,
			want: Point{T: t0 + hour + 1, P: 500}, wantOK: true,
		},
		{
			name: "a snapshot of a balance of zero after a balance",
			last: Point{T: t0, P: 500}, hasLast: true,
			now: t0 + minute, points: 0,
			want: Point{T: t0 + minute, P: 0, D: -500}, wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := nextSample(tt.last, tt.hasLast, tt.now, tt.points, tt.delta, tt.reason)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("nextSample(%+v, %v, %d, %d, %d, %q) = %+v, %v, want %+v, %v",
					tt.last, tt.hasLast, tt.now, tt.points, tt.delta, tt.reason, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
