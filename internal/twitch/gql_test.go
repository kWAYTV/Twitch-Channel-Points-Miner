package twitch

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

var raid = twitchapi.GQLPersistedOperation{
	OperationName: "JoinRaid",
	Variables:     map[string]interface{}{"input": map[string]interface{}{"raidID": "raid-1"}},
}

const raidVariables = `{"input":{"raidID":"raid-1"}}`

func TestClient_PostGQL(t *testing.T) {
	t.Run("an operation", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`{"data":{"joinRaid":{"raidID":"raid-1"}}}`))

		got, err := client.PostGQL(raid)

		if err != nil {
			t.Fatalf("PostGQL error = %v, want none", err)
		}
		want := map[string]interface{}{"data": map[string]interface{}{"joinRaid": map[string]interface{}{"raidID": "raid-1"}}}
		if !equal(got, want) {
			t.Errorf("PostGQL = %v, want %v", got, want)
		}
		if want := []gqlRequest{{Operation: "JoinRaid", Variables: raidVariables}}; !equal(fake.twitch.Requests, want) {
			t.Errorf("twitch received %v, want %v", fake.twitch.Requests, want)
		}
		if want := gqlHeaders(); !equal(fake.twitch.Headers, want) {
			t.Errorf("request headers = %v, want %v", fake.twitch.Headers, want)
		}
	})

	t.Run("no payload", func(t *testing.T) {
		client, fake := setup(t)

		got, err := client.PostGQL(nil)

		if err != nil {
			t.Fatalf("PostGQL error = %v, want none", err)
		}
		if want := (map[string]interface{}{}); !equal(got, want) {
			t.Errorf("PostGQL = %v, want %v", got, want)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("JoinRaid", "", down("connection refused"))

		got, err := client.PostGQL(raid)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("PostGQL error = %v, want %s", err, want)
		}
		if got != nil {
			t.Errorf("PostGQL = %v, want nothing", got)
		}
		want := []string{
			"debug: Client version updated to " + buildID,
			`debug: GQL request failed: Post "https://gql.twitch.tv/gql": connection refused`,
		}
		if !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("an answer that is not JSON", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`<html>`))

		_, err := client.PostGQL(raid)

		if want := "invalid character '<' looking for beginning of value"; errorText(err) != want {
			t.Errorf("PostGQL error = %v, want %s", err, want)
		}
	})

	t.Run("a 503 answer in JSON", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", code(503, `{"error":"Service Unavailable"}`))

		got, err := client.PostGQL(raid)

		if err != nil {
			t.Fatalf("PostGQL error = %v, want none: the status is not read", err)
		}
		if want := (map[string]interface{}{"error": "Service Unavailable"}); !equal(got, want) {
			t.Errorf("PostGQL = %v, want %v", got, want)
		}
	})

	t.Run("an account that is not logged in", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.Token = ""
		fake.twitch.answer("JoinRaid", "", ok(`{}`))

		got, err := client.PostGQL(raid)

		if err != nil {
			t.Fatalf("PostGQL error = %v, want none: the status is not read", err)
		}
		if want := (map[string]interface{}{"error": "Unauthorized"}); !equal(got, want) {
			t.Errorf("PostGQL = %v, want %v", got, want)
		}
	})

	t.Run("debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("JoinRaid", "", ok(`{}`))

		_, _ = client.PostGQL(raid)

		want := []string{
			"debug: Client version updated to " + buildID,
			"debug: GQL JoinRaid (join raid) | Status 200",
		}
		if !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("deep debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Deep = true
		fake.twitch.answer("JoinRaid", "", ok(` {"data":{}} `))

		got, err := client.PostGQL(raid)

		if err != nil {
			t.Fatalf("PostGQL error = %v, want none", err)
		}
		if want := (map[string]interface{}{"data": map[string]interface{}{}}); !equal(got, want) {
			t.Errorf("PostGQL = %v, want %v", got, want)
		}
		want := []string{
			"deep: GQL JoinRaid | Status 200 | Headers: map[Authorization:[OAuth token] Client-Id:[ue6666qo983tsx6so1t0vnawi233wa] Client-Session-Id:[session] Client-Version:[" + buildID + "] Content-Type:[application/json] User-Agent:[agent] X-Device-Id:[device]]" +
				` | Request: {"operationName":"JoinRaid","variables":{"input":{"raidID":"raid-1"}},"extensions":{"persistedQuery":{"version":0,"sha256Hash":""}}} | Response: {"data":{}}`,
		}
		if !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("deep debug logging on, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.logger.Deep = true
		fake.twitch.answer("JoinRaid", "", ok(`{"data":{}}`))

		_, _ = client.PostGQL(raid)

		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("deep debug logging on, an answer that is not JSON", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Deep = true
		fake.twitch.answer("JoinRaid", "", ok(`<html>`))

		_, err := client.PostGQL(raid)

		if want := "invalid character '<' looking for beginning of value"; errorText(err) != want {
			t.Errorf("PostGQL error = %v, want %s", err, want)
		}
	})
}

type raidAnswer struct {
	Data struct {
		JoinRaid struct {
			RaidID string `json:"raidID"`
		} `json:"joinRaid"`
	} `json:"data"`
}

func TestClient_PostGQLDecode(t *testing.T) {
	t.Run("an operation", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`{"data":{"joinRaid":{"raidID":"raid-1"}}}`))
		var got raidAnswer

		err := client.PostGQLDecode(raid, &got)

		if err != nil {
			t.Fatalf("PostGQLDecode error = %v, want none", err)
		}
		if got.Data.JoinRaid.RaidID != "raid-1" {
			t.Errorf("decoded raid id = %q, want %q", got.Data.JoinRaid.RaidID, "raid-1")
		}
		if want := []string{raidVariables}; !equal(fake.twitch.sent("JoinRaid"), want) {
			t.Errorf("twitch received %v, want %v", fake.twitch.sent("JoinRaid"), want)
		}
	})

	t.Run("no payload", func(t *testing.T) {
		client, fake := setup(t)
		var got raidAnswer

		err := client.PostGQLDecode(nil, &got)

		if err != nil {
			t.Fatalf("PostGQLDecode error = %v, want none", err)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("nowhere to decode into", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`<html>`))

		err := client.PostGQLDecode(raid, nil)

		if err != nil {
			t.Fatalf("PostGQLDecode error = %v, want none: the answer is not read", err)
		}
		if want := []string{raidVariables}; !equal(fake.twitch.sent("JoinRaid"), want) {
			t.Errorf("twitch received %v, want %v", fake.twitch.sent("JoinRaid"), want)
		}
	})

	t.Run("nowhere to decode into, deep debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Deep = true
		fake.twitch.answer("JoinRaid", "", ok(`<html>`))

		err := client.PostGQLDecode(raid, nil)

		if err != nil {
			t.Fatalf("PostGQLDecode error = %v, want none: the answer is not read", err)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", down("connection refused"))
		var got raidAnswer

		err := client.PostGQLDecode(raid, &got)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("PostGQLDecode error = %v, want %s", err, want)
		}
	})
}
