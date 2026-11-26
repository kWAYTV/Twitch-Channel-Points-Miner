package miner

import (
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

type chatUpdate struct {
	Login  string
	Online bool
}

// fakeChatRooms records every chat the miner asks to update or leave.
type fakeChatRooms struct {
	mu      sync.Mutex
	Updates []chatUpdate
	Left    []string
	LeftAll int
}

var _ chatRooms = (*fakeChatRooms)(nil)

func (f *fakeChatRooms) Update(streamer *domain.Streamer, online bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Updates = append(f.Updates, chatUpdate{Login: streamer.Username, Online: online})
}

func (f *fakeChatRooms) Leave(streamer *domain.Streamer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Left = append(f.Left, streamer.Username)
}

func (f *fakeChatRooms) LeaveAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.LeftAll++
}
