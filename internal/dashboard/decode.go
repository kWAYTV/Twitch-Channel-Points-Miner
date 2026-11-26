package dashboard

import (
	"bytes"
	"encoding/json"
)

// parsePoint reads one line of a points file. Files are append-only, so a torn
// last line simply fails to parse.
func parsePoint(line []byte) (Point, bool) {
	var p Point
	if json.Unmarshal(line, &p) != nil || p.T == 0 {
		return Point{}, false
	}
	return p, true
}

// parseEvent reads one line of an events file.
func parseEvent(line []byte) (Event, bool) {
	var e Event
	if json.Unmarshal(line, &e) != nil || e.Type == "" {
		return Event{}, false
	}
	return e, true
}

// lastPoint returns the last point that parses among the lines of data.
func lastPoint(data []byte) (Point, bool) {
	lines := bytes.Split(data, []byte{'\n'})
	for i := len(lines) - 1; i >= 0; i-- {
		if p, ok := parsePoint(lines[i]); ok {
			return p, true
		}
	}
	return Point{}, false
}
