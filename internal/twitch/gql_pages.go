package twitch

import (
	"github.com/kWAYTV/twitch-points-miner/internal/jsonvalue"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// pagesEnd is how a walk through the pages of a listing ended.
type pagesEnd int

const (
	// pagesDone: the last page was read, or the visitor stopped the walk.
	pagesDone pagesEnd = iota
	// pagesMissing: an answer had no listing where one was expected.
	pagesMissing
	// pagesCutShort: a page with nothing on it said that another follows.
	pagesCutShort
)

// eachEdge walks a listing that Twitch serves in pages, calling visit for
// every edge until visit stops the walk or fails, or the pages run out.
// dataPath is where the listing sits in each answer.
//
// The next page is asked for with the cursor of the last edge that has one,
// so a page without a cursor ends the walk: asking again would return the
// same page forever.
func (t *Client) eachEdge(op twitchapi.GQLPersistedOperation, dataPath string, visit func(edge map[string]interface{}) (stop bool, err error)) (pagesEnd, error) {
	cursor := ""
	for {
		resp, err := t.PostGQL(op.With(map[string]interface{}{"cursor": cursor}))
		if err != nil {
			return pagesDone, err
		}
		data, _ := jsonvalue.At(resp, dataPath).(map[string]interface{})
		if data == nil {
			return pagesMissing, nil
		}
		edges, _ := data["edges"].([]interface{})
		cursor = ""
		for _, raw := range edges {
			edge, _ := raw.(map[string]interface{})
			if c, ok := edge["cursor"].(string); ok {
				cursor = c
			}
			if stop, err := visit(edge); err != nil || stop {
				return pagesDone, err
			}
		}
		pageInfo, _ := data["pageInfo"].(map[string]interface{})
		more, _ := pageInfo["hasNextPage"].(bool)
		switch {
		case !more:
			return pagesDone, nil
		case len(edges) == 0:
			return pagesCutShort, nil
		case cursor == "":
			return pagesDone, nil
		}
	}
}
