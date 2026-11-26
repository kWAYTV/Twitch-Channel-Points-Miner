package domain

import (
	"fmt"
	"strings"
	"time"
)

type Stream struct {
	BroadcastID       string
	Title             string
	Game              map[string]interface{}
	Tags              []map[string]interface{}
	DropsTags         bool
	Campaigns         []interface{}
	CampaignIDs       []string
	CampaignsResolved bool
	DropsActive       bool
	ViewersCount      int
	SpadeURL          string
	PlaylistURL       string
	Payload           []map[string]interface{}

	WatchStreakMissing bool
	MinuteWatched      float64
	WatchCount         int
	CreatedAt          time.Time
	StreamUpAt         time.Time
	lastUpdate         time.Time
	lastMinuteUpdate   time.Time
}

func NewStream() *Stream {
	return &Stream{
		Game:               map[string]interface{}{},
		Tags:               []map[string]interface{}{},
		Campaigns:          []interface{}{},
		CampaignIDs:        []string{},
		WatchStreakMissing: true,
	}
}

// StartNewBroadcast forgets what was known about the broadcast before: the
// watch streak is open again, nothing has been watched, and the playlist
// address belongs to the old broadcast.
func (s *Stream) StartNewBroadcast() {
	s.WatchStreakMissing = true
	s.ResetWatchProgress()
	s.PlaylistURL = ""
}

func (s *Stream) Update(broadcastID, title string, game map[string]interface{}, tags []map[string]interface{}, viewers int, createdAt time.Time, dropID string) {
	s.updateAt(broadcastID, title, game, tags, viewers, createdAt, dropID, time.Now())
}

func (s *Stream) updateAt(broadcastID, title string, game map[string]interface{}, tags []map[string]interface{}, viewers int, createdAt time.Time, dropID string, now time.Time) {
	if s.BroadcastID != "" && s.BroadcastID != broadcastID {
		s.StartNewBroadcast()
	}
	s.BroadcastID = broadcastID
	s.Title = strings.TrimSpace(title)
	s.Game = game
	s.Tags = tags
	s.ViewersCount = viewers
	if !createdAt.IsZero() {
		s.CreatedAt = createdAt
	}
	s.DropsTags = false
	for _, tag := range tags {
		if id, ok := tag["id"].(string); ok && id == dropID && len(game) > 0 {
			s.DropsTags = true
			break
		}
	}
	s.lastUpdate = now
}

func (s *Stream) UpdateRequired() bool {
	return s.updateRequiredAt(time.Now())
}

func (s *Stream) updateRequiredAt(now time.Time) bool {
	return s.lastUpdate.IsZero() || now.Sub(s.lastUpdate) >= 120*time.Second
}

func (s *Stream) UpdateMinuteWatched() {
	s.updateMinuteWatchedAt(time.Now())
}

func (s *Stream) updateMinuteWatchedAt(now time.Time) {
	if !s.lastMinuteUpdate.IsZero() {
		duration := now.Sub(s.lastMinuteUpdate).Minutes()
		// if it's been more than 2 minutes since last minute-watched update,
		// stream probably was not being continuously watched, so reset to 0
		// (likely got pre-empted by something higher priority)
		if duration > 2.0 {
			s.MinuteWatched = 0
		} else {
			s.MinuteWatched += duration
		}
	}
	s.lastMinuteUpdate = now
}

func (s *Stream) ResetWatchProgress() {
	if s == nil {
		return
	}
	s.MinuteWatched = 0
	s.WatchCount = 0
	s.CreatedAt = time.Time{}
	s.lastMinuteUpdate = time.Time{}
}

func (s *Stream) LastUpdateAgo() time.Duration {
	return s.lastUpdateAgoAt(time.Now())
}

func (s *Stream) lastUpdateAgoAt(now time.Time) time.Duration {
	if s == nil || s.lastUpdate.IsZero() {
		return 0
	}
	return now.Sub(s.lastUpdate)
}

func (s *Stream) StreamUpElapsed() bool {
	return s.streamUpElapsedAt(time.Now())
}

func (s *Stream) streamUpElapsedAt(now time.Time) bool {
	if s == nil || s.StreamUpAt.IsZero() {
		return true
	}
	return now.Sub(s.StreamUpAt) > 2*time.Minute
}

func (s *Stream) String() string {
	return fmt.Sprintf("%s (%s)", s.Title, s.GameName())
}

func (s *Stream) GameName() string {
	if s == nil || s.Game == nil {
		return ""
	}
	if v, ok := s.Game["displayName"].(string); ok {
		return strings.TrimSpace(v)
	}
	if v, ok := s.Game["name"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}
