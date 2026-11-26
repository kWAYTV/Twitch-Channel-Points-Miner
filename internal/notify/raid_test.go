package notify

import (
	"reflect"
	"testing"
)

func TestDescribeRaid(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a raid joined", "Joining raid from Alice to Bob", described{Message{Headline: "Joined the raid from Alice to Bob"}, true}},
		{"upper case", "JOINING RAID FROM Alice TO Bob", described{Message{Headline: "Joined the raid from Alice to Bob"}, true}},
		{"no target", "Joining raid from Alice", described{}},
		{"no source", "Joining raid from to Bob", described{}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeRaid(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeRaid(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}
