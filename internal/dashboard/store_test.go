package dashboard

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// setup opens a store on a folder of its own, with the clock at t0.
func setup(t *testing.T) (*Store, *fakeClock, string) {
	t.Helper()
	dir := t.TempDir()
	clock := newFakeClock(t0)
	return openStore(t, dir, clock), clock, dir
}

func openStore(t *testing.T, dir string, clock *fakeClock) *Store {
	t.Helper()
	store, err := newStore(dir, clock.Now)
	if err != nil {
		t.Fatalf("newStore(%q): %v", dir, err)
	}
	return store
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// replaceWithFolder puts a folder where a data file goes, so reading it fails.
func replaceWithFolder(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("remove %s: %v", name, err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", name, err)
	}
}

func fileNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestNewStore(t *testing.T) {
	t.Run("a folder that does not exist yet", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "data", "analytics")

		store, err := newStore(dir, newFakeClock(t0).Now)

		if err != nil {
			t.Fatalf("newStore(%q) error = %v, want nil", dir, err)
		}
		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("%q is not a folder: %v", dir, err)
		}
	})

	t.Run("a path that is a file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "analytics", "")
		path := filepath.Join(dir, "analytics")

		store, err := newStore(path, newFakeClock(t0).Now)

		if store != nil {
			t.Errorf("newStore(%q) = %+v, want nil", path, store)
		}
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) || pathErr.Op != "mkdir" || pathErr.Path != path {
			t.Errorf("newStore(%q) error = %v, want a mkdir error for that path", path, err)
		}
	})

	t.Run("a streamer recorded by an earlier run", func(t *testing.T) {
		earlier, clock, dir := setup(t)
		earlier.RecordPoint("alice", 1460, 450, "WATCH_STREAK")
		clock.Set(t0 + hour)

		store := openStore(t, dir, clock)

		want := []StreamerInfo{{Name: "alice", Points: 1460, SessionGain: 0, Online: false, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("a points file with a torn last line", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "alice.points.jsonl", `{"t":1700000000000,"p":1460}`+"\n"+`{"t":17000`)

		store := openStore(t, dir, newFakeClock(t0+hour))

		want := []StreamerInfo{{Name: "alice", Points: 1460, Updated: t0}}
		if got := store.Streamers(); !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("a points file longer than the tail that is read", func(t *testing.T) {
		dir := t.TempDir()
		lines := strings.Repeat(`{"t":1700000000000,"p":1000,"d":10,"r":"WATCH"}`+"\n", 200)
		writeFile(t, dir, "alice.points.jsonl", lines+`{"t":1700000060000,"p":1460}`+"\n")

		store := openStore(t, dir, newFakeClock(t0+hour))

		want := []StreamerInfo{{Name: "alice", Points: 1460, Updated: t0 + minute}}
		if got := store.Streamers(); !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty points file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "alice.points.jsonl", "")

		store := openStore(t, dir, newFakeClock(t0))

		want := []StreamerInfo{{Name: "alice"}}
		if got := store.Streamers(); !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("a points file whose name is not a login", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "not a login.points.jsonl", `{"t":1700000000000,"p":1460}`+"\n")

		store := openStore(t, dir, newFakeClock(t0))

		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})

	t.Run("an events file with no points file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "alice.events.jsonl", `{"t":1700000000000,"type":"BET_WIN","msg":"Won a bet on alice."}`+"\n")

		store := openStore(t, dir, newFakeClock(t0))

		if got, want := store.Streamers(), []StreamerInfo{}; !reflect.DeepEqual(got, want) {
			t.Errorf("Streamers() = %+v, want %+v", got, want)
		}
	})
}

func TestCleanName(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"a plain login", "alice", "alice"},
		{"digits and an underscore", "alice_tv_42", "alice_tv_42"},
		{"capitals", "Alice", "alice"},
		{"spaces around it", "  alice\n", "alice"},
		{"an empty string", "", ""},
		{"only spaces", "   ", ""},
		{"63 characters", strings.Repeat("a", 63), strings.Repeat("a", 63)},
		{"64 characters", strings.Repeat("a", 64), strings.Repeat("a", 64)},
		{"65 characters", strings.Repeat("a", 65), ""},
		{"a path", "../evil", ""},
		{"a space inside", "not a login", ""},
		{"a hyphen", "alice-tv", ""},
		{"a dot", "alice.tv", ""},
		{"a letter outside ASCII", "ålice", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanName(tt.raw); got != tt.want {
				t.Errorf("cleanName(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestIsNameByte(t *testing.T) {
	tests := []struct {
		name string
		c    byte
		want bool
	}{
		{"a", 'a', true},
		{"z", 'z', true},
		{"0", '0', true},
		{"9", '9', true},
		{"an underscore", '_', true},
		{"the byte before a", '`', false},
		{"the byte after z", '{', false},
		{"the byte before 0", '/', false},
		{"the byte after 9", ':', false},
		{"a capital", 'A', false},
		{"a space", ' ', false},
		{"a byte outside ASCII", 0xC3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNameByte(tt.c); got != tt.want {
				t.Errorf("isNameByte(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}
