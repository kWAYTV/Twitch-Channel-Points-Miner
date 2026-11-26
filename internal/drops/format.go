package drops

import (
	"fmt"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

func formatClaimLine(drop twitch.ClaimedDrop) string {
	reward := rewardLabel(drop.RewardName)
	campaign := campaignLabel(drop.CampaignName)
	if drop.AccountNotConnected {
		account := "the game's account"
		if drop.GameName != "" {
			account = "your " + drop.GameName + " account"
		}
		return fmt.Sprintf("Drop %s (%s) is ready but not received: connect %s at twitch.tv/drops/inventory", reward, campaign, account)
	}
	progress := formatDropProgress(drop.CurrentValue, drop.RequiredValue)
	percent := progressPercent(drop.CurrentValue, drop.RequiredValue)
	return fmt.Sprintf("Claim %s (%s) %s (%d%%)", reward, campaign, progress, percent)
}

// rewardLabel names a drop's reward, also when Twitch gave it no name.
func rewardLabel(name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	return "Drop"
}

// campaignLabel names a drop's campaign, also when Twitch gave it no name.
func campaignLabel(name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	return "Unknown Campaign"
}

// gameLabel is the game's name as shown, with a stand-in when Twitch gave none.
func gameLabel(name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	return "Unknown Game"
}

// formatUnconnectedLine says that a game's drops are not collected and where
// to connect the account.
func formatUnconnectedLine(status twitch.DropStatus) string {
	link := strings.TrimSpace(status.AccountLinkURL)
	if link == "" {
		link = "https://www.twitch.tv/drops/inventory"
	}
	return fmt.Sprintf("Drops for %s are not collected: your Twitch account is not connected to the game. Connect it at %s", gameLabel(status.GameName), link)
}

func formatDropProgress(current, required int) string {
	if required > 0 {
		return fmt.Sprintf("%d/%d", current, required)
	}
	return fmt.Sprintf("%d", current)
}

func formatDropProgressBar(current, required int) string {
	const width = 20
	percent := progressPercent(current, required)
	if percent > 100 {
		percent = 100
	}
	filled := (percent * width) / 100
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func formatDropStatusLine(status twitch.DropStatus) string {
	game := strings.TrimSpace(status.GameName)
	if game == "" {
		game = strings.TrimSpace(status.CampaignName)
	}
	title := formatDropStatusTitle(game, campaignLabel(status.CampaignName), rewardLabel(status.RewardName))
	channel := strings.TrimSpace(status.ChannelName)
	if channel == "" {
		channel = "unknown_streamer"
	}

	current := status.CurrentValue
	required := status.RequiredValue
	if current < 0 {
		current = 0
	}
	remaining := 0
	if required > current {
		remaining = required - current
	}
	if required <= 0 {
		return fmt.Sprintf(
			"DROP_PROGRESS %s %d%% | %s / %s / %dm watched / %s",
			formatDropProgressBar(current, required),
			progressPercent(current, required),
			title,
			channel,
			current,
			dropStatusLabel(status),
		)
	}
	return fmt.Sprintf(
		"DROP_PROGRESS %s %d%% | %s / %s / %dm of %dm / remaining %dm / %s",
		formatDropProgressBar(current, required),
		progressPercent(current, required),
		title,
		channel,
		current,
		required,
		remaining,
		dropStatusLabel(status),
	)
}

func formatDropStatusTitle(game, campaign, reward string) string {
	parts := make([]string, 0, 3)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		for _, existing := range parts {
			if strings.EqualFold(existing, value) {
				return
			}
		}
		parts = append(parts, value)
	}
	add(game)
	add(campaign)
	add(reward)
	return strings.Join(parts, " / ")
}

func dropStatusLabel(status twitch.DropStatus) string {
	if status.Claimed {
		return "CLAIMED"
	}
	if status.Claimable || (status.RequiredValue > 0 && status.CurrentValue >= status.RequiredValue) {
		return "CLAIMABLE"
	}
	return "IN_PROGRESS"
}

func progressPercent(current, required int) int {
	if required <= 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	percent := (current * 100) / required
	if percent < 0 {
		return 0
	}
	return percent
}
