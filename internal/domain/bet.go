package domain

type Strategy string

const (
	StrategyMostVoted  Strategy = "MOST_VOTED"
	StrategyHighOdds   Strategy = "HIGH_ODDS"
	StrategyPercentage Strategy = "PERCENTAGE"
	StrategySmartMoney Strategy = "SMART_MONEY"
	StrategySmart      Strategy = "SMART"
	// StrategyKelly bets only on an outcome the pool underprices, and sizes
	// the stake by the Kelly criterion.
	StrategyKelly   Strategy = "KELLY"
	StrategyNumber1 Strategy = "NUMBER_1"
	StrategyNumber2 Strategy = "NUMBER_2"
	StrategyNumber3 Strategy = "NUMBER_3"
	StrategyNumber4 Strategy = "NUMBER_4"
	StrategyNumber5 Strategy = "NUMBER_5"
	StrategyNumber6 Strategy = "NUMBER_6"
	StrategyNumber7 Strategy = "NUMBER_7"
	StrategyNumber8 Strategy = "NUMBER_8"
)

type DelayMode string

const (
	DelayModeFromStart  DelayMode = "FROM_START"
	DelayModeFromEnd    DelayMode = "FROM_END"
	DelayModePercentage DelayMode = "PERCENTAGE"
)

type Condition string

const (
	ConditionGT  Condition = "GT"
	ConditionLT  Condition = "LT"
	ConditionGTE Condition = "GTE"
	ConditionLTE Condition = "LTE"
)

type OutcomeKey string

const (
	OutcomePercentageUsers OutcomeKey = "PERCENTAGE_USERS"
	OutcomeOdds            OutcomeKey = "ODDS"
	OutcomeOddsPercentage  OutcomeKey = "ODDS_PERCENTAGE"
	OutcomeTopPoints       OutcomeKey = "TOP_POINTS"
	OutcomeTotalUsers      OutcomeKey = "TOTAL_USERS"
	OutcomeTotalPoints     OutcomeKey = "TOTAL_POINTS"
	OutcomeDecisionUsers   OutcomeKey = "DECISION_USERS"
	OutcomeDecisionPoints  OutcomeKey = "DECISION_POINTS"
)

type FilterCondition struct {
	By    OutcomeKey `json:"by"`
	Where Condition  `json:"where"`
	Value *float64   `json:"value"`
}

type BetSettings struct {
	Strategy        Strategy         `json:"strategy,omitempty"`
	Percentage      *int             `json:"percentage,omitempty"`
	PercentageGap   *int             `json:"percentage_gap,omitempty"`
	MaxPoints       *int             `json:"max_points,omitempty"`
	MinimumPoints   *int             `json:"minimum_points,omitempty"`
	StealthMode     *bool            `json:"stealth_mode,omitempty"`
	FilterCondition *FilterCondition `json:"filter_condition,omitempty"`
	Delay           *float64         `json:"delay,omitempty"`
	DelayMode       DelayMode        `json:"delay_mode,omitempty"`
}

// What a bet setting is when the config leaves it out.
const (
	DefaultBetStrategy      = StrategyKelly
	DefaultBetPercentage    = 5
	DefaultBetPercentageGap = 20
	DefaultBetMaxPoints     = 50000
	DefaultBetMinimumPoints = 0
	DefaultBetStealthMode   = false
	// Halfway through the window: a channel can close betting before the
	// window ends, and a bet left for the last seconds is then never placed.
	DefaultBetDelayMode = DelayModePercentage
	DefaultBetDelay     = 0.5
)

func (b *BetSettings) Default() {
	if b.Strategy == "" {
		b.Strategy = DefaultBetStrategy
	}
	if b.Percentage == nil {
		v := DefaultBetPercentage
		b.Percentage = &v
	}
	if b.PercentageGap == nil {
		v := DefaultBetPercentageGap
		b.PercentageGap = &v
	}
	if b.MaxPoints == nil {
		v := DefaultBetMaxPoints
		b.MaxPoints = &v
	}
	if b.MinimumPoints == nil {
		v := DefaultBetMinimumPoints
		b.MinimumPoints = &v
	}
	if b.StealthMode == nil {
		v := DefaultBetStealthMode
		b.StealthMode = &v
	}
	if b.DelayMode == "" {
		b.DelayMode = DefaultBetDelayMode
	}
	if b.Delay == nil {
		d := DefaultBetDelay
		b.Delay = &d
	}
}
