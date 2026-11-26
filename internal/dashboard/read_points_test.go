package dashboard

import (
	"bufio"
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// threePoints records alice at t0, a minute later and two minutes later.
func threePoints(t *testing.T) *Store {
	t.Helper()
	store, clock, _ := setup(t)
	store.RecordPoint("alice", 1000, 0, "")
	clock.Set(t0 + minute)
	store.RecordPoint("alice", 1010, 10, "WATCH")
	clock.Set(t0 + 2*minute)
	store.RecordPoint("alice", 1460, 450, "WATCH_STREAK")
	return store
}

func TestStore_Points(t *testing.T) {
	first := Point{T: t0, P: 1000}
	second := Point{T: t0 + minute, P: 1010, D: 10, R: "WATCH"}
	third := Point{T: t0 + 2*minute, P: 1460, D: 450, R: "WATCH_STREAK"}

	t.Run("since the beginning", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("alice", 0)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if want := []Point{first, second, third}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
		if prev != nil {
			t.Errorf("prev = %+v, want nil", prev)
		}
	})

	t.Run("since exactly the second point", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("alice", t0+minute)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if want := []Point{second, third}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
		if !reflect.DeepEqual(prev, &first) {
			t.Errorf("prev = %+v, want %+v", prev, &first)
		}
	})

	t.Run("since a millisecond after the second point", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("alice", t0+minute+1)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if want := []Point{third}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
		if !reflect.DeepEqual(prev, &second) {
			t.Errorf("prev = %+v, want %+v", prev, &second)
		}
	})

	t.Run("since a millisecond before the second point", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("alice", t0+minute-1)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if want := []Point{second, third}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
		if !reflect.DeepEqual(prev, &first) {
			t.Errorf("prev = %+v, want %+v", prev, &first)
		}
	})

	t.Run("since after the last point", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("alice", t0+hour)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if points != nil {
			t.Errorf("points = %+v, want nil", points)
		}
		if !reflect.DeepEqual(prev, &third) {
			t.Errorf("prev = %+v, want %+v", prev, &third)
		}
	})

	t.Run("a name nobody recorded", func(t *testing.T) {
		store := threePoints(t)

		points, prev, err := store.Points("bob", 0)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if points != nil {
			t.Errorf("points = %+v, want nil", points)
		}
		if prev != nil {
			t.Errorf("prev = %+v, want nil", prev)
		}
	})

	t.Run("a file with blank, foreign and torn lines around one point", func(t *testing.T) {
		store, _, dir := setup(t)
		writeFile(t, dir, "alice.points.jsonl", "\n"+
			`not json`+"\n"+
			`{"p":5}`+"\n"+
			`{"t":1700000000000,"p":1000}`+"\n"+
			"\n"+
			`{"t":17000`)

		points, prev, err := store.Points("alice", 0)

		if err != nil {
			t.Fatalf("Points error = %v, want nil", err)
		}
		if want := []Point{first}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
		if prev != nil {
			t.Errorf("prev = %+v, want nil", prev)
		}
	})

	t.Run("a line longer than a megabyte", func(t *testing.T) {
		store, _, dir := setup(t)
		writeFile(t, dir, "alice.points.jsonl", `{"t":1700000000000,"p":1000}`+"\n"+strings.Repeat("x", 1024*1024+1)+"\n")

		points, _, err := store.Points("alice", 0)

		if !errors.Is(err, bufio.ErrTooLong) {
			t.Errorf("Points error = %v, want %v", err, bufio.ErrTooLong)
		}
		if want := []Point{first}; !reflect.DeepEqual(points, want) {
			t.Errorf("points = %+v, want %+v", points, want)
		}
	})

	t.Run("a points file that can't be read", func(t *testing.T) {
		store, _, dir := setup(t)
		replaceWithFolder(t, dir, "alice.points.jsonl")

		points, prev, err := store.Points("alice", 0)

		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) || pathErr.Path != filepath.Join(dir, "alice.points.jsonl") {
			t.Errorf("Points error = %v, want a read error for the points file", err)
		}
		if points != nil {
			t.Errorf("points = %+v, want nil", points)
		}
		if prev != nil {
			t.Errorf("prev = %+v, want nil", prev)
		}
	})
}
