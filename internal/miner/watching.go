package miner

import (
	"errors"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/stoppable"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

const (
	minuteWatchFailureBackoffThreshold = 3
	minuteWatchFailureBackoffBase      = 5 * time.Minute
	minuteWatchFailureBackoffMax       = 30 * time.Minute
)

func streamerWatchBackedOff(streamer *domain.Streamer, now time.Time) bool {
	if streamer == nil || streamer.WatchBackoffUntil.IsZero() {
		return false
	}
	return now.Before(streamer.WatchBackoffUntil)
}

func recordMinuteWatchFailure(streamer *domain.Streamer, now time.Time) {
	if streamer == nil {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	streamer.WatchFailureCount++
	if streamer.WatchFailureCount < minuteWatchFailureBackoffThreshold {
		return
	}
	multiplier := streamer.WatchFailureCount - minuteWatchFailureBackoffThreshold + 1
	backoff := time.Duration(multiplier) * minuteWatchFailureBackoffBase
	if backoff > minuteWatchFailureBackoffMax {
		backoff = minuteWatchFailureBackoffMax
	}
	streamer.WatchBackoffUntil = now.Add(backoff)
}

func recordMinuteWatchSuccess(streamer *domain.Streamer) {
	if streamer == nil {
		return
	}
	streamer.WatchFailureCount = 0
	streamer.WatchBackoffUntil = time.Time{}
}

func (m *Miner) minuteWatcher(streamers []*domain.Streamer, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		m.beat()
		watchList := m.pickStreamersToWatch(streamers)
		if len(watchList) == 0 {
			if stoppable.Wait(20*time.Second, stop) {
				return
			}
			continue
		}

		interval := watchInterval(len(watchList))
		for _, streamer := range watchList {
			select {
			case <-stop:
				return
			default:
			}

			if !m.watchMinute(streamer) {
				continue
			}

			if stoppable.Wait(m.human.Around(interval, watchIntervalSpread), stop) {
				return
			}
		}
	}
}

// logMinuteWatchFailure reports a failed minute. A failure that keeps
// repeating is tried every few seconds, so it is reported the first time and
// then once per back-off, with when the next try is.
func (m *Miner) logMinuteWatchFailure(streamer *domain.Streamer, err error, counted bool) {
	name := m.styledStreamerName(streamer)
	switch {
	case !counted || streamer.WatchFailureCount == 1:
		m.logger.Errorf("minute watch %s: %v", name, err)
	case streamer.WatchFailureCount >= minuteWatchFailureBackoffThreshold:
		m.logger.Errorf("minute watch %s: failed %d times, next try at %s: %v", name, streamer.WatchFailureCount, streamer.WatchBackoffUntil.Format("15:04"), err)
	}
}

// watchMinute reports one watched minute for the streamer and records how it went.
// It returns false when the streamer turned out to be offline before anything was sent.
func (m *Miner) watchMinute(streamer *domain.Streamer) bool {
	if streamer.Stream != nil && streamer.Stream.LastUpdateAgo() > 10*time.Minute {
		wasOnline := streamer.IsOnline
		if _, err := m.checkStreamerOnlineWithIdentityRepair(streamer); err != nil {
			m.logger.Printf("online check %s: %v", m.styledStreamerName(streamer), err)
		}
		if !streamer.IsOnline {
			// The check marks the streamer offline itself. Going offline is
			// setPresence's to handle, so that it is announced and a finished
			// streak is carried over, as for any other way of going offline.
			streamer.IsOnline = wasOnline
			m.setPresence(streamer, false, "stale-stream")
			return false
		}
	}

	before := markStream(streamer)
	if err := m.sendMinuteWatchedWithIdentityRepair(streamer); err != nil {
		countWatchFailure := true
		if errors.Is(err, twitch.ErrStreamerOffline) {
			countWatchFailure = false
			live, liveErr := m.twitch.IsStreamLive(streamer.ChannelID)
			if liveErr != nil {
				m.logger.Printf("live check %s: %v", m.styledStreamerName(streamer), liveErr)
			}
			if !live {
				m.setPresence(streamer, false, "minute-watch")
			} else {
				m.logger.Printf("minute watch %s: transient offline response, keeping online", m.styledStreamerName(streamer))
			}
		}
		if countWatchFailure {
			recordMinuteWatchFailure(streamer, m.now())
		}
		m.logMinuteWatchFailure(streamer, err, countWatchFailure)
	} else {
		recordMinuteWatchSuccess(streamer)
		m.syncStreakSince(streamer, before)
	}
	m.syncActiveStreakWatch(streamer, m.now())
	return true
}

// watchIntervalSpread is how far the pause between two watched minutes may
// vary: a fifth shorter or longer.
const watchIntervalSpread = 0.2

// periodSpread is how far the wait between two periodic checks may vary.
const periodSpread = 0.2

// loadStagger is the longest a streamer's loading is held back at startup,
// so that the requests do not all leave in the same instant.
const loadStagger = 1200 * time.Millisecond

func watchInterval(count int) time.Duration {
	if count <= 0 {
		return 20 * time.Second
	}
	interval := time.Duration(float64(20*time.Second) / float64(count))
	if interval < 5*time.Second {
		return 5 * time.Second
	}
	return interval
}
