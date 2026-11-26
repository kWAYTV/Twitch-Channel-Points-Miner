package notify

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/notify/notifytest"
)

const aliceOnlineMatrixBody = `{"msgtype":"m.text","body":"🟢 Alice is online\nPlaying Valorant · Balance 1.2k points","format":"org.matrix.custom.html","formatted_body":"🟢 <b>Alice is online</b><br>Playing Valorant · Balance 1.2k points"}`

// aliceOnlineOnMatrix is the request aliceOnline arrives as, under this transaction id.
func aliceOnlineOnMatrix(txnID string) notifytest.MatrixRequest {
	return notifytest.MatrixRequest{
		Method:        "PUT",
		Path:          "/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/" + txnID,
		Authorization: "Bearer syt_secret",
		ContentType:   "application/json",
		Body:          aliceOnlineMatrixBody,
	}
}

// numberedTxnIDs names messages txn-1, txn-2 and so on.
func numberedTxnIDs() func() string {
	sent := 0
	return func() string {
		sent++
		return "txn-" + strconv.Itoa(sent)
	}
}

func setupMatrix(t *testing.T, eventNames ...string) (*MatrixNotifier, *notifytest.FakeMatrix, *fakeLog) {
	t.Helper()
	matrix := notifytest.NewFakeMatrix(t)
	log := &fakeLog{}
	settings := MatrixSettings{Homeserver: matrix.URL, AccessToken: matrix.Token, RoomID: matrix.RoomID, Events: eventNames}
	return newMatrixNotifier(settings, numberedTxnIDs(), log.Printf), matrix, log
}

func TestNewMatrixNotifier(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings MatrixSettings
	}{
		{"no homeserver", MatrixSettings{AccessToken: "syt_secret", RoomID: "!room:example.org"}},
		{"no access token", MatrixSettings{Homeserver: "https://matrix.example.org", RoomID: "!room:example.org"}},
		{"no room id", MatrixSettings{Homeserver: "https://matrix.example.org", AccessToken: "syt_secret"}},
		{"only spaces for a homeserver", MatrixSettings{Homeserver: "  ", AccessToken: "syt_secret", RoomID: "!room:example.org"}},
		{"only spaces for an access token", MatrixSettings{Homeserver: "https://matrix.example.org", AccessToken: "  ", RoomID: "!room:example.org"}},
		{"only spaces for a room id", MatrixSettings{Homeserver: "https://matrix.example.org", AccessToken: "syt_secret", RoomID: "  "}},
		{"no settings", MatrixSettings{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := &fakeLog{}

			got := newMatrixNotifier(tc.settings, numberedTxnIDs(), log.Printf)

			if got != nil {
				t.Errorf("newMatrixNotifier(%+v) = %+v, want nil", tc.settings, got)
			}
			if len(log.Lines) != 0 {
				t.Errorf("logged %q, want nothing", log.Lines)
			}
		})
	}

	t.Run("a homeserver, an access token and a room id", func(t *testing.T) {
		settings := MatrixSettings{Homeserver: "https://matrix.example.org", AccessToken: "syt_secret", RoomID: "!room:example.org"}
		if got := NewMatrixNotifier(settings); got == nil {
			t.Errorf("NewMatrixNotifier(%+v) = nil, want a notifier", settings)
		}
	})

	t.Run("a homeserver that is no http address", func(t *testing.T) {
		log := &fakeLog{}
		settings := MatrixSettings{Homeserver: "matrix.example.org", AccessToken: "syt_secret", RoomID: "!room:example.org"}

		got := newMatrixNotifier(settings, numberedTxnIDs(), log.Printf)

		if got != nil {
			t.Errorf("newMatrixNotifier(%+v) = %+v, want nil", settings, got)
		}
		want := []string{"matrix: notifications are off: homeserver has to be an address starting with http:// or https://"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})
}

func TestMatrixNotifier_Send(t *testing.T) {
	t.Run("an event with no events list", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := matrix.Requests(), []notifytest.MatrixRequest{aliceOnlineOnMatrix("txn-1")}; !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
		if len(log.Lines) != 0 {
			t.Errorf("logged %q, want nothing", log.Lines)
		}
	})

	t.Run("a homeserver with a slash at the end", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		settings := MatrixSettings{Homeserver: matrix.URL + "/", AccessToken: matrix.Token, RoomID: matrix.RoomID}
		notifier := newMatrixNotifier(settings, numberedTxnIDs(), (&fakeLog{}).Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := matrix.Requests(), []notifytest.MatrixRequest{aliceOnlineOnMatrix("txn-1")}; !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("settings with spaces around them", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		settings := MatrixSettings{Homeserver: " " + matrix.URL + " ", AccessToken: " " + matrix.Token + "\n", RoomID: " " + matrix.RoomID + " "}
		notifier := newMatrixNotifier(settings, numberedTxnIDs(), (&fakeLog{}).Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := matrix.Requests(), []notifytest.MatrixRequest{aliceOnlineOnMatrix("txn-1")}; !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("a room id with characters a path reserves", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		matrix.RoomID = "!a/b?c d#e:example.org"
		settings := MatrixSettings{Homeserver: matrix.URL, AccessToken: matrix.Token, RoomID: "!a/b?c d#e:example.org"}
		notifier := newMatrixNotifier(settings, numberedTxnIDs(), (&fakeLog{}).Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []notifytest.MatrixRequest{{
			Method:        "PUT",
			Path:          "/_matrix/client/v3/rooms/%21a%2Fb%3Fc%20d%23e:example.org/send/m.room.message/txn-1",
			Authorization: "Bearer syt_secret",
			ContentType:   "application/json",
			Body:          aliceOnlineMatrixBody,
		}}
		if got := matrix.Requests(); !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("text with characters HTML and JSON reserve", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t)

		notifier.Send(Message{"nick <3 & you", []string{`"<b>hi</b> \ bye"`}}, events.EventChatMention)

		want := []notifytest.MatrixRequest{{
			Method:        "PUT",
			Path:          "/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/txn-1",
			Authorization: "Bearer syt_secret",
			ContentType:   "application/json",
			Body:          `{"msgtype":"m.text","body":"💬 nick <3 & you\n\"<b>hi</b> \\ bye\"","format":"org.matrix.custom.html","formatted_body":"💬 <b>nick &lt;3 &amp; you</b><br>\"&lt;b&gt;hi&lt;/b&gt; \\ bye\""}`,
		}}
		if got := matrix.Requests(); !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("a quiet event with no events list", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t)

		notifier.Send(aliceOnline, events.EventGainForWatch)

		if got := matrix.Requests(); len(got) != 0 {
			t.Errorf("matrix received %+v, want nothing", got)
		}
	})

	t.Run("a quiet event named in the events list", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t, "gain_for_watch")

		notifier.Send(Message{"+10 points on Alice", []string{"Watching"}}, events.EventGainForWatch)

		want := []notifytest.MatrixRequest{{
			Method:        "PUT",
			Path:          "/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/txn-1",
			Authorization: "Bearer syt_secret",
			ContentType:   "application/json",
			Body:          `{"msgtype":"m.text","body":"➕ +10 points on Alice\nWatching","format":"org.matrix.custom.html","formatted_body":"➕ <b>+10 points on Alice</b><br>Watching"}`,
		}}
		if got := matrix.Requests(); !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("a quiet event with a star in the events list", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t, "*")

		notifier.Send(aliceOnline, events.EventGainForWatch)

		want := []notifytest.MatrixRequest{{
			Method:        "PUT",
			Path:          "/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/txn-1",
			Authorization: "Bearer syt_secret",
			ContentType:   "application/json",
			Body:          `{"msgtype":"m.text","body":"➕ Alice is online\nPlaying Valorant · Balance 1.2k points","format":"org.matrix.custom.html","formatted_body":"➕ <b>Alice is online</b><br>Playing Valorant · Balance 1.2k points"}`,
		}}
		if got := matrix.Requests(); !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("an event missing from the events list", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t, "streamer_offline", "BET_LOSE")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got := matrix.Requests(); len(got) != 0 {
			t.Errorf("matrix received %+v, want nothing", got)
		}
	})

	t.Run("no event", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t)

		notifier.Send(aliceOnline, "")

		if got := matrix.Requests(); len(got) != 0 {
			t.Errorf("matrix received %+v, want nothing", got)
		}
	})

	t.Run("a notifier that is not set up", func(t *testing.T) {
		var notifier *MatrixNotifier

		notifier.Send(aliceOnline, events.EventStreamerOnline)
	})

	t.Run("the same event twice", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t)
		notifier.Send(aliceOnline, events.EventStreamerOnline)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := matrix.Requests(), []notifytest.MatrixRequest{aliceOnlineOnMatrix("txn-1"), aliceOnlineOnMatrix("txn-2")}; !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})

	t.Run("the same event twice from a notifier that names its own messages", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		notifier := NewMatrixNotifier(MatrixSettings{Homeserver: matrix.URL, AccessToken: matrix.Token, RoomID: matrix.RoomID})
		notifier.Send(aliceOnline, events.EventStreamerOnline)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		got := matrix.Requests()
		if len(got) != 2 {
			t.Fatalf("matrix received %+v, want two requests", got)
		}
		shape := regexp.MustCompile(`^/_matrix/client/v3/rooms/%21room:example\.org/send/m\.room\.message/[0-9a-f]{32}$`)
		for _, request := range got {
			if !shape.MatchString(request.Path) {
				t.Errorf("matrix received a request for %q, want one matching %s", request.Path, shape)
			}
		}
		if got[0].Path == got[1].Path {
			t.Errorf("both requests went to %q, want two transaction ids", got[0].Path)
		}
	})

	t.Run("another account's access token", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		log := &fakeLog{}
		settings := MatrixSettings{Homeserver: matrix.URL, AccessToken: "syt_other", RoomID: matrix.RoomID}
		notifier := newMatrixNotifier(settings, numberedTxnIDs(), log.Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: Invalid access token passed. (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
		if got := matrix.Requests(); len(got) != 0 {
			t.Errorf("matrix kept %+v, want nothing", got)
		}
	})

	t.Run("a room the account is not in", func(t *testing.T) {
		matrix := notifytest.NewFakeMatrix(t)
		log := &fakeLog{}
		settings := MatrixSettings{Homeserver: matrix.URL, AccessToken: matrix.Token, RoomID: "!other:example.org"}
		notifier := newMatrixNotifier(settings, numberedTxnIDs(), log.Printf)

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: User not in room !other:example.org (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("the homeserver answering 401", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)
		matrix.Refuse(401, "M_UNKNOWN_TOKEN", "Access token has expired")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: Access token has expired (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("the homeserver answering 500", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)
		matrix.Refuse(500, "M_UNKNOWN", "Internal server error")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: Internal server error (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("the homeserver answering 500 with no reason", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)
		matrix.RefuseWith(500, "<html>Internal Server Error</html>")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: 500 Internal Server Error (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("the homeserver closing the connection", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)
		matrix.Drop()

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: request failed (network error or timeout) (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("a second refusal", func(t *testing.T) {
		notifier, matrix, log := setupMatrix(t)
		matrix.Refuse(401, "M_UNKNOWN_TOKEN", "Access token has expired")
		notifier.Send(aliceOnline, events.EventStreamerOnline)
		matrix.Refuse(500, "M_UNKNOWN", "Internal server error")

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		want := []string{"matrix: notification not delivered: Access token has expired (further failures are not logged)"}
		if !reflect.DeepEqual(log.Lines, want) {
			t.Errorf("logged %q, want %q", log.Lines, want)
		}
	})

	t.Run("an event after a refusal", func(t *testing.T) {
		notifier, matrix, _ := setupMatrix(t)
		matrix.Drop()
		notifier.Send(aliceOnline, events.EventStreamerOnline)
		matrix.Accept()

		notifier.Send(aliceOnline, events.EventStreamerOnline)

		if got, want := matrix.Requests(), []notifytest.MatrixRequest{aliceOnlineOnMatrix("txn-2")}; !reflect.DeepEqual(got, want) {
			t.Errorf("matrix received %+v, want %+v", got, want)
		}
	})
}

func TestMatrixSendURL(t *testing.T) {
	for _, tc := range []struct {
		name       string
		homeserver string
		roomID     string
		want       string
		wantOK     bool
	}{
		{"an https homeserver", "https://matrix.example.org", "!room:example.org", "https://matrix.example.org/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"an http homeserver with a port", "http://localhost:8008", "!room:example.org", "http://localhost:8008/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"a slash at the end", "https://matrix.example.org/", "!room:example.org", "https://matrix.example.org/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"two slashes at the end", "https://matrix.example.org//", "!room:example.org", "https://matrix.example.org/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"a homeserver under a path", "https://example.org/matrix/", "!room:example.org", "https://example.org/matrix/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"a scheme in capitals", "HTTPS://matrix.example.org", "!room:example.org", "HTTPS://matrix.example.org/_matrix/client/v3/rooms/%21room:example.org/send/m.room.message/", true},
		{"a room id with characters a path reserves", "https://matrix.example.org", "!a/b?c d#e:example.org", "https://matrix.example.org/_matrix/client/v3/rooms/%21a%2Fb%3Fc%20d%23e:example.org/send/m.room.message/", true},
		{"a room alias", "https://matrix.example.org", "#miner:example.org", "https://matrix.example.org/_matrix/client/v3/rooms/%23miner:example.org/send/m.room.message/", true},
		{"no scheme", "matrix.example.org", "!room:example.org", "", false},
		{"no scheme and a port", "matrix.example.org:8448", "!room:example.org", "", false},
		{"another scheme", "ftp://matrix.example.org", "!room:example.org", "", false},
		{"a scheme and no host", "https://", "!room:example.org", "", false},
		{"an address that can't be read", "https://matrix.example.org:port", "!room:example.org", "", false},
		{"no homeserver", "", "!room:example.org", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := matrixSendURL(tc.homeserver, tc.roomID)
			if got != tc.want || ok != tc.wantOK {
				t.Errorf("matrixSendURL(%q, %q) = %q, %v, want %q, %v", tc.homeserver, tc.roomID, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestMatrixBody(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		want    string
	}{
		{"a headline and details", events.EventStreamerOnline, aliceOnline, aliceOnlineMatrixBody},
		{
			"a headline alone",
			events.EventStreamerOnline,
			Message{Headline: "Alice is online"},
			`{"msgtype":"m.text","body":"🟢 Alice is online","format":"org.matrix.custom.html","formatted_body":"🟢 <b>Alice is online</b>"}`,
		},
		{
			"characters HTML and JSON reserve",
			events.EventChatMention,
			Message{"nick <3 & you", []string{`"<b>hi</b> \ bye"`}},
			`{"msgtype":"m.text","body":"💬 nick <3 & you\n\"<b>hi</b> \\ bye\"","format":"org.matrix.custom.html","formatted_body":"💬 <b>nick &lt;3 &amp; you</b><br>\"&lt;b&gt;hi&lt;/b&gt; \\ bye\""}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(matrixBody(tc.event, tc.message)); got != tc.want {
				t.Errorf("matrixBody(%q, %+v) = %s, want %s", tc.event, tc.message, got, tc.want)
			}
		})
	}
}

func TestFormatMatrix(t *testing.T) {
	for _, tc := range []struct {
		name      string
		event     events.Event
		message   Message
		wantPlain string
		wantHTML  string
	}{
		{
			"a headline and details",
			events.EventStreamerOnline,
			Message{"Alice is online", []string{"Playing Valorant", "Balance 1.2k points"}},
			"🟢 Alice is online\nPlaying Valorant · Balance 1.2k points",
			"🟢 <b>Alice is online</b><br>Playing Valorant · Balance 1.2k points",
		},
		{"a headline alone", events.EventStreamerOnline, Message{Headline: "Alice is online"}, "🟢 Alice is online", "🟢 <b>Alice is online</b>"},
		{"no headline", events.EventStartup, Message{}, "▶️ Session started", "▶️ <b>Session started</b>"},
		{"details without a headline", events.EventBetWin, Message{"", []string{"Chose Yes"}}, "📈 Bet won\nChose Yes", "📈 <b>Bet won</b><br>Chose Yes"},
		{"an event without a glyph", "NONSENSE", Message{Headline: "Something changed"}, "Something changed", "<b>Something changed</b>"},
		{
			"characters HTML reserves",
			events.EventChatMention,
			Message{"nick <3 & you", []string{`"<b>hi</b> & bye"`}},
			"💬 nick <3 & you\n\"<b>hi</b> & bye\"",
			"💬 <b>nick &lt;3 &amp; you</b><br>\"&lt;b&gt;hi&lt;/b&gt; &amp; bye\"",
		},
		{
			"a headline of exactly the limit",
			events.EventBetLose,
			Message{Headline: strings.Repeat("é", 256)},
			"📉 " + strings.Repeat("é", 256),
			"📉 <b>" + strings.Repeat("é", 256) + "</b>",
		},
		{
			"a headline one rune over the limit",
			events.EventBetLose,
			Message{Headline: strings.Repeat("é", 257)},
			"📉 " + strings.Repeat("é", 255) + "…",
			"📉 <b>" + strings.Repeat("é", 255) + "…</b>",
		},
		{
			"details that fill the HTML exactly",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4082)}},
			"📉 x\n" + strings.Repeat("é", 4082),
			"📉 <b>x</b><br>" + strings.Repeat("é", 4082),
		},
		{
			"details one rune too long for the HTML",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4083)}},
			"📉 x\n" + strings.Repeat("é", 4083),
			"📉 <b>x</b><br>" + strings.Repeat("é", 4081) + "…",
		},
		{
			"details that fill the plain text exactly",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4092)}},
			"📉 x\n" + strings.Repeat("é", 4092),
			"📉 <b>x</b><br>" + strings.Repeat("é", 4081) + "…",
		},
		{
			"details one rune too long for the plain text",
			events.EventBetLose,
			Message{"x", []string{strings.Repeat("é", 4093)}},
			"📉 x\n" + strings.Repeat("é", 4091) + "…",
			"📉 <b>x</b><br>" + strings.Repeat("é", 4081) + "…",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plain, html := formatMatrix(tc.event, tc.message)
			if plain != tc.wantPlain {
				t.Errorf("formatMatrix(%q, %+v) plain text = %q, want %q", tc.event, tc.message, plain, tc.wantPlain)
			}
			if html != tc.wantHTML {
				t.Errorf("formatMatrix(%q, %+v) HTML = %q, want %q", tc.event, tc.message, html, tc.wantHTML)
			}
		})
	}
}

func TestPlainMessage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   events.Event
		message Message
		limit   int
		want    string
	}{
		{"details that fill the limit exactly", events.EventBetLose, Message{"x", []string{"abcdef"}}, 10, "📉 x\nabcdef"},
		{"details one rune over the limit", events.EventBetLose, Message{"x", []string{"abcdefg"}}, 10, "📉 x\nabcde…"},
		{"a limit with room for one rune of details", events.EventBetLose, Message{"x", []string{"abcdefg"}}, 5, "📉 x\n…"},
		{"a limit with no room for details", events.EventBetLose, Message{"x", []string{"abcdefg"}}, 4, "📉 x"},
		{"no details", events.EventBetLose, Message{Headline: "x"}, 10, "📉 x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := plainMessage(tc.event, tc.message, tc.limit); got != tc.want {
				t.Errorf("plainMessage(%q, %+v, %d) = %q, want %q", tc.event, tc.message, tc.limit, got, tc.want)
			}
		})
	}
}
