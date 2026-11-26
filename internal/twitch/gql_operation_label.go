package twitch

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// operationLabel names a GQL payload for the debug log: one operation, or a
// batch of up to three names and a count of the rest.
func operationLabel(payload interface{}, includeNote bool) string {
	names := operationNames(payload)
	if len(names) == 0 {
		return "unknown"
	}
	if len(names) == 1 {
		name := names[0]
		if includeNote {
			if note := operationNote(name); note != "" {
				return fmt.Sprintf("%s (%s)", name, note)
			}
		}
		return name
	}
	display := names
	if len(display) > 3 {
		display = display[:3]
	}
	label := strings.Join(display, ", ")
	if len(names) > len(display) {
		label = fmt.Sprintf("%s +%d", label, len(names)-len(display))
	}
	return fmt.Sprintf("batch[%s]", label)
}

func operationNames(payload interface{}) []string {
	var names []string
	seen := map[string]struct{}{}
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	switch p := payload.(type) {
	case twitchapi.GQLPersistedOperation:
		add(p.OperationName)
		return names
	case *twitchapi.GQLPersistedOperation:
		if p != nil {
			add(p.OperationName)
		}
		return names
	case []twitchapi.GQLPersistedOperation:
		for i := range p {
			add(p[i].OperationName)
		}
		return names
	case []*twitchapi.GQLPersistedOperation:
		for _, op := range p {
			if op != nil {
				add(op.OperationName)
			}
		}
		return names
	case map[string]interface{}:
		if name, ok := p["operationName"].(string); ok {
			add(name)
		}
		return names
	case []interface{}:
		for _, raw := range p {
			if m, ok := raw.(map[string]interface{}); ok {
				if name, ok := m["operationName"].(string); ok {
					add(name)
				}
			}
		}
		return names
	}

	// fallback for other struct-like payloads
	val := reflect.ValueOf(payload)
	for val.IsValid() && val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return names
		}
		val = val.Elem()
	}
	if !val.IsValid() {
		return names
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName("OperationName")
		if field.IsValid() && field.Kind() == reflect.String {
			add(field.String())
		}
		return names
	}
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			for item.IsValid() && item.Kind() == reflect.Pointer {
				if item.IsNil() {
					break
				}
				item = item.Elem()
			}
			if !item.IsValid() || item.Kind() != reflect.Struct {
				continue
			}
			field := item.FieldByName("OperationName")
			if field.IsValid() && field.Kind() == reflect.String {
				add(field.String())
			}
		}
	}
	return names
}

func operationNote(name string) string {
	switch name {
	case "WithIsStreamLiveQuery":
		return "check live"
	case "VideoPlayerStreamInfoOverlayChannel":
		return "stream info"
	case "RewardList":
		return "reward list"
	case "PlaybackAccessToken":
		return "playback token"
	case "ChannelPointsContext":
		return "points/bonus"
	case "ClaimCommunityPoints":
		return "claim bonus"
	case "CommunityMomentCallout_Claim":
		return "claim moment"
	case "JoinRaid":
		return "join raid"
	case "DropsPage_ClaimDropRewards":
		return "claim drop"
	case "DropsHighlightService_AvailableDrops":
		return "available drops"
	case "Inventory":
		return "inventory"
	case "GetIDFromLogin":
		return "resolve id"
	case "ChannelFollows":
		return "followers"
	case "UserPointsContribution":
		return "community goal"
	case "ContributeCommunityPointsCommunityGoal":
		return "contribute goal"
	case "MakePrediction":
		return "make prediction"
	case "ClipsCards__User":
		return "clips"
	case "FilterableVideoTower_Videos":
		return "vods"
	default:
		return ""
	}
}
