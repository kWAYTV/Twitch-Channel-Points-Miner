package miner

import (
	"reflect"
	"testing"
)

func TestNormalizeNames(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []string
	}{
		{"no names", nil, []string{}},
		{"an empty list", []string{}, []string{}},
		{"lower-case names", []string{"valorant", "wow"}, []string{"valorant", "wow"}},
		{"mixed case", []string{"Valorant"}, []string{"valorant"}},
		{"spaces around a name", []string{"  valorant  "}, []string{"valorant"}},
		{"spaces inside a name", []string{"Just Chatting"}, []string{"just chatting"}},
		{"a blank name", []string{"valorant", "   ", "", "wow"}, []string{"valorant", "wow"}},
		{"a repeat in another case", []string{"valorant", "wow", " VALORANT "}, []string{"valorant", "wow"}},
		{"characters outside ASCII", []string{"Pokémon ÉCARLATE"}, []string{"pokémon écarlate"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeNames(tt.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeNames(%q) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

func TestNameSet(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  map[string]struct{}
	}{
		{"no names", nil, map[string]struct{}{}},
		{"two names", []string{"foo", "bar"}, map[string]struct{}{"foo": {}, "bar": {}}},
		{"a repeated name", []string{"foo", "foo"}, map[string]struct{}{"foo": {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nameSet(tt.names); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nameSet(%q) = %v, want %v", tt.names, got, tt.want)
			}
		})
	}
}

func TestPositions(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  map[string]int
	}{
		{"no names", nil, map[string]int{}},
		{"three names", []string{"valorant", "wow", "chess"}, map[string]int{"valorant": 0, "wow": 1, "chess": 2}},
		{"a repeated name keeps its last place", []string{"valorant", "wow", "valorant"}, map[string]int{"valorant": 2, "wow": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := positions(tt.names); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("positions(%q) = %v, want %v", tt.names, got, tt.want)
			}
		})
	}
}
