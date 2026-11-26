package dashboard

import "net/http"

func (s *Server) handleStreamers(w http.ResponseWriter, r *http.Request) {
	list := s.store.Streamers()
	writeJSON(w, r, map[string]interface{}{
		"now":       s.store.now().UnixMilli(),
		"totals":    totalsOf(list),
		"streamers": list,
	})
}

func (s *Server) handleStreamer(w http.ResponseWriter, r *http.Request) {
	name := cleanName(r.PathValue("name"))
	if name == "" || !s.store.Known(name) {
		http.Error(w, "unknown streamer", http.StatusNotFound)
		return
	}
	span, ok := windowFor(r.URL.Query().Get("range"), s.store.now().UnixMilli())
	if !ok {
		http.Error(w, "range must be one of 24h, 7d, 30d, all", http.StatusBadRequest)
		return
	}
	points, previous, err := s.store.Points(name, span.since)
	if err != nil {
		http.Error(w, "failed to read points", http.StatusInternalServerError)
		return
	}
	stored, err := s.store.Events(name, span.since)
	if err != nil {
		http.Error(w, "failed to read events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, r, streamerHistory(name, span, points, previous, withMarkers(stored, points)))
}
