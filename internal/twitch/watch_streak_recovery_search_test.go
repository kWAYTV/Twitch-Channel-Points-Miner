package twitch

import "testing"

func TestClient_RecoverStreak_Search(t *testing.T) {
	t.Run("a clip of the missed stream from the last day", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("missed-1", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
	})

	t.Run("a clip of another stream", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("another-stream", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true: the last check shows the streak as it was")
		}
		if len(fake.twitch.Tracked) != 0 {
			t.Errorf("events tracked = %v, want none", fake.twitch.Tracked)
		}
	})

	t.Run("a clip of the missed stream on the second page", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("another-stream", "c1", true)))
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", "c1"), ok(clipPage("missed-1", "c2", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		if len(fake.twitch.Tracked) != 2 {
			t.Errorf("events tracked = %v, want the clip's two", fake.twitch.Tracked)
		}
	})

	t.Run("a clip of the missed stream from the last week", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_WEEK", ""), ok(clipPage("missed-1", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		if len(fake.twitch.Tracked) != 2 {
			t.Errorf("events tracked = %v, want the clip's two", fake.twitch.Tracked)
		}
	})

	t.Run("a vod of the missed stream", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("FilterableVideoTower_Videos", videosVariables(""), ok(videoPage("missed-1")))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		if len(fake.twitch.Tracked) != 1 {
			t.Errorf("events tracked = %v, want the vod's first minute", fake.twitch.Tracked)
		}
	})

	t.Run("an empty page that claims a next page", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(`{"data":{"user":{"clips":{"edges":[],"pageInfo":{"hasNextPage":true}}}}}`))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		_, err := client.RecoverStreak(newStreamer())

		if want := "no eligible clips or vods found to save streak expiring at " + streakExpiry; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if want := []string{"emoji :ambulance: streak recovery Streamer: empty clips edges but hasNextPage == true with cursor "}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("clips down", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), down("connection refused"))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		saved, err := client.RecoverStreak(newStreamer())

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("an answer with no clips", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(`{"data":{"user":null}}`))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		_, err := client.RecoverStreak(newStreamer())

		if want := "missing clips data for channel 123456"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})

	t.Run("an answer with no videos", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("FilterableVideoTower_Videos", videosVariables(""), ok(`{"data":{"user":null}}`))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		_, err := client.RecoverStreak(newStreamer())

		if want := "missing videos data for channel 123456"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})
}
