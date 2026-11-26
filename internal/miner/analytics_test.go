package miner

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestSetAnalyticsSink(t *testing.T) {
	t.Run("a logged event", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)

		h.miner.logger.Eventf(events.EventStreamerOnline, "%sAlice%s is Online!", colorGreen, colorReset)

		if want := []recordedEvent{{Event: "STREAMER_ONLINE", Message: "Alice is online."}}; !reflect.DeepEqual(sink.Events, want) {
			t.Errorf("recorded events = %+v, want %+v", sink.Events, want)
		}
	})

	t.Run("a plain log line", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)

		h.miner.logger.Printf("Alice is Online!")

		if sink.Events != nil {
			t.Errorf("recorded events = %+v, want none", sink.Events)
		}
	})

	t.Run("a logged event after the sink is removed", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		SetAnalyticsSink(nil)

		h.miner.logger.Eventf(events.EventStreamerOnline, "Alice is Online!")

		if sink.Events != nil {
			t.Errorf("recorded events = %+v, want none", sink.Events)
		}
	})

	t.Run("a logged event after another sink takes over", func(t *testing.T) {
		h := setup(t, Options{})
		first := installAnalyticsSink(t)
		second := installAnalyticsSink(t)

		h.miner.logger.Eventf(events.EventStreamerOnline, "Alice is Online!")

		if first.Events != nil {
			t.Errorf("events recorded by the first sink = %+v, want none", first.Events)
		}
		if want := []recordedEvent{{Event: "STREAMER_ONLINE", Message: "Alice is online."}}; !reflect.DeepEqual(second.Events, want) {
			t.Errorf("events recorded by the second sink = %+v, want %+v", second.Events, want)
		}
	})

	t.Run("a balance change after the sink is removed", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		SetAnalyticsSink(nil)

		h.miner.logPointsDelta(&domain.Streamer{Username: "alice", ChannelPoints: 750}, -250, "PREDICTION")

		if sink.Points != nil {
			t.Errorf("recorded points = %+v, want none", sink.Points)
		}
	})
}

func TestMiner_recordAnalyticsSnapshots(t *testing.T) {
	loaded := func(login string, points int, online bool) *domain.Streamer {
		return &domain.Streamer{Username: login, ChannelPoints: points, IsOnline: online, PointsInit: true}
	}

	t.Run("an online and an offline streamer", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)

		h.miner.recordAnalyticsSnapshots([]*domain.Streamer{loaded("alice", 100, true), loaded("bob", 50, false)})

		want := []recordedSnapshot{
			{Streamer: "alice", Points: 100, Online: true},
			{Streamer: "bob", Points: 50, Online: false},
		}
		if !reflect.DeepEqual(sink.Snapshots, want) {
			t.Errorf("recorded snapshots = %+v, want %+v", sink.Snapshots, want)
		}
	})

	t.Run("a streamer whose balance never loaded, and a missing one", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		unloaded := &domain.Streamer{Username: "bob", IsOnline: true}

		h.miner.recordAnalyticsSnapshots([]*domain.Streamer{unloaded, nil, loaded("alice", 100, true)})

		want := []recordedSnapshot{{Streamer: "alice", Points: 100, Online: true}}
		if !reflect.DeepEqual(sink.Snapshots, want) {
			t.Errorf("recorded snapshots = %+v, want %+v", sink.Snapshots, want)
		}
	})

	t.Run("no sink installed", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.recordAnalyticsSnapshots([]*domain.Streamer{loaded("alice", 100, true)})
	})
}

func TestMiner_analyticsSnapshotter(t *testing.T) {
	t.Run("stopped before it starts", func(t *testing.T) {
		h := setup(t, Options{})
		sink := installAnalyticsSink(t)
		alice := &domain.Streamer{Username: "alice", ChannelPoints: 100, IsOnline: true, PointsInit: true}
		stop := make(chan struct{})
		close(stop)

		h.miner.analyticsSnapshotter([]*domain.Streamer{alice}, stop)

		want := []recordedSnapshot{{Streamer: "alice", Points: 100, Online: true}}
		if !reflect.DeepEqual(sink.Snapshots, want) {
			t.Errorf("recorded snapshots = %+v, want %+v", sink.Snapshots, want)
		}
	})
}
