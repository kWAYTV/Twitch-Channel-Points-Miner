package domain

import (
	"reflect"
	"testing"
	"time"
)

var (
	now       = time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	startedAt = time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
)

func TestNewStream(t *testing.T) {
	t.Run("a new stream", func(t *testing.T) {
		want := &Stream{
			Game:               map[string]interface{}{},
			Tags:               []map[string]interface{}{},
			Campaigns:          []interface{}{},
			CampaignIDs:        []string{},
			WatchStreakMissing: true,
		}
		if got := NewStream(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewStream() = %+v, want %+v", got, want)
		}
	})
}

func TestStream_updateAt(t *testing.T) {
	t.Run("a first update", func(t *testing.T) {
		stream := &Stream{WatchStreakMissing: true}

		stream.updateAt("broadcast-1", "  Speedrun\n", map[string]interface{}{"displayName": "Game"}, []map[string]interface{}{{"id": "english"}}, 100, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:        "broadcast-1",
			Title:              "Speedrun",
			Game:               map[string]interface{}{"displayName": "Game"},
			Tags:               []map[string]interface{}{{"id": "english"}},
			ViewersCount:       100,
			WatchStreakMissing: true,
			CreatedAt:          startedAt,
			lastUpdate:         now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("the same broadcast again", func(t *testing.T) {
		stream := &Stream{
			BroadcastID:   "broadcast-1",
			Title:         "Speedrun",
			ViewersCount:  100,
			PlaylistURL:   "https://example.test/playlist.m3u8",
			MinuteWatched: 3,
			WatchCount:    2,
			CreatedAt:     startedAt,
		}

		stream.updateAt("broadcast-1", "Speedrun, part two", nil, nil, 50, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:   "broadcast-1",
			Title:         "Speedrun, part two",
			ViewersCount:  50,
			PlaylistURL:   "https://example.test/playlist.m3u8",
			MinuteWatched: 3,
			WatchCount:    2,
			CreatedAt:     startedAt,
			lastUpdate:    now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("the same broadcast with no start time", func(t *testing.T) {
		stream := &Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt}

		stream.updateAt("broadcast-1", "Speedrun", nil, nil, 50, time.Time{}, "drops", now)

		want := &Stream{BroadcastID: "broadcast-1", Title: "Speedrun", ViewersCount: 50, CreatedAt: startedAt, lastUpdate: now}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("a new broadcast", func(t *testing.T) {
		stream := &Stream{
			BroadcastID:   "broadcast-1",
			Title:         "Speedrun",
			PlaylistURL:   "https://example.test/playlist.m3u8",
			MinuteWatched: 3,
			WatchCount:    2,
			CreatedAt:     startedAt,
		}
		stream.updateMinuteWatchedAt(now.Add(-time.Minute))

		stream.updateAt("broadcast-2", "Speedrun", nil, nil, 50, startedAt.Add(time.Hour), "drops", now)

		want := &Stream{
			BroadcastID:        "broadcast-2",
			Title:              "Speedrun",
			ViewersCount:       50,
			WatchStreakMissing: true,
			CreatedAt:          startedAt.Add(time.Hour),
			lastUpdate:         now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("a new broadcast with no start time", func(t *testing.T) {
		stream := &Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt}

		stream.updateAt("broadcast-2", "Speedrun", nil, nil, 50, time.Time{}, "drops", now)

		want := &Stream{BroadcastID: "broadcast-2", Title: "Speedrun", ViewersCount: 50, WatchStreakMissing: true, lastUpdate: now}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("the drops tag on a stream with a game", func(t *testing.T) {
		stream := &Stream{}

		stream.updateAt("broadcast-1", "Speedrun", map[string]interface{}{"displayName": "Game"}, []map[string]interface{}{{"id": "english"}, {"id": "drops"}}, 100, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:  "broadcast-1",
			Title:        "Speedrun",
			Game:         map[string]interface{}{"displayName": "Game"},
			Tags:         []map[string]interface{}{{"id": "english"}, {"id": "drops"}},
			DropsTags:    true,
			ViewersCount: 100,
			CreatedAt:    startedAt,
			lastUpdate:   now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("the drops tag on a stream with no game", func(t *testing.T) {
		stream := &Stream{}

		stream.updateAt("broadcast-1", "Speedrun", map[string]interface{}{}, []map[string]interface{}{{"id": "drops"}}, 100, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:  "broadcast-1",
			Title:        "Speedrun",
			Game:         map[string]interface{}{},
			Tags:         []map[string]interface{}{{"id": "drops"}},
			ViewersCount: 100,
			CreatedAt:    startedAt,
			lastUpdate:   now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("a tag whose id is not a string", func(t *testing.T) {
		stream := &Stream{}

		stream.updateAt("broadcast-1", "Speedrun", map[string]interface{}{"displayName": "Game"}, []map[string]interface{}{{"id": 7}}, 100, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:  "broadcast-1",
			Title:        "Speedrun",
			Game:         map[string]interface{}{"displayName": "Game"},
			Tags:         []map[string]interface{}{{"id": 7}},
			ViewersCount: 100,
			CreatedAt:    startedAt,
			lastUpdate:   now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("the drops tag has gone", func(t *testing.T) {
		stream := &Stream{BroadcastID: "broadcast-1", DropsTags: true}

		stream.updateAt("broadcast-1", "Speedrun", map[string]interface{}{"displayName": "Game"}, nil, 100, startedAt, "drops", now)

		want := &Stream{
			BroadcastID:  "broadcast-1",
			Title:        "Speedrun",
			Game:         map[string]interface{}{"displayName": "Game"},
			ViewersCount: 100,
			CreatedAt:    startedAt,
			lastUpdate:   now,
		}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})
}

func TestStream_Update(t *testing.T) {
	t.Run("an updated stream needs no update", func(t *testing.T) {
		stream := NewStream()

		stream.Update("broadcast-1", "Speedrun", nil, nil, 100, startedAt, "drops")

		if got := stream.UpdateRequired(); got != false {
			t.Errorf("UpdateRequired() = %t, want false", got)
		}
	})
}

func TestStream_UpdateRequired(t *testing.T) {
	t.Run("a stream never updated", func(t *testing.T) {
		stream := NewStream()
		if got := stream.UpdateRequired(); got != true {
			t.Errorf("UpdateRequired() = %t, want true", got)
		}
	})
}

func TestStream_updateRequiredAt(t *testing.T) {
	t.Run("a stream never updated", func(t *testing.T) {
		stream := &Stream{}
		if got := stream.updateRequiredAt(now); got != true {
			t.Errorf("updateRequiredAt() = %t, want true", got)
		}
	})

	for _, row := range []struct {
		name  string
		after time.Duration
		want  bool
	}{
		{"the moment it was updated", 0, false},
		{"one second under two minutes later", 119 * time.Second, false},
		{"exactly two minutes later", 120 * time.Second, true},
		{"one second over two minutes later", 121 * time.Second, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			stream := &Stream{}
			stream.updateAt("broadcast-1", "Speedrun", nil, nil, 100, startedAt, "drops", now)

			if got := stream.updateRequiredAt(now.Add(row.after)); got != row.want {
				t.Errorf("updateRequiredAt(%s later) = %t, want %t", row.after, got, row.want)
			}
		})
	}
}

func TestStream_UpdateMinuteWatched(t *testing.T) {
	t.Run("the first call counts nothing", func(t *testing.T) {
		stream := NewStream()

		stream.UpdateMinuteWatched()

		if stream.MinuteWatched != 0 {
			t.Errorf("MinuteWatched = %v, want 0", stream.MinuteWatched)
		}
	})
}

func TestStream_updateMinuteWatchedAt(t *testing.T) {
	t.Run("the first call counts nothing", func(t *testing.T) {
		stream := &Stream{}

		stream.updateMinuteWatchedAt(now)

		want := &Stream{lastMinuteUpdate: now}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	for _, row := range []struct {
		name  string
		after time.Duration
		want  float64
	}{
		{"ninety seconds later", 90 * time.Second, 4.5},
		{"exactly two minutes later", 120 * time.Second, 5},
		{"one second over two minutes later: starts again", 121 * time.Second, 0},
		{"the same moment", 0, 3},
	} {
		t.Run(row.name, func(t *testing.T) {
			stream := &Stream{}
			stream.updateMinuteWatchedAt(now)
			stream.MinuteWatched = 3

			stream.updateMinuteWatchedAt(now.Add(row.after))

			want := &Stream{MinuteWatched: row.want, lastMinuteUpdate: now.Add(row.after)}
			if !reflect.DeepEqual(stream, want) {
				t.Errorf("stream = %+v, want %+v", stream, want)
			}
		})
	}
}

func TestStream_ResetWatchProgress(t *testing.T) {
	t.Run("a stream with progress", func(t *testing.T) {
		stream := &Stream{BroadcastID: "broadcast-1", Title: "Speedrun", MinuteWatched: 3, WatchCount: 2, CreatedAt: startedAt}
		stream.updateAt("broadcast-1", "Speedrun", nil, nil, 100, startedAt, "drops", now)
		stream.updateMinuteWatchedAt(now)

		stream.ResetWatchProgress()

		want := &Stream{BroadcastID: "broadcast-1", Title: "Speedrun", ViewersCount: 100, lastUpdate: now}
		if !reflect.DeepEqual(stream, want) {
			t.Errorf("stream = %+v, want %+v", stream, want)
		}
	})

	t.Run("no stream", func(t *testing.T) {
		var stream *Stream

		stream.ResetWatchProgress()

		if stream != nil {
			t.Errorf("stream = %+v, want nil", stream)
		}
	})
}

func TestStream_LastUpdateAgo(t *testing.T) {
	t.Run("a stream never updated", func(t *testing.T) {
		stream := NewStream()
		if got := stream.LastUpdateAgo(); got != 0 {
			t.Errorf("LastUpdateAgo() = %s, want 0", got)
		}
	})
}

func TestStream_lastUpdateAgoAt(t *testing.T) {
	t.Run("updated five minutes ago", func(t *testing.T) {
		stream := &Stream{}
		stream.updateAt("broadcast-1", "Speedrun", nil, nil, 100, startedAt, "drops", now.Add(-5*time.Minute))

		if got := stream.lastUpdateAgoAt(now); got != 5*time.Minute {
			t.Errorf("lastUpdateAgoAt() = %s, want %s", got, 5*time.Minute)
		}
	})

	t.Run("a stream never updated", func(t *testing.T) {
		stream := &Stream{}
		if got := stream.lastUpdateAgoAt(now); got != 0 {
			t.Errorf("lastUpdateAgoAt() = %s, want 0", got)
		}
	})

	t.Run("no stream", func(t *testing.T) {
		var stream *Stream
		if got := stream.lastUpdateAgoAt(now); got != 0 {
			t.Errorf("lastUpdateAgoAt() = %s, want 0", got)
		}
	})
}

func TestStream_StreamUpElapsed(t *testing.T) {
	t.Run("a stream that came up long ago", func(t *testing.T) {
		stream := &Stream{StreamUpAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}
		if got := stream.StreamUpElapsed(); got != true {
			t.Errorf("StreamUpElapsed() = %t, want true", got)
		}
	})

	t.Run("a stream that comes up far in the future", func(t *testing.T) {
		stream := &Stream{StreamUpAt: time.Date(2200, time.January, 1, 0, 0, 0, 0, time.UTC)}
		if got := stream.StreamUpElapsed(); got != false {
			t.Errorf("StreamUpElapsed() = %t, want false", got)
		}
	})
}

func TestStream_streamUpElapsedAt(t *testing.T) {
	for _, row := range []struct {
		name   string
		stream *Stream
		want   bool
	}{
		{"just came up", &Stream{StreamUpAt: now}, false},
		{"up one second under two minutes ago", &Stream{StreamUpAt: now.Add(-119 * time.Second)}, false},
		{"up exactly two minutes ago", &Stream{StreamUpAt: now.Add(-120 * time.Second)}, false},
		{"up one second over two minutes ago", &Stream{StreamUpAt: now.Add(-121 * time.Second)}, true},
		{"never came up", &Stream{}, true},
		{"no stream", nil, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := row.stream.streamUpElapsedAt(now); got != row.want {
				t.Errorf("streamUpElapsedAt() = %t, want %t", got, row.want)
			}
		})
	}
}

func TestStream_EncodePayload(t *testing.T) {
	t.Run("a payload of one event", func(t *testing.T) {
		stream := &Stream{Payload: []map[string]interface{}{{"channel_id": "123", "minutes_logged": 1.5}}}

		got, err := stream.EncodePayload()

		// [{"channel_id":"123","minutes_logged":1.5}]
		want := map[string]string{"data": "W3siY2hhbm5lbF9pZCI6IjEyMyIsIm1pbnV0ZXNfbG9nZ2VkIjoxLjV9XQ=="}
		if err != nil {
			t.Fatalf("EncodePayload() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("EncodePayload() = %v, want %v", got, want)
		}
	})

	t.Run("a payload that encodes to a plus sign", func(t *testing.T) {
		stream := &Stream{Payload: []map[string]interface{}{{"k": "~~~"}}}

		got, err := stream.EncodePayload()

		// [{"k":"~~~"}]
		want := map[string]string{"data": "W3siayI6In5+fiJ9XQ=="}
		if err != nil {
			t.Fatalf("EncodePayload() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("EncodePayload() = %v, want %v", got, want)
		}
	})

	t.Run("an empty payload", func(t *testing.T) {
		stream := &Stream{Payload: []map[string]interface{}{}}

		got, err := stream.EncodePayload()

		// []
		want := map[string]string{"data": "W10="}
		if err != nil {
			t.Fatalf("EncodePayload() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("EncodePayload() = %v, want %v", got, want)
		}
	})

	t.Run("no payload", func(t *testing.T) {
		stream := &Stream{}

		got, err := stream.EncodePayload()

		// null
		want := map[string]string{"data": "bnVsbA=="}
		if err != nil {
			t.Fatalf("EncodePayload() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("EncodePayload() = %v, want %v", got, want)
		}
	})

	t.Run("a payload JSON cannot hold", func(t *testing.T) {
		stream := &Stream{Payload: []map[string]interface{}{{"channel": make(chan int)}}}

		got, err := stream.EncodePayload()

		if got != nil {
			t.Errorf("EncodePayload() = %v, want nil", got)
		}
		if err == nil || err.Error() != "json: unsupported type: chan int" {
			t.Errorf("EncodePayload() error = %v, want %q", err, "json: unsupported type: chan int")
		}
	})
}

func TestStream_String(t *testing.T) {
	t.Run("a title and a game", func(t *testing.T) {
		stream := &Stream{Title: "Speedrun", Game: map[string]interface{}{"displayName": "Game"}}
		if got := stream.String(); got != "Speedrun (Game)" {
			t.Errorf("String() = %q, want %q", got, "Speedrun (Game)")
		}
	})

	t.Run("an empty stream", func(t *testing.T) {
		stream := &Stream{}
		if got := stream.String(); got != " ()" {
			t.Errorf("String() = %q, want %q", got, " ()")
		}
	})
}

func TestStream_GameName(t *testing.T) {
	for _, row := range []struct {
		name   string
		stream *Stream
		want   string
	}{
		{"a display name", &Stream{Game: map[string]interface{}{"displayName": "My Game", "name": "my-game"}}, "My Game"},
		{"only a name", &Stream{Game: map[string]interface{}{"name": "my-game"}}, "my-game"},
		{"a display name that is not a string", &Stream{Game: map[string]interface{}{"displayName": 7, "name": "my-game"}}, "my-game"},
		{"an empty display name", &Stream{Game: map[string]interface{}{"displayName": "", "name": "my-game"}}, ""},
		{"a display name with spaces around it", &Stream{Game: map[string]interface{}{"displayName": "  My Game "}}, "My Game"},
		{"a name with spaces around it", &Stream{Game: map[string]interface{}{"name": "  my-game "}}, "my-game"},
		{"a name outside ASCII", &Stream{Game: map[string]interface{}{"displayName": "ポケモン"}}, "ポケモン"},
		{"a game with no names", &Stream{Game: map[string]interface{}{"id": "1"}}, ""},
		{"an empty game", &Stream{Game: map[string]interface{}{}}, ""},
		{"no game", &Stream{}, ""},
		{"no stream", nil, ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := row.stream.GameName(); got != row.want {
				t.Errorf("GameName() = %q, want %q", got, row.want)
			}
		})
	}
}
