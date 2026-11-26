package twitch

import "testing"

func TestClient_ClaimBonus(t *testing.T) {
	const claimRequest = `{"operationName":"ClaimCommunityPoints","variables":{"input":{"channelID":"123456","claimID":"claim-1"}},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"46aaeebe02c99afdf4fc97c7c0cba964124bf6b0af229395f1f6d1feed05b3d0"}}}`

	t.Run("a bonus", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if err != nil {
			t.Fatalf("ClaimBonus error = %v, want none", err)
		}
		if want := []string{claimVariables}; !equal(fake.twitch.sent("ClaimCommunityPoints"), want) {
			t.Errorf("claims sent = %v, want %v", fake.twitch.sent("ClaimCommunityPoints"), want)
		}
		wantHeaders := gqlHeaders()
		wantHeaders.Set("Cookie", "auth-token=token; persistent=user-id")
		if !equal(fake.twitch.Headers, wantHeaders) {
			t.Errorf("request headers = %v, want %v", fake.twitch.Headers, wantHeaders)
		}
	})

	t.Run("an account with no user id", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.ID = ""
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if err != nil {
			t.Fatalf("ClaimBonus error = %v, want none", err)
		}
		wantHeaders := gqlHeaders()
		wantHeaders.Set("Cookie", "auth-token=token")
		if !equal(fake.twitch.Headers, wantHeaders) {
			t.Errorf("request headers = %v, want %v", fake.twitch.Headers, wantHeaders)
		}
	})

	t.Run("an account that is not logged in", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.Token = ""

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := `claim bonus status 401: {"error":"Unauthorized"}`; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
		wantHeaders := gqlHeaders()
		wantHeaders.Set("Authorization", "OAuth ")
		if !equal(fake.twitch.Headers, wantHeaders) {
			t.Errorf("request headers = %v, want %v, with no cookie", fake.twitch.Headers, wantHeaders)
		}
	})

	t.Run("no claim id", func(t *testing.T) {
		client, fake := setup(t)

		err := client.ClaimBonus(newStreamer(), "")

		if want := "missing claim id"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("a streamer with no channel id", func(t *testing.T) {
		client, _ := setup(t)
		streamer := newStreamer()
		streamer.ChannelID = ""

		err := client.ClaimBonus(streamer, "claim-1")

		if want := "missing streamer channel id"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		client, _ := setup(t)

		err := client.ClaimBonus(nil, "claim-1")

		if want := "missing streamer channel id"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ClaimCommunityPoints", "", down("connection refused"))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := `claim bonus request failed: Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("a 500 answer", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ClaimCommunityPoints", "", code(500, " oops\n"))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := "claim bonus status 500: oops"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("a 500 answer, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.twitch.answer("ClaimCommunityPoints", "", code(500, "oops"))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := "claim bonus status 500"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("an answer that is not JSON", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ClaimCommunityPoints", "", ok("<html>"))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := "claim bonus decode: invalid character '<' looking for beginning of value"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("a claim Twitch refuses", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ClaimCommunityPoints", "", ok(`{"data":{"claimCommunityPoints":{"status":"NOT_FOUND"}}}`))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := "claim bonus status NOT_FOUND (resp=map[data:map[claimCommunityPoints:map[status:NOT_FOUND]]])"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("a claim Twitch refuses, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.twitch.answer("ClaimCommunityPoints", "", ok(`{"data":{"claimCommunityPoints":{"status":"NOT_FOUND"}}}`))

		err := client.ClaimBonus(newStreamer(), "claim-1")

		if want := "claim bonus status NOT_FOUND"; errorText(err) != want {
			t.Errorf("ClaimBonus error = %v, want %s", err, want)
		}
	})

	t.Run("debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		_ = client.ClaimBonus(newStreamer(), "claim-1")

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: ClaimCommunityPoints status=200",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("deep debug logging on", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.logger.Deep = true
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		_ = client.ClaimBonus(newStreamer(), "claim-1")

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: ClaimCommunityPoints status=200",
			"deep: ClaimCommunityPoints status=200 headers=map[Authorization:[OAuth token] Client-Id:[ue6666qo983tsx6so1t0vnawi233wa] Client-Session-Id:[session] Client-Version:[" + buildID + "] Content-Type:[application/json] Cookie:[auth-token=token; persistent=user-id] User-Agent:[agent] X-Device-Id:[device]]" +
				" req=" + claimRequest + " resp=" + claimSucceeded,
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("deep debug logging on, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		fake.logger.Debug = true
		fake.logger.Deep = true
		fake.twitch.answer("ClaimCommunityPoints", "", ok(claimSucceeded))

		_ = client.ClaimBonus(newStreamer(), "claim-1")

		wantLines := []string{
			"debug: Client version updated to " + buildID,
			"debug: ClaimCommunityPoints status=200",
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})
}

func TestClaimBonusFailure(t *testing.T) {
	tests := []struct {
		name       string
		result     map[string]interface{}
		anonymized bool
		want       string
	}{
		{
			"a claimed bonus",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "SUCCESS"}}},
			false,
			"<nil>",
		},
		{
			"a bonus claimed before",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "ALREADY_CLAIMED"}}},
			false,
			"<nil>",
		},
		{
			"a status in lower case",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "success"}}},
			false,
			"<nil>",
		},
		{
			"an empty status",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": ""}}},
			false,
			"<nil>",
		},
		{"an empty answer", map[string]interface{}{}, false, "<nil>"},
		{"no answer", nil, false, "<nil>"},
		{
			"another status",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "not_found"}}},
			false,
			"claim bonus status NOT_FOUND (resp=map[data:map[claimCommunityPoints:map[status:not_found]]])",
		},
		{
			"another status, anonymized",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "not_found"}}},
			true,
			"claim bonus status NOT_FOUND",
		},
		{
			"GQL errors",
			map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": "service timeout"}}},
			false,
			"claim bonus gql errors: [map[message:service timeout]]",
		},
		{
			"GQL errors, anonymized",
			map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": "service timeout"}}},
			true,
			"claim bonus gql errors",
		},
		{
			"GQL errors beside a claimed bonus",
			map[string]interface{}{
				"errors": []interface{}{},
				"data":   map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"status": "SUCCESS"}},
			},
			false,
			"claim bonus gql errors: []",
		},
		{
			"an error message",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"error": map[string]interface{}{"message": "too late"}}}},
			false,
			"claim bonus error: too late (resp=map[data:map[claimCommunityPoints:map[error:map[message:too late]]]])",
		},
		{
			"an error message, anonymized",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"error": map[string]interface{}{"message": "too late"}}}},
			true,
			"claim bonus error",
		},
		{
			"an error with no message",
			map[string]interface{}{"data": map[string]interface{}{"claimCommunityPoints": map[string]interface{}{"error": map[string]interface{}{"code": "TOO_LATE"}}}},
			false,
			"<nil>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorText(claimBonusFailure(tt.result, tt.anonymized)); got != tt.want {
				t.Errorf("claimBonusFailure(%v, %t) = %s, want %s", tt.result, tt.anonymized, got, tt.want)
			}
		})
	}
}
