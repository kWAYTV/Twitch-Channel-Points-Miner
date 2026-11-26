package miner

import (
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/streak"
)

// pickStreamersToWatch chooses which online streamers get the watch slots this round.
func (m *Miner) pickStreamersToWatch(streamers []*domain.Streamer) []*domain.Streamer {
	now := m.now()
	candidates := m.watchCandidates(streamers, now)
	streakReady := streakReadyCandidates(candidates)
	slots := newWatchSlots(candidates)

	m.keepActiveStreaks(slots, streamers, now)

	// With a game priority list, a streak on a game outside it does not jump the queue.
	skipEarlyStreak := len(m.gamePriority) > 0 && !anyPriorityGame(streakReady)
	for _, priority := range m.watchPriorities {
		if slots.full() {
			break
		}
		if priority == watchPriorityStreak && skipEarlyStreak {
			continue
		}
		m.pickByPriority(slots, priority, streamers, now)
	}

	slots.ensureStreak(streakReady)
	if skipEarlyStreak {
		slots.preferPriorityGameFirst()
	}
	if !slots.full() {
		slots.pick(candidates, true, nil, "FALLBACK")
	}

	watchList := make([]*domain.Streamer, 0, len(slots.picked))
	for _, idx := range slots.picked {
		watchList = append(watchList, streamers[idx])
		m.syncActiveStreakWatch(streamers[idx], now)
	}
	m.logWatchQueueSlots(slots, streamers, now)
	return watchList
}

// keepActiveStreaks gives a slot to every streamer whose watch streak is already in
// progress, and forgets the ones that no longer qualify.
func (m *Miner) keepActiveStreaks(slots *watchSlots, streamers []*domain.Streamer, now time.Time) {
	m.activeStreakMu.Lock()
	defer m.activeStreakMu.Unlock()
	if len(m.activeStreakWatches) == 0 {
		return
	}
	activeStreaks := make([]watchCandidate, 0, len(m.activeStreakWatches))
	for _, c := range slots.candidates {
		s := streamers[c.idx]
		if s == nil {
			continue
		}
		key := streamerKey(s)
		watch, ok := m.activeStreakWatches[key]
		if !ok {
			continue
		}
		if !m.shouldKeepActiveStreak(s, now) || !streak.ActiveWatchMatches(s, watch) {
			delete(m.activeStreakWatches, key)
			continue
		}
		activeStreaks = append(activeStreaks, c)
	}
	slots.pick(activeStreaks, false, nil, "ACTIVE_STREAK")
}

// dropsCandidates lists the streamers worth watching for drops: the stream
// carries drops and drop claims are on for the streamer. With a game priority
// list, only streams of those games count, so a drop of a game nobody asked
// for never takes a slot. Neither does a game the account is not connected to,
// whose drops could not be received.
func (m *Miner) dropsCandidates(candidates []watchCandidate, streamers []*domain.Streamer) []watchCandidate {
	drops := make([]watchCandidate, 0, len(candidates))
	for _, c := range candidates {
		s := streamers[c.idx]
		if s == nil || s.Stream == nil || !s.Settings.ClaimDrops {
			continue
		}
		if !s.Stream.DropsTags && len(s.Stream.CampaignIDs) == 0 {
			continue
		}
		if len(m.gamePriority) > 0 && !c.priorityGame {
			continue
		}
		if m.twitch != nil && m.twitch.DropsUnavailableFor(s.Stream.GameName()) {
			continue
		}
		drops = append(drops, c)
	}
	return drops
}

// pickByPriority fills free slots with the candidates one watch priority asks for.
func (m *Miner) pickByPriority(slots *watchSlots, priority watchPriority, streamers []*domain.Streamer, now time.Time) {
	candidates := slots.candidates
	switch priority {
	case watchPriorityOrder:
		slots.pick(candidates, false, nil, "ORDER")
	case watchPriorityStreak:
		streaks := make([]watchCandidate, 0, len(candidates))
		for _, c := range candidates {
			if m.shouldPrioritizeStreak(streamers[c.idx], now) {
				streaks = append(streaks, c)
			}
		}
		slots.pick(streaks, true, nil, "STREAK")
	case watchPriorityDrops:
		drops := m.dropsCandidates(candidates, streamers)
		// Twitch counts drop progress on one channel at a time, so one slot
		// goes to drops and the other stays free to earn elsewhere.
		if len(drops) > 0 && !slots.holdsAnyOf(drops) {
			slots.add(sortWatchCandidates(drops, nil, true)[0], "DROPS")
		}
	case watchPrioritySubscribed:
		subscribed := make([]watchCandidate, 0, len(candidates))
		for _, c := range candidates {
			s := streamers[c.idx]
			if s == nil {
				continue
			}
			if s.HasActiveMultipliers() {
				subscribed = append(subscribed, c)
			}
		}
		slots.pick(subscribed, true, func(a, b watchCandidate) bool {
			return streamers[a.idx].TotalMultiplier() > streamers[b.idx].TotalMultiplier()
		}, "SUBSCRIBED")
	case watchPriorityPointsAscending:
		slots.pick(candidates, true, func(a, b watchCandidate) bool {
			return streamers[a.idx].ChannelPoints < streamers[b.idx].ChannelPoints
		}, "POINTS_ASC")
	case watchPriorityPointsDescending:
		slots.pick(candidates, true, func(a, b watchCandidate) bool {
			return streamers[a.idx].ChannelPoints > streamers[b.idx].ChannelPoints
		}, "POINTS_DESC")
	}
}

//	func (m *Miner) logAvailableCampaigns() {
//		if m.logger == nil || m.twitch == nil {
//			return
//		}
//		summaries := m.twitch.AvailableCampaignSummaries()
//		if len(summaries) == 0 {
//			m.logger.Printf("Active drop campaigns: none detected")
//			return
//		}
//		m.logger.Printf("Active drop campaigns (%d):", len(summaries))
//		for _, summary := range summaries {
//			m.logger.Printf(" - %s", summary)
//		}
//	}
func (m *Miner) gamePreference(streamer *domain.Streamer) (int, bool) {
	baseRank := len(m.gamePriority) + 1
	if streamer == nil || streamer.Stream == nil {
		return baseRank, false
	}
	name := domain.GameKey(streamer.Stream.GameName())
	if name == "" {
		return baseRank, false
	}
	if _, ok := m.gameExclusions[name]; ok {
		return 0, true
	}
	if idx, ok := m.gamePriorityIndex[name]; ok {
		return idx, false
	}
	return baseRank, false
}
