package miner

import (
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// peakTracker counts how many calls are running at once and remembers the most it saw.
type peakTracker struct {
	running atomic.Int32
	peak    atomic.Int32
}

func (p *peakTracker) enter() {
	now := p.running.Add(1)
	for {
		old := p.peak.Load()
		if now <= old || p.peak.CompareAndSwap(old, now) {
			return
		}
	}
}

func (p *peakTracker) leave() {
	p.running.Add(-1)
}

func TestForEachIndex(t *testing.T) {
	t.Run("nothing to visit", func(t *testing.T) {
		var calls atomic.Int32

		forEachIndex(0, 4, func(int) { calls.Add(1) })

		if got := calls.Load(); got != 0 {
			t.Errorf("calls = %d, want 0", got)
		}
	})

	t.Run("more indexes than the limit", func(t *testing.T) {
		var mu sync.Mutex
		var visited []int

		forEachIndex(10, 4, func(i int) {
			mu.Lock()
			visited = append(visited, i)
			mu.Unlock()
		})

		sort.Ints(visited)
		want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		if !reflect.DeepEqual(visited, want) {
			t.Errorf("visited = %v, want %v", visited, want)
		}
	})

	t.Run("as many at once as the limit allows", func(t *testing.T) {
		var allRunning sync.WaitGroup
		allRunning.Add(4)
		release := make(chan struct{})
		finished := make(chan struct{})

		go func() {
			forEachIndex(4, 4, func(int) {
				allRunning.Done()
				<-release
			})
			close(finished)
		}()
		reached := make(chan struct{})
		go func() {
			allRunning.Wait()
			close(reached)
		}()

		select {
		case <-reached:
		case <-time.After(5 * time.Second):
			t.Errorf("4 calls never ran at once")
		}
		close(release)
		<-finished
	})

	t.Run("never more at once than the limit", func(t *testing.T) {
		var tracker peakTracker

		forEachIndex(50, 4, func(int) {
			tracker.enter()
			runtime.Gosched()
			tracker.leave()
		})

		if got := tracker.peak.Load(); got > 4 {
			t.Errorf("most at once = %d, want at most 4", got)
		}
	})

	t.Run("a limit of zero", func(t *testing.T) {
		var tracker peakTracker
		var calls atomic.Int32

		forEachIndex(3, 0, func(int) {
			tracker.enter()
			calls.Add(1)
			runtime.Gosched()
			tracker.leave()
		})

		if got := calls.Load(); got != 3 {
			t.Errorf("calls = %d, want 3", got)
		}
		if got := tracker.peak.Load(); got != 1 {
			t.Errorf("most at once = %d, want 1", got)
		}
	})

	t.Run("a negative limit", func(t *testing.T) {
		var tracker peakTracker

		forEachIndex(3, -2, func(int) {
			tracker.enter()
			runtime.Gosched()
			tracker.leave()
		})

		if got := tracker.peak.Load(); got != 1 {
			t.Errorf("most at once = %d, want 1", got)
		}
	})
}
