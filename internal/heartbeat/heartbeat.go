// Package heartbeat is how a running miner shows it is still working: it
// writes the time to a file each watch round, and anything outside the
// process, such as a container health check, reads how old that time is.
package heartbeat

import (
	"os"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/jsonfile"
)

// File is the heartbeat's name inside the data folder.
const File = "heartbeat"

// Write records a beat at the given time.
func Write(path string, at time.Time) error {
	return jsonfile.Write(path, []byte(at.UTC().Format(time.RFC3339Nano)), 0o644)
}

// Age is how long before now the last beat was. The error is one that
// os.IsNotExist recognises when there has been no beat yet.
func Age(path string, now time.Time) (time.Duration, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	at, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(raw)))
	if err != nil {
		return 0, err
	}
	return now.Sub(at), nil
}
