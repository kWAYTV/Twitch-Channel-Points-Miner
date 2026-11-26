package streak

import (
	"testing"
	"time"
)

func TestOrNow(t *testing.T) {
	t.Run("a set time", func(t *testing.T) {
		set := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)

		if got := orNow(set); !got.Equal(set) {
			t.Errorf("orNow(%s) = %s, want the same time", set, got)
		}
	})

	t.Run("a time never set", func(t *testing.T) {
		before := time.Now()

		got := orNow(time.Time{})

		if after := time.Now(); got.Before(before) || got.After(after) {
			t.Errorf("orNow(zero) = %s, want between %s and %s", got, before, after)
		}
	})
}
