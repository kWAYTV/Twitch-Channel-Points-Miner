package dashboard

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStore_Events(t *testing.T) {
	t.Run("its own events, then the ones that name no streamer", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("DROP_CLAIM", "Claimed a drop.")
		clock.Set(t0 + minute)
		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		events, err := store.Events("alice", 0)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		want := []Event{
			{T: t0 + minute, Type: "BET_WIN", Msg: "Won a bet on alice."},
			{T: t0, Type: "DROP_CLAIM", Msg: "Claimed a drop."},
		}
		if !reflect.DeepEqual(events, want) {
			t.Errorf("events = %+v, want %+v", events, want)
		}
	})

	t.Run("since exactly the event", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		events, err := store.Events("alice", t0)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		if want := []Event{{T: t0, Type: "BET_WIN", Msg: "Won a bet on alice."}}; !reflect.DeepEqual(events, want) {
			t.Errorf("events = %+v, want %+v", events, want)
		}
	})

	t.Run("since a millisecond after the event", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		events, err := store.Events("alice", t0+1)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		if events != nil {
			t.Errorf("events = %+v, want nil", events)
		}
	})

	t.Run("since a millisecond before the event", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		events, err := store.Events("alice", t0-1)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		if want := []Event{{T: t0, Type: "BET_WIN", Msg: "Won a bet on alice."}}; !reflect.DeepEqual(events, want) {
			t.Errorf("events = %+v, want %+v", events, want)
		}
	})

	t.Run("a name with no events", func(t *testing.T) {
		store, _, _ := setup(t)

		events, err := store.Events("alice", 0)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		if events != nil {
			t.Errorf("events = %+v, want nil", events)
		}
	})

	t.Run("a file with blank, foreign and torn lines around one event", func(t *testing.T) {
		store, _, dir := setup(t)
		writeFile(t, dir, "alice.events.jsonl", "\n"+
			`not json`+"\n"+
			`{"t":1700000000000,"msg":"No type."}`+"\n"+
			`{"t":1700000000000,"type":"BET_WIN","msg":"Won a bet on alice."}`+"\n"+
			`{"t":17000`)

		events, err := store.Events("alice", 0)

		if err != nil {
			t.Fatalf("Events error = %v, want nil", err)
		}
		if want := []Event{{T: t0, Type: "BET_WIN", Msg: "Won a bet on alice."}}; !reflect.DeepEqual(events, want) {
			t.Errorf("events = %+v, want %+v", events, want)
		}
	})

	t.Run("an events file that can't be read", func(t *testing.T) {
		store, _, dir := setup(t)
		replaceWithFolder(t, dir, "alice.events.jsonl")

		events, err := store.Events("alice", 0)

		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) || pathErr.Path != filepath.Join(dir, "alice.events.jsonl") {
			t.Errorf("Events error = %v, want a read error for the events file", err)
		}
		if events != nil {
			t.Errorf("events = %+v, want nil", events)
		}
	})

	t.Run("the file of events that name no streamer can't be read", func(t *testing.T) {
		store, _, dir := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("BET_WIN", "Won a bet on alice.")
		replaceWithFolder(t, dir, "_global.events.jsonl")

		events, err := store.Events("alice", 0)

		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) || pathErr.Path != filepath.Join(dir, "_global.events.jsonl") {
			t.Errorf("Events error = %v, want a read error for the shared events file", err)
		}
		if events != nil {
			t.Errorf("events = %+v, want nil", events)
		}
	})
}
