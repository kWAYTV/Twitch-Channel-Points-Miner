package dashboard

import (
	"io"
	"os"
	"path/filepath"
)

// Points loads a streamer's samples with T >= since. prev is the last sample
// before since (nil if none), so a chart can start from a known balance.
func (s *Store) Points(name string, since int64) (points []Point, prev *Point, err error) {
	err = readLines(filepath.Join(s.dir, name+pointsSuffix), func(line []byte) {
		p, ok := parsePoint(line)
		if !ok {
			return
		}
		if p.T < since {
			prev = &p
			return
		}
		points = append(points, p)
	})
	return points, prev, err
}

func readLastPoint(path string) (Point, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Point{}, false
	}
	defer f.Close()
	const tail = 4096
	if info, err := f.Stat(); err == nil && info.Size() > tail {
		_, _ = f.Seek(-tail, io.SeekEnd)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return Point{}, false
	}
	return lastPoint(data)
}
