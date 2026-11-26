package pubsub

import (
	"context"
	"fmt"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
)

// The topic families. A topic is a family, a dot, and the ID of the account
// or of a channel.
const (
	communityPointsUserTopic    = "community-points-user-v1"
	predictionsUserTopic        = "predictions-user-v1"
	playbackTopic               = "video-playback-by-id"
	raidTopic                   = "raid"
	predictionsChannelTopic     = "predictions-channel-v1"
	momentsChannelTopic         = "community-moments-channel-v1"
	communityPointsChannelTopic = "community-points-channel-v1"
)

func topicOf(family, id string) string {
	return family + "." + id
}

// topicChannelID is the channel ID a topic of the family ends in, or "" for a
// topic of another family.
func topicChannelID(topic, family string) string {
	id, ok := strings.CutPrefix(topic, family+".")
	if !ok {
		return ""
	}
	return id
}

func (p *Client) buildTopics() ([]string, error) {
	userID := p.account.UserID()
	if userID == "" {
		return nil, fmt.Errorf("no user id for pubsub")
	}
	return topicsFor(userID, p.streamers), nil
}

// topicsFor lists the topics to listen to: the account's own, then each
// streamer's by what its settings ask for. A streamer without a channel ID has none.
func topicsFor(userID string, streamers []*domain.Streamer) []string {
	topics := []string{}
	seen := make(map[string]struct{})

	addTopic := func(topic string) {
		if topic == "" {
			return
		}
		if _, ok := seen[topic]; ok {
			return
		}
		seen[topic] = struct{}{}
		topics = append(topics, topic)
	}

	addTopic(topicOf(communityPointsUserTopic, userID))

	shouldListenPredictionUser := false
	for _, s := range streamers {
		if s.Settings.MakePredictions {
			shouldListenPredictionUser = true
			break
		}
	}
	if shouldListenPredictionUser {
		addTopic(topicOf(predictionsUserTopic, userID))
	}

	for _, s := range streamers {
		if s.ChannelID == "" {
			continue
		}
		addTopic(topicOf(playbackTopic, s.ChannelID))
		if s.Settings.FollowRaid {
			addTopic(topicOf(raidTopic, s.ChannelID))
		}
		if s.Settings.MakePredictions {
			addTopic(topicOf(predictionsChannelTopic, s.ChannelID))
		}
		if s.Settings.ClaimMoments {
			addTopic(topicOf(momentsChannelTopic, s.ChannelID))
		}
		if s.Settings.CommunityGoals {
			addTopic(topicOf(communityPointsChannelTopic, s.ChannelID))
		}
	}

	return topics
}

// topicNeedsAuth says whether a topic is the account's own, which Twitch only
// serves with the account's token.
func topicNeedsAuth(topic string) bool {
	family := topicFamily(topic)
	return family == communityPointsUserTopic || family == predictionsUserTopic
}

// topicFamily is a topic without the ID it ends in.
func topicFamily(topic string) string {
	return strings.SplitN(topic, ".", 2)[0]
}

func (p *Client) listenTopics(ctx context.Context, conn *websocket.Conn, topics []string) error {
	for _, t := range topics {
		data := map[string]interface{}{"topics": []string{t}}
		if topicNeedsAuth(t) {
			data["auth_token"] = p.account.AuthToken()
		}
		payload := map[string]interface{}{
			"type":  "LISTEN",
			"nonce": random.String(16),
			"data":  data,
		}
		if p.anonymizer.Enabled() {
			p.debugf("PubSub LISTEN %s", topicFamily(t))
		} else {
			p.debugf("PubSub LISTEN %s", t)
		}
		if err := wsjson.Write(ctx, conn, payload); err != nil {
			return err
		}
	}
	return nil
}

func chunkTopics(topics []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		return [][]string{topics}
	}
	var batches [][]string
	for len(topics) > 0 {
		end := chunkSize
		if len(topics) < chunkSize {
			end = len(topics)
		}
		batches = append(batches, topics[:end])
		topics = topics[end:]
	}
	return batches
}
