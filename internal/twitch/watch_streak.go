package twitch

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func currentStreamHasCompletedWatchStreak(createdAt, achievementAt time.Time) bool {
	if createdAt.IsZero() || achievementAt.IsZero() {
		return false
	}
	return achievementAt.After(createdAt)
}

// outerMilestone is the watch streak's state in a reward list, or nil when the
// answer has none.
func outerMilestone(resp *gqlRewardListResponse) *gqlWatchStreakMilestone {
	if resp == nil || resp.Data.Channel == nil || resp.Data.Channel.Self == nil {
		return nil
	}
	return resp.Data.Channel.Self.WatchStreakMilestone
}

func extractWatchStreakAchievementAt(resp *gqlRewardListResponse) time.Time {
	milestone := outerMilestone(resp)
	if milestone == nil || milestone.WatchStreakMilestone == nil {
		return time.Time{}
	}
	return parseRFC3339Timestamp(milestone.WatchStreakMilestone.AchievementTimestamp)
}

func extractWatchStreakExpiresAt(resp *gqlRewardListResponse) time.Time {
	milestone := outerMilestone(resp)
	if milestone == nil {
		return time.Time{}
	}
	return parseRFC3339Timestamp(milestone.ExpiresAt)
}

func extractWatchStreakLength(resp *gqlRewardListResponse) int {
	milestone := outerMilestone(resp)
	if milestone == nil || milestone.WatchStreakMilestone == nil {
		return -1
	}
	out, err := strconv.Atoi(milestone.WatchStreakMilestone.Value)
	if err != nil {
		return -1
	}
	return out
}

// rewardList fetches a channel's reward list, which carries the streak state.
// It reports false when there is no channel to ask about or Twitch did not
// answer.
func (t *Client) rewardList(channelID string) (*gqlRewardListResponse, bool) {
	if channelID == "" {
		return nil, false
	}
	op := twitchapi.GQLOperations.RewardList.With(map[string]interface{}{"channelID": channelID})
	var resp gqlRewardListResponse
	if err := t.PostGQLDecode(op, &resp); err != nil {
		t.debugf("RewardList lookup failed for channel %s: %v", channelID, err)
		return nil, false
	}
	return &resp, true
}

func (t *Client) rewardListAchievementAt(channelID string) time.Time {
	resp, ok := t.rewardList(channelID)
	if !ok {
		return time.Time{}
	}
	return extractWatchStreakAchievementAt(resp)
}

func (t *Client) LogExtraInfo(channelID string, online bool) string {
	resp, ok := t.rewardList(channelID)
	if !ok {
		return ""
	}
	extrainfoSuffix := fmt.Sprintf(" | streak length %d", extractWatchStreakLength(resp))
	if expiresAt := extractWatchStreakExpiresAt(resp); !expiresAt.IsZero() {
		extrainfoSuffix += fmt.Sprintf(" | expires at %s", expiresAt.Local())
	}
	if online {
		if createdAt := t.fallbackStreamCreatedAt(channelID); !createdAt.IsZero() {
			extrainfoSuffix += fmt.Sprintf(" | started at %s", createdAt.Local())
		}
		if achievementAt := extractWatchStreakAchievementAt(resp); !achievementAt.IsZero() {
			extrainfoSuffix += fmt.Sprintf(" | streak achievement timestamp %s", achievementAt.Local())
		}
	}
	return extrainfoSuffix
}
