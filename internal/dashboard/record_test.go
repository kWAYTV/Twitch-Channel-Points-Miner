package dashboard

import (
	"os"
	"reflect"
	"sync"
	"testing"
)

func pointsOf(t *testing.T, store *Store, name string) []Point {
	t.Helper()
	points, _, err := store.Points(name, 0)
	if err != nil {
		t.Fatalf("Points(%q, 0): %v", name, err)
	}
	return points
}

func TestStore_RecordPoint(t *testing.T) {
	t.Run("a first balance", func(t *testing.T) {
		store, _, dir := setup(t)

		store.RecordPoint("alice", 1000, 0, "")

		if got, want := pointsOf(t, store, "alice"), []Point{{T: t0, P: 1000}}; !reflect.DeepEqual(got, want) {
			t.Errorf("Points = %+v, want %+v", got, want)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1000, SessionGain: 0, Online: false, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
		if got, want := fileNames(t, dir), []string{"alice.points.jsonl"}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("a gain with its reason", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		clock.Set(t0 + minute)

		store.RecordPoint("alice", 1010, 10, "WATCH")

		wantPoints := []Point{{T: t0, P: 1000}, {T: t0 + minute, P: 1010, D: 10, R: "WATCH"}}
		if got := pointsOf(t, store, "alice"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1010, SessionGain: 10, Updated: t0 + minute}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a loss", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		clock.Set(t0 + minute)

		store.RecordPoint("alice", 750, -250, "PREDICTION")

		wantPoints := []Point{{T: t0, P: 1000}, {T: t0 + minute, P: 750, D: -250, R: "PREDICTION"}}
		if got := pointsOf(t, store, "alice"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 750, SessionGain: -250, Updated: t0 + minute}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a gain as the first thing recorded", func(t *testing.T) {
		store, _, _ := setup(t)

		store.RecordPoint("alice", 1010, 10, "WATCH")

		wantPoints := []Point{{T: t0, P: 1010, D: 10, R: "WATCH"}}
		if got := pointsOf(t, store, "alice"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1010, SessionGain: 10, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a gain after a restart", func(t *testing.T) {
		earlier, clock, dir := setup(t)
		earlier.RecordPoint("alice", 1000, 0, "")
		clock.Set(t0 + hour)
		store := openStore(t, dir, clock)

		store.RecordPoint("alice", 1010, 10, "WATCH")

		wantPoints := []Point{{T: t0, P: 1000}, {T: t0 + hour, P: 1010, D: 10, R: "WATCH"}}
		if got := pointsOf(t, store, "alice"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1010, SessionGain: 10, Updated: t0 + hour}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("the same balance again, with no delta and no reason", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordPoint("alice", 1000, 0, "")
		clock.Set(t0 + minute)

		store.RecordPoint("alice", 1000, 0, "")

		if got, want := pointsOf(t, store, "alice"), []Point{{T: t0, P: 1000}}; !reflect.DeepEqual(got, want) {
			t.Errorf("Points = %+v, want %+v", got, want)
		}
	})

	t.Run("a name in capitals with spaces around it", func(t *testing.T) {
		store, _, dir := setup(t)

		store.RecordPoint(" Alice ", 1000, 0, "")

		if got, want := pointsOf(t, store, "alice"), []Point{{T: t0, P: 1000}}; !reflect.DeepEqual(got, want) {
			t.Errorf("Points = %+v, want %+v", got, want)
		}
		if got, want := fileNames(t, dir), []string{"alice.points.jsonl"}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("a name that is not a login", func(t *testing.T) {
		store, _, dir := setup(t)

		store.RecordPoint("../evil", 1000, 0, "")

		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
		if got, want := fileNames(t, dir), []string{}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("the folder is gone", func(t *testing.T) {
		store, _, dir := setup(t)
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("remove %s: %v", dir, err)
		}

		store.RecordPoint("alice", 1000, 0, "")

		if got := pointsOf(t, store, "alice"); got != nil {
			t.Errorf("Points = %+v, want nil", got)
		}
		wantSummary := []StreamerInfo{{Name: "alice", Points: 1000, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})
}

func TestStore_RecordSnapshot(t *testing.T) {
	t.Run("a first balance, online", func(t *testing.T) {
		store, _, _ := setup(t)

		store.RecordSnapshot("bob", 500, true)

		if got, want := pointsOf(t, store, "bob"), []Point{{T: t0, P: 500}}; !reflect.DeepEqual(got, want) {
			t.Errorf("Points = %+v, want %+v", got, want)
		}
		wantSummary := []StreamerInfo{{Name: "bob", Points: 500, SessionGain: 0, Online: true, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("an unchanged balance five minutes later, now offline", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordSnapshot("bob", 500, true)
		clock.Set(t0 + 5*minute)

		store.RecordSnapshot("bob", 500, false)

		if got, want := pointsOf(t, store, "bob"), []Point{{T: t0, P: 500}}; !reflect.DeepEqual(got, want) {
			t.Errorf("Points = %+v, want %+v", got, want)
		}
		wantSummary := []StreamerInfo{{Name: "bob", Points: 500, Online: false, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("an unchanged balance an hour later", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordSnapshot("bob", 500, true)
		clock.Set(t0 + hour)

		store.RecordSnapshot("bob", 500, true)

		wantPoints := []Point{{T: t0, P: 500}, {T: t0 + hour, P: 500}}
		if got := pointsOf(t, store, "bob"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "bob", Points: 500, Online: true, Updated: t0 + hour}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a changed balance", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordSnapshot("bob", 500, true)
		clock.Set(t0 + 5*minute)

		store.RecordSnapshot("bob", 300, true)

		wantPoints := []Point{{T: t0, P: 500}, {T: t0 + 5*minute, P: 300, D: -200}}
		if got := pointsOf(t, store, "bob"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "bob", Points: 300, SessionGain: -200, Online: true, Updated: t0 + 5*minute}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a balance that changed while the miner was off", func(t *testing.T) {
		earlier, clock, dir := setup(t)
		earlier.RecordSnapshot("bob", 500, true)
		clock.Set(t0 + hour)
		store := openStore(t, dir, clock)

		store.RecordSnapshot("bob", 540, true)

		wantPoints := []Point{{T: t0, P: 500}, {T: t0 + hour, P: 540, D: 40}}
		if got := pointsOf(t, store, "bob"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
		wantSummary := []StreamerInfo{{Name: "bob", Points: 540, SessionGain: 0, Online: true, Updated: t0 + hour}}
		if got := store.Streamers(); !reflect.DeepEqual(got, wantSummary) {
			t.Errorf("Streamers() = %+v, want %+v", got, wantSummary)
		}
	})

	t.Run("a name that is not a login", func(t *testing.T) {
		store, _, dir := setup(t)

		store.RecordSnapshot("../evil", 500, true)

		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
		if got, want := fileNames(t, dir), []string{}; !reflect.DeepEqual(got, want) {
			t.Errorf("files = %v, want %v", got, want)
		}
	})

	t.Run("eight snapshots of one new balance at once record it once", func(t *testing.T) {
		store, clock, _ := setup(t)
		store.RecordSnapshot("bob", 500, true)
		clock.Set(t0 + 5*minute)
		var ready, done sync.WaitGroup
		ready.Add(1)
		for i := 0; i < 8; i++ {
			done.Add(1)
			go func() {
				defer done.Done()
				ready.Wait()
				store.RecordSnapshot("bob", 300, true)
			}()
		}

		ready.Done()
		done.Wait()

		wantPoints := []Point{{T: t0, P: 500}, {T: t0 + 5*minute, P: 300, D: -200}}
		if got := pointsOf(t, store, "bob"); !reflect.DeepEqual(got, wantPoints) {
			t.Errorf("Points = %+v, want %+v", got, wantPoints)
		}
	})
}
