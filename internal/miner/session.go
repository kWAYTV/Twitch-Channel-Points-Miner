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
	m.startedAt = time.Now()
	m.logger.Printf("%s | %s", project.Name, project.VersionLabel())
	m.logger.Println(project.URL)
	sessionID := newSessionID()
	m.logger.EmojiEventf(":green_circle:", events.EventStartup, "Start session: '%s'", sessionID)
	m.stop = make(chan struct{})
	m.initialPoints = make(map[string]int)

	tw, err := twitch.NewClient(m.Username, twitchapi.GetUserAgent("CHROME"), m.Password, m.logger, m.anonymizer)
	if err != nil {
		m.logger.Fatalf("failed to create twitch client: %v", err)
	}
	tw.SetGameChangeHandler(m.handleGameChange)
	m.twitch = tw
	if err := m.twitch.Login(m.Username); err != nil {
		m.logger.Fatalf("login failed: %v", err)
	}
	m.loadWarmStartCache()

	var targets []string
	if useFollowers {
		follows, err := m.twitch.GetFollowers(100, order)
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

	loadStartedAt := time.Now()
	m.logger.EmojiPrintf(":hourglass_flowing_sand:", "Loading data for %d streamer(s). Please wait...", len(targets))
	streamerObjs := m.loadStreamers(targets)
	m.saveWarmStartCacheSnapshot(streamerObjs)
	m.logLoadSummary(streamerObjs, time.Since(loadStartedAt))

	claimDropsEnabled := m.shouldClaimDrops(streamerObjs)
	dropClaimer := drops.NewClaimer(m.twitch, m.logger, m.showDropsProgress)
	if m.ClaimDropsStartup && claimDropsEnabled {
		dropClaimer.Claim("startup drop claim failed")
	}

	m.streamers = streamerObjs

	// background loops
	if claimDropsEnabled {
		go dropClaimer.Run(m.stop)
	}
	go m.contextRefresher(streamerObjs, m.stop)
	go m.minuteWatcher(streamerObjs, m.stop)
	go m.startPubSub(streamerObjs, m.stop)
	go m.streakRecovery(streamerObjs, m.stop)
	go m.analyticsSnapshotter(m.stop)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	m.shutdown(sessionID)
}

func (m *Miner) startPubSub(streamers []*domain.Streamer, stop <-chan struct{}) {
	client := pubsub.NewClient(
		m.twitch,
		m.logger,
		m.anonymizer,
		streamers,
		m.DisableSSLCertVerification,
		m.handlePubSubGain,
		m.handlePubSubPresence,
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
	m.logger.EmojiEventf(":stop_sign:", events.EventShutdown, "Ending session: '%s'", sessionID)
	duration := formatDuration(time.Since(m.startedAt))
	m.logger.EmojiPrintf(":hourglass:", "Duration %s", duration)
	totalPointsChange := 0
	for _, s := range m.streamers {
		totalPointsChange += s.ChannelPoints - m.initialPoints[streamerInitialPointsKey(s)]
	}
	if m.anonymizeLogs() {
		m.logger.EmojiPrintf(":chart_with_upwards_trend:", "Total Points gained: [hidden]")
	} else {
		totalSign := "+"
		totalColor := colorGreen
		if totalPointsChange < 0 {
			totalPointsChange = -totalPointsChange
			totalSign = "-"
			totalColor = colorRed
		}
		m.logger.EmojiPrintf(":chart_with_upwards_trend:", "Total Points gained: %s%s%d%s", totalColor, totalSign, totalPointsChange, colorReset)
	}
	for _, s := range m.streamers {
		initial := m.initialPoints[streamerInitialPointsKey(s)]
		total := s.ChannelPoints - initial
		if total == 0 && len(s.History) == 0 {
			continue
		}
		signColor := colorGreen
		sign := "+"
		if total < 0 {
			signColor = colorRed
			sign = "-"
			total = -total
		}
		points := m.formattedStreamerPoints(s)
		if m.anonymizeLogs() {
			m.logger.EmojiPrintf(":moneybag:", "%s (%s%s%s points), Total Points [hidden]", m.styledStreamerName(s), colorCyan, points, colorReset)
			if s.History != nil {
				for reason, entry := range s.History {
					m.logger.Printf("                         %s (%d times, [hidden])", reason, entry.Count)
				}
			}
		} else {
			m.logger.EmojiPrintf(":moneybag:", "%s (%s%s%s points), Total Points %s%s%d%s", m.styledStreamerName(s), colorCyan, points, colorReset, signColor, sign, total, colorReset)
			if s.History != nil {
				for reason, entry := range s.History {
					m.logger.Printf("                         %s (%d times, %d gained)", reason, entry.Count, entry.Amount)
				}
			}
		}
	}
	m.chat.LeaveAll()
	os.Exit(0)
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
