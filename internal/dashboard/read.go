package dashboard

import (
	"bufio"
	"os"
	"sort"
)

// Known reports whether name is a recorded streamer.
func (s *Store) Known(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.streamers[name]
	return ok
}

// Streamers returns the live summary of every known streamer, sorted by name.
func (s *Store) Streamers() []StreamerInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]StreamerInfo, 0, len(s.streamers))
	for name, st := range s.streamers {
		info := StreamerInfo{Name: name, Points: st.last.P, Online: st.online, Updated: st.last.T}
		if st.hasSession {
			info.SessionGain = st.last.P - st.sessionP
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// readLines calls fn for every line of a file; a missing file is not an error.
// Files are append-only, so a torn last line simply fails to parse and is skipped.
func readLines(path string, fn func(line []byte)) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := sc.Bytes(); len(line) > 0 {
			fn(line)
		}
	}
	return sc.Err()
}
