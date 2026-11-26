package notify

import (
	"reflect"
	"testing"
)

func TestDescribeSession(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a session starting", "Start session: 'abc'", described{Message{Headline: "Session started"}, true}},
		{"a session ending", "Ending session: 'abc'", described{Message{Headline: "Session ended"}, true}},
		{"a session starting in upper case", "START SESSION: 'abc'", described{Message{Headline: "Session started"}, true}},
		{
			"streamers loaded with some online",
			"19 streamers loaded in 3.3 seconds: 3 online (Arrowcs, Fl0m, Bob), 16 offline",
			described{Message{"19 streamers loaded", []string{"3 online: Arrowcs, Fl0m, Bob", "16 offline", "Took 3.3 seconds"}}, true},
		},
		{
			"streamers loaded with none online",
			"19 streamers loaded in 3.3 seconds: 0 online, 19 offline",
			described{Message{"19 streamers loaded", []string{"0 online", "19 offline", "Took 3.3 seconds"}}, true},
		},
		{
			"one streamer loaded",
			"1 streamer loaded in 0.4 seconds: 0 online, 1 offline",
			described{Message{"1 streamer loaded", []string{"0 online", "1 offline", "Took 0.4 seconds"}}, true},
		},
		{
			"streamers loaded with a full stop at the end",
			"2 streamers loaded in 1 second: 1 online (Bob), 1 offline.",
			described{Message{"2 streamers loaded", []string{"1 online: Bob", "1 offline", "Took 1 second"}}, true},
		},
		{
			"streamers loaded with no names in the brackets",
			"2 streamers loaded in 1 second: 1 online (), 1 offline",
			described{Message{"2 streamers loaded", []string{"1 online", "1 offline", "Took 1 second"}}, true},
		},
		{"streamers loaded with no counts", "streamers loaded in 1 second", described{}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeSession(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeSession(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}
