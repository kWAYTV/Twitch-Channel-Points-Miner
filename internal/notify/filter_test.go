package notify

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

func TestWantsEvent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		selected map[events.Event]struct{}
		event    events.Event
		want     bool
	}{
		{"no selection and an ordinary event", nil, events.EventStreamerOnline, true},
		{"no selection and a watch tick", nil, events.EventGainForWatch, false},
		{"no selection and drop progress", nil, events.EventDropStatus, false},
		{"no selection and a bet the filter skipped", nil, events.EventBetFilters, false},
		{"an empty selection and an ordinary event", map[events.Event]struct{}{}, events.EventStreamerOnline, true},
		{"an empty selection and a watch tick", map[events.Event]struct{}{}, events.EventGainForWatch, false},
		{"a star and a watch tick", map[events.Event]struct{}{"*": {}}, events.EventGainForWatch, true},
		{"ALL and a watch tick", map[events.Event]struct{}{"ALL": {}}, events.EventGainForWatch, true},
		{"a named watch tick", map[events.Event]struct{}{events.EventGainForWatch: {}}, events.EventGainForWatch, true},
		{"a named event", map[events.Event]struct{}{events.EventBetWin: {}}, events.EventBetWin, true},
		{"an event that is not named", map[events.Event]struct{}{events.EventBetWin: {}}, events.EventStreamerOnline, false},
		{"no selection and no event", nil, "", false},
		{"a star and no event", map[events.Event]struct{}{"*": {}}, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := wantsEvent(tc.selected, tc.event); got != tc.want {
				t.Errorf("wantsEvent(%v, %q) = %v, want %v", tc.selected, tc.event, got, tc.want)
			}
		})
	}
}

func TestSelectedEvents(t *testing.T) {
	for _, tc := range []struct {
		name  string
		names []string
		want  map[events.Event]struct{}
	}{
		{"two names", []string{"STREAMER_ONLINE", "BET_LOSE"}, map[events.Event]struct{}{events.EventStreamerOnline: {}, events.EventBetLose: {}}},
		{"a name in lower case with spaces around it", []string{" streamer_online "}, map[events.Event]struct{}{events.EventStreamerOnline: {}}},
		{"the same name twice", []string{"BET_LOSE", "bet_lose"}, map[events.Event]struct{}{events.EventBetLose: {}}},
		{"a blank name among others", []string{"", "  ", "BET_LOSE"}, map[events.Event]struct{}{events.EventBetLose: {}}},
		{"a star", []string{"*"}, map[events.Event]struct{}{"*": {}}},
		{"a name that is no event", []string{"nonsense"}, map[events.Event]struct{}{"NONSENSE": {}}},
		{"an empty list", []string{}, map[events.Event]struct{}{}},
		{"no list", nil, map[events.Event]struct{}{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := selectedEvents(tc.names); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("selectedEvents(%q) = %v, want %v", tc.names, got, tc.want)
			}
		})
	}
}
