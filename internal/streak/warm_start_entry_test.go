package streak

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestEntryFromStreamer(t *testing.T) {
	t.Run("a live streamer whose streak is collected", func(t *testing.T) {
		streamer := &domain.Streamer{
			Username:  "streamer",
			ChannelID: "123",
			IsOnline:  true,
			Stream:    &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt},
		}
		want := WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "streamer",
			ChannelID:       "123",
			BroadcastID:     "broadcast-1",
			StreamCreatedAt: startedAt,
			CheckedAt:       now,
			IsOnline:        true,
		}
		if got := entryFromStreamer("account", streamer, now); got != want {
			t.Errorf("entryFromStreamer() = %+v, want %+v", got, want)
		}
	})

	t.Run("a live streamer whose streak is missing", func(t *testing.T) {
		streamer := &domain.Streamer{
			Username:  "streamer",
			ChannelID: "123",
			IsOnline:  true,
			Stream:    &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt, WatchStreakMissing: true},
		}
		want := WarmStartEntry{
			AccountName:        "account",
			StreamerLogin:      "streamer",
			ChannelID:          "123",
			BroadcastID:        "broadcast-1",
			StreamCreatedAt:    startedAt,
			WatchStreakMissing: true,
			CheckedAt:          now,
			IsOnline:           true,
		}
		if got := entryFromStreamer("account", streamer, now); got != want {
			t.Errorf("entryFromStreamer() = %+v, want %+v", got, want)
		}
	})

	t.Run("an offline streamer with no stream", func(t *testing.T) {
		streamer := &domain.Streamer{Username: "streamer"}
		want := WarmStartEntry{AccountName: "account", StreamerLogin: "streamer", WatchStreakMissing: true, CheckedAt: now}
		if got := entryFromStreamer("account", streamer, now); got != want {
			t.Errorf("entryFromStreamer() = %+v, want %+v", got, want)
		}
	})

	t.Run("names in capitals with spaces around them", func(t *testing.T) {
		streamer := &domain.Streamer{
			Username:  " Streamer ",
			ChannelID: " 123 ",
			IsOnline:  true,
			Stream:    &domain.Stream{BroadcastID: " broadcast-1 ", CreatedAt: startedAt},
		}
		want := WarmStartEntry{
			AccountName:     "account",
			StreamerLogin:   "streamer",
			ChannelID:       "123",
			BroadcastID:     "broadcast-1",
			StreamCreatedAt: startedAt,
			CheckedAt:       now,
			IsOnline:        true,
		}
		if got := entryFromStreamer(" Account ", streamer, now); got != want {
			t.Errorf("entryFromStreamer() = %+v, want %+v", got, want)
		}
	})
}

func TestEntriesForAccount(t *testing.T) {
	t.Run("entries of two accounts", func(t *testing.T) {
		stored := []WarmStartEntry{
			{AccountName: "account", StreamerLogin: "alpha", BroadcastID: "broadcast-1"},
			{AccountName: "someone-else", StreamerLogin: "beta", BroadcastID: "broadcast-2"},
		}
		want := map[string]WarmStartEntry{
			"alpha": {AccountName: "account", StreamerLogin: "alpha", BroadcastID: "broadcast-1"},
		}
		if got := entriesForAccount(stored, "account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("no account name keeps every account", func(t *testing.T) {
		stored := []WarmStartEntry{
			{AccountName: "account", StreamerLogin: "alpha"},
			{AccountName: "someone-else", StreamerLogin: "beta"},
		}
		want := map[string]WarmStartEntry{
			"alpha": {AccountName: "account", StreamerLogin: "alpha"},
			"beta":  {AccountName: "someone-else", StreamerLogin: "beta"},
		}
		if got := entriesForAccount(stored, ""); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("account names in capitals with spaces around them", func(t *testing.T) {
		stored := []WarmStartEntry{{AccountName: " ACCOUNT ", StreamerLogin: "alpha"}}
		want := map[string]WarmStartEntry{"alpha": {AccountName: " ACCOUNT ", StreamerLogin: "alpha"}}
		if got := entriesForAccount(stored, "  Account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("a login in capitals with spaces around it", func(t *testing.T) {
		stored := []WarmStartEntry{{AccountName: "account", StreamerLogin: " Alpha "}}
		want := map[string]WarmStartEntry{"alpha": {AccountName: "account", StreamerLogin: " Alpha "}}
		if got := entriesForAccount(stored, "account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("an entry with no login", func(t *testing.T) {
		stored := []WarmStartEntry{{AccountName: "account", StreamerLogin: "  ", BroadcastID: "broadcast-1"}}
		want := map[string]WarmStartEntry{}
		if got := entriesForAccount(stored, "account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("the same login twice: the later entry", func(t *testing.T) {
		stored := []WarmStartEntry{
			{AccountName: "account", StreamerLogin: "alpha", BroadcastID: "broadcast-1"},
			{AccountName: "account", StreamerLogin: "alpha", BroadcastID: "broadcast-2"},
		}
		want := map[string]WarmStartEntry{
			"alpha": {AccountName: "account", StreamerLogin: "alpha", BroadcastID: "broadcast-2"},
		}
		if got := entriesForAccount(stored, "account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing stored", func(t *testing.T) {
		want := map[string]WarmStartEntry{}
		if got := entriesForAccount(nil, "account"); !reflect.DeepEqual(got, want) {
			t.Errorf("entriesForAccount() = %#v, want %#v", got, want)
		}
	})
}

func TestEntryResolved(t *testing.T) {
	streamer := &domain.Streamer{Username: "streamer", Stream: &domain.Stream{BroadcastID: "broadcast-1"}}

	for _, row := range []struct {
		name  string
		entry WarmStartEntry
		want  bool
	}{
		{"a collected streak checked an hour ago", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-time.Hour), IsOnline: true}, true},
		{"an offline entry", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-time.Hour), IsOnline: false}, false},
		{"a streak still missing", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-time.Hour), IsOnline: true, WatchStreakMissing: true}, false},
		{"another broadcast", WarmStartEntry{BroadcastID: "broadcast-2", CheckedAt: now.Add(-time.Hour), IsOnline: true}, false},
		{"never checked", WarmStartEntry{BroadcastID: "broadcast-1", IsOnline: true}, false},
		{"checked one second under six hours ago", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-6*time.Hour + time.Second), IsOnline: true}, true},
		{"checked exactly six hours ago", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-6 * time.Hour), IsOnline: true}, true},
		{"checked one second over six hours ago", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(-6*time.Hour - time.Second), IsOnline: true}, false},
		{"checked this moment", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now, IsOnline: true}, true},
		{"checked an hour from now", WarmStartEntry{BroadcastID: "broadcast-1", CheckedAt: now.Add(time.Hour), IsOnline: true}, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := entryResolved(row.entry, streamer, now); got != row.want {
				t.Errorf("entryResolved() = %t, want %t", got, row.want)
			}
		})
	}
}

func TestSameWarmStartStream(t *testing.T) {
	for _, row := range []struct {
		name     string
		entry    WarmStartEntry
		streamer *domain.Streamer
		want     bool
	}{
		{
			"the same broadcast",
			WarmStartEntry{BroadcastID: "broadcast-1"},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1"}}, true,
		},
		{
			"another broadcast",
			WarmStartEntry{BroadcastID: "broadcast-1"},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-2"}}, false,
		},
		{
			"the same broadcast with spaces around the ids",
			WarmStartEntry{BroadcastID: " broadcast-1"},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1 "}}, true,
		},
		{
			"the same broadcast with another start time",
			WarmStartEntry{BroadcastID: "broadcast-1", StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt.Add(time.Minute)}}, true,
		},
		{
			"another broadcast with the same start time",
			WarmStartEntry{BroadcastID: "broadcast-1", StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-2", CreatedAt: startedAt}}, false,
		},
		{
			"only the entry has a broadcast id, the same start time",
			WarmStartEntry{BroadcastID: "broadcast-1", StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt}}, true,
		},
		{
			"only the stream has a broadcast id, the same start time",
			WarmStartEntry{StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt}}, true,
		},
		{
			"a broadcast id of only spaces, the same start time",
			WarmStartEntry{BroadcastID: "  ", StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{BroadcastID: "broadcast-1", CreatedAt: startedAt}}, true,
		},
		{
			"no broadcast ids, the same start time in another zone",
			WarmStartEntry{StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt.In(time.FixedZone("CEST", 2*60*60))}}, true,
		},
		{
			"no broadcast ids, another start time",
			WarmStartEntry{StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt.Add(time.Minute)}}, false,
		},
		{
			"no broadcast ids, the entry has no start time",
			WarmStartEntry{},
			&domain.Streamer{Stream: &domain.Stream{CreatedAt: startedAt}}, false,
		},
		{
			"no broadcast ids, the stream has no start time",
			WarmStartEntry{StreamCreatedAt: startedAt},
			&domain.Streamer{Stream: &domain.Stream{}}, false,
		},
		{"nothing to compare", WarmStartEntry{}, &domain.Streamer{Stream: &domain.Stream{}}, false},
		{"no stream", WarmStartEntry{BroadcastID: "broadcast-1"}, &domain.Streamer{}, false},
		{"no streamer", WarmStartEntry{BroadcastID: "broadcast-1"}, nil, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := SameWarmStartStream(row.entry, row.streamer); got != row.want {
				t.Errorf("SameWarmStartStream() = %t, want %t", got, row.want)
			}
		})
	}
}
