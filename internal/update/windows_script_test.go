package update

import "testing"

func TestWindowsUpdaterScript(t *testing.T) {
	t.Run("a plain path and two arguments", func(t *testing.T) {
		got := windowsUpdaterScript(`C:\Miner\miner.exe`, `C:\Miner\miner-update-1`, `C:\Miner`, []string{"-config", "my config.json"})

		want := `@echo off
setlocal
set "TARGET=C:\Miner\miner.exe"
set "NEWFILE=C:\Miner\miner-update-1"
set "WORKDIR=C:\Miner"
cd /D "%WORKDIR%"
:wait
ping 127.0.0.1 -n 2 >nul 2>nul
:loop
move /Y "%NEWFILE%" "%TARGET%" >nul 2>nul
if errorlevel 1 (
  ping 127.0.0.1 -n 3 >nul 2>nul
  goto loop
)
start "" /b "%TARGET%" "-config" "my config.json"
start "" /b cmd /c "del /q ""%~f0"""
exit /b
`
		if got != want {
			t.Errorf("script:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("a path with batch characters and no arguments", func(t *testing.T) {
		got := windowsUpdaterScript(`C:\R&D\100%\miner.exe`, `C:\R&D\100%\miner-update-1`, `C:\R&D\100%`, nil)

		want := `@echo off
setlocal
set "TARGET=C:\R^&D\100%%\miner.exe"
set "NEWFILE=C:\R^&D\100%%\miner-update-1"
set "WORKDIR=C:\R^&D\100%%"
cd /D "%WORKDIR%"
:wait
ping 127.0.0.1 -n 2 >nul 2>nul
:loop
move /Y "%NEWFILE%" "%TARGET%" >nul 2>nul
if errorlevel 1 (
  ping 127.0.0.1 -n 3 >nul 2>nul
  goto loop
)
start "" /b "%TARGET%"
start "" /b cmd /c "del /q ""%~f0"""
exit /b
`
		if got != want {
			t.Errorf("script:\n%s\nwant:\n%s", got, want)
		}
	})
}

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no arguments", nil, ""},
		{"an empty list", []string{}, ""},
		{"one argument", []string{"-debug"}, ` "-debug"`},
		{"two arguments", []string{"-config", "run.json"}, ` "-config" "run.json"`},
		{"an argument with a space", []string{"my config.json"}, ` "my config.json"`},
		{"an empty argument", []string{""}, ` ""`},
		{"an argument outside ASCII", []string{"żółw.json"}, ` "żółw.json"`},
		{"a Windows path (today: backslashes doubled)", []string{`C:\cfg\run.json`}, ` "C:\\cfg\\run.json"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatArgs(tt.args); got != tt.want {
				t.Errorf("formatArgs(%q) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}

	t.Run("a Windows path (bug: should keep single backslashes)", func(t *testing.T) {
		t.Skip("bug: arguments are quoted as Go strings, so the relaunched miner gets C:\\\\cfg\\\\run.json")
		args := []string{`C:\cfg\run.json`}
		if got, want := formatArgs(args), ` "C:\cfg\run.json"`; got != want {
			t.Errorf("formatArgs(%q) = %q, want %q", args, got, want)
		}
	})
}

func TestEscapeForBatch(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"a plain path", `C:\Miner\miner.exe`, `C:\Miner\miner.exe`},
		{"a caret", `C:\a^b`, `C:\a^^b`},
		{"an ampersand", `C:\R&D`, `C:\R^&D`},
		{"a pipe", `C:\a|b`, `C:\a^|b`},
		{"a less-than sign", `C:\a<b`, `C:\a^<b`},
		{"a greater-than sign", `C:\a>b`, `C:\a^>b`},
		{"a percent sign", `C:\100%`, `C:\100%%`},
		{"a double quote", `C:\a"b`, `C:\a""b`},
		{"an escaped-looking pair", `C:\a^&b`, `C:\a^^^&b`},
		{"a path with a space", `C:\My Miner\miner.exe`, `C:\My Miner\miner.exe`},
		{"a path outside ASCII", `C:\Górnik\miner.exe`, `C:\Górnik\miner.exe`},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeForBatch(tt.path); got != tt.want {
				t.Errorf("escapeForBatch(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
