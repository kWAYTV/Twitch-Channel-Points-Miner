package update

// progressWriter logs a download's progress about every 5% of total bytes.
type progressWriter struct {
	total   int64
	written int64
	step    int64
	// logged is how many bytes the last progress line reported.
	logged int64
	logf   func(format string, args ...interface{})
}

func newProgressWriter(total int64, logf func(format string, args ...interface{})) *progressWriter {
	step := total / 20 // ~5% steps
	if step == 0 {
		step = total
	}
	return &progressWriter{total: total, step: step, logf: logf}
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.written += int64(n)
	if p.total > 0 && p.written >= p.step {
		p.logProgress()
		for p.written >= p.step {
			p.step += p.total / 20
			if p.total/20 == 0 {
				p.step = p.written + 1
			}
		}
	}
	return n, nil
}

func (p *progressWriter) Done() {
	if p.total > 0 && p.written < p.total {
		p.written = p.total
	}
	// The last write may already have reported the whole download.
	if p.logged < p.written {
		p.logProgress()
	}
}

func (p *progressWriter) logProgress() {
	if p.total == 0 {
		return
	}
	percent := float64(p.written) * 100 / float64(p.total)
	if percent > 100 {
		percent = 100
	}
	p.logged = p.written
	p.logf("auto-update: download %.0f%% (%.1f/%.1f MB)", percent, bytesToMB(p.written), bytesToMB(p.total))
}

func bytesToMB(b int64) float64 {
	return float64(b) / (1024 * 1024)
}
