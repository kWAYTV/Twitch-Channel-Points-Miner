package stoppable

import (
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	t.Run("a wait that runs out", func(t *testing.T) {
		started := time.Now()

		stopped := Wait(2*time.Millisecond, make(chan struct{}))

		if stopped {
			t.Errorf("Wait = stopped, want it to run out")
		}
		if waited := time.Since(started); waited < 2*time.Millisecond {
			t.Errorf("waited %s, want at least 2ms", waited)
		}
	})

	t.Run("stopped before it starts", func(t *testing.T) {
		stop := make(chan struct{})
		close(stop)

		if stopped := Wait(time.Hour, stop); !stopped {
			t.Errorf("Wait = ran out, want stopped")
		}
	})

	t.Run("stopped while waiting", func(t *testing.T) {
		stop := make(chan struct{})
		done := make(chan bool)
		go func() { done <- Wait(time.Hour, stop) }()

		close(stop)

		if stopped := <-done; !stopped {
			t.Errorf("Wait = ran out, want stopped")
		}
	})

	t.Run("no time to wait, already stopped", func(t *testing.T) {
		stop := make(chan struct{})
		close(stop)

		if stopped := Wait(0, stop); stopped {
			t.Errorf("Wait(0) = stopped, want false: there was nothing to cut short")
		}
	})

	t.Run("a negative time", func(t *testing.T) {
		if stopped := Wait(-time.Second, make(chan struct{})); stopped {
			t.Errorf("Wait(-1s) = stopped, want false")
		}
	})

	t.Run("no stop channel", func(t *testing.T) {
		if stopped := Wait(time.Millisecond, nil); stopped {
			t.Errorf("Wait = stopped, want it to run out")
		}
	})
}
