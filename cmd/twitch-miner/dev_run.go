package main

import "strings"

// isGoRunExecutable reports whether the executable was built by `go run` or
// `go test`, which put it in a go-build folder or under the temp directory.
func isGoRunExecutable(path, tempDir string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "go-build") {
		return true
	}
	temp := strings.ToLower(tempDir)
	return strings.HasPrefix(lower, temp)
}
