package notify

import (
	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// quietEvents are high-frequency or low-signal. An empty events list sends
// every other event; name these (or "*") to opt in.
var quietEvents = map[events.Event]struct{}{
	events.EventGainForWatch: {},
	events.EventDropStatus:   {},
	events.EventBetFilters:   {},
}

// wantsEvent reports whether a notifier with the given selection should deliver
// event. An empty selection means every event except quietEvents. "*" or "ALL"
// means every event, including the quiet ones.
func wantsEvent(selected map[events.Event]struct{}, event events.Event) bool {
	if event == "" {
		return false
	}
	if len(selected) == 0 {
		_, quiet := quietEvents[event]
		return !quiet
	}
	if _, ok := selected["*"]; ok {
		return true
	}
	if _, ok := selected[events.Event("ALL")]; ok {
		return true
	}
	_, ok := selected[event]
	return ok
}

// selectedEvents reads a settings events list into a selection. Names are
// matched without regard to case or surrounding space, and blank names are dropped.
func selectedEvents(names []string) map[events.Event]struct{} {
	selected := make(map[events.Event]struct{})
	for _, raw := range names {
		if event := events.NormalizeEventName(raw); event != "" {
			selected[event] = struct{}{}
		}
	}
	return selected
}
