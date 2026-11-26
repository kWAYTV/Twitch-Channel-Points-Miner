package twitch

import (
	"fmt"
	"testing"
	"time"
)

const (
	recoverySpade = "POST https://spade.twitch.tv/track"

	// A streak of three that expires unless a clip or vod of stream missed-1 is watched.
	streakExpiring = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"},"missedStreams":[{"broadcastIdentifiers":[{"id":"missed-1"}]}],"expiresAt":"2026-11-27T00:00:00Z"}}}}}`
	streakSaved    = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"},"missedStreams":[]}}}}}`
	streakLost     = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"0"},"missedStreams":[]}}}}}`

	noClips  = `{"data":{"user":{"clips":{"edges":[],"pageInfo":{"hasNextPage":false}}}}}`
	noVideos = `{"data":{"user":{"videos":{"edges":[],"pageInfo":{"hasNextPage":false}}}}}`
)

// The client writes these in the machine's own zone.
var (
	streakExpiry = time.Date(2026, time.November, 27, 0, 0, 0, 0, time.UTC).Local().String()
	clipMadeAt   = time.Date(2026, time.November, 25, 10, 0, 0, 0, time.UTC).Local().String()
)

func clipsVariables(period, cursor string) string {
	return fmt.Sprintf(`{"criteria":{"filter":%q},"cursor":%q,"limit":20,"login":"streamer"}`, period, cursor)
}

func videosVariables(cursor string) string {
	return fmt.Sprintf(`{"channelOwnerLogin":"streamer","cursor":%q,"limit":20,"videoSort":"TIME"}`, cursor)
}

// clipPage is a page of one clip, cut from the stream with the given broadcast id.
func clipPage(broadcastID, cursor string, hasNext bool) string {
	return fmt.Sprintf(`{"data":{"user":{"clips":{"edges":[{"cursor":%q,"node":{"id":"clip-id","slug":"clip-slug","url":"https://clips.example/clip-slug","createdAt":"2026-11-25T10:00:00Z","broadcastIdentifier":{"id":%q}}}],"pageInfo":{"hasNextPage":%t}}}}}`, cursor, broadcastID, hasNext)
}

// videoPage is a page of one vod of the stream with the given broadcast id.
func videoPage(broadcastID string) string {
	return fmt.Sprintf(`{"data":{"user":{"videos":{"edges":[{"cursor":"v1","node":{"id":"vod-id","publishedAt":"2026-11-25T10:00:00Z","broadcastIdentifier":{"id":%q}}}],"pageInfo":{"hasNextPage":false}}}}}`, broadcastID)
}

// serveNothingToWatch makes the fake answer every listing with an empty page, and
// accept whatever is reported as watched.
func serveNothingToWatch(twitch *fakeTwitch) {
	twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(noClips))
	twitch.answer("ClipsCards__User", clipsVariables("LAST_WEEK", ""), ok(noClips))
	twitch.answer("FilterableVideoTower_Videos", videosVariables(""), ok(noVideos))
	twitch.serve(recoverySpade, code(204, ""))
}

func TestClient_RecoverStreak(t *testing.T) {
	t.Run("a streak that is not expiring", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("missed-1", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false: nothing needed saving")
		}
		if len(fake.twitch.Tracked) != 0 {
			t.Errorf("events tracked = %v, want none", fake.twitch.Tracked)
		}
	})

	t.Run("nothing of the missed stream to watch", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		saved, err := client.RecoverStreak(newStreamer())

		if want := "no eligible clips or vods found to save streak expiring at " + streakExpiry; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
		if len(fake.twitch.Tracked) != 0 {
			t.Errorf("events tracked = %v, want none", fake.twitch.Tracked)
		}
		if len(fake.clock.Slept) != 0 {
			t.Errorf("waits = %v, want none", fake.clock.Slept)
		}
	})

	t.Run("a clip watched, the streak saved only at the last check", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("missed-1", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		if want := []time.Duration{5 * time.Second, 10 * time.Second}; !equal(fake.clock.Slept, want) {
			t.Errorf("waits = %v, want %v", fake.clock.Slept, want)
		}
	})

	t.Run("a clip watched, the streak lost at the last check", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("missed-1", "c1", false)))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakExpiring), ok(streakLost))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false: the streak is back to zero")
		}
	})

	t.Run("a streamer with no channel id", func(t *testing.T) {
		client, fake := setup(t)
		streamer := newStreamer()
		streamer.ChannelID = ""

		saved, err := client.RecoverStreak(streamer)

		if want := "missing streamer channel id"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		client, _ := setup(t)

		_, err := client.RecoverStreak(nil)

		if want := "missing streamer channel id"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})

	t.Run("reward list down", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, down("connection refused"))

		saved, err := client.RecoverStreak(newStreamer())

		if want := `RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("reward list down at the last check", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), down("connection refused"))

		_, err := client.RecoverStreak(newStreamer())

		if want := `RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})

	t.Run("a reward list with no viewer", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(`{"data":{"channel":{"self":null}}}`))

		_, err := client.RecoverStreak(newStreamer())

		if want := "RewardList response does not have data.channel.self for channel 123456"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})

	t.Run("an expiring streak with no list of missed streams", func(t *testing.T) {
		client, fake := setup(t)
		serveNothingToWatch(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(`{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"},"expiresAt":"2026-11-27T00:00:00Z"}}}}}`))

		_, err := client.RecoverStreak(newStreamer())

		if want := "RewardList response does not have missedStreams for channel 123456"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
	})
}

func TestExpiryRefreshWait(t *testing.T) {
	tests := []struct {
		name  string
		clips int
		vods  int
		want  time.Duration
	}{
		{"one clip", 1, 0, 10 * time.Second},
		{"two clips", 2, 0, 5 * time.Second},
		{"nothing watched", 0, 0, 0},
		{"one clip and a vod", 1, 1, 0},
		{"one vod", 0, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expiryRefreshWait(tt.clips, tt.vods); got != tt.want {
				t.Errorf("expiryRefreshWait(%d, %d) = %s, want %s", tt.clips, tt.vods, got, tt.want)
			}
		})
	}
}
