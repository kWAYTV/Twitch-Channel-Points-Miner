package notify

import (
	"reflect"
	"testing"
)

func TestDescribeGain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		head   string
		game   string
		streak string
		want   described
	}{
		{"points gained for watching", "+50 → Alice (12.4k points) - Reason: WATCH", "", "", described{Message{"+50 points on Alice", []string{"Watching", "Balance 12.4k points"}}, true}},
		{"points lost", "-500 → Alice (1k points) - Reason: PREDICTION", "", "", described{Message{"−500 points on Alice", []string{"Prediction", "Balance 1k points"}}, true}},
		{"points lost with a minus sign", "−500 → Alice (1k points) - Reason: PREDICTION", "", "", described{Message{"−500 points on Alice", []string{"Prediction", "Balance 1k points"}}, true}},
		{"an arrow in ASCII", "+50 -> Alice (12.4k points) - Reason: WATCH", "", "", described{Message{"+50 points on Alice", []string{"Watching", "Balance 12.4k points"}}, true}},
		{"points lost with an arrow in ASCII", "-500 -> Alice (1k points) - Reason: PREDICTION", "", "", described{Message{"−500 points on Alice", []string{"Prediction", "Balance 1k points"}}, true}},
		{"a game and a watch streak", "+50 → Alice (12.4k points) - Reason: WATCH", "Valorant", "3", described{Message{"+50 points on Alice", []string{"Watching Valorant", "Watch streak 3", "Balance 12.4k points"}}, true}},
		{"no reason", "+50 → Alice (12.4k points)", "", "", described{Message{"+50 points on Alice", []string{"Balance 12.4k points"}}, true}},
		{"no balance", "+50 → Alice - Reason: WATCH", "", "", described{Message{"+50 points on Alice", []string{"Watching"}}, true}},
		{"a reason in lower case", "+50 → Alice (12.4k points) - reason: watch", "", "", described{Message{"+50 points on Alice", []string{"Watching", "Balance 12.4k points"}}, true}},
		{"an amount without a sign", "50 → Alice (12.4k points) - Reason: WATCH", "", "", described{}},
		{"no amount", "→ Alice (12.4k points) - Reason: WATCH", "", "", described{}},
		{"nothing after the arrow", "+50 →", "", "", described{}},
		{"no arrow", "+50 Alice (12.4k points)", "", "", described{}},
		{"an empty line", "", "", "", described{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, ok := describeGain(tc.head, tc.game, tc.streak)
			if got := (described{message, ok}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("describeGain(%q, %q, %q) = %+v, want %+v", tc.head, tc.game, tc.streak, got, tc.want)
			}
		})
	}
}

func TestGainDetails(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		game   string
		streak string
		want   []string
	}{
		{"watching", "WATCH", "", "", []string{"Watching", ""}},
		{"watching a game", "WATCH", "Valorant", "", []string{"Watching Valorant", ""}},
		{"watching with a streak", "WATCH", "", "3", []string{"Watching", "Watch streak 3"}},
		{"watching in lower case", "watch", "", "", []string{"Watching", ""}},
		{"a watch streak", "WATCH_STREAK", "", "", []string{"Watch streak", ""}},
		{"a watch streak with its length", "WATCH_STREAK", "", "4", []string{"Watch streak 4", ""}},
		{"a watch streak with a game", "WATCH_STREAK", "Valorant", "", []string{"Watch streak", "Playing Valorant"}},
		{"a bonus claim", "CLAIM", "", "", []string{"Bonus claim", ""}},
		{"a bonus claim with a streak", "CLAIM", "Valorant", "3", []string{"Bonus claim", "Watch streak 3"}},
		{"a raid", "RAID", "", "", []string{"Raid", "", ""}},
		{"a raid with a game and a streak", "RAID", "Valorant", "3", []string{"Raid", "Playing Valorant", "Watch streak 3"}},
		{"no reason", "", "", "", []string{"", ""}},
		{"no reason with a game and a streak", "", "Valorant", "3", []string{"Playing Valorant", "Watch streak 3"}},
		{"only spaces for a reason", "   ", "", "", []string{"", ""}},
		{"another reason", "PREDICTION", "", "", []string{"Prediction", "", ""}},
		{"another reason with a game and a streak", "PREDICTION", "Valorant", "3", []string{"Prediction", "Playing Valorant", "Watch streak 3"}},
		{"another reason of several words", "SUB_GIFT_BONUS", "", "", []string{"Sub gift bonus", "", ""}},
		{"a reason followed by more words", "WATCH and more", "", "", []string{"Watching", ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := gainDetails(tc.reason, tc.game, tc.streak); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("gainDetails(%q, %q, %q) = %q, want %q", tc.reason, tc.game, tc.streak, got, tc.want)
			}
		})
	}

	t.Run("another reason that starts outside ASCII", func(t *testing.T) {
		want := []string{"Étoile", "", ""}
		if got := gainDetails("ÉTOILE", "", ""); !reflect.DeepEqual(got, want) {
			t.Errorf(`gainDetails("ÉTOILE", "", "") = %q, want %q`, got, want)
		}
	})
}

func TestUnsignedAmount(t *testing.T) {
	type amount struct {
		digits string
		lost   bool
	}
	for _, tc := range []struct {
		name string
		in   string
		want amount
	}{
		{"a gain", "+50", amount{"50", false}},
		{"a loss", "-50", amount{"50", true}},
		{"a loss with a minus sign", "−50", amount{"50", true}},
		{"no sign", "50", amount{"50", false}},
		{"spaces around it", " -50 ", amount{"50", true}},
		{"a thousands separator", "+1,234", amount{"1,234", false}},
		{"only a sign", "-", amount{"", true}},
		{"an empty amount", "", amount{"", false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			digits, lost := unsignedAmount(tc.in)
			if got := (amount{digits, lost}); got != tc.want {
				t.Errorf("unsignedAmount(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsSignedAmount(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want bool
	}{
		{"a plus and a digit", "+5", true},
		{"a hyphen and a digit", "-5", true},
		{"a minus sign and a digit", "−5", true},
		{"a sign and several digits", "+1,234", true},
		{"a sign and a digit outside ASCII", "+٣", true},
		{"a digit without a sign", "50", false},
		{"a sign and a letter", "+a", false},
		{"two signs", "+-5", false},
		{"only a sign", "+", false},
		{"only a digit", "5", false},
		{"an empty amount", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSignedAmount(tc.in); got != tc.want {
				t.Errorf("isSignedAmount(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
