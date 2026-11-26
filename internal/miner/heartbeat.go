package miner

import "github.com/kWAYTV/twitch-points-miner/internal/heartbeat"

// beat records that the miner is still doing its rounds, for a health check
// outside the process. A beat that cannot be written is not worth stopping
// for: the check then reports the miner as stuck, which is the safe mistake.
func (m *Miner) beat() {
	if m.heartbeatPath == "" {
		return
	}
	_ = heartbeat.Write(m.heartbeatPath, m.now())
}
