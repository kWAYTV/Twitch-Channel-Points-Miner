package miner

import (
	"sync"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

type recordedPoint struct {
	Streamer string
	Points   int
	Delta    int
	Reason   string
}

type recordedSnapshot struct {
	Streamer string
	Points   int
	Online   bool
}

type recordedEvent struct {
	Event   string
	Message string
}

// fakeAnalyticsSink keeps everything the miner records.
type fakeAnalyticsSink struct {
	mu        sync.Mutex
	Points    []recordedPoint
	Snapshots []recordedSnapshot
	Events    []recordedEvent
	Bets      []prediction.SettledBet
}

var _ AnalyticsSink = (*fakeAnalyticsSink)(nil)

func (f *fakeAnalyticsSink) RecordPoint(streamer string, points, delta int, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Points = append(f.Points, recordedPoint{Streamer: streamer, Points: points, Delta: delta, Reason: reason})
}

func (f *fakeAnalyticsSink) RecordSnapshot(streamer string, points int, online bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Snapshots = append(f.Snapshots, recordedSnapshot{Streamer: streamer, Points: points, Online: online})
}

func (f *fakeAnalyticsSink) RecordEvent(event, message string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Events = append(f.Events, recordedEvent{Event: event, Message: message})
}

func (f *fakeAnalyticsSink) RecordBet(bet prediction.SettledBet) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Bets = append(f.Bets, bet)
}

// installAnalyticsSink installs a fresh sink for one test and removes it afterwards.
func installAnalyticsSink(t *testing.T) *fakeAnalyticsSink {
	t.Helper()
	sink := &fakeAnalyticsSink{}
	SetAnalyticsSink(sink)
	t.Cleanup(func() { SetAnalyticsSink(nil) })
	return sink
}
