package dashboard

import (
	"reflect"
	"testing"
)

func TestDownsample(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
		max    int
		want   []Point
	}{
		{
			name:   "no points",
			points: nil,
			max:    4,
			want:   nil,
		},
		{
			name:   "one point fewer than the limit",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}},
		},
		{
			name:   "exactly the limit",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}},
		},
		{
			name:   "one point over the limit, a spike before a dip",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 30}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 5, P: 30}},
		},
		{
			name:   "a dip before a spike",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 5}, {T: 3, P: 20}, {T: 4, P: 50}, {T: 5, P: 30}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 5}, {T: 4, P: 50}, {T: 5, P: 30}},
		},
		{
			name:   "a flat stretch",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 10}, {T: 3, P: 10}, {T: 4, P: 10}, {T: 5, P: 10}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 10}, {T: 5, P: 10}},
		},
		{
			name:   "the lowest and the highest balance twice each",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 5}, {T: 3, P: 5}, {T: 4, P: 50}, {T: 5, P: 50}, {T: 6, P: 30}},
			max:    4,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 5}, {T: 4, P: 50}, {T: 6, P: 30}},
		},
		{
			name:   "a spike and a dip at the two ends",
			points: []Point{{T: 1, P: 999}, {T: 2, P: 10}, {T: 3, P: 20}, {T: 4, P: 30}, {T: 5, P: -5}},
			max:    4,
			want:   []Point{{T: 1, P: 999}, {T: 2, P: 10}, {T: 4, P: 30}, {T: 5, P: -5}},
		},
		{
			name: "a limit of six, two stretches",
			points: []Point{
				{T: 1, P: 10},
				{T: 2, P: 20}, {T: 3, P: 90}, {T: 4, P: 15},
				{T: 5, P: 40}, {T: 6, P: 2}, {T: 7, P: 60},
				{T: 8, P: 30},
			},
			max:  6,
			want: []Point{{T: 1, P: 10}, {T: 3, P: 90}, {T: 4, P: 15}, {T: 6, P: 2}, {T: 7, P: 60}, {T: 8, P: 30}},
		},
		{
			name:   "an odd limit of five",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 40}, {T: 6, P: 30}},
			max:    5,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 6, P: 30}},
		},
		{
			name:   "exactly an odd limit of five",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 30}},
			max:    5,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 30}},
		},
		{
			name:   "a limit of three is raised to four",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 30}},
			max:    3,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 5, P: 30}},
		},
		{
			name:   "a limit of zero is raised to four",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}, {T: 5, P: 30}},
			max:    0,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 5, P: 30}},
		},
		{
			name:   "four points under a limit of zero",
			points: []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}},
			max:    0,
			want:   []Point{{T: 1, P: 10}, {T: 2, P: 50}, {T: 3, P: 5}, {T: 4, P: 20}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Downsample(tt.points, tt.max); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Downsample(%+v, %d) = %+v, want %+v", tt.points, tt.max, got, tt.want)
			}
		})
	}
}
