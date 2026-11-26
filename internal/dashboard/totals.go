package dashboard

type totals struct {
	Points      int `json:"points"`
	SessionGain int `json:"session_gain"`
	Online      int `json:"online"`
	Count       int `json:"count"`
}

// totalsOf adds up the summaries of every streamer.
func totalsOf(list []StreamerInfo) totals {
	t := totals{Count: len(list)}
	for _, st := range list {
		t.Points += st.Points
		t.SessionGain += st.SessionGain
		if st.Online {
			t.Online++
		}
	}
	return t
}
