package streak

import (
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

var (
	now       = time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	startedAt = time.Date(2026, time.April, 10, 11, 30, 0, 0, time.UTC)
)

func TestActiveWatchMatches(t *testing.T) {
	for _, row := range []struct {
		name     string
		streamer *domain.Streamer
		watch    ActiveWatch
		want     bool
	}{
		{
			"the same broadcast",
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1"}},
			ActiveWatch{BroadcastID: "broadcast-1"}, true,
		},
		{
			"another broadcast",
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-2"}},
			ActiveWatch{BroadcastID: "broadcast-1"}, false,
		},
		{
			"the same broadcast with another start time",
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt.Add(time.Minute)}},
			ActiveWatch{BroadcastID: "broadcast-1", CreatedAt: startedAt}, true,
		},
		{
			"another broadcast with the same start time",
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-2", CreatedAt: startedAt}},
			ActiveWatch{BroadcastID: "broadcast-1", CreatedAt: startedAt}, false,
		},
		{
			"only the stream has a broadcast id",
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt}},
			ActiveWatch{CreatedAt: startedAt}, false,
		},
		{
			"only the watch has a broadcast id",
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt}},
			ActiveWatch{BroadcastID: "broadcast-1", CreatedAt: startedAt}, false,
		},
		{
			"no broadcast ids, the same start time",
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt}},
			ActiveWatch{CreatedAt: startedAt}, true,
		},
		{
			"no broadcast ids, the same start time in another zone",
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt.In(time.FixedZone("CEST", 2*60*60))}},
			ActiveWatch{CreatedAt: startedAt}, true,
		},
		{
			"no broadcast ids, another start time",
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt.Add(time.Second)}},
			ActiveWatch{CreatedAt: startedAt}, false,
		},
		{
			"no broadcast ids, the stream has no start time",
			&domain.Streamer{Stream: &domain.Stream{}},
			ActiveWatch{CreatedAt: startedAt}, false,
		},
		{
			"no broadcast ids, the watch has no start time",
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt}},
			ActiveWatch{}, false,
		},
		{"nothing to compare", &domain.Streamer{Stream: &domain.Stream{}}, ActiveWatch{}, false},
		{"no stream", &domain.Streamer{}, ActiveWatch{BroadcastID: "broadcast-1"}, false},
		{"no streamer", nil, ActiveWatch{BroadcastID: "broadcast-1"}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := ActiveWatchMatches(row.streamer, row.watch); got != row.want {
				t.Errorf("ActiveWatchMatches() = %t, want %t", got, row.want)
			}
		})
	}
}

func TestStreamStartSurvivesOffline(t *testing.T) {
	offlineAt := now

	for _, row := range []struct {
		name  string
		start time.Time
		want  bool
	}{
		{"started after going offline", offlineAt.Add(time.Minute), true},
		{"started as it went offline", offlineAt, true},
		{"started one second before going offline", offlineAt.Add(-time.Second), true},
		{"started one second under two minutes before", offlineAt.Add(-119 * time.Second), true},
		{"started exactly two minutes before", offlineAt.Add(-120 * time.Second), true},
		{"started one second over two minutes before", offlineAt.Add(-121 * time.Second), false},
		{"started an hour before", offlineAt.Add(-time.Hour), false},
		{"no start time", time.Time{}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := StreamStartSurvivesOffline(row.start, offlineAt); got != row.want {
				t.Errorf("StreamStartSurvivesOffline(%s, %s) = %t, want %t", row.start, offlineAt, got, row.want)
			}
		})
	}

	t.Run("never went offline", func(t *testing.T) {
		if got := StreamStartSurvivesOffline(startedAt, time.Time{}); got != false {
			t.Errorf("StreamStartSurvivesOffline(%s, zero) = %t, want false", startedAt, got)
		}
	})
}
