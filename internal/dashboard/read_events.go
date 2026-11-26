package dashboard

import "path/filepath"

// Events loads a streamer's events with T >= since, including unattributed ones.
func (s *Store) Events(name string, since int64) ([]Event, error) {
	var events []Event
	for _, n := range []string{name, globalName} {
		err := readLines(filepath.Join(s.dir, n+eventsSuffix), func(line []byte) {
			e, ok := parseEvent(line)
			if !ok || e.T < since {
				return
			}
			events = append(events, e)
		})
		if err != nil {
			return nil, err
		}
	}
	return events, nil
}
