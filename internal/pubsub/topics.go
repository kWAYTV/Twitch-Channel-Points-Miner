package pubsub

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
)

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

	addTopic(fmt.Sprintf("community-points-user-v1.%s", userID))

	shouldListenPredictionUser := false
	for _, s := range streamers {
		if s.Settings.MakePredictions {
			shouldListenPredictionUser = true
			break
		}
	}
	if shouldListenPredictionUser {
		addTopic(fmt.Sprintf("predictions-user-v1.%s", userID))
	}

	for _, s := range streamers {
		if s.ChannelID == "" {
			continue
		}
		addTopic(fmt.Sprintf("video-playback-by-id.%s", s.ChannelID))
		if s.Settings.FollowRaid {
			addTopic(fmt.Sprintf("raid.%s", s.ChannelID))
		}
		if s.Settings.MakePredictions {
			addTopic(fmt.Sprintf("predictions-channel-v1.%s", s.ChannelID))
		}
		if s.Settings.ClaimMoments {
			addTopic(fmt.Sprintf("community-moments-channel-v1.%s", s.ChannelID))
		}
		if s.Settings.CommunityGoals {
			addTopic(fmt.Sprintf("community-points-channel-v1.%s", s.ChannelID))
		}
	}

	return topics
}

// topicNeedsAuth says whether a topic is the account's own, which Twitch only
// serves with the account's token.
func topicNeedsAuth(topic string) bool {
	return strings.HasPrefix(topic, "community-points-user-v1.") || strings.HasPrefix(topic, "predictions-user-v1.")
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
		if p.anonymizeLogs() {
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
