package twitch

import (
	"encoding/json"
	"testing"
	"time"
)

// rewardList decodes a RewardList answer as the client does.
func rewardList(t *testing.T, answer string) *gqlRewardListResponse {
	t.Helper()
	if answer == "" {
		return nil
	}
	var resp gqlRewardListResponse
	if err := json.Unmarshal([]byte(answer), &resp); err != nil {
		t.Fatalf("reward list %s: %v", answer, err)
	}
	return &resp
}

func TestCurrentStreamHasCompletedWatchStreak(t *testing.T) {
	start := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		createdAt     time.Time
		achievementAt time.Time
		want          bool
	}{
		{"reached six minutes into the stream", start, start.Add(6 * time.Minute), true},
		{"reached a nanosecond into the stream", start, start.Add(time.Nanosecond), true},
		{"reached as the stream started", start, start, false},
		{"reached a nanosecond before the stream", start, start.Add(-time.Nanosecond), false},
		{"reached the day before", start, start.Add(-24 * time.Hour), false},
		{"no stream start", time.Time{}, start, false},
		{"never reached", start, time.Time{}, false},
		{"neither known", time.Time{}, time.Time{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := currentStreamHasCompletedWatchStreak(tt.createdAt, tt.achievementAt); got != tt.want {
				t.Errorf("currentStreamHasCompletedWatchStreak(%s, %s) = %t, want %t", tt.createdAt, tt.achievementAt, got, tt.want)
			}
		})
	}
}

func TestExtractWatchStreakAchievementAt(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   time.Time
	}{
		{
			"a streak reached",
			`{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3","achievementTimestamp":"2026-03-01T10:06:00Z"}}}}}}`,
			time.Date(2026, time.March, 1, 10, 6, 0, 0, time.UTC),
		},
		{"a streak never reached", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3","achievementTimestamp":null}}}}}}`, time.Time{}},
		{"a time that is not a time", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"achievementTimestamp":"yesterday"}}}}}}`, time.Time{}},
		{"an expiring streak with no milestone", `{"data":{"channel":{"self":{"watchStreakMilestone":{"expiresAt":"2026-03-02T00:00:00Z"}}}}}`, time.Time{}},
		{"no streak", `{"data":{"channel":{"self":{"watchStreakMilestone":null}}}}`, time.Time{}},
		{"no viewer", `{"data":{"channel":{"self":null}}}`, time.Time{}},
		{"no channel", `{"data":{"channel":null}}`, time.Time{}},
		{"no answer", ``, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractWatchStreakAchievementAt(rewardList(t, tt.answer)); !got.Equal(tt.want) {
				t.Errorf("extractWatchStreakAchievementAt(%s) = %s, want %s", tt.answer, got, tt.want)
			}
		})
	}
}

func TestExtractWatchStreakExpiresAt(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   time.Time
	}{
		{
			"an expiring streak",
			`{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"},"expiresAt":"2026-03-02T00:00:00Z"}}}}}`,
			time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			"an expiring streak with no milestone",
			`{"data":{"channel":{"self":{"watchStreakMilestone":{"expiresAt":"2026-03-02T00:00:00Z"}}}}}`,
			time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC),
		},
		{"a streak that is not expiring", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"},"expiresAt":null}}}}}`, time.Time{}},
		{"a time that is not a time", `{"data":{"channel":{"self":{"watchStreakMilestone":{"expiresAt":"tomorrow"}}}}}`, time.Time{}},
		{"no streak", `{"data":{"channel":{"self":{"watchStreakMilestone":null}}}}`, time.Time{}},
		{"no viewer", `{"data":{"channel":{"self":null}}}`, time.Time{}},
		{"no channel", `{"data":{"channel":null}}`, time.Time{}},
		{"no answer", ``, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractWatchStreakExpiresAt(rewardList(t, tt.answer)); !got.Equal(tt.want) {
				t.Errorf("extractWatchStreakExpiresAt(%s) = %s, want %s", tt.answer, got, tt.want)
			}
		})
	}
}

func TestExtractWatchStreakLength(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   int
	}{
		{"a streak of three", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3"}}}}}}`, 3},
		{"a streak of zero", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"0"}}}}}}`, 0},
		{"a length that is not a number", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"three"}}}}}}`, -1},
		{"an empty length", `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":""}}}}}}`, -1},
		{"an expiring streak with no milestone", `{"data":{"channel":{"self":{"watchStreakMilestone":{"expiresAt":"2026-03-02T00:00:00Z"}}}}}`, -1},
		{"no streak", `{"data":{"channel":{"self":{"watchStreakMilestone":null}}}}`, -1},
		{"no viewer", `{"data":{"channel":{"self":null}}}`, -1},
		{"no channel", `{"data":{"channel":null}}`, -1},
		{"no answer", ``, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractWatchStreakLength(rewardList(t, tt.answer)); got != tt.want {
				t.Errorf("extractWatchStreakLength(%s) = %d, want %d", tt.answer, got, tt.want)
			}
		})
	}
}

// The times in the text are in the machine's own zone, as the client writes them.
func TestClient_LogExtraInfo(t *testing.T) {
	const expiringStreakOfThree = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3","achievementTimestamp":"2026-11-26T10:06:00Z"},"expiresAt":"2026-11-27T00:00:00Z"}}}}}`
	expiresAt := time.Date(2026, time.November, 27, 0, 0, 0, 0, time.UTC).Local().String()
	startedAt := tenOClock.Local().String()
	reachedAt := tenOClock.Add(6 * time.Minute).Local().String()

	t.Run("an offline streamer with a streak", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))

		got := client.LogExtraInfo("123456", false)

		if want := " | streak length 3"; got != want {
			t.Errorf("LogExtraInfo = %q, want %q", got, want)
		}
	})

	t.Run("an offline streamer with an expiring streak", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("RewardList", rewardListVariables, ok(expiringStreakOfThree))

		got := client.LogExtraInfo("123456", false)

		if want := " | streak length 3 | expires at " + expiresAt; got != want {
			t.Errorf("LogExtraInfo = %q, want %q", got, want)
		}
	})

	t.Run("an online streamer with an expiring streak", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("RewardList", rewardListVariables, ok(expiringStreakOfThree))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(liveSinceTen))

		got := client.LogExtraInfo("123456", true)

		want := " | streak length 3 | expires at " + expiresAt + " | started at " + startedAt + " | streak achievement timestamp " + reachedAt
		if got != want {
			t.Errorf("LogExtraInfo = %q, want %q", got, want)
		}
	})

	t.Run("an online streamer whose stream start is unknown", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("RewardList", rewardListVariables, ok(streakAtSixPastTen))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(`{"data":{"user":{"stream":null}}}`))

		got := client.LogExtraInfo("123456", true)

		if want := " | streak length 3 | streak achievement timestamp " + reachedAt; got != want {
			t.Errorf("LogExtraInfo = %q, want %q", got, want)
		}
	})

	t.Run("an online streamer with no streak", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("RewardList", rewardListVariables, ok(`{"data":{"channel":{"self":{"watchStreakMilestone":null}}}}`))
		fake.twitch.answer("WithIsStreamLiveQuery", liveVariables, ok(liveSinceTen))

		got := client.LogExtraInfo("123456", true)

		if want := " | streak length -1 | started at " + startedAt; got != want {
			t.Errorf("LogExtraInfo = %q, want %q", got, want)
		}
	})

	t.Run("no channel id", func(t *testing.T) {
		client, fake := setup(t)

		got := client.LogExtraInfo("", true)

		if got != "" {
			t.Errorf("LogExtraInfo = %q, want nothing", got)
		}
		if len(fake.twitch.Requests) != 0 {
			t.Errorf("twitch received %v, want nothing", fake.twitch.Requests)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.logger.Debug = true
		fake.twitch.answer("RewardList", rewardListVariables, down("connection refused"))

		got := client.LogExtraInfo("123456", true)

		if got != "" {
			t.Errorf("LogExtraInfo = %q, want nothing", got)
		}
		wantLines := []string{
			"debug: Client version updated to " + buildID,
			`debug: GQL request failed: Post "https://gql.twitch.tv/gql": connection refused`,
			`debug: RewardList lookup failed for channel 123456: Post "https://gql.twitch.tv/gql": connection refused`,
		}
		if !equal(fake.logger.Lines, wantLines) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, wantLines)
		}
	})
}
