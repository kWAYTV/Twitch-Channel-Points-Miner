package miner

import (
	"errors"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func streamerKey(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	if id := strings.TrimSpace(streamer.ChannelID); id != "" {
		return id
	}
	return domain.LoginKey(streamer.Username)
}

func (m *Miner) applyStreamerLoginChange(streamer *domain.Streamer, login string) bool {
	if streamer == nil {
		return false
	}
	login = domain.LoginKey(login)
	if login == "" || strings.EqualFold(streamer.Username, login) {
		return false
	}

	oldLogin := domain.LoginKey(streamer.Username)
	if m != nil && m.chat != nil {
		m.chat.Leave(streamer)
	}

	streamer.Username = login
	streamer.StreamerURL = twitchapi.ChannelURL(login)
	if streamer.Stream != nil {
		streamer.Stream.SpadeURL = ""
	}
	if m != nil && m.logger != nil {
		if m.anonymizer.Enabled() {
			m.logger.Printf("identity refresh %s", m.styledStreamerName(streamer))
		} else {
			m.logger.Printf("identity refresh %s -> %s for channel %s", oldLogin, streamer.Username, streamer.ChannelID)
		}
	}

	if m != nil && m.initialPoints != nil && oldLogin != "" {
		stableKey := streamerKey(streamer)
		if stableKey != "" {
			m.initialPointsMu.Lock()
			if initial, ok := m.initialPoints[oldLogin]; ok {
				m.initialPoints[stableKey] = initial
			}
			m.initialPointsMu.Unlock()
		}
	}
	if m != nil && m.chat != nil {
		m.chat.Update(streamer, streamer.IsOnline)
	}
	if m != nil {
		m.syncWarmStartCacheFromStreamer(streamer)
	}
	return true
}

func shouldRepairStreamerIdentity(err error) bool {
	return errors.Is(err, twitch.ErrChannelNotFound)
}

func (m *Miner) repairStreamerIdentity(streamer *domain.Streamer, err error) bool {
	if m == nil || m.twitch == nil || streamer == nil || !shouldRepairStreamerIdentity(err) {
		return false
	}
	user, lookupErr := m.twitch.GetUserByID(streamer.ChannelID)
	if lookupErr != nil {
		if m.logger != nil {
			m.logger.Printf("identity refresh %s: %v", m.styledStreamerName(streamer), lookupErr)
		}
		return false
	}
	if user == nil || user.Login == "" {
		return false
	}
	return m.applyStreamerLoginChange(streamer, user.Login)
}

// withIdentityRepair asks Twitch something about a streamer. When Twitch no
// longer knows the channel by its login, the login is looked up again and the
// question asked once more.
func withIdentityRepair[T any](m *Miner, streamer *domain.Streamer, call func() (T, error)) (T, error) {
	result, err := call()
	if err == nil || !m.repairStreamerIdentity(streamer, err) {
		return result, err
	}
	return call()
}

func (m *Miner) loadChannelPointsContextWithIdentityRepair(streamer *domain.Streamer) (int, error) {
	return withIdentityRepair(m, streamer, func() (int, error) {
		return m.twitch.LoadChannelPointsContext(streamer)
	})
}

func (m *Miner) checkStreamerOnlineWithIdentityRepair(streamer *domain.Streamer) (bool, error) {
	return withIdentityRepair(m, streamer, func() (bool, error) {
		return m.twitch.CheckStreamerOnline(streamer)
	})
}

func (m *Miner) updateStreamWithIdentityRepair(streamer *domain.Streamer) error {
	_, err := withIdentityRepair(m, streamer, func() (struct{}, error) {
		return struct{}{}, m.twitch.UpdateStream(streamer)
	})
	return err
}

func (m *Miner) sendMinuteWatchedWithIdentityRepair(streamer *domain.Streamer) error {
	_, err := withIdentityRepair(m, streamer, func() (struct{}, error) {
		return struct{}{}, m.twitch.SendMinuteWatched(streamer)
	})
	return err
}
