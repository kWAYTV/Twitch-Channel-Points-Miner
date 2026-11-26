package miner

import (
	"reflect"
	"testing"
)

func assertSlots(t *testing.T, slots *watchSlots, wantPicked []int, wantReasons map[int]string) {
	t.Helper()
	if !reflect.DeepEqual(slots.picked, wantPicked) {
		t.Errorf("picked = %v, want %v", slots.picked, wantPicked)
	}
	if !reflect.DeepEqual(slots.reasons, wantReasons) {
		t.Errorf("reasons = %v, want %v", slots.reasons, wantReasons)
	}
}

func TestWatchSlots_full(t *testing.T) {
	a := watchCandidate{idx: 0, position: 0}
	b := watchCandidate{idx: 1, position: 1}

	t.Run("no picks", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{a, b})

		if slots.full() {
			t.Errorf("full() = true, want false")
		}
	})

	t.Run("one pick", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{a, b})
		slots.take(a, "ORDER")

		if slots.full() {
			t.Errorf("full() = true, want false")
		}
	})

	t.Run("two picks", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{a, b})
		slots.take(a, "ORDER")
		slots.take(b, "ORDER")

		if !slots.full() {
			t.Errorf("full() = false, want true")
		}
	})
}

func TestWatchSlots_add(t *testing.T) {
	chess := watchCandidate{idx: 0, position: 0, game: "chess"}
	chessToo := watchCandidate{idx: 1, position: 1, game: "chess"}
	poker := watchCandidate{idx: 2, position: 2, game: "poker"}
	noGame := watchCandidate{idx: 3, position: 3}

	t.Run("a first candidate", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, poker})

		slots.add(chess, "ORDER", false)

		assertSlots(t, slots, []int{0}, map[int]string{0: "ORDER"})
	})

	t.Run("a candidate without a reason", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, poker})

		slots.add(chess, "", false)

		assertSlots(t, slots, []int{0}, map[int]string{})
	})

	t.Run("a candidate already picked", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, poker})
		slots.add(chess, "ORDER", false)

		slots.add(chess, "DROPS", false)

		assertSlots(t, slots, []int{0}, map[int]string{0: "ORDER"})
	})

	t.Run("a third candidate", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, poker, noGame})
		slots.add(chess, "ORDER", false)
		slots.add(poker, "ORDER", false)

		slots.add(noGame, "DROPS", true)

		assertSlots(t, slots, []int{0, 2}, map[int]string{0: "ORDER", 2: "ORDER"})
	})

	t.Run("a picked game again while another game waits", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, chessToo, poker})
		slots.add(chess, "ORDER", false)

		slots.add(chessToo, "ORDER", false)

		assertSlots(t, slots, []int{0}, map[int]string{0: "ORDER"})
	})

	t.Run("a picked game again while no other game waits", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, chessToo, noGame})
		slots.add(chess, "ORDER", false)

		slots.add(chessToo, "ORDER", false)

		assertSlots(t, slots, []int{0, 1}, map[int]string{0: "ORDER", 1: "ORDER"})
	})

	t.Run("a picked game again, forced", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, chessToo, poker})
		slots.add(chess, "ORDER", false)

		slots.add(chessToo, "STREAK", true)

		assertSlots(t, slots, []int{0, 1}, map[int]string{0: "ORDER", 1: "STREAK"})
	})

	t.Run("a candidate with no game after one with no game", func(t *testing.T) {
		alsoNoGame := watchCandidate{idx: 4, position: 4}
		slots := newWatchSlots([]watchCandidate{noGame, alsoNoGame, poker})
		slots.add(noGame, "ORDER", false)

		slots.add(alsoNoGame, "ORDER", false)

		assertSlots(t, slots, []int{3, 4}, map[int]string{3: "ORDER", 4: "ORDER"})
	})
}

func TestWatchSlots_take(t *testing.T) {
	chess := watchCandidate{idx: 0, position: 0, game: "chess"}
	chessToo := watchCandidate{idx: 1, position: 1, game: "chess"}
	poker := watchCandidate{idx: 2, position: 2, game: "poker"}

	t.Run("a picked game again while another game waits", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, chessToo, poker})
		slots.take(chess, "ORDER")

		slots.take(chessToo, "FORCE_STREAK_SLOT2")

		assertSlots(t, slots, []int{0, 1}, map[int]string{0: "ORDER", 1: "FORCE_STREAK_SLOT2"})
	})

	t.Run("a candidate without a reason", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{chess, poker})

		slots.take(poker, "")

		assertSlots(t, slots, []int{2}, map[int]string{})
	})
}

func TestWatchSlots_pick(t *testing.T) {
	first := watchCandidate{idx: 0, position: 0, rank: 3, game: "chess"}
	second := watchCandidate{idx: 1, position: 1, rank: 1, game: "poker"}
	third := watchCandidate{idx: 2, position: 2, rank: 0, game: "go"}

	t.Run("an empty list", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{first, second, third})

		slots.pick(nil, true, nil, "ORDER", false)

		assertSlots(t, slots, []int{}, map[int]string{})
	})

	t.Run("in list order", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{first, second, third})

		slots.pick([]watchCandidate{third, second, first}, false, nil, "ORDER", false)

		assertSlots(t, slots, []int{0, 1}, map[int]string{0: "ORDER", 1: "ORDER"})
	})

	t.Run("by game rank", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{first, second, third})

		slots.pick([]watchCandidate{first, second, third}, true, nil, "DROPS", false)

		assertSlots(t, slots, []int{2, 1}, map[int]string{2: "DROPS", 1: "DROPS"})
	})

	t.Run("by its own order first", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{first, second, third})
		highestIdxFirst := func(a, b watchCandidate) bool { return a.idx > b.idx }

		slots.pick([]watchCandidate{first, second}, true, highestIdxFirst, "POINTS_DESC", false)

		assertSlots(t, slots, []int{1, 0}, map[int]string{1: "POINTS_DESC", 0: "POINTS_DESC"})
	})

	t.Run("one slot left", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{first, second, third})
		slots.take(third, "STREAK")

		slots.pick([]watchCandidate{first, second}, false, nil, "ORDER", false)

		assertSlots(t, slots, []int{2, 0}, map[int]string{2: "STREAK", 0: "ORDER"})
	})

	t.Run("two of one game while another game waits", func(t *testing.T) {
		chessToo := watchCandidate{idx: 3, position: 3, game: "chess"}
		slots := newWatchSlots([]watchCandidate{first, chessToo, second})

		slots.pick([]watchCandidate{first, chessToo}, false, nil, "DROPS", false)

		assertSlots(t, slots, []int{0}, map[int]string{0: "DROPS"})
	})

	t.Run("two of one game, forced", func(t *testing.T) {
		chessToo := watchCandidate{idx: 3, position: 3, game: "chess"}
		slots := newWatchSlots([]watchCandidate{first, chessToo, second})

		slots.pick([]watchCandidate{first, chessToo}, false, nil, "STREAK", true)

		assertSlots(t, slots, []int{0, 3}, map[int]string{0: "STREAK", 3: "STREAK"})
	})
}

func TestWatchSlots_ensureStreak(t *testing.T) {
	plain := watchCandidate{idx: 0, position: 0, rank: 2, game: "chess"}
	other := watchCandidate{idx: 1, position: 1, rank: 2, game: "poker"}
	streak := watchCandidate{idx: 2, position: 2, rank: 2, game: "chess", isStreakReady: true}

	t.Run("no pending streak", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{plain, other})
		slots.take(plain, "ORDER")
		slots.take(other, "ORDER")

		slots.ensureStreak(nil)

		assertSlots(t, slots, []int{0, 1}, map[int]string{0: "ORDER", 1: "ORDER"})
	})

	t.Run("nothing picked yet", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{plain, streak})

		slots.ensureStreak([]watchCandidate{streak})

		assertSlots(t, slots, []int{}, map[int]string{})
	})

	t.Run("a streak already picked", func(t *testing.T) {
		streakToo := watchCandidate{idx: 3, position: 3, rank: 0, game: "go", isStreakReady: true}
		slots := newWatchSlots([]watchCandidate{plain, streak, streakToo})
		slots.take(plain, "ORDER")
		slots.take(streak, "STREAK")

		slots.ensureStreak([]watchCandidate{streak, streakToo})

		assertSlots(t, slots, []int{0, 2}, map[int]string{0: "ORDER", 2: "STREAK"})
	})

	t.Run("a free slot", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{plain, other, streak})
		slots.take(plain, "ORDER")

		slots.ensureStreak([]watchCandidate{streak})

		assertSlots(t, slots, []int{0, 2}, map[int]string{0: "ORDER", 2: "FORCE_STREAK_SLOT2"})
	})

	t.Run("both slots taken", func(t *testing.T) {
		slots := newWatchSlots([]watchCandidate{plain, other, streak})
		slots.take(plain, "SUBSCRIBED")
		slots.take(other, "ORDER")

		slots.ensureStreak([]watchCandidate{streak})

		assertSlots(t, slots, []int{0, 2}, map[int]string{0: "SUBSCRIBED", 1: "ORDER", 2: "FORCE_STREAK_SLOT2"})
	})

	t.Run("two pending streaks", func(t *testing.T) {
		betterRanked := watchCandidate{idx: 3, position: 3, rank: 0, game: "go", isStreakReady: true}
		slots := newWatchSlots([]watchCandidate{plain, streak, betterRanked})
		slots.take(plain, "ORDER")

		slots.ensureStreak([]watchCandidate{streak, betterRanked})

		assertSlots(t, slots, []int{0, 3}, map[int]string{0: "ORDER", 3: "FORCE_STREAK_SLOT2"})
	})
}

func TestWatchSlots_preferPriorityGameFirst(t *testing.T) {
	plain := watchCandidate{idx: 0, position: 0}
	streakOffList := watchCandidate{idx: 1, position: 1, isStreakReady: true}
	streakOffListToo := watchCandidate{idx: 2, position: 2, isStreakReady: true}
	streakOnList := watchCandidate{idx: 3, position: 3, isStreakReady: true, priorityGame: true}
	tests := []struct {
		name   string
		picked []watchCandidate
		want   []int
	}{
		{"one pick", []watchCandidate{streakOffList}, []int{1}},
		{"a streak off the list, then a plain pick", []watchCandidate{streakOffList, plain}, []int{0, 1}},
		{"a streak off the list, then a streak on it", []watchCandidate{streakOffList, streakOnList}, []int{3, 1}},
		{"two streaks off the list", []watchCandidate{streakOffList, streakOffListToo}, []int{1, 2}},
		{"a streak on the list, then a plain pick", []watchCandidate{streakOnList, plain}, []int{3, 0}},
		{"a plain pick, then a streak off the list", []watchCandidate{plain, streakOffList}, []int{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := newWatchSlots([]watchCandidate{plain, streakOffList, streakOffListToo, streakOnList})
			for _, c := range tt.picked {
				slots.take(c, "ORDER")
			}

			slots.preferPriorityGameFirst()

			if !reflect.DeepEqual(slots.picked, tt.want) {
				t.Errorf("picked = %v, want %v", slots.picked, tt.want)
			}
		})
	}
}
