package streak

import (
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const WarmStartFreshness = 6 * time.Hour

type WarmStartEntry struct {
	AccountName        string    `json:"account_name"`
	StreamerLogin      string    `json:"streamer_login"`
	ChannelID          string    `json:"channel_id,omitempty"`
	BroadcastID        string    `json:"broadcast_id,omitempty"`
	StreamCreatedAt    time.Time `json:"stream_created_at,omitempty"`
	WatchStreakMissing bool      `json:"watch_streak_missing"`
	CheckedAt          time.Time `json:"checked_at,omitempty"`
	IsOnline           bool      `json:"is_online"`
}

// entryFromStreamer is what the cache remembers about a streamer checked at
// checkedAt.
func entryFromStreamer(accountName string, streamer *domain.Streamer, checkedAt time.Time) WarmStartEntry {
	entry := WarmStartEntry{
		AccountName:        domain.LoginKey(accountName),
		StreamerLogin:      domain.LoginKey(streamer.Username),
		ChannelID:          strings.TrimSpace(streamer.ChannelID),
		CheckedAt:          checkedAt,
		IsOnline:           streamer.IsOnline,
		WatchStreakMissing: true,
	}
	if streamer.Stream != nil {
		entry.BroadcastID = strings.TrimSpace(streamer.Stream.BroadcastID)
		entry.StreamCreatedAt = streamer.Stream.CreatedAt
		entry.WatchStreakMissing = streamer.Stream.WatchStreakMissing
	}
	return entry
}

// entriesForAccount keys the stored entries by streamer login, keeping those
// of accountName. An empty accountName keeps every account's entries.
func entriesForAccount(stored []WarmStartEntry, accountName string) map[string]WarmStartEntry {
	entries := make(map[string]WarmStartEntry)
	accountKey := domain.LoginKey(accountName)
	for _, entry := range stored {
		loginKey := domain.LoginKey(entry.StreamerLogin)
		if loginKey == "" {
			continue
		}
		if accountKey != "" && domain.LoginKey(entry.AccountName) != accountKey {
			continue
		}
		entries[loginKey] = entry
	}
	return entries
}

// entryResolved reports whether entry still proves, at now, that the watch
// streak of the stream the streamer is on has been collected.
func entryResolved(entry WarmStartEntry, streamer *domain.Streamer, now time.Time) bool {
	if !entry.IsOnline || entry.WatchStreakMissing {
		return false
	}
	if entry.CheckedAt.IsZero() || now.Sub(entry.CheckedAt) > WarmStartFreshness {
		return false
	}
	return SameWarmStartStream(entry, streamer)
}

func SameWarmStartStream(entry WarmStartEntry, streamer *domain.Streamer) bool {
	if streamer == nil || streamer.Stream == nil {
		return false
	}
	broadcastID := strings.TrimSpace(streamer.Stream.BroadcastID)
	entryBroadcastID := strings.TrimSpace(entry.BroadcastID)
	if entryBroadcastID != "" && broadcastID != "" {
		return entryBroadcastID == broadcastID
	}
	if !entry.StreamCreatedAt.IsZero() && !streamer.Stream.CreatedAt.IsZero() {
		return entry.StreamCreatedAt.Equal(streamer.Stream.CreatedAt)
	}
	return false
}
