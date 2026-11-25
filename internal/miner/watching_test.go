package miner

import (
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestWatchInterval(t *testing.T) {
	m := &Miner{}
	if got := m.watchInterval(0); got != 20*time.Second {
		t.Fatalf("zero count got %s", got)
	}
	if got := m.watchInterval(2); got != 10*time.Second {
		t.Fatalf("two watchers got %s", got)
	}
	if got := m.watchInterval(10); got != 5*time.Second {
		t.Fatalf("min interval cap expected 5s got %s", got)
	}
}

func TestRecordMinuteWatchFailureBacksOffAfterRepeatedFailures(t *testing.T) {
	now := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)
	streamer := &domain.Streamer{Username: "oldlogin", ChannelID: "123456"}

	recordMinuteWatchFailure(streamer, now)
	recordMinuteWatchFailure(streamer, now.Add(10*time.Second))
	if !streamer.WatchBackoffUntil.IsZero() {
		t.Fatalf("backoff should not start before threshold")
	}

	recordMinuteWatchFailure(streamer, now.Add(20*time.Second))
	if !streamer.WatchBackoffUntil.After(now) {
		t.Fatalf("expected backoff to start after repeated failures")
	}
	if streamer.WatchFailureCount != 3 {
		t.Fatalf("failure count got %d", streamer.WatchFailureCount)
	}
}

func TestRecordMinuteWatchSuccessClearsBackoff(t *testing.T) {
	streamer := &domain.Streamer{
		Username:          "oldlogin",
		WatchFailureCount: 3,
		WatchBackoffUntil: time.Now().Add(5 * time.Minute),
	}

	recordMinuteWatchSuccess(streamer)
	if streamer.WatchFailureCount != 0 {
		t.Fatalf("failure count got %d", streamer.WatchFailureCount)
	}
	if !streamer.WatchBackoffUntil.IsZero() {
		t.Fatalf("backoff should be cleared")
	}
}

func TestStreamerWatchBackedOffExpires(t *testing.T) {
	now := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)
	streamer := &domain.Streamer{
		Username:          "stuck",
		WatchBackoffUntil: now.Add(5 * time.Minute),
	}

	if !streamerWatchBackedOff(streamer, now.Add(time.Minute)) {
		t.Fatalf("streamer should be backed off before expiry")
	}
	if streamerWatchBackedOff(streamer, now.Add(6*time.Minute)) {
		t.Fatalf("streamer should be selectable after expiry")
	}
}

func TestSleepWithStop(t *testing.T) {
	m := &Miner{}
	stop := make(chan struct{})
	done := make(chan bool, 1)
	go func() {
		done <- m.sleepWithStop(150*time.Millisecond, stop)
	}()
	time.Sleep(20 * time.Millisecond)
	close(stop)
	if !<-done {
		t.Fatalf("expected sleep to abort on stop")
	}

	start := time.Now()
	if m.sleepWithStop(50*time.Millisecond, make(chan struct{})) {
		t.Fatalf("expected no stop signal")
	}
	if elapsed := time.Since(start); elapsed < 45*time.Millisecond {
		t.Fatalf("sleep returned too early: %s", elapsed)
	}
}
