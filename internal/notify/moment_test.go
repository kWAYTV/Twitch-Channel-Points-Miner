package notify

import (
	"reflect"
	"testing"
)

func TestDescribeMoment(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a moment claimed", "Alice Claimed Moment", described{Message{Headline: "Claimed a moment on Alice"}, true}},
		{"lower case", "Alice claimed moment", described{Message{Headline: "Claimed a moment on Alice"}, true}},
		{"no name", " Claimed Moment", described{Message{Headline: "Claimed a moment"}, true}},
		{"words after the marker", "Alice Claimed Moment today", described{}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeMoment(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeMoment(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}
