package miner

import (
	"strings"
)

// normalizeNames trims and lower-cases each name, and drops blanks and repeats.
func normalizeNames(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, raw := range values {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

func nameSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

// positions maps each name to its place in the list.
func positions(names []string) map[string]int {
	index := make(map[string]int, len(names))
	for i, name := range names {
		index[name] = i
	}
	return index
}
