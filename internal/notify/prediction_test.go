package notify

import (
	"reflect"
	"testing"
)

func TestDescribeBetResult(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{
			"a win",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: WIN, Gained: +500",
			described{Message{"Won 500 points on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a win with no points",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: WIN",
			described{Message{"Won the bet on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a loss",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: LOSE, Lost: -500",
			described{Message{"Lost 500 points on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a loss with no points",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: LOSE",
			described{Message{"Lost the bet on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a refund",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: REFUND, Refunded: +500",
			described{Message{"Refunded 500 points on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a refund with no points",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: REFUND",
			described{Message{"Bet refunded on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"another result",
			"EventPrediction: bob - overtime? - Decision: Yes - Result: CANCELED",
			described{Message{"Bet finished on bob", []string{"CANCELED", `"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"no result",
			"EventPrediction: bob - overtime? - Decision: Yes",
			described{Message{"Bet finished on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"a decision with its letter and colour",
			"EventPrediction: bob - overtime? - Decision: A: Yes (BLUE) - Result: WIN, Gained: +500",
			described{Message{"Won 500 points on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{
			"no decision",
			"EventPrediction: bob - overtime? - Result: WIN, Gained: +500",
			described{Message{"Won 500 points on bob", []string{`"overtime?"`}}, true},
		},
		{
			"no title",
			"EventPrediction: bob - Decision: Yes - Result: WIN, Gained: +500",
			described{Message{"Won 500 points on bob", []string{"Chose Yes"}}, true},
		},
		{
			"a title with a dash in it",
			"EventPrediction: bob - win - or lose? - Decision: Yes - Result: WIN, Gained: +500",
			described{Message{"Won 500 points on bob", []string{`"win or lose?"`, "Chose Yes"}}, true},
		},
		{
			"only a name",
			"EventPrediction: bob",
			described{Message{Headline: "Bet finished on bob"}, true},
		},
		{
			"a prefix in lower case",
			"eventprediction: bob - overtime? - decision: Yes - result: win, gained: +500",
			described{Message{"Won 500 points on bob", []string{`"overtime?"`, "Chose Yes"}}, true},
		},
		{"no name", "EventPrediction:", described{}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeBetResult(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeBetResult(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}

func TestSplitResult(t *testing.T) {
	type result struct {
		kind, pts string
	}
	for _, tc := range []struct {
		name string
		in   string
		want result
	}{
		{"a win with points", "WIN +500", result{"WIN", "500"}},
		{"a loss with points", "LOSE -500", result{"LOSE", "500"}},
		{"a loss with a minus sign", "LOSE −500", result{"LOSE", "500"}},
		{"a refund with points", "REFUND +500", result{"REFUND", "500"}},
		{"a win with no points", "WIN", result{"WIN", ""}},
		{"a win in lower case", "win +500", result{"WIN", "500"}},
		{"words between the result and the points", "WIN and gained +500", result{"WIN", "500"}},
		{"another result", "CANCELED by a moderator", result{"", "CANCELED by a moderator"}},
		{"an empty result", "", result{"", ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kind, pts := splitResult(tc.in)
			if got := (result{kind, pts}); got != tc.want {
				t.Errorf("splitResult(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestDescribeBetPlaced(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{
			"an amount and an outcome",
			"Place 1,234 points on: Yes (BLUE), Points: 50,000, Users: 120 (45.20%), Odds: 1.80 (55.56%) for bob",
			described{Message{"Bet placed on bob", []string{"1,234 points on Yes"}}, true},
		},
		{
			"an outcome without an amount",
			"Place bet on: Yes (BLUE), Points: 10, Users: 2 (50.00%), Odds: 2 (50%) for bob",
			described{Message{"Bet placed on bob", []string{"On Yes"}}, true},
		},
		{
			"an amount without an outcome",
			"Place 1,234 points on: , Points: 50,000 for bob",
			described{Message{"Bet placed on bob", []string{"1,234 points"}}, true},
		},
		{
			"neither an amount nor an outcome",
			"Place bet on: , Points: 50,000 for bob",
			described{Message{Headline: "Bet placed on bob"}, true},
		},
		{
			"an outcome that holds the word for",
			"Place 10 points on: Yes for sure for bob",
			described{Message{"Bet placed on bob", []string{"10 points on Yes for sure"}}, true},
		},
		{
			"upper case",
			"PLACE 10 POINTS ON: Yes FOR bob",
			described{Message{"Bet placed on bob", []string{"10 points on Yes"}}, true},
		},
		{"no streamer", "Place 10 points on: Yes", described{}},
		{"a bet that is scheduled", "Place bet after 4m0s for Alice", described{}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeBetPlaced(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeBetPlaced(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}

func TestDescribeBetNotice(t *testing.T) {
	for _, tc := range []struct {
		name string
		head string
		want described
	}{
		{"a bet that is scheduled", "Place bet after 4m0s for Alice", described{Message{Headline: "Bet on Alice in 4m0s"}, true}},
		{"a bet that is scheduled for nobody", "Place bet after 4m0s", described{}},
		{"a bet that is skipped", "Skip bet for Alice: balance below minimum_points", described{Message{"Bet skipped on Alice", []string{"The balance is below the minimum"}}, true}},
		{"a bet that is skipped for no reason", "Skip bet for Alice: ", described{}},
		{"a bet that is skipped with no colon", "Skip bet for Alice", described{}},
		{"a prediction that failed", "prediction Alice: timeout.", described{Message{"Failed prediction on Alice", []string{"timeout"}}, true}},
		{"a prediction that failed in upper case", "PREDICTION Alice: timeout", described{Message{"Failed prediction on Alice", []string{"timeout"}}, true}},
		{"another line", "Alice is Online!", described{}},
		{"an empty line", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeBetNotice(tc.head)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeBetNotice(%q) = %+v, want %+v", tc.head, got, tc.want)
			}
		})
	}
}

func TestSkipReason(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		want   string
	}{
		{"no outcome chosen", "no outcome selected", "No outcome was selected"},
		{"a stake under Twitch's minimum", "stake below twitch minimum", "The stake is below Twitch's minimum"},
		{"a balance under the minimum", "balance below minimum_points", "The balance is below the minimum"},
		{"a filter that did not match", "filter condition not met", "The bet filter did not match"},
		{"a prediction no longer active", "event status is LOCKED", "The prediction is no longer active"},
		{"a known reason in upper case", "NO OUTCOME", "No outcome was selected"},
		{"another reason", "something_else happened", "Something else happened"},
		{"only spaces", "   ", ""},
		{"an empty reason", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := skipReason(tc.reason); got != tc.want {
				t.Errorf("skipReason(%q) = %q, want %q", tc.reason, got, tc.want)
			}
		})
	}

	t.Run("another reason that starts outside ASCII", func(t *testing.T) {
		if got, want := skipReason("école fermée"), "École fermée"; got != want {
			t.Errorf(`skipReason("école fermée") = %q, want %q`, got, want)
		}
	})
}

func TestShortOutcome(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"a plain outcome", "Yes", "Yes"},
		{"an outcome with its colour", "Yes (BLUE)", "Yes"},
		{"an outcome with its letter", "A: Yes", "Yes"},
		{"an outcome with its number", "#2: Maybe", "Maybe"},
		{"an outcome with its letter and colour", "A: Yes (BLUE)", "Yes"},
		{"an outcome with its statistics", "Yes (BLUE), Points: 50,000, Users: 120 (45.20%), Odds: 1.80 (55.56%)", "Yes"},
		{"statistics in lower case", "Yes, points: 50,000", "Yes"},
		{"a colour in lower case", "Yes (blue)", "Yes (blue)"},
		{"a letter in lower case", "a: Yes", "a: Yes"},
		{"brackets that are not a colour", "Over (2.5)", "Over (2.5)"},
		{"spaces around it", "  Yes  ", "Yes"},
		{"only statistics", ", Points: 50,000", ""},
		{"an empty outcome", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shortOutcome(tc.in); got != tc.want {
				t.Errorf("shortOutcome(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
