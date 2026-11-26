package notify

import (
	"reflect"
	"testing"
)

func TestDescribeMention(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a mention", "nick at #alice wrote: hello there", described{Message{"nick mentioned you in #alice", []string{`"hello there"`}}, true}},
		{"upper case", "nick AT #alice WROTE: hello there", described{Message{"nick mentioned you in #alice", []string{`"hello there"`}}, true}},
		{"a text that holds the marker", "nick at #alice wrote: he wrote: hi", described{Message{"nick mentioned you in #alice", []string{`"he wrote: hi"`}}, true}},
		{"a text outside ASCII", "nick at #alice wrote: İyi günler 👋", described{Message{"nick mentioned you in #alice", []string{`"İyi günler 👋"`}}, true}},
		{"an empty text", "nick at #alice wrote: ", described{Message{"nick mentioned you in #alice", []string{`""`}}, true}},
		{"no channel", "nick wrote: hello there", described{}},
		{"an empty channel", "nick at # wrote: hello there", described{}},
		{"no nick", " at #alice wrote: hello there", described{}},
		{"no text marker", "nick at #alice said hello", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeMention(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeMention(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}
