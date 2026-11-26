package miner

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/drops"
	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/project"
	"github.com/kWAYTV/twitch-points-miner/internal/pubsub"
	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

func (m *Miner) run(streamers []string, useFollowers bool, order domain.FollowersOrder) {
	m.startedAt = m.now()
	m.logger.Printf("%s | %s", project.Name, project.VersionLabel())
	m.logger.Println(project.URL)
	sessionID := newSessionID()
	m.logger.EmojiEventf(":green_circle:", events.EventStartup, "Start session: '%s'", sessionID)
	m.stop = make(chan struct{})
	m.initialPoints = make(map[string]int)

	tw, err := twitch.NewClient(m.Username, twitchapi.AndroidTVUserAgent, m.logger, m.anonymizer)
	if err != nil {
		m.logger.Fatalf("failed to create twitch client: %v", err)
	}
	tw.SetGameChangeHandler(m.handleGameChange)
	m.twitch = tw
	if err := tw.Login(m.Username); err != nil {
		m.logger.Fatalf("login failed: %v", err)
	}
	m.loadWarmStartCache()

	var targets []string
	if useFollowers {
		follows, err := tw.GetFollowers(100, order)
		if err != nil {
			m.logger.Fatalf("failed to load followers: %v", err)
		}
		targets = follows
	} else {
		targets = streamers
	}

	filteredTargets, excludedCount := m.filterExcludedTargets(targets)
	if excludedCount > 0 {
		m.logger.EmojiPrintf(":no_entry_sign:", "Excluded %d streamer(s) via streamers_exclude", excludedCount)
	}
	targets = filteredTargets

	loadStartedAt := m.now()
	m.logger.EmojiPrintf(":hourglass_flowing_sand:", "Loading data for %d streamer(s). Please wait...", len(targets))
	streamerObjs := m.loadStreamers(targets)
	m.saveWarmStartCacheSnapshot(streamerObjs)
	m.logLoadSummary(streamerObjs, m.now().Sub(loadStartedAt))

	claimDropsEnabled := m.shouldClaimDrops(streamerObjs)
	tw.ExcludeGamesFromDropClaims(m.gameExclusions)
	dropClaimer := drops.NewClaimer(tw, m.logger, m.showDropsProgress)
	if claimDropsEnabled {
		dropClaimer.Claim("startup drop claim failed")
	}

	m.streamers = streamerObjs

	// background loops
	if claimDropsEnabled {
		go dropClaimer.Run(m.stop, m.human)
	}
	go m.contextRefresher(streamerObjs, m.stop)
	go m.minuteWatcher(streamerObjs, m.stop)
	go m.startPubSub(tw, streamerObjs, m.stop)
	go m.streakRecovery(streamerObjs, m.stop)
	go m.analyticsSnapshotter(streamerObjs, m.stop)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	m.shutdown(sessionID)
}

func (m *Miner) startPubSub(tw *twitch.Client, streamers []*domain.Streamer, stop <-chan struct{}) {
	client := pubsub.NewClient(
		tw,
		m.logger,
		m.anonymizer,
		streamers,
		pubsub.Handlers{
			Gain:       m.handlePubSubGain,
			Presence:   m.handlePubSubPresence,
			BetSettled: recordSettledBet,
		},
		m.human,
	)
	client.Start(stop)
}

func (m *Miner) shutdown(sessionID string) {
	select {
	case <-m.stop:
	default:
		close(m.stop)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
	m.logSessionEnd(sessionID, m.streamers, m.now().Sub(m.startedAt))
	if m.chat != nil {
		m.chat.LeaveAll()
	}
	os.Exit(0)
}

// logSessionEnd reports how long the session ran and what each streamer's balance gained.
func (m *Miner) logSessionEnd(sessionID string, streamers []*domain.Streamer, ran time.Duration) {
	m.logger.EmojiEventf(":stop_sign:", events.EventShutdown, "Ending session: '%s'", sessionID)
	duration := formatDuration(ran)
	m.logger.EmojiPrintf(":hourglass:", "Duration %s", duration)
	totalPointsChange := 0
	for _, s := range streamers {
		totalPointsChange += s.ChannelPoints - m.initialPoints[streamerKey(s)]
	}
	if m.anonymizer.Enabled() {
		m.logger.EmojiPrintf(":chart_with_upwards_trend:", "Total Points gained: [hidden]")
	} else {
		totalColor, totalSign, total := signedChange(totalPointsChange)
		m.logger.EmojiPrintf(":chart_with_upwards_trend:", "Total Points gained: %s%s%d%s", totalColor, totalSign, total, colorReset)
	}
	for _, s := range streamers {
		change := s.ChannelPoints - m.initialPoints[streamerKey(s)]
		if change == 0 && len(s.History) == 0 {
			continue
		}
		signColor, sign, total := signedChange(change)
		who := m.nameWithPoints(s)
		if m.anonymizer.Enabled() {
			m.logger.EmojiPrintf(":moneybag:", "%s, Total Points [hidden]", who)
			if s.History != nil {
				for reason, entry := range s.History {
					m.logger.Printf("                         %s (%d times, [hidden])", reason, entry.Count)
				}
			}
		} else {
			m.logger.EmojiPrintf(":moneybag:", "%s, Total Points %s%s%d%s", who, signColor, sign, total, colorReset)
			if s.History != nil {
				for reason, entry := range s.History {
					m.logger.Printf("                         %s (%d times, %d gained)", reason, entry.Count, entry.Amount)
				}
			}
		}
	}
}

// newSessionID creates a UUID-like string for session logging.
func newSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	// variant and version bits per RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
