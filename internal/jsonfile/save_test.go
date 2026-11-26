package jsonfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func saved(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back %s: %v", path, err)
	}
	return string(raw)
}

func TestSaveJSON(t *testing.T) {
	t.Run("an object with a string and a number", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, map[string]interface{}{"foo": "bar", "num": 1})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "{\n  \"foo\": \"bar\",\n  \"num\": 1\n}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("a struct with a nested list", func(t *testing.T) {
		type streak struct {
			Channel string `json:"channel"`
			Days    []int  `json:"days"`
		}
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, streak{Channel: "alice", Days: []int{1, 2}})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "{\n  \"channel\": \"alice\",\n  \"days\": [\n    1,\n    2\n  ]\n}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("an empty object", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, map[string]interface{}{})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "{}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("nil", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, nil)

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "null"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("text outside ASCII and HTML characters", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, map[string]interface{}{"title": "añadido <🎁>"})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "{\n  \"title\": \"añadido \\u003c🎁\\u003e\"\n}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("over a longer file that is already there", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")
		if err := os.WriteFile(path, []byte(`{"old": "a much longer value than the new one"}`), 0o644); err != nil {
			t.Fatalf("write the old file: %v", err)
		}

		err := SaveJSON(path, map[string]interface{}{"n": 1})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}
		if got, want := saved(t, path), "{\n  \"n\": 1\n}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("a value JSON can't hold", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")

		err := SaveJSON(path, map[string]interface{}{"ch": make(chan int)})

		if got, want := errorText(err), "json: unsupported type: chan int"; got != want {
			t.Errorf("SaveJSON error = %q, want %q", got, want)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, fs.ErrNotExist) {
			t.Errorf("a file was written: stat error = %v, want not exist", statErr)
		}
	})

	t.Run("a folder that does not exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "out.json")

		err := SaveJSON(path, map[string]interface{}{"n": 1})

		if err != nil {
			t.Fatalf("SaveJSON error = %v, want the folder created", err)
		}
		if got, want := saved(t, path), "{\n  \"n\": 1\n}"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("a value JSON can't hold, over a file already there", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.json")
		if err := os.WriteFile(path, []byte("kept"), 0o644); err != nil {
			t.Fatal(err)
		}

		_ = SaveJSON(path, map[string]interface{}{"ch": make(chan int)})

		if got, want := saved(t, path), "kept"; got != want {
			t.Errorf("file = %q, want it left as %q", got, want)
		}
	})

	t.Run("nothing else left in the folder", func(t *testing.T) {
		dir := t.TempDir()

		if err := SaveJSON(filepath.Join(dir, "out.json"), map[string]interface{}{"n": 1}); err != nil {
			t.Fatalf("SaveJSON error = %v, want nil", err)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 || entries[0].Name() != "out.json" {
			t.Errorf("folder holds %v, want out.json alone", entries)
		}
	})
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestWrite(t *testing.T) {
	t.Run("content as given", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "notes.txt")

		err := Write(path, []byte("as written"), 0o600)

		if err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}
		if got, want := saved(t, path), "as written"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("over a file already there", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "notes.txt")
		if err := os.WriteFile(path, []byte("a longer text that was there before"), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := Write(path, []byte("new"), 0o644); err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}

		if got, want := saved(t, path), "new"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
	})

	t.Run("in folders that do not exist", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a", "b", "notes.txt")

		if err := Write(path, []byte("deep"), 0o644); err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}

		if got, want := saved(t, path), "deep"; got != want {
			t.Errorf("file = %q, want %q", got, want)
		}
		entries, err := os.ReadDir(filepath.Join(dir, "a", "b"))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Errorf("folder holds %v, want notes.txt alone", entries)
		}
	})

	t.Run("where a folder cannot be made", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "blocker")
		if err := os.WriteFile(blocker, []byte("a file"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := Write(filepath.Join(blocker, "notes.txt"), []byte("x"), 0o644)

		if err == nil {
			t.Errorf("Write error = nil, want an error: a file stands where the folder should be")
		}
	})
}
