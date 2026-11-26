package twitch

import (
	"fmt"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// spadeTrackURL is fixed here so recovery does not look up the spade URL of
// streamers that are offline.
const spadeTrackURL = "https://spade.twitch.tv/track"

// streakRecovery is one attempt to save a watch streak that is about to expire,
// by watching a clip or a vod of a stream the account missed.
type streakRecovery struct {
	client     *Client
	streamer   *domain.Streamer
	name       string
	rewardList twitchapi.GQLPersistedOperation
	// missed holds the broadcast ids of the streams that would have kept the streak.
	missed []string
	// length is the streak length before anything was watched.
	length int
	clips  int
	vods   int
}

// RecoverStreak reports whether the streamer's watch streak is safe after trying
// to save it. It returns false with no error when the streak is not expiring.
func (t *Client) RecoverStreak(streamer *domain.Streamer) (bool, error) {
	if streamer == nil || streamer.ChannelID == "" {
		return false, fmt.Errorf("missing streamer channel id")
	}
	rewardList := twitchapi.GQLOperations.RewardList.With(map[string]interface{}{"channelID": streamer.ChannelID})
	r := &streakRecovery{client: t, streamer: streamer, name: t.anonymizer.Shown(streamer), rewardList: rewardList}

	before, err := r.loadRewardList()
	if err != nil {
		return false, err
	}
	if extractWatchStreakExpiresAt(before).IsZero() {
		// streak not expiring ... or some missing data in getting expiresAt
		return false, nil
	}
	milestone := outerMilestone(before)
	if milestone == nil || milestone.MissedStreams == nil {
		return false, fmt.Errorf("RewardList response does not have missedStreams for channel %s", streamer.ChannelID)
	}
	for _, stream := range milestone.MissedStreams {
		for _, id := range stream.BroadcastIdentifiers {
			r.missed = append(r.missed, id.ID)
		}
	}
	r.length = extractWatchStreakLength(before)

	// Clips of the last day first, then of the last week (a clip just over 24 hours
	// old may still be from a missed stream), then vods.
	for _, source := range []string{"LAST_DAY", "LAST_WEEK", "videos"} {
		saved, settled, err := r.watchMissedFrom(source)
		if err != nil || settled {
			return saved, err
		}
	}
	r.waitForExpiryRefresh()

	after, err := r.loadRewardList()
	if err != nil {
		return false, err
	}
	if saved, settled := r.settledBy(after); settled {
		return saved, nil
	}
	expiresAt := extractWatchStreakExpiresAt(after)
	return false, fmt.Errorf("no eligible clips or vods found to save streak expiring at %s", expiresAt.Local())
}

// loadRewardList fetches the streamer's reward list, which carries the streak state.
func (r *streakRecovery) loadRewardList() (*gqlRewardListResponse, error) {
	var resp gqlRewardListResponse
	if err := r.client.PostGQLDecode(r.rewardList, &resp); err != nil {
		return nil, fmt.Errorf("RewardList lookup failed for channel %s: %v", r.streamer.ChannelID, err)
	}
	if resp.Data.Channel == nil || resp.Data.Channel.Self == nil {
		return nil, fmt.Errorf("RewardList response does not have data.channel.self for channel %s", r.streamer.ChannelID)
	}
	return &resp, nil
}

// settledBy reports whether the streak stopped expiring, and if so whether it was
// saved: its length is still what it was before the recovery.
func (r *streakRecovery) settledBy(resp *gqlRewardListResponse) (saved, settled bool) {
	if !extractWatchStreakExpiresAt(resp).IsZero() {
		return false, false
	}
	// streak not expiring ... or some missing data in getting expiresAt
	return r.length == extractWatchStreakLength(resp), true
}

// checkSettled loads the reward list again and reports the outcome as settledBy does.
func (r *streakRecovery) checkSettled() (saved, settled bool, err error) {
	resp, err := r.loadRewardList()
	if err != nil {
		return false, false, err
	}
	saved, settled = r.settledBy(resp)
	return saved, settled, nil
}

// waitForExpiryRefresh gives Twitch a moment to refresh the streak expiry after
// clips were watched and no vod was.
func (r *streakRecovery) waitForExpiryRefresh() {
	if wait := expiryRefreshWait(r.clips, r.vods); wait > 0 {
		r.client.pause(wait)
	}
}

// expiryRefreshWait is how long to wait before the last streak check: nothing
// once a vod was watched, longest after a single clip.
func expiryRefreshWait(clips, vods int) time.Duration {
	if vods != 0 {
		return 0
	}
	if clips == 1 {
		return 10 * time.Second
	} else if clips > 1 {
		return 5 * time.Second
	}
	return 0
}

func (r *streakRecovery) logf(format string, args ...interface{}) {
	if r.client.logger != nil {
		r.client.logger.EmojiPrintf(":ambulance:", "streak recovery %s: "+format, append([]interface{}{r.name}, args...)...)
	}
}
