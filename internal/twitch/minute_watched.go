package twitch

import "github.com/kWAYTV/twitch-points-miner/internal/domain"

func (t *Client) SendMinuteWatched(streamer *domain.Streamer) error {
	if err := t.UpdateStream(streamer); err != nil {
		return err
	}
	t.reportPlaylistCheck(streamer, t.sendWatchPlaylist(streamer))
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

// reportPlaylistCheck logs a failed playlist check. The minute is still
// counted without it, so a failure that repeats is logged once, and again only
// after a check has worked in between.
func (t *Client) reportPlaylistCheck(streamer *domain.Streamer, err error) {
	if err == nil {
		t.playlistFailures.Delete(streamer.Username)
		return
	}
	if _, reported := t.playlistFailures.LoadOrStore(streamer.Username, struct{}{}); reported || t.logger == nil {
		return
	}
	t.logger.Errorf("minute watch %s: %v", t.anonymizer.Shown(streamer), err)
}
