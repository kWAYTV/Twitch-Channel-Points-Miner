package notify

import (
	"reflect"
	"strings"
	"testing"
)

// described is what a describer answers: the message, and whether the line was its kind.
type described struct {
	message Message
	ok      bool
}

func TestNewMessage(t *testing.T) {
	for _, tc := range []struct {
		name     string
		headline string
		details  []string
		want     Message
	}{
		{"a headline and two details", "Alice is online", []string{"Playing Valorant", "Watch streak 3"}, Message{"Alice is online", []string{"Playing Valorant", "Watch streak 3"}}},
		{"a headline alone", "Alice is online", nil, Message{Headline: "Alice is online"}},
		{"spaces around every part", "  Alice is online ", []string{" Playing Valorant  "}, Message{"Alice is online", []string{"Playing Valorant"}}},
		{"an empty detail between two others", "Alice is online", []string{"Playing Valorant", "", "Watch streak 3"}, Message{"Alice is online", []string{"Playing Valorant", "Watch streak 3"}}},
		{"only blank details", "Alice is online", []string{"", "   "}, Message{Headline: "Alice is online"}},
		{"an empty headline", "", []string{"Playing Valorant"}, Message{"", []string{"Playing Valorant"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := newMessage(tc.headline, tc.details...); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("newMessage(%q, %q) = %+v, want %+v", tc.headline, tc.details, got, tc.want)
			}
		})
	}
}

func TestMessage_Sentence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		message Message
		want    string
	}{
		{"a headline and two details", Message{"+50 points on Alice", []string{"Watching Valorant", "Balance 12.4k points"}}, "+50 points on Alice. Watching Valorant. Balance 12.4k points."},
		{"a headline alone", Message{Headline: "Alice is online"}, "Alice is online."},
		{"a headline that ends its own sentence", Message{"Alice is online!", []string{"Really?"}}, "Alice is online! Really?"},
		{"a blank detail", Message{"Alice is online", []string{"  ", "Watch streak 3"}}, "Alice is online. Watch streak 3."},
		{"details without a headline", Message{"", []string{"Watch streak 3"}}, "Watch streak 3."},
		{"one rune under the limit", Message{Headline: strings.Repeat("é", 498)}, strings.Repeat("é", 498) + "."},
		{"exactly the limit", Message{Headline: strings.Repeat("é", 499)}, strings.Repeat("é", 499) + "."},
		{"one rune over the limit", Message{Headline: strings.Repeat("é", 500)}, strings.Repeat("é", 499) + "…"},
		{"an empty message", Message{}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.message.Sentence(); got != tc.want {
				t.Errorf("%+v.Sentence() = %q, want %q", tc.message, got, tc.want)
			}
		})
	}
}

func TestEndSentence(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"no punctuation at the end", "Alice is online", "Alice is online."},
		{"a full stop at the end", "Alice is online.", "Alice is online."},
		{"an exclamation mark at the end", "Alice is online!", "Alice is online!"},
		{"a question mark at the end", "Alice is online?", "Alice is online?"},
		{"a quote at the end", `"overtime?"`, `"overtime?".`},
		{"a letter outside ASCII at the end", "Playing Pokémon Rubí", "Playing Pokémon Rubí."},
		{"an ellipsis at the end", "Alice is…", "Alice is…."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := endSentence(tc.in); got != tc.want {
				t.Errorf("endSentence(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSentence(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		want string
	}{
		{"a console line", "🥳 Alice (1.2k points) is Online! | Playing: Valorant | streak length 3", "Alice is online. Playing Valorant. Balance 1.2k points. Watch streak 3."},
		{"a mention over the limit", "nick at #alice wrote: " + strings.Repeat("é", 500), `nick mentioned you in #alice. "` + strings.Repeat("é", 468) + "…"},
		{"an empty line", "", ""},

		{"a stored sentence: online", "Alice is online. Playing Valorant. Balance 1.2k points. Watch streak 3.", "Alice is online. Playing Valorant. Balance 1.2k points. Watch streak 3."},
		{"a stored sentence: offline", "Alice is offline. Balance 1.2k points.", "Alice is offline. Balance 1.2k points."},
		{"a stored sentence: online with no details", "Alice is online.", "Alice is online."},
		{"a stored sentence: points for watching", "+50 points on Alice. Watching Valorant. Balance 12.4k points.", "+50 points on Alice. Watching Valorant. Balance 12.4k points."},
		{"a stored sentence: points for a watch streak", "+450 points on Alice. Watch streak 4. Balance 1.5k points.", "+450 points on Alice. Watch streak 4. Balance 1.5k points."},
		{"a stored sentence: points for a bonus claim", "+50 points on Alice. Bonus claim. Balance 1.5k points.", "+50 points on Alice. Bonus claim. Balance 1.5k points."},
		{"a stored sentence: points lost", "−500 points on Alice. Prediction. Balance 1k points.", "−500 points on Alice. Prediction. Balance 1k points."},
		{"a stored sentence: a bet placed", "Bet placed on bob. 1,234 points on Yes.", "Bet placed on bob. 1,234 points on Yes."},
		{"a stored sentence: a bet won", `Won 500 points on bob. "will alice win?". Chose Yes.`, `Won 500 points on bob. "will alice win?". Chose Yes.`},
		{"a stored sentence: a bet lost", `Lost 500 points on bob. "overtime?". Chose No.`, `Lost 500 points on bob. "overtime?". Chose No.`},
		{"a stored sentence: a bet refunded", `Refunded 500 points on bob. "overtime?". Chose No.`, `Refunded 500 points on bob. "overtime?". Chose No.`},
		{"a stored sentence: a bet scheduled", "Bet on Alice in 4m0s.", "Bet on Alice in 4m0s."},
		{"a stored sentence: a bet skipped", "Bet skipped on Alice. The balance is below the minimum.", "Bet skipped on Alice. The balance is below the minimum."},
		{"a stored sentence: a prediction failed", "Failed prediction on Alice. timeout.", "Failed prediction on Alice. timeout."},
		{"a stored sentence: a raid", "Joined the raid from Alice to Bob.", "Joined the raid from Alice to Bob."},
		{"a stored sentence: a moment", "Claimed a moment on Alice.", "Claimed a moment on Alice."},
		{"a stored sentence: a drop claimed", "Claimed Cool Hat. Summer Drops campaign. Progress 60/60 (100%).", "Claimed Cool Hat. Summer Drops campaign. Progress 60/60 (100%)."},
		{"a stored sentence: drop progress", "Drop 40% complete. Game / Camp / chan. 12m watched of 30m.", "Drop 40% complete. Game / Camp / chan. 12m watched of 30m."},
		{"a stored sentence: a mention", `nick mentioned you in #alice. "hello <b>there</b>".`, `nick mentioned you in #alice. "hello <b>there</b>".`},
		{"a stored sentence: a session started", "Session started.", "Session started."},
		{"a stored sentence: a session ended", "Session ended.", "Session ended."},
		{"a stored sentence: streamers loaded", "19 streamers loaded. 3 online: Arrowcs, Fl0m, Bob. 16 offline. Took 3.3 seconds.", "19 streamers loaded. 3 online: Arrowcs, Fl0m, Bob. 16 offline. Took 3.3 seconds."},
		{"a stored sentence: a line no rule knows", "Something nobody planned for. Playing Valorant.", "Something nobody planned for. Playing Valorant."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Sentence(tc.line); got != tc.want {
				t.Errorf("Sentence(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		want Message
	}{
		{"a session starting", "Start session: 'abc'", Message{Headline: "Session started"}},
		{"a streamer going online", "Alice (1.2k points) is Online!", Message{"Alice is online", []string{"Balance 1.2k points"}}},
		{"a streamer going offline", "Alice (1.2k points) is Offline!", Message{"Alice is offline", []string{"Balance 1.2k points"}}},
		{"points gained", "+50 → Alice (12.4k points) - Reason: WATCH", Message{"+50 points on Alice", []string{"Watching", "Balance 12.4k points"}}},
		{"a bet result", "EventPrediction: bob - overtime? - Decision: No - Result: LOSE, Gained: -500", Message{"Lost 500 points on bob", []string{`"overtime?"`, "Chose No"}}},
		{"a bet placed", "Place 1,234 points on: Yes (BLUE), Points: 50,000 for bob", Message{"Bet placed on bob", []string{"1,234 points on Yes"}}},
		{"a bet scheduled", "Place bet after 4m0s for Alice", Message{Headline: "Bet on Alice in 4m0s"}},
		{"a raid", "Joining raid from Alice to Bob", Message{Headline: "Joined the raid from Alice to Bob"}},
		{"a moment", "Alice Claimed Moment", Message{Headline: "Claimed a moment on Alice"}},
		{"a drop claimed", "Claim Cool Hat (Summer Drops) 60/60 (100%)", Message{"Claimed Cool Hat", []string{"Summer Drops campaign", "Progress 60/60 (100%)"}}},
		{"drop progress", "DROP_PROGRESS [####----------------] 40% | Game / Camp / chan / 12m of 30m / remaining 18m / active", Message{"Drop 40% complete", []string{"Game / Camp / chan", "12m watched of 30m"}}},
		{"a mention", "nick at #alice wrote: hello <b>there</b>", Message{"nick mentioned you in #alice", []string{`"hello <b>there</b>"`}}},
		{"a line no rule knows", "Something nobody planned for | Game: Valorant | checked twice", Message{"Something nobody planned for", []string{"Playing Valorant", "checked twice"}}},
		{
			"a game, a watch streak and timestamps after the line",
			"Alice (1.2k points) is Online! | Playing: Valorant | streak length 3 | expires at 2026-11-25 15:04:05.123 +0100 CET | started at 2026-11-25 12:00:00 +0100 CET | streak achievement timestamp 2026-11-25 13:00:00 +0100 CET",
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points", "Watch streak 3"}},
		},
		{"a game after points gained", "+50 → Alice (12.4k points) - Reason: WATCH | Game: Valorant", Message{"+50 points on Alice", []string{"Watching Valorant", "Balance 12.4k points"}}},
		{"a glyph in front", "🥳 Alice is Online!", Message{Headline: "Alice is online"}},
		{"colour codes", "\x1b[38;5;46mAlice\x1b[0m is Online!", Message{Headline: "Alice is online"}},
		{"runs of spaces", "  Alice   is \t Online!  ", Message{Headline: "Alice is online"}},
		{"only a glyph", "🥳", Message{}},
		{"only spaces", "   ", Message{}},
		{"an empty line", "", Message{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Describe(tc.line); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Describe(%q) = %+v, want %+v", tc.line, got, tc.want)
			}
		})
	}
}

func TestDescribeOther(t *testing.T) {
	for _, tc := range []struct {
		name   string
		head   string
		game   string
		streak string
		other  []string
		want   Message
	}{
		{"a line alone", "Something changed on Alice", "", "", nil, Message{Headline: "Something changed on Alice"}},
		{"a game", "Something changed on Alice", "Valorant", "", nil, Message{"Something changed on Alice", []string{"Playing Valorant"}}},
		{"a game the line already names", "Alice now plays VALORANT", "Valorant", "", nil, Message{Headline: "Alice now plays VALORANT"}},
		{"a watch streak", "Something changed on Alice", "", "3", nil, Message{"Something changed on Alice", []string{"Watch streak 3"}}},
		{"a watch streak the line already names", "Watch streak kept on Alice", "", "3", nil, Message{Headline: "Watch streak kept on Alice"}},
		{"other parts", "Something changed on Alice", "", "", []string{"first", "second"}, Message{"Something changed on Alice", []string{"first", "second"}}},
		{"a game, a watch streak and other parts", "Something changed on Alice", "Valorant", "3", []string{"first"}, Message{"Something changed on Alice", []string{"Playing Valorant", "Watch streak 3", "first"}}},
		{"an empty line", "", "", "", nil, Message{Headline: "Something changed"}},
		{"an empty line with a game", "", "Valorant", "", nil, Message{"Something changed", []string{"Playing Valorant"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := describeOther(tc.head, tc.game, tc.streak, tc.other); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeOther(%q, %q, %q, %q) = %+v, want %+v", tc.head, tc.game, tc.streak, tc.other, got, tc.want)
			}
		})
	}
}

func TestSplitNotifyPipes(t *testing.T) {
	type pipes struct {
		head, game, streak string
		other              []string
	}
	for _, tc := range []struct {
		name string
		in   string
		want pipes
	}{
		{"a line alone", "Alice is Online!", pipes{head: "Alice is Online!"}},
		{"a game after playing", "Alice is Online! | Playing: Valorant", pipes{head: "Alice is Online!", game: "Valorant"}},
		{"a game after game", "Alice is Online! | Game: Valorant", pipes{head: "Alice is Online!", game: "Valorant"}},
		{"a game label in upper case", "Alice is Online! | PLAYING: Valorant", pipes{head: "Alice is Online!", game: "Valorant"}},
		{"a game with a colon in its name", "Alice is Online! | Game: Halo: Reach", pipes{head: "Alice is Online!", game: "Halo: Reach"}},
		{"two games", "Alice is Online! | Playing: Valorant | Game: Chess", pipes{head: "Alice is Online!", game: "Valorant"}},
		{"an empty game", "Alice is Online! | Game: ", pipes{head: "Alice is Online!"}},
		{"a watch streak of three", "Alice is Online! | streak length 3", pipes{head: "Alice is Online!", streak: "3"}},
		{"a watch streak of one", "Alice is Online! | streak length 1", pipes{head: "Alice is Online!", streak: "1"}},
		{"a watch streak of zero", "Alice is Online! | streak length 0", pipes{head: "Alice is Online!"}},
		{"a watch streak below zero", "Alice is Online! | streak length -1", pipes{head: "Alice is Online!"}},
		{"a watch streak that is no number", "Alice is Online! | streak length long", pipes{head: "Alice is Online!"}},
		{"when the streak expires", "Alice is Online! | expires at 2026-11-25 15:04:05 +0100 CET", pipes{head: "Alice is Online!"}},
		{"when the stream started", "Alice is Online! | started at 2026-11-25 12:00:00 +0100 CET", pipes{head: "Alice is Online!"}},
		{"when the streak was reached", "Alice is Online! | streak achievement timestamp 2026-11-25 13:00:00 +0100 CET", pipes{head: "Alice is Online!"}},
		{"another part", "Alice is Online! | 3 viewers", pipes{head: "Alice is Online!", other: []string{"3 viewers"}}},
		{"an empty part", "Alice is Online! |  | 3 viewers", pipes{head: "Alice is Online!", other: []string{"3 viewers"}}},
		{
			"every kind of part",
			"Alice is Online! | Playing: Valorant | streak length 3 | expires at noon | 3 viewers | partner",
			pipes{"Alice is Online!", "Valorant", "3", []string{"3 viewers", "partner"}},
		},
		{"an empty line", "", pipes{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			head, game, streak, other := splitNotifyPipes(tc.in)
			if got := (pipes{head, game, streak, other}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("splitNotifyPipes(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestSplitPoints(t *testing.T) {
	type subject struct {
		name, points string
	}
	for _, tc := range []struct {
		name string
		in   string
		want subject
	}{
		{"a name and a balance", "Alice (1.2k points)", subject{"Alice", "1.2k points"}},
		{"a balance of one point", "Alice (1 point)", subject{"Alice", "1 point"}},
		{"a balance in upper case", "Alice (1.2K POINTS)", subject{"Alice", "1.2K POINTS"}},
		{"brackets in the name", "Alice (EU) (1.2k points)", subject{"Alice (EU)", "1.2k points"}},
		{"a name alone", "Alice", subject{"Alice", ""}},
		{"brackets that hold no balance", "Alice (live)", subject{"Alice (live)", ""}},
		{"words after the balance", "Alice (1.2k points) now", subject{"Alice (1.2k points) now", ""}},
		{"a balance alone", "(1.2k points)", subject{"(1.2k points)", ""}},
		{"an empty subject", "", subject{"", ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name, points := splitPoints(tc.in)
			if got := (subject{name, points}); got != tc.want {
				t.Errorf("splitPoints(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestBalance(t *testing.T) {
	for _, tc := range []struct {
		name   string
		points string
		want   string
	}{
		{"some points", "1.2k points", "Balance 1.2k points"},
		{"no points", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := balance(tc.points); got != tc.want {
				t.Errorf("balance(%q) = %q, want %q", tc.points, got, tc.want)
			}
		})
	}
}
