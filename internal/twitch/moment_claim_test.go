package twitch

import "testing"

func TestClient_ClaimMoment(t *testing.T) {
	t.Run("a moment", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("CommunityMomentCallout_Claim", "", ok(`{"data":{"claimCommunityMoment":{"moment":{"id":"moment-1"}}}}`))

		err := client.ClaimMoment(newStreamer(), "moment-1")

		if err != nil {
			t.Fatalf("ClaimMoment error = %v, want none", err)
		}
		want := []string{`{"input":{"momentID":"moment-1"}}`}
		if got := fake.twitch.sent("CommunityMomentCallout_Claim"); !equal(got, want) {
			t.Errorf("moment claims sent = %v, want %v", got, want)
		}
	})

	t.Run("no moment id", func(t *testing.T) {
		client, fake := setup(t)

		err := client.ClaimMoment(newStreamer(), "")

		if err != nil {
			t.Fatalf("ClaimMoment error = %v, want none", err)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("CommunityMomentCallout_Claim", "", down("connection refused"))

		err := client.ClaimMoment(newStreamer(), "moment-1")

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("ClaimMoment error = %v, want %s", err, want)
		}
	})
}
