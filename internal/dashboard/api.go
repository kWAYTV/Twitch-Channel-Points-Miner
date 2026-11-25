package dashboard

import (
	"net/http"
	"time"
)

var ranges = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"all": 0,
}

type totals struct {
	Points      int `json:"points"`
	SessionGain int `json:"session_gain"`
	Online      int `json:"online"`
	Count       int `json:"count"`
}

func (s *Server) handleStreamers(w http.ResponseWriter, r *http.Request) {
	list := s.store.Streamers()
	t := totals{Count: len(list)}
	for _, st := range list {
		t.Points += st.Points
		t.SessionGain += st.SessionGain
		if st.Online {
			t.Online++
		}
	}
	writeJSON(w, r, map[string]interface{}{
		"now":       s.store.now().UnixMilli(),
		"totals":    t,
		"streamers": list,
	})
}

// window is the stretch of history a request asks for.
type window struct {
	name  string
	since int64
	now   int64
}

// windowFor resolves a range name such as "7d". An empty name means the last day.
func (s *Server) windowFor(rangeName string) (window, bool) {
	if rangeName == "" {
		rangeName = "24h"
	}
	span, ok := ranges[rangeName]
	if !ok {
		return window{}, false
	}
	w := window{name: rangeName, now: s.store.now().UnixMilli()}
	if span > 0 {
		w.since = w.now - span.Milliseconds()
	}
	return w, true
}

func (s *Server) handleStreamer(w http.ResponseWriter, r *http.Request) {
	name := cleanName(r.PathValue("name"))
	if name == "" || !s.store.Known(name) {
		http.Error(w, "unknown streamer", http.StatusNotFound)
		return
	}
	span, ok := s.windowFor(r.URL.Query().Get("range"))
	if !ok {
		http.Error(w, "range must be one of 24h, 7d, 30d, all", http.StatusBadRequest)
		return
	}
	points, previous, err := s.store.Points(name, span.since)
	if err != nil {
		http.Error(w, "failed to read points", http.StatusInternalServerError)
		return
	}
	events, err := s.eventsWithMarkers(name, span.since, points)
	if err != nil {
		http.Error(w, "failed to read events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, r, streamerHistory(name, span, points, previous, events))
}
