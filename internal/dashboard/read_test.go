package dashboard

import (
	"reflect"
	"testing"
)

func TestStore_Known(t *testing.T) {
	t.Run("a recorded streamer", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		if got := store.Known("alice"); got != true {
			t.Errorf("Known(%q) = %v, want true", "alice", got)
		}
	})

	t.Run("a name nobody recorded", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")

		if got := store.Known("bob"); got != false {
			t.Errorf("Known(%q) = %v, want false", "bob", got)
		}
	})
}

func TestStore_Streamers(t *testing.T) {
	t.Run("no streamers", func(t *testing.T) {
		store, _, _ := setup(t)

		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("three streamers recorded out of order", func(t *testing.T) {
		store, _, _ := setup(t)
		store.RecordPoint("carol", 300, 0, "")
		store.RecordPoint("alice", 100, 0, "")
		store.RecordPoint("bob", 200, 0, "")

		want := []StreamerInfo{
			{Name: "alice", Points: 100, Updated: t0},
			{Name: "bob", Points: 200, Updated: t0},
			{Name: "carol", Points: 300, Updated: t0},
		}
		if got := store.Streamers(); !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})
}
