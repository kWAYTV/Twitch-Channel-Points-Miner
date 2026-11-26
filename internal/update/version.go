package update

import (
	"strconv"
	"strings"
)

func normalizeVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "v")
	if idx := strings.IndexAny(raw, " -"); idx > 0 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, "+"); idx > 0 {
		raw = raw[:idx]
	}
	return raw
}

func compareVersions(a, b string) int {
	parse := func(v string) []int {
		parts := strings.Split(v, ".")
		out := make([]int, len(parts))
		for i, p := range parts {
			val, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				val = 0
			}
			out[i] = val
		}
		return out
	}

	as := parse(a)
	bs := parse(b)
	max := len(as)
	if len(bs) > max {
		max = len(bs)
	}
	for i := 0; i < max; i++ {
		av, bv := 0, 0
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}
