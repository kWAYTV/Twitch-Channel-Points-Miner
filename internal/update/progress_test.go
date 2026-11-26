package update

import (
	"reflect"
	"testing"
)

const mb = 1024 * 1024

func checkLog(t *testing.T, log *logLines, want []string) {
	t.Helper()
	if !reflect.DeepEqual(log.Lines, want) {
		t.Errorf("logged %q, want %q", log.Lines, want)
	}
}

func TestProgressWriter_Write(t *testing.T) {
	t.Run("less than one step", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)

		n, err := progress.Write(make([]byte, 99))

		if n != 99 || err != nil {
			t.Errorf("Write = %d, %v, want 99, <nil>", n, err)
		}
		checkLog(t, log, nil)
	})

	t.Run("exactly one step", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)

		progress.Write(make([]byte, 100))

		checkLog(t, log, []string{"auto-update: download 5% (0.0/0.0 MB)"})
	})

	t.Run("the byte that completes a step", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)
		progress.Write(make([]byte, 99))

		progress.Write(make([]byte, 1))

		checkLog(t, log, []string{"auto-update: download 5% (0.0/0.0 MB)"})
	})

	t.Run("a second step", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)
		progress.Write(make([]byte, 100))

		progress.Write(make([]byte, 100))

		checkLog(t, log, []string{
			"auto-update: download 5% (0.0/0.0 MB)",
			"auto-update: download 10% (0.0/0.0 MB)",
		})
	})

	t.Run("several steps in one write", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)

		progress.Write(make([]byte, 1000))

		checkLog(t, log, []string{"auto-update: download 50% (0.0/0.0 MB)"})
	})

	t.Run("a small write after several steps at once", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)
		progress.Write(make([]byte, 1000))

		progress.Write(make([]byte, 99))

		checkLog(t, log, []string{"auto-update: download 50% (0.0/0.0 MB)"})
	})

	t.Run("one megabyte of twenty", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(20*mb, log.Logf)

		progress.Write(make([]byte, mb))

		checkLog(t, log, []string{"auto-update: download 5% (1.0/20.0 MB)"})
	})

	t.Run("more than the total", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)

		progress.Write(make([]byte, 3000))

		checkLog(t, log, []string{"auto-update: download 100% (0.0/0.0 MB)"})
	})

	t.Run("a total too small to split into steps", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(5, log.Logf)

		progress.Write(make([]byte, 5))

		checkLog(t, log, []string{"auto-update: download 100% (0.0/0.0 MB)"})
	})

	t.Run("part of a total too small to split into steps", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(5, log.Logf)

		progress.Write(make([]byte, 4))

		checkLog(t, log, nil)
	})

	t.Run("an empty write", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2000, log.Logf)

		n, err := progress.Write(nil)

		if n != 0 || err != nil {
			t.Errorf("Write = %d, %v, want 0, <nil>", n, err)
		}
		checkLog(t, log, nil)
	})

	t.Run("no total", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(0, log.Logf)

		progress.Write(make([]byte, 100))

		checkLog(t, log, nil)
	})
}

func TestProgressWriter_Done(t *testing.T) {
	t.Run("after every byte", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2*mb, log.Logf)
		progress.Write(make([]byte, 2*mb))

		progress.Done()

		checkLog(t, log, []string{
			"auto-update: download 100% (2.0/2.0 MB)",
			"auto-update: download 100% (2.0/2.0 MB)",
		})
	})

	t.Run("after half the bytes", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2*mb, log.Logf)
		progress.Write(make([]byte, mb))

		progress.Done()

		checkLog(t, log, []string{
			"auto-update: download 50% (1.0/2.0 MB)",
			"auto-update: download 100% (2.0/2.0 MB)",
		})
	})

	t.Run("after no bytes", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(2*mb, log.Logf)

		progress.Done()

		checkLog(t, log, []string{"auto-update: download 100% (2.0/2.0 MB)"})
	})

	t.Run("no total", func(t *testing.T) {
		log := &logLines{}
		progress := newProgressWriter(0, log.Logf)

		progress.Done()

		checkLog(t, log, nil)
	})
}

func TestBytesToMB(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  float64
	}{
		{"nothing", 0, 0},
		{"half a megabyte", 524288, 0.5},
		{"one megabyte", 1048576, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToMB(tt.bytes); got != tt.want {
				t.Errorf("bytesToMB(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}
