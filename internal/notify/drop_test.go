package notify

import (
	"reflect"
	"testing"
)

func TestDescribeDropClaim(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a drop claimed", "Claim Cool Hat (Summer Drops) 60/60 (100%)", described{Message{"Claimed Cool Hat", []string{"Summer Drops campaign", "Progress 60/60 (100%)"}}, true}},
		{"upper case", "CLAIM Cool Hat (Summer Drops) 60/60", described{Message{"Claimed Cool Hat", []string{"Summer Drops campaign", "Progress 60/60"}}, true}},
		{"no progress", "Claim Cool Hat (Summer Drops)", described{}},
		{"no campaign", "Claim Cool Hat 60/60", described{}},
		{"an empty campaign", "Claim Cool Hat () 60/60", described{}},
		{"no reward", "Claim (Summer Drops) 60/60", described{}},
		{"another line", "Alice Claimed Moment", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeDropClaim(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeDropClaim(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}

func TestDescribeDropProgress(t *testing.T) {
	for _, tc := range []struct {
		name  string
		head  string
		other []string
		want  described
	}{
		{
			"progress with time watched and time left",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Game / Camp / chan / 12m of 30m / remaining 18m / active"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / chan", "12m watched of 30m"}}, true},
		},
		{
			"progress with time watched",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Game / Camp / chan / 12m of 30m"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / chan", "12m watched of 30m"}}, true},
		},
		{
			"progress with no time watched",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Game / Camp / chan"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / chan"}}, true},
		},
		{
			"a last part that is not a time",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Game / Camp / 2 of 3"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / 2 of 3"}}, true},
		},
		{
			"a description of one part",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Summer Drops"},
			described{Message{"Drop 40% complete", []string{"Summer Drops"}}, true},
		},
		{
			"a description in two parts",
			"DROP_PROGRESS [####----------------] 40%",
			[]string{"Game / Camp / chan", "active"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / chan | active"}}, true},
		},
		{
			"a description after the bar on the same line",
			"DROP_PROGRESS [####----------------] 40% | Game / Camp / chan / 12m of 30m / remaining 18m",
			[]string{"ignored"},
			described{Message{"Drop 40% complete", []string{"Game / Camp / chan", "12m watched of 30m"}}, true},
		},
		{
			"no description",
			"DROP_PROGRESS [####----------------] 40%",
			nil,
			described{Message{Headline: "Drop 40% complete"}, true},
		},
		{
			"a full bar",
			"DROP_PROGRESS [####################] 100%",
			nil,
			described{Message{Headline: "Drop 100% complete"}, true},
		},
		{
			"lower case",
			"drop_progress [] 0%",
			nil,
			described{Message{Headline: "Drop 0% complete"}, true},
		},
		{"no percentage", "DROP_PROGRESS [####----------------]", nil, described{}},
		{"another line", "Claim Cool Hat (Summer Drops) 60/60", nil, described{}},
		{"an empty line", "", nil, described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeDropProgress(tc.head, tc.other)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeDropProgress(%q, %q) = %+v, want %+v", tc.head, tc.other, got, tc.want)
			}
		})
	}
}

func TestDescribeDropNotConnected(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{
			"a game with where to connect it",
			"Drops for WARDOGS are not collected: your Twitch account is not connected to the game. Connect it at https://account.wardogs.com",
			described{Message{"Drops for WARDOGS are not collected", []string{"Your Twitch account is not connected to the game", "Connect it at https://account.wardogs.com"}}, true},
		},
		{
			"a game of several words",
			"Drops for World of Tanks are not collected: your Twitch account is not connected to the game. Connect it at https://www.twitch.tv/drops/inventory",
			described{Message{"Drops for World of Tanks are not collected", []string{"Your Twitch account is not connected to the game", "Connect it at https://www.twitch.tv/drops/inventory"}}, true},
		},
		{
			"no address",
			"Drops for WARDOGS are not collected: your Twitch account is not connected to the game.",
			described{Message{"Drops for WARDOGS are not collected", []string{"Your Twitch account is not connected to the game"}}, true},
		},
		{
			"upper case",
			"DROPS FOR WARDOGS ARE NOT COLLECTED: YOUR TWITCH ACCOUNT IS NOT CONNECTED TO THE GAME. CONNECT IT AT https://account.wardogs.com",
			described{Message{"Drops for WARDOGS are not collected", []string{"Your Twitch account is not connected to the game", "Connect it at https://account.wardogs.com"}}, true},
		},
		{"no game", "Drops for  are not collected: your Twitch account is not connected to the game.", described{}},
		{"a drop claimed", "Claim Cool Hat (Summer Drops) 60/60 (100%)", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeDropNotConnected(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeDropNotConnected(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}
