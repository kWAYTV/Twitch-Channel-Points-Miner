package logging

import (
	"bytes"
	"errors"
	"testing"
)

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestDualWriter_Write(t *testing.T) {
	type written struct {
		n   int
		err string
	}
	errText := func(err error) string {
		if err == nil {
			return ""
		}
		return err.Error()
	}

	t.Run("a console and a file", func(t *testing.T) {
		console, file := &bytes.Buffer{}, &bytes.Buffer{}

		n, err := dualWriter{console: console, file: file}.Write([]byte("\x1b[38;5;46mAlice\x1b[0m is live\n"))

		if got, want := (written{n, errText(err)}), (written{28, ""}); got != want {
			t.Errorf("Write = %+v, want %+v", got, want)
		}
		if got, want := console.String(), "\x1b[38;5;46mAlice\x1b[0m is live\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
		if got, want := file.String(), "Alice is live\n"; got != want {
			t.Errorf("file holds %q, want %q", got, want)
		}
	})

	t.Run("only a console", func(t *testing.T) {
		console := &bytes.Buffer{}

		n, err := dualWriter{console: console}.Write([]byte("hello\n"))

		if got, want := (written{n, errText(err)}), (written{6, ""}); got != want {
			t.Errorf("Write = %+v, want %+v", got, want)
		}
		if got, want := console.String(), "hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})

	t.Run("only a file", func(t *testing.T) {
		file := &bytes.Buffer{}

		n, err := dualWriter{file: file}.Write([]byte("hello\n"))

		if got, want := (written{n, errText(err)}), (written{6, ""}); got != want {
			t.Errorf("Write = %+v, want %+v", got, want)
		}
		if got, want := file.String(), "hello\n"; got != want {
			t.Errorf("file holds %q, want %q", got, want)
		}
	})

	t.Run("a console that fails", func(t *testing.T) {
		file := &bytes.Buffer{}

		n, err := dualWriter{console: brokenWriter{}, file: file}.Write([]byte("hello\n"))

		if got, want := (written{n, errText(err)}), (written{0, "disk full"}); got != want {
			t.Errorf("Write = %+v, want %+v", got, want)
		}
		if got := file.String(); got != "" {
			t.Errorf("file holds %q, want nothing", got)
		}
	})

	t.Run("a file that fails", func(t *testing.T) {
		console := &bytes.Buffer{}

		n, err := dualWriter{console: console, file: brokenWriter{}}.Write([]byte("hello\n"))

		if got, want := (written{n, errText(err)}), (written{0, "disk full"}); got != want {
			t.Errorf("Write = %+v, want %+v", got, want)
		}
		if got, want := console.String(), "hello\n"; got != want {
			t.Errorf("console shows %q, want %q", got, want)
		}
	})
}

func TestLogFileName(t *testing.T) {
	for _, tc := range []struct {
		name      string
		username  string
		anonymize bool
		want      string
	}{
		{"a user name", "tester", false, "tester.log"},
		{"a user name with spaces around it", "  tester ", false, "tester.log"},
		{"a user name with characters a file can't have", `te/st:er`, false, "te_st_er.log"},
		{"a user name outside ASCII", "tëster", false, "tëster.log"},
		{"no user name", "", false, "miner.log"},
		{"only spaces for a user name", "   ", false, "miner.log"},
		{"a user name and anonymous logs", "tester", true, "miner.log"},
		{"no user name and anonymous logs", "", true, "miner.log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := logFileName(tc.username, tc.anonymize); got != tc.want {
				t.Errorf("logFileName(%q, %v) = %q, want %q", tc.username, tc.anonymize, got, tc.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"a plain name", "tester_01", "tester_01"},
		{"a slash", "a/b", "a_b"},
		{"a backslash", `a\b`, "a_b"},
		{"a colon", "a:b", "a_b"},
		{"an asterisk", "a*b", "a_b"},
		{"a question mark", "a?b", "a_b"},
		{"a double quote", `a"b`, "a_b"},
		{"a less-than sign", "a<b", "a_b"},
		{"a greater-than sign", "a>b", "a_b"},
		{"a pipe", "a|b", "a_b"},
		{"every character a file can't have", `/\:*?"<>|`, "_________"},
		{"letters outside ASCII", "tëster 日本", "tëster 日本"},
		{"a dot and a space", "a. b", "a. b"},
		{"an empty name", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitizeFilename(tc.in); got != tc.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
