package dashboard

import "net/http"

// latestBetsShown is how many of the newest bets a response lists.
const latestBetsShown = 50

// handleBets answers with how each strategy has done over every recorded
// bet, and the newest bets themselves.
func (s *Server) handleBets(w http.ResponseWriter, r *http.Request) {
	bets, err := s.store.Bets(0)
	if err != nil {
		http.Error(w, "failed to read bets", http.StatusInternalServerError)
		return
	}
	writeJSON(w, r, map[string]interface{}{
		"strategies": strategyTotals(bets),
		"bets":       latestBets(bets, latestBetsShown),
	})
}
