package twitch

import "testing"

func TestClient_JoinRaid(t *testing.T) {
	t.Run("a raid", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`{"data":{"joinRaid":{"raidID":"raid-1"}}}`))

		err := client.JoinRaid(newStreamer(), "raid-1")

		if err != nil {
			t.Fatalf("JoinRaid error = %v, want none", err)
		}
		if want := []string{`{"input":{"raidID":"raid-1"}}`}; !equal(fake.twitch.sent("JoinRaid"), want) {
			t.Errorf("raids joined = %v, want %v", fake.twitch.sent("JoinRaid"), want)
		}
	})

	t.Run("no raid id", func(t *testing.T) {
		client, fake := setup(t)

		err := client.JoinRaid(newStreamer(), "")

		if want := "missing raid id"; errorText(err) != want {
			t.Errorf("JoinRaid error = %v, want %s", err, want)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", down("connection refused"))

		err := client.JoinRaid(newStreamer(), "raid-1")

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("JoinRaid error = %v, want %s", err, want)
		}
	})
}
