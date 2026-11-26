package twitch

import (
	"errors"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func TestClient_EachEdge(t *testing.T) {
	const (
		path       = "data.user.follows"
		firstPage  = `{"cursor":"","limit":50,"order":"DESC"}`
		secondPage = `{"cursor":"c2","limit":50,"order":"DESC"}`
	)
	follows := func() twitchapi.GQLPersistedOperation {
		return twitchapi.GQLOperations.ChannelFollows.With(map[string]interface{}{"limit": 50, "order": "DESC"})
	}
	// cursors collects the cursor of every edge visited, "-" for an edge without one.
	cursors := func(seen *[]string) func(map[string]interface{}) (bool, error) {
		return func(edge map[string]interface{}) (bool, error) {
			cursor, found := edge["cursor"].(string)
			if !found {
				cursor = "-"
			}
			*seen = append(*seen, cursor)
			return false, nil
		}
	}

	tests := []struct {
		name      string
		first     reply
		second    reply
		wantSeen  []string
		wantAsked []string
		wantEnd   pagesEnd
		wantErr   string
	}{
		{
			name:      "one page",
			first:     ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c1"},{"cursor":"c2"}],"pageInfo":{"hasNextPage":false}}}}}`),
			wantSeen:  []string{"c1", "c2"},
			wantAsked: []string{firstPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "two pages, the second asked for with the last cursor of the first",
			first:     ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c1"},{"cursor":"c2"}],"pageInfo":{"hasNextPage":true}}}}}`),
			second:    ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c3"}],"pageInfo":{"hasNextPage":false}}}}}`),
			wantSeen:  []string{"c1", "c2", "c3"},
			wantAsked: []string{firstPage, secondPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "a last edge without a cursor, after one that has it",
			first:     ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c2"},{"node":{}}],"pageInfo":{"hasNextPage":true}}}}}`),
			second:    ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c3"}],"pageInfo":{"hasNextPage":false}}}}}`),
			wantSeen:  []string{"c2", "-", "c3"},
			wantAsked: []string{firstPage, secondPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "a page without a cursor that claims a next page",
			first:     ok(`{"data":{"user":{"follows":{"edges":[{"node":{}}],"pageInfo":{"hasNextPage":true}}}}}`),
			wantSeen:  []string{"-"},
			wantAsked: []string{firstPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "an empty page that claims a next page",
			first:     ok(`{"data":{"user":{"follows":{"edges":[],"pageInfo":{"hasNextPage":true}}}}}`),
			wantAsked: []string{firstPage},
			wantEnd:   pagesCutShort,
			wantErr:   "<nil>",
		},
		{
			name:      "an empty last page",
			first:     ok(`{"data":{"user":{"follows":{"edges":[],"pageInfo":{"hasNextPage":false}}}}}`),
			wantAsked: []string{firstPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "an edge that is not an object",
			first:     ok(`{"data":{"user":{"follows":{"edges":["broken",{"cursor":"c2"}],"pageInfo":{"hasNextPage":false}}}}}`),
			wantSeen:  []string{"-", "c2"},
			wantAsked: []string{firstPage},
			wantEnd:   pagesDone,
			wantErr:   "<nil>",
		},
		{
			name:      "an answer with no listing",
			first:     ok(`{"data":{"user":null}}`),
			wantAsked: []string{firstPage},
			wantEnd:   pagesMissing,
			wantErr:   "<nil>",
		},
		{
			name:      "a listing that is not an object",
			first:     ok(`{"data":{"user":{"follows":"none"}}}`),
			wantAsked: []string{firstPage},
			wantEnd:   pagesMissing,
			wantErr:   "<nil>",
		},
		{
			name:      "twitch down for the second page",
			first:     ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c2"}],"pageInfo":{"hasNextPage":true}}}}}`),
			second:    down("connection refused"),
			wantSeen:  []string{"c2"},
			wantAsked: []string{firstPage, secondPage},
			wantEnd:   pagesDone,
			wantErr:   `Post "https://gql.twitch.tv/gql": connection refused`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fake := setup(t)
			fake.twitch.answer("ChannelFollows", firstPage, tt.first)
			fake.twitch.answer("ChannelFollows", secondPage, tt.second)
			var seen []string

			end, err := client.eachEdge(follows(), path, cursors(&seen))

			if errorText(err) != tt.wantErr {
				t.Errorf("eachEdge error = %v, want %s", err, tt.wantErr)
			}
			if end != tt.wantEnd {
				t.Errorf("eachEdge ended with %d, want %d", end, tt.wantEnd)
			}
			if !equal(seen, tt.wantSeen) {
				t.Errorf("edges visited = %q, want %q", seen, tt.wantSeen)
			}
			if asked := fake.twitch.sent("ChannelFollows"); !equal(asked, tt.wantAsked) {
				t.Errorf("pages asked for = %q, want %q", asked, tt.wantAsked)
			}
		})
	}

	const twoPages = `{"data":{"user":{"follows":{"edges":[{"cursor":"c1"},{"cursor":"c2"}],"pageInfo":{"hasNextPage":true}}}}}`

	t.Run("a visitor that stops the walk", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(twoPages))
		visited := 0

		end, err := client.eachEdge(follows(), path, func(map[string]interface{}) (bool, error) {
			visited++
			return true, nil
		})

		if err != nil || end != pagesDone {
			t.Errorf("eachEdge = %d, %v, want %d and no error", end, err, pagesDone)
		}
		if visited != 1 {
			t.Errorf("edges visited = %d, want 1", visited)
		}
		if asked := fake.twitch.sent("ChannelFollows"); len(asked) != 1 {
			t.Errorf("pages asked for = %q, want the first only", asked)
		}
	})

	t.Run("a visitor that fails", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(twoPages))
		failure := errors.New("clip unwatchable")
		visited := 0

		_, err := client.eachEdge(follows(), path, func(map[string]interface{}) (bool, error) {
			visited++
			return false, failure
		})

		if err != failure {
			t.Errorf("eachEdge error = %v, want %v", err, failure)
		}
		if visited != 1 {
			t.Errorf("edges visited = %d, want 1", visited)
		}
	})

	t.Run("the operation passed in keeps its variables", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(twoPages))
		fake.twitch.answer("ChannelFollows", secondPage, ok(`{"data":{"user":{"follows":{"edges":[],"pageInfo":{"hasNextPage":false}}}}}`))
		op := follows()

		_, _ = client.eachEdge(op, path, func(map[string]interface{}) (bool, error) { return false, nil })

		if _, found := op.Variables["cursor"]; found {
			t.Errorf("variables = %v, want no cursor left in them", op.Variables)
		}
	})
}
