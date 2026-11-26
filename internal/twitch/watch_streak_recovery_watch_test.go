package twitch

import (
	"regexp"
	"testing"
	"time"
)

const vodMinuteEvents = `[{"event":"minute-watched","properties":{"broadcast_id":null,"channel":"streamer","channel_id":"123456","content_mode":"vod","live":false,"player":"site","user_id":"user-id","vod_id":"vod-id"}}]`

var (
	clipPlayEvents = regexp.MustCompile(`^\[\{"event":"video-play","properties":\{"channel_id":"123456","clip_slug":"clip-slug","content_mode":"clip","live":false,"location":"vod","minutes_logged":0,"play_session_id":"[0-9A-Za-z]{32}","player":"site","url":"https://clips\.example/clip-slug","user_id":"user-id","vod_id":"clip-id","vod_timestamp":0,"vod_type":"clip"\}\}\]$`)
	clipFiveEvents = regexp.MustCompile(`^\[\{"event":"n_second_play","properties":\{"channel_id":"123456","clip_slug":"clip-slug","content_mode":"clip","live":false,"location":"vod","minutes_logged":0,"platform":"web","play_session_id":"[0-9A-Za-z]{32}","player":"site","seconds_after_play":5,"url":"https://clips\.example/clip-slug","user_id":"user-id","vod_id":"clip-id","vod_timestamp":4\.9,"vod_type":"clip"\}\}\]$`)
)

// trackedOnSpade reports whether got is one post to spade.twitch.tv per pattern,
// each with events that match its pattern.
func trackedOnSpade(got []spadePost, patterns ...*regexp.Regexp) bool {
	var events []string
	for _, post := range got {
		if post.URL != "https://spade.twitch.tv/track" {
			return false
		}
		events = append(events, post.Events)
	}
	return matchAll(events, patterns...)
}

// serveMissedClip makes the fake list one clip of the missed stream in the last day.
func serveMissedClip(twitch *fakeTwitch) {
	serveNothingToWatch(twitch)
	twitch.answer("ClipsCards__User", clipsVariables("LAST_DAY", ""), ok(clipPage("missed-1", "c1", false)))
}

// serveMissedVOD makes the fake list one vod of the missed stream, and no clips.
func serveMissedVOD(twitch *fakeTwitch) {
	serveNothingToWatch(twitch)
	twitch.answer("FilterableVideoTower_Videos", videosVariables(""), ok(videoPage("missed-1")))
}

func TestClient_RecoverStreak_Watch(t *testing.T) {
	t.Run("a clip that saves the streak", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedClip(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		if !trackedOnSpade(fake.twitch.Tracked, clipPlayEvents, clipFiveEvents) {
			t.Errorf("events tracked = %v, want a post matching %v, then one matching %v", fake.twitch.Tracked, clipPlayEvents, clipFiveEvents)
		}
		if want := []time.Duration{5 * time.Second}; !equal(fake.clock.Slept, want) {
			t.Errorf("waits = %v, want %v", fake.clock.Slept, want)
		}
		if want := []string{"emoji :ambulance: streak recovery Streamer: watching clip clip-slug from " + clipMadeAt}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a clip that saves the streak, anonymized", func(t *testing.T) {
		client, fake := setupAnonymized(t)
		serveMissedClip(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		_, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if want := []string{"emoji :ambulance: streak recovery Streamer1: watching clip clip-slug from " + clipMadeAt}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a clip after which the streak is back to zero", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedClip(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakLost))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("a clip, spade refusing the play", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedClip(fake.twitch)
		fake.twitch.serve(recoverySpade, code(500, "nope"))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if want := "video-play post failed for channel 123456: minute watched failed: 500 nope"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
		if len(fake.clock.Slept) != 0 {
			t.Errorf("waits = %v, want none", fake.clock.Slept)
		}
	})

	t.Run("a clip, spade refusing the five-second mark", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedClip(fake.twitch)
		fake.twitch.serve(recoverySpade, code(204, ""), code(500, "nope"))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if want := "n_second_play post failed for channel 123456: minute watched failed: 500 nope"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("a clip, reward list down afterwards", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedClip(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), down("connection refused"))

		saved, err := client.RecoverStreak(newStreamer())

		if want := `RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("a vod that saves the streak in its second minute", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedVOD(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if err != nil {
			t.Fatalf("RecoverStreak error = %v, want none", err)
		}
		if !saved {
			t.Errorf("RecoverStreak = false, want true")
		}
		wantTracked := []spadePost{
			{URL: "https://spade.twitch.tv/track", Events: vodMinuteEvents},
			{URL: "https://spade.twitch.tv/track", Events: vodMinuteEvents},
		}
		if !equal(fake.twitch.Tracked, wantTracked) {
			t.Errorf("events tracked = %v, want %v", fake.twitch.Tracked, wantTracked)
		}
		if want := []time.Duration{5 * time.Second, 55 * time.Second, 5 * time.Second}; !equal(fake.clock.Slept, want) {
			t.Errorf("waits = %v, want %v", fake.clock.Slept, want)
		}
		wantLines := []string{
			"emoji :ambulance: streak recovery Streamer: watching vod minute 0 from " + clipMadeAt,
			"emoji :ambulance: streak recovery Streamer: watching vod minute 1 from " + clipMadeAt,
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})

	t.Run("a vod that never saves the streak", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedVOD(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring))

		saved, err := client.RecoverStreak(newStreamer())

		if want := "no eligible clips or vods found to save streak expiring at " + streakExpiry; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
		if len(fake.twitch.Tracked) != 6 {
			t.Errorf("%d minutes tracked, want 6", len(fake.twitch.Tracked))
		}
		wantSlept := []time.Duration{
			5 * time.Second, 55 * time.Second, 5 * time.Second, 55 * time.Second, 5 * time.Second, 55 * time.Second,
			5 * time.Second, 55 * time.Second, 5 * time.Second, 55 * time.Second, 5 * time.Second, 55 * time.Second,
		}
		if !equal(fake.clock.Slept, wantSlept) {
			t.Errorf("waits = %v, want %v", fake.clock.Slept, wantSlept)
		}
	})

	t.Run("a vod, spade refusing the minute", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedVOD(fake.twitch)
		fake.twitch.serve(recoverySpade, code(500, "nope"))
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), ok(streakSaved))

		saved, err := client.RecoverStreak(newStreamer())

		if want := "minute-watched post failed for channel 123456: minute watched failed: 500 nope"; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
	})

	t.Run("a vod, reward list down after the first minute", func(t *testing.T) {
		client, fake := setup(t)
		serveMissedVOD(fake.twitch)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakExpiring), down("connection refused"))

		saved, err := client.RecoverStreak(newStreamer())

		if want := `RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("RecoverStreak error = %v, want %s", err, want)
		}
		if saved {
			t.Errorf("RecoverStreak = true, want false")
		}
		if want := []time.Duration{5 * time.Second}; !equal(fake.clock.Slept, want) {
			t.Errorf("waits = %v, want %v", fake.clock.Slept, want)
		}
	})
}
