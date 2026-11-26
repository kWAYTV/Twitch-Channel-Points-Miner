package miner

import (
	"reflect"
	"testing"
)

func TestDefaultWatchPriorities(t *testing.T) {
	want := []watchPriority{watchPriorityStreak, watchPriorityDrops, watchPriorityOrder}
	if got := defaultWatchPriorities(); !reflect.DeepEqual(got, want) {
		t.Errorf("defaultWatchPriorities() = %v, want %v", got, want)
	}
}

func TestParseWatchPriorities(t *testing.T) {
	defaults := []watchPriority{watchPriorityStreak, watchPriorityDrops, watchPriorityOrder}
	tests := []struct {
		name  string
		names []string
		want  []watchPriority
	}{
		{"no names", nil, defaults},
		{"an empty list", []string{}, defaults},
		{"ORDER", []string{"ORDER"}, []watchPriority{watchPriorityOrder}},
		{"STREAK", []string{"STREAK"}, []watchPriority{watchPriorityStreak}},
		{"DROPS", []string{"DROPS"}, []watchPriority{watchPriorityDrops}},
		{"SUBSCRIBED", []string{"SUBSCRIBED"}, []watchPriority{watchPrioritySubscribed}},
		{"SUBS", []string{"SUBS"}, []watchPriority{watchPrioritySubscribed}},
		{"MULTIPLIER", []string{"MULTIPLIER"}, []watchPriority{watchPrioritySubscribed}},
		{"POINTS_ASC", []string{"POINTS_ASC"}, []watchPriority{watchPriorityPointsAscending}},
		{"POINTS_ASCENDING", []string{"POINTS_ASCENDING"}, []watchPriority{watchPriorityPointsAscending}},
		{"POINTS_DESC", []string{"POINTS_DESC"}, []watchPriority{watchPriorityPointsDescending}},
		{"POINTS_DESCENDING", []string{"POINTS_DESCENDING"}, []watchPriority{watchPriorityPointsDescending}},
		{"lower case with spaces around", []string{"  drops "}, []watchPriority{watchPriorityDrops}},
		{"several names keep their order", []string{"POINTS_DESC", "ORDER", "STREAK"}, []watchPriority{watchPriorityPointsDescending, watchPriorityOrder, watchPriorityStreak}},
		{"a repeated name", []string{"DROPS", "ORDER", "DROPS"}, []watchPriority{watchPriorityDrops, watchPriorityOrder}},
		{"two names for one priority", []string{"SUBS", "MULTIPLIER"}, []watchPriority{watchPrioritySubscribed}},
		{"an unknown name among known ones", []string{"DROPS", "ignored", "ORDER"}, []watchPriority{watchPriorityDrops, watchPriorityOrder}},
		{"only unknown names", []string{"foo", ""}, defaults},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseWatchPriorities(tt.names); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseWatchPriorities(%q) = %v, want %v", tt.names, got, tt.want)
			}
		})
	}
}
