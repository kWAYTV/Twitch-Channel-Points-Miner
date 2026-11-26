package drops

import (
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/twitch"
)

func TestFormatClaimLine(t *testing.T) {
	tests := []struct {
		name string
		drop twitch.ClaimedDrop
		want string
	}{
		{
			"every field set",
			twitch.ClaimedDrop{RewardName: "Token", CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180},
			"Claim Token (Summer) 180/180 (100%)",
		},
		{
			"a drop whose game account is not connected",
			twitch.ClaimedDrop{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 180, RequiredValue: 180, AccountNotConnected: true},
			"Drop Token (Summer) is ready but not received: connect your Tanks account at twitch.tv/drops/inventory",
		},
		{
			"a drop whose game account is not connected, the game unknown",
			twitch.ClaimedDrop{RewardName: "Token", CampaignName: "Summer", AccountNotConnected: true},
			"Drop Token (Summer) is ready but not received: connect the game's account at twitch.tv/drops/inventory",
		},
		{
			"no reward name",
			twitch.ClaimedDrop{CampaignName: "Summer", CurrentValue: 180, RequiredValue: 180},
			"Claim Drop (Summer) 180/180 (100%)",
		},
		{
			"no campaign name",
			twitch.ClaimedDrop{RewardName: "Token", CurrentValue: 180, RequiredValue: 180},
			"Claim Token (Unknown Campaign) 180/180 (100%)",
		},
		{
			"no required minutes",
			twitch.ClaimedDrop{RewardName: "Token", CampaignName: "Summer", CurrentValue: 30},
			"Claim Token (Summer) 30 (100%)",
		},
		{
			"nothing set",
			twitch.ClaimedDrop{},
			"Claim Drop (Unknown Campaign) 0 (0%)",
		},
		{
			"names outside ASCII",
			twitch.ClaimedDrop{RewardName: "Żeton", CampaignName: "夏", CurrentValue: 180, RequiredValue: 180},
			"Claim Żeton (夏) 180/180 (100%)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatClaimLine(tt.drop); got != tt.want {
				t.Errorf("formatClaimLine(%+v) = %q, want %q", tt.drop, got, tt.want)
			}
		})
	}
}

func TestFormatDropProgress(t *testing.T) {
	tests := []struct {
		name              string
		current, required int
		want              string
	}{
		{"part of the way", 3, 10, "3/10"},
		{"one required minute", 0, 1, "0/1"},
		{"no required minutes", 5, 0, "5"},
		{"negative required minutes", 5, -1, "5"},
		{"negative minutes watched", -2, 10, "-2/10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropProgress(tt.current, tt.required); got != tt.want {
				t.Errorf("formatDropProgress(%d, %d) = %q, want %q", tt.current, tt.required, got, tt.want)
			}
		})
	}
}

func TestFormatDropProgressBar(t *testing.T) {
	tests := []struct {
		name              string
		current, required int
		want              string
	}{
		{"a third of the way", 63, 180, "[#######-------------]"},
		{"just under one step", 8, 180, "[--------------------]"},
		{"exactly one step", 9, 180, "[#-------------------]"},
		{"one minute short", 179, 180, "[###################-]"},
		{"exactly finished", 180, 180, "[####################]"},
		{"past the end", 250, 180, "[####################]"},
		{"nothing watched or required", 0, 0, "[--------------------]"},
		{"watched with nothing required", 5, 0, "[####################]"},
		{"negative minutes watched", -5, 180, "[--------------------]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropProgressBar(tt.current, tt.required); got != tt.want {
				t.Errorf("formatDropProgressBar(%d, %d) = %q, want %q", tt.current, tt.required, got, tt.want)
			}
		})
	}
}

func TestFormatDropStatusLine(t *testing.T) {
	tests := []struct {
		name   string
		status twitch.DropStatus
		want   string
	}{
		{
			"a drop in progress",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Summer / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"a campaign named after its game",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Tanks", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"no game name",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Summer / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"no campaign name",
			twitch.DropStatus{RewardName: "Token", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Unknown Campaign / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"no reward name",
			twitch.DropStatus{CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Summer / Drop / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"no channel name",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Summer / Token / unknown_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"names in spaces",
			twitch.DropStatus{RewardName: " Token ", CampaignName: " Summer ", GameName: " Tanks ", ChannelName: " example_streamer ", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Tanks / Summer / Token / example_streamer / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"names outside ASCII",
			twitch.DropStatus{RewardName: "Żeton", CampaignName: "夏", GameName: "Czołgi", ChannelName: "strumień", CurrentValue: 63, RequiredValue: 180},
			"DROP_PROGRESS [#######-------------] 35% | Czołgi / 夏 / Żeton / strumień / 63m of 180m / remaining 117m / IN_PROGRESS",
		},
		{
			"nothing set",
			twitch.DropStatus{},
			"DROP_PROGRESS [--------------------] 0% | Unknown Campaign / Drop / unknown_streamer / 0m watched / IN_PROGRESS",
		},
		{
			"no required minutes",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 30},
			"DROP_PROGRESS [####################] 100% | Tanks / Summer / Token / example_streamer / 30m watched / IN_PROGRESS",
		},
		{
			"negative minutes watched",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: -5, RequiredValue: 180},
			"DROP_PROGRESS [--------------------] 0% | Tanks / Summer / Token / example_streamer / 0m of 180m / remaining 180m / IN_PROGRESS",
		},
		{
			"exactly finished",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 180, RequiredValue: 180},
			"DROP_PROGRESS [####################] 100% | Tanks / Summer / Token / example_streamer / 180m of 180m / remaining 0m / CLAIMABLE",
		},
		{
			"watched past the end",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 200, RequiredValue: 180},
			"DROP_PROGRESS [####################] 111% | Tanks / Summer / Token / example_streamer / 200m of 180m / remaining 0m / CLAIMABLE",
		},
		{
			"a claimed drop",
			twitch.DropStatus{RewardName: "Token", CampaignName: "Summer", GameName: "Tanks", ChannelName: "example_streamer", CurrentValue: 180, RequiredValue: 180, Claimed: true},
			"DROP_PROGRESS [####################] 100% | Tanks / Summer / Token / example_streamer / 180m of 180m / remaining 0m / CLAIMED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropStatusLine(tt.status); got != tt.want {
				t.Errorf("formatDropStatusLine(%+v)\n got %q\nwant %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestFormatDropStatusTitle(t *testing.T) {
	tests := []struct {
		name                   string
		game, campaign, reward string
		want                   string
	}{
		{"three different names", "Tanks", "Summer", "Token", "Tanks / Summer / Token"},
		{"a campaign named after its game", "Tanks", "Tanks", "Token", "Tanks / Token"},
		{"a campaign named after its game in other capitals", "Tanks", "TANKS", "Token", "Tanks / Token"},
		{"a reward named after its game", "Tanks", "Summer", "tanks", "Tanks / Summer"},
		{"three equal names", "Tanks", "Tanks", "Tanks", "Tanks"},
		{"no game", "", "Summer", "Token", "Summer / Token"},
		{"a game of spaces", "  ", "Summer", "Token", "Summer / Token"},
		{"names in spaces", " Tanks ", " Summer ", " Token ", "Tanks / Summer / Token"},
		{"no names", "", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropStatusTitle(tt.game, tt.campaign, tt.reward); got != tt.want {
				t.Errorf("formatDropStatusTitle(%q, %q, %q) = %q, want %q", tt.game, tt.campaign, tt.reward, got, tt.want)
			}
		})
	}
}

func TestDropStatusLabel(t *testing.T) {
	tests := []struct {
		name   string
		status twitch.DropStatus
		want   string
	}{
		{"claimed", twitch.DropStatus{Claimed: true}, "CLAIMED"},
		{"claimed and still marked claimable", twitch.DropStatus{Claimed: true, Claimable: true}, "CLAIMED"},
		{"marked claimable", twitch.DropStatus{Claimable: true, CurrentValue: 3, RequiredValue: 10}, "CLAIMABLE"},
		{"one minute short", twitch.DropStatus{CurrentValue: 9, RequiredValue: 10}, "IN_PROGRESS"},
		{"exactly finished", twitch.DropStatus{CurrentValue: 10, RequiredValue: 10}, "CLAIMABLE"},
		{"watched past the end", twitch.DropStatus{CurrentValue: 11, RequiredValue: 10}, "CLAIMABLE"},
		{"watched with nothing required", twitch.DropStatus{CurrentValue: 5}, "IN_PROGRESS"},
		{"nothing set", twitch.DropStatus{}, "IN_PROGRESS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropStatusLabel(tt.status); got != tt.want {
				t.Errorf("dropStatusLabel(%+v) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestProgressPercent(t *testing.T) {
	tests := []struct {
		name              string
		current, required int
		want              int
	}{
		{"half way", 5, 10, 50},
		{"a third, rounded down", 1, 3, 33},
		{"nothing watched", 0, 10, 0},
		{"exactly finished", 10, 10, 100},
		{"watched past the end", 25, 10, 250},
		{"negative minutes watched", -5, 10, 0},
		{"watched with nothing required", 1, 0, 100},
		{"nothing watched or required", 0, 0, 0},
		{"negative minutes with nothing required", -1, 0, 0},
		{"negative required minutes", 5, -3, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progressPercent(tt.current, tt.required); got != tt.want {
				t.Errorf("progressPercent(%d, %d) = %d, want %d", tt.current, tt.required, got, tt.want)
			}
		})
	}
}
