// Package devrun tells a build made by "go run" or "go test" from an installed program.
package devrun

import "strings"

// IsGoRunExecutable reports whether the executable was built by `go run` or
// `go test`, which put it in a go-build folder or under the temp directory.
func IsGoRunExecutable(path, tempDir string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "go-build") {
		return true
	}
	temp := strings.TrimRight(strings.ToLower(tempDir), `/\`)
	rest, under := strings.CutPrefix(lower, temp)
	return under && (rest == "" || rest[0] == '/' || rest[0] == '\\')
}
