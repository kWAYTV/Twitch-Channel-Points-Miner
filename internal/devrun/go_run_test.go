package devrun

import "testing"

func TestIsGoRunExecutable(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		tempDir string
		want    bool
	}{
		{name: "an installed program", path: "/opt/miner/twitch-miner", tempDir: "/tmp", want: false},
		{name: "in a go-build folder", path: "/home/me/.cache/go-build/ab/twitch-miner", tempDir: "/tmp", want: true},
		{name: "in a go-build folder in upper case", path: `D:\GO-BUILD123\b001\exe\miner.exe`, tempDir: `C:\Temp`, want: true},
		{name: "under the temp folder", path: "/tmp/run/twitch-miner", tempDir: "/tmp", want: true},
		{name: "under the temp folder in another case", path: `c:\users\me\appdata\local\temp\miner.exe`, tempDir: `C:\Users\Me\AppData\Local\Temp`, want: true},
		{name: "under the temp folder, the path in another case", path: `C:\Users\Me\AppData\Local\Temp\miner.exe`, tempDir: `c:\users\me\appdata\local\temp`, want: true},
		{name: "no path", path: "", tempDir: "/tmp", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGoRunExecutable(tt.path, tt.tempDir); got != tt.want {
				t.Errorf("IsGoRunExecutable(%q, %q) = %v, want %v", tt.path, tt.tempDir, got, tt.want)
			}
		})
	}

	t.Run("a folder named like the temp folder with more letters", func(t *testing.T) {
		if got := IsGoRunExecutable("/tmpfiles/twitch-miner", "/tmp"); got != false {
			t.Errorf(`IsGoRunExecutable("/tmpfiles/twitch-miner", "/tmp") = %v, want false`, got)
		}
	})
}

func TestIsGoRunExecutable_updaterExamples(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		tempDir string
		want    bool
	}{
		{"an installed binary", "/opt/miner/miner", "/tmp", false},
		{"a binary in a go-build folder", "/home/me/.cache/go-build/ab/miner", "/tmp", true},
		{"a binary in a GO-BUILD folder", `D:\Cache\GO-BUILD123\miner.exe`, `C:\Temp`, true},
		{"a binary under the temp folder", "/tmp/run/miner", "/tmp", true},
		{"a binary under the temp folder in other capitals", `c:\temp\run\miner.exe`, `C:\Temp`, true},
		{"a binary beside the temp folder", "/var/tmp/miner", "/tmp", false},
		{"no temp folder", "/opt/miner/miner", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGoRunExecutable(tt.path, tt.tempDir); got != tt.want {
				t.Errorf("IsGoRunExecutable(%q, %q) = %t, want %t", tt.path, tt.tempDir, got, tt.want)
			}
		})
	}
}
