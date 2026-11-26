package twitch

import "github.com/kWAYTV/twitch-points-miner/internal/domain"

func (t *Client) SendMinuteWatched(streamer *domain.Streamer) error {
	if err := t.UpdateStream(streamer); err != nil {
		return err
	}
	if err := t.sendWatchPlaylist(streamer); err != nil && t.logger != nil {
		name := displayName(streamer.Username)
		if t.anonymizer != nil && t.anonymizer.Enabled() {
			name = t.anonymizer.StreamerName(streamer)
		}
		t.logger.Errorf("minute watch %s: %v", name, err)
	}
	if streamer.Stream.SpadeURL == "" {
		if err := t.GetSpadeURL(streamer); err != nil {
			return err
		}
	}
	streamer.Stream.UpdateMinuteWatched()
	err := t.sendSpadePayload(streamer.Stream.SpadeURL, streamer.Username, streamer.Stream.Payload)
	if err == nil {
		streamer.Stream.UpdateMinuteWatched()
	}
	return err
}
