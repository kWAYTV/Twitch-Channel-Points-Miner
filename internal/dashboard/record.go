package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RecordPoint appends a balance sample for a streamer.
func (s *Store) RecordPoint(streamer string, points, delta int, reason string) {
	name := cleanName(streamer)
	if name == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordLocked(name, points, delta, reason)
}

// RecordSnapshot records the current balance (skipped while unchanged) and online state.
func (s *Store) RecordSnapshot(streamer string, points int, online bool) {
	name := cleanName(streamer)
	if name == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordLocked(name, points, 0, "").online = online
}

func (s *Store) recordLocked(name string, points, delta int, reason string) *streamerState {
	st := s.streamers[name]
	if st == nil {
		st = &streamerState{}
		s.streamers[name] = st
	}
	if !st.hasSession {
		st.sessionP, st.hasSession = points-delta, true
	}
	p, ok := nextSample(st.last, st.hasLast, s.now().UnixMilli(), points, delta, reason)
	if !ok {
		return st
	}
	st.last, st.hasLast = p, true
	s.appendLine(name+pointsSuffix, p)
	return st
}

func (s *Store) appendLine(file string, v interface{}) {
	line, err := json.Marshal(v)
	if err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(s.dir, file), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
	_ = f.Close()
}
