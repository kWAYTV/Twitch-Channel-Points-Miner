package pubsub

import (
	"sync"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/prediction"
)

type claimed struct {
	Streamer string
	ID       string
}

type bet struct {
	EventID   string
	OutcomeID string
	Amount    int
}

// fakeAccount is an Account that answers from its fields and records what it
// is asked to do on Twitch.
type fakeAccount struct {
	mu sync.Mutex

	ID    string
	Token string
	Login string
	Live  bool

	LiveErr    error
	ContextErr error
	BonusErr   error
	MomentErr  error
	RaidErr    error
	BetErr     error

	ContextLoads      []string
	BonusClaims       []claimed
	MomentClaims      []claimed
	RaidsJoined       []claimed
	Bets              []bet
	GoalContributions []string

	// ContextLoaded, when set, receives the streamer of each context load.
	ContextLoaded chan string
}

func (a *fakeAccount) UserID() string    { return a.ID }
func (a *fakeAccount) AuthToken() string { return a.Token }
func (a *fakeAccount) LoginName() string { return a.Login }

func (a *fakeAccount) IsStreamLive(string) (bool, error) {
	return a.Live, a.LiveErr
}

func (a *fakeAccount) LoadChannelPointsContext(streamer *domain.Streamer) (int, error) {
	a.mu.Lock()
	a.ContextLoads = append(a.ContextLoads, streamer.Username)
	a.mu.Unlock()
	if a.ContextLoaded != nil {
		a.ContextLoaded <- streamer.Username
	}
	return 0, a.ContextErr
}

func (a *fakeAccount) ClaimBonus(streamer *domain.Streamer, claimID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.BonusClaims = append(a.BonusClaims, claimed{streamer.Username, claimID})
	return a.BonusErr
}

func (a *fakeAccount) ClaimMoment(streamer *domain.Streamer, momentID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.MomentClaims = append(a.MomentClaims, claimed{streamer.Username, momentID})
	return a.MomentErr
}

func (a *fakeAccount) JoinRaid(streamer *domain.Streamer, raidID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.RaidsJoined = append(a.RaidsJoined, claimed{streamer.Username, raidID})
	return a.RaidErr
}

func (a *fakeAccount) MakePrediction(event *prediction.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Bets = append(a.Bets, bet{event.EventID, event.Decision.OutcomeID, event.Decision.Amount})
	return a.BetErr
}

func (a *fakeAccount) ContributeToCommunityGoals(streamer *domain.Streamer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.GoalContributions = append(a.GoalContributions, streamer.Username)
}
