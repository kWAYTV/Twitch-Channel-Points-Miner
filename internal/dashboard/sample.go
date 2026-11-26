package dashboard

// nextSample decides the point to store for a balance reading. A reading with
// no delta and no reason is a snapshot: its delta is worked out from the last
// sample, and it is skipped (false) while the balance is unchanged and the last
// sample is younger than unchangedInterval.
func nextSample(last Point, hasLast bool, now int64, points, delta int, reason string) (Point, bool) {
	if delta == 0 && reason == "" && hasLast {
		if last.P == points && now-last.T < unchangedInterval.Milliseconds() {
			return Point{}, false
		}
		delta = points - last.P
	}
	return Point{T: now, P: points, D: delta, R: reason}, true
}
