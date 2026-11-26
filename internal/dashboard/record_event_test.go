package dashboard

import (
	"reflect"
	"strings"
	"testing"
)

func eventsOf(t *testing.T, store *Store, name string) []Event {
	t.Helper()
	events, err := store.Events(name, 0)
	if err != nil {
		t.Fatalf("Events(%q, 0): %v", name, err)
	}
	return events
}

func TestStore_RecordEvent(t *testing.T) {
	t.Run("a bet won on a recorded streamer", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordPoint("bob", 500, 0, "")
		clock.Set(t0 + minute)

		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		want := []Event{{T: t0 + minute, Type: "BET_WIN", Msg: "Won a bet on alice."}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
		if got := eventsOf(t, store, "bob"); got != nil {
			t.Errorf("Events(bob) = %+v, want nil", got)
		}
	})

	t.Run("the same event twice", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		store.RecordEvent("BET_WIN", "Won a bet on alice.")

		want := []Event{
			{T: t0, Type: "BET_WIN", Msg: "Won a bet on alice."},
			{T: t0, Type: "BET_WIN", Msg: "Won a bet on alice."},
		}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
	})

	t.Run("a streamer going online", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		clock.Set(t0 + minute)

		store.RecordEvent("STREAMER_ONLINE", "alice is online.")

		want := []Event{{T: t0 + minute, Type: "STREAMER_ONLINE", Msg: "alice is online."}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1000, Online: true, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a streamer going offline", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordSnapshot("alice", 1000, true)
		clock.Set(t0 + minute)

		store.RecordEvent("STREAMER_OFFLINE", "alice went offline.")

		want := []Event{{T: t0 + minute, Type: "STREAMER_OFFLINE", Msg: "alice went offline."}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1000, Online: false, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("someone going online whom nobody recorded", func(t *testing.T) {
		store, _, dir := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		store.RecordEvent("STREAMER_ONLINE", "carol is online.")

		if got := eventsOf(t, store, "alice"); got != nil {
			t.Errorf("Events(alice) = %+v, want nil", got)
		}
		if got, want := fileNames(t, dir), []string{"alice.points.jsonl"}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("a drop claim that names no streamer", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		store.RecordPoint("bob", 500, 0, "")
		clock.Set(t0 + minute)

		store.RecordEvent("DROP_CLAIM", "Claimed a drop.")

		want := []Event{{T: t0 + minute, Type: "DROP_CLAIM", Msg: "Claimed a drop."}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
		if got := eventsOf(t, store, "bob"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(bob) = %+v, want %+v", got, want)
		}
	})

	t.Run("a type that is not kept", func(t *testing.T) {
		store, _, dir := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		store.RecordEvent("GAIN_FOR_WATCH", "Gained 10 points on alice.")

		if got := eventsOf(t, store, "alice"); got != nil {
			t.Errorf("Events(alice) = %+v, want nil", got)
		}
		if got, want := fileNames(t, dir), []string{"alice.points.jsonl"}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("a type in lower case and a message with spaces around it", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		store.RecordEvent(" bet_lose ", "  Lost a bet on alice.\n")

		want := []Event{{T: t0, Type: "BET_LOSE", Msg: "Lost a bet on alice."}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
	})

	t.Run("a message of 301 characters", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		store.RecordEvent("MOMENT_CLAIM", "alice "+strings.Repeat("x", 295))

		want := []Event{{T: t0, Type: "MOMENT_CLAIM", Msg: "alice " + strings.Repeat("x", 294) + "…"}}
		if got := eventsOf(t, store, "alice"); !reflect.DeepEqual(got, want) {
			t.Errorf("Events(alice) = %+v, want %+v", got, want)
		}
	})
}
