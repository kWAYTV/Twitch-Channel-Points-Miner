package twitch

import (
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const (
	overlayVariables    = `{"channel":"streamer"}`
	liveVariables       = `{"id":"123456"}`
	rewardListVariables = `{"channelID":"123456","shouldIncludeAllSuspendedStreaks":false}`

	overlayLive        = `{"data":{"user":{"stream":{"id":"broadcast-1","createdAt":"2026-11-26T10:00:00Z","viewersCount":42,"tags":[{"id":"tag-1"}]},"broadcastSettings":{"title":" Speedrun night ","game":{"id":"game-1","name":"Tanks","displayName":"Tanks!"}}}}}`
	overlayNoStartTime = `{"data":{"user":{"stream":{"id":"broadcast-1","viewersCount":42,"tags":[{"id":"tag-1"}]},"broadcastSettings":{"title":" Speedrun night ","game":{"id":"game-1","name":"Tanks","displayName":"Tanks!"}}}}}`
	overlayNoGame      = `{"data":{"user":{"stream":{"id":"broadcast-1","createdAt":"2026-11-26T10:00:00Z","viewersCount":42,"tags":[]},"broadcastSettings":{"title":"Speedrun night","game":null}}}}`
	overlayOffline     = `{"data":{"user":{"stream":null,"broadcastSettings":{"title":"Speedrun night","game":null}}}}`
	overlayNoUser      = `{"data":{"user":null}}`
	liveSinceTen       = `{"data":{"user":{"stream":{"createdAt":"2026-11-26T10:00:00Z"}}}}`
	streakAtSixPastTen = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3","achievementTimestamp":"2026-11-26T10:06:00Z"}}}}}}`
	streakAtNine       = `{"data":{"channel":{"self":{"watchStreakMilestone":{"watchStreakMilestone":{"value":"3","achievementTimestamp":"2026-11-26T09:00:00Z"}}}}}}`
)

var tenOClock = time.Date(2026, time.November, 26, 10, 0, 0, 0, time.UTC)

// streamFacts is what a stream update writes onto a streamer's stream.
type streamFacts struct {
	BroadcastID string
	Title       string
	Game        map[string]interface{}
	Tags        []map[string]interface{}
	Viewers     int
	CreatedAt   time.Time
	DropsTags   bool
}

func factsOf(stream *domain.Stream) streamFacts {
	return streamFacts{
		BroadcastID: stream.BroadcastID,
		Title:       stream.Title,
		Game:        stream.Game,
		Tags:        stream.Tags,
		Viewers:     stream.ViewersCount,
		CreatedAt:   stream.CreatedAt,
		DropsTags:   stream.DropsTags,
	}
}

var liveFacts = streamFacts{
	BroadcastID: "broadcast-1",
	Title:       "Speedrun night",
	Game:        map[string]interface{}{"id": "game-1", "name": "Tanks", "displayName": "Tanks!"},
	Tags:        []map[string]interface{}{{"id": "tag-1"}},
	Viewers:     42,
	CreatedAt:   tenOClock,
}

func TestParseRFC3339Timestamp(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Time
	}{
		{"a time in UTC", "2026-03-01T10:00:00Z", time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)},
		{"a time with a fraction of a second", "2026-03-01T10:00:00.123456789Z", time.Date(2026, time.March, 1, 10, 0, 0, 123456789, time.UTC)},
		{"a time two hours ahead of UTC", "2026-03-01T10:00:00+02:00", time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)},
		{"a time with padding", " 2026-03-01T10:00:00Z\n", time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)},
		{"a time with no zone", "2026-03-01T10:00:00", time.Time{}},
		{"a date", "2026-03-01", time.Time{}},
		{"words", "not-a-time", time.Time{}},
		{"only spaces", "   ", time.Time{}},
		{"nothing", "", time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRFC3339Timestamp(tt.value); !got.Equal(tt.want) {
				t.Errorf("parseRFC3339Timestamp(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}
