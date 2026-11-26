package miner

import (
	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

// chatRooms joins and leaves streamers' chats.
type chatRooms interface {
	Update(streamer *domain.Streamer, online bool)
	Leave(streamer *domain.Streamer)
	LeaveAll()
}
