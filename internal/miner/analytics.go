package miner

import (
	"sync/atomic"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/notify"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

// AnalyticsSink receives point changes and log events for the analytics dashboard.
// Implementations must be safe for concurrent use and must not block for long.
type AnalyticsSink interface {
	RecordPoint(streamer string, points, delta int, reason string)
	RecordSnapshot(streamer string, points int, online bool)
	RecordEvent(event, message string)
	RecordBet(bet prediction.SettledBet)
}

const analyticsSnapshotInterval = 5 * time.Minute

type analyticsSinkHolder struct{ sink AnalyticsSink }

var analyticsSink atomic.Pointer[analyticsSinkHolder]

// SetAnalyticsSink installs the process-wide analytics sink (nil disables it)
// and subscribes it to logged events.
func SetAnalyticsSink(sink AnalyticsSink) {
	if sink == nil {
		analyticsSink.Store(nil)
		logging.Subscribe("analytics", nil)
		return
	}
	analyticsSink.Store(&analyticsSinkHolder{sink: sink})
	logging.Subscribe("analytics", func(event events.Event, message notify.Message) {
		sink.RecordEvent(string(event), message.Sentence())
	})
}

func currentAnalyticsSink() AnalyticsSink {
	if h := analyticsSink.Load(); h != nil {
		return h.sink
	}
	return nil
}

// recordSettledBet passes a bet whose result is known on to the dashboard.
func recordSettledBet(bet prediction.SettledBet) {
	if sink := currentAnalyticsSink(); sink != nil {
		sink.RecordBet(bet)
	}
}

func (m *Miner) analyticsRecordPoint(streamer *domain.Streamer, delta int, reason string) {
	if streamer == nil {
		return
	}
	if sink := currentAnalyticsSink(); sink != nil {
		sink.RecordPoint(streamer.Username, streamer.ChannelPoints, delta, reason)
	}
}

// analyticsSnapshotter records every streamer's balance and online state now and
// then every five minutes, catching changes that never pass through logPointsDelta.
func (m *Miner) analyticsSnapshotter(streamers []*domain.Streamer, stop <-chan struct{}) {
	ticker := time.NewTicker(analyticsSnapshotInterval)
	defer ticker.Stop()
	for {
		m.recordAnalyticsSnapshots(streamers)
		select {
		case <-stop:
			return
		case <-ticker.C:
		}
	}
}

func (m *Miner) recordAnalyticsSnapshots(streamers []*domain.Streamer) {
	if sink := currentAnalyticsSink(); sink != nil {
		for _, s := range streamers {
			if s != nil && s.PointsInit {
				sink.RecordSnapshot(s.Username, s.ChannelPoints, s.IsOnline)
			}
		}
	}
}
