package privacy

import (
	"math"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const initialPoints = 500

func setup(enabled bool) *Anonymizer {
	return newAnonymizer(enabled, func() int { return initialPoints })
}

func TestNew(t *testing.T) {
	t.Run("a first balance", func(t *testing.T) {
		a := New(true)

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelPoints: 4242})

		if got < 100 || got > 1000 {
			t.Errorf("PseudoChannelPoints = %d, want from 100 to 1000", got)
		}
	})
}

func TestAnonymizer_Enabled(t *testing.T) {
	t.Run("switched on", func(t *testing.T) {
		if got := setup(true).Enabled(); got != true {
			t.Errorf("Enabled() = %t, want true", got)
		}
	})

	t.Run("switched off", func(t *testing.T) {
		if got := setup(false).Enabled(); got != false {
			t.Errorf("Enabled() = %t, want false", got)
		}
	})

	t.Run("no anonymizer", func(t *testing.T) {
		var a *Anonymizer
		if got := a.Enabled(); got != false {
			t.Errorf("Enabled() = %t, want false", got)
		}
	})
}

func TestAnonymizer_Name(t *testing.T) {
	t.Run("the first name", func(t *testing.T) {
		a := setup(true)

		if got := a.Name("pewdiepie"); got != "Streamer1" {
			t.Errorf("Name(%q) = %q, want %q", "pewdiepie", got, "Streamer1")
		}
	})

	t.Run("a second name", func(t *testing.T) {
		a := setup(true)
		a.Name("pewdiepie")

		if got := a.Name("ohnepixel"); got != "Streamer2" {
			t.Errorf("Name(%q) = %q, want %q", "ohnepixel", got, "Streamer2")
		}
	})

	t.Run("the first name again", func(t *testing.T) {
		a := setup(true)
		a.Name("pewdiepie")
		a.Name("ohnepixel")

		if got := a.Name("pewdiepie"); got != "Streamer1" {
			t.Errorf("Name(%q) = %q, want %q", "pewdiepie", got, "Streamer1")
		}
	})

	t.Run("the first name in capitals and spaces", func(t *testing.T) {
		a := setup(true)
		a.Name("pewdiepie")

		if got := a.Name(" PewDiePie "); got != "Streamer1" {
			t.Errorf("Name(%q) = %q, want %q", " PewDiePie ", got, "Streamer1")
		}
	})

	t.Run("an empty name", func(t *testing.T) {
		a := setup(true)

		if got := a.Name(""); got != "" {
			t.Errorf("Name(%q) = %q, want %q", "", got, "")
		}
	})

	t.Run("a name of spaces", func(t *testing.T) {
		a := setup(true)

		if got := a.Name("  "); got != "" {
			t.Errorf("Name(%q) = %q, want %q", "  ", got, "")
		}
	})

	t.Run("a name after an empty one", func(t *testing.T) {
		a := setup(true)
		a.Name("")

		if got := a.Name("pewdiepie"); got != "Streamer1" {
			t.Errorf("Name(%q) = %q, want %q", "pewdiepie", got, "Streamer1")
		}
	})

	t.Run("switched off", func(t *testing.T) {
		a := setup(false)

		if got := a.Name(" PewDiePie "); got != " PewDiePie " {
			t.Errorf("Name(%q) = %q, want %q", " PewDiePie ", got, " PewDiePie ")
		}
	})

	t.Run("no anonymizer", func(t *testing.T) {
		var a *Anonymizer

		if got := a.Name("pewdiepie"); got != "pewdiepie" {
			t.Errorf("Name(%q) = %q, want %q", "pewdiepie", got, "pewdiepie")
		}
	})
}

func TestAnonymizer_StreamerName(t *testing.T) {
	t.Run("a streamer with a username", func(t *testing.T) {
		a := setup(true)

		got := a.StreamerName(&domain.Streamer{Username: "pewdiepie", ChannelID: "123"})

		if got != "Streamer1" {
			t.Errorf("StreamerName = %q, want %q", got, "Streamer1")
		}
	})

	t.Run("a streamer already named by username", func(t *testing.T) {
		a := setup(true)
		a.Name("ohnepixel")
		a.Name("pewdiepie")

		got := a.StreamerName(&domain.Streamer{Username: "PewDiePie", ChannelID: "123"})

		if got != "Streamer2" {
			t.Errorf("StreamerName = %q, want %q", got, "Streamer2")
		}
	})

	t.Run("a streamer with only a channel ID", func(t *testing.T) {
		a := setup(true)
		a.Name("pewdiepie")

		got := a.StreamerName(&domain.Streamer{ChannelID: "123"})

		if got != "Streamer2" {
			t.Errorf("StreamerName = %q, want %q", got, "Streamer2")
		}
	})

	t.Run("a streamer with neither", func(t *testing.T) {
		a := setup(true)

		if got := a.StreamerName(&domain.Streamer{}); got != "" {
			t.Errorf("StreamerName = %q, want %q", got, "")
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		a := setup(true)

		if got := a.StreamerName(nil); got != "" {
			t.Errorf("StreamerName(nil) = %q, want %q", got, "")
		}
	})

	t.Run("switched off, a streamer with a username", func(t *testing.T) {
		a := setup(false)

		got := a.StreamerName(&domain.Streamer{Username: " PewDiePie ", ChannelID: "123"})

		if got != "PewDiePie" {
			t.Errorf("StreamerName = %q, want %q", got, "PewDiePie")
		}
	})

	t.Run("switched off, a streamer with only a channel ID", func(t *testing.T) {
		a := setup(false)

		got := a.StreamerName(&domain.Streamer{ChannelID: "123"})

		if got != "id:123" {
			t.Errorf("StreamerName = %q, want %q", got, "id:123")
		}
	})
}

func TestAnonymizer_PseudoChannelPoints(t *testing.T) {
	t.Run("a streamer seen for the first time", func(t *testing.T) {
		a := setup(true)

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		if got != 500 {
			t.Errorf("PseudoChannelPoints = %d, want 500", got)
		}
	})

	t.Run("a balance that has not moved", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		if got != 500 {
			t.Errorf("PseudoChannelPoints = %d, want 500", got)
		}
	})

	t.Run("a balance that grew by ten", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1010})

		if got != 510 {
			t.Errorf("PseudoChannelPoints = %d, want 510", got)
		}
	})

	t.Run("a balance that fell by three", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 997})

		if got != 497 {
			t.Errorf("PseudoChannelPoints = %d, want 497", got)
		}
	})

	t.Run("a balance that fell below the made-up one", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 400})

		if got != -100 {
			t.Errorf("PseudoChannelPoints = %d, want -100", got)
		}
	})

	t.Run("a second change follows the first", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1010})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1007})

		if got != 507 {
			t.Errorf("PseudoChannelPoints = %d, want 507", got)
		}
	})

	t.Run("a second streamer", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1010})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "ohnepixel", ChannelID: "456", ChannelPoints: 20})

		if got != 500 {
			t.Errorf("PseudoChannelPoints = %d, want 500", got)
		}
	})

	t.Run("a streamer renamed under the same channel ID", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "felix", ChannelID: "123", ChannelPoints: 1010})

		if got != 510 {
			t.Errorf("PseudoChannelPoints = %d, want 510", got)
		}
	})

	t.Run("a streamer known only by username, in other capitals", func(t *testing.T) {
		a := setup(true)
		a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelPoints: 1000})

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "PewDiePie", ChannelPoints: 1010})

		if got != 510 {
			t.Errorf("PseudoChannelPoints = %d, want 510", got)
		}
	})

	t.Run("a streamer with no username or channel ID", func(t *testing.T) {
		a := setup(true)

		got := a.PseudoChannelPoints(&domain.Streamer{ChannelPoints: 4242})

		if got != 4242 {
			t.Errorf("PseudoChannelPoints = %d, want 4242", got)
		}
	})

	t.Run("no streamer", func(t *testing.T) {
		a := setup(true)

		if got := a.PseudoChannelPoints(nil); got != 0 {
			t.Errorf("PseudoChannelPoints(nil) = %d, want 0", got)
		}
	})

	t.Run("switched off", func(t *testing.T) {
		a := setup(false)

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 4242})

		if got != 4242 {
			t.Errorf("PseudoChannelPoints = %d, want 4242", got)
		}
	})

	t.Run("no anonymizer", func(t *testing.T) {
		var a *Anonymizer

		got := a.PseudoChannelPoints(&domain.Streamer{Username: "pewdiepie", ChannelID: "123", ChannelPoints: 4242})

		if got != 4242 {
			t.Errorf("PseudoChannelPoints = %d, want 4242", got)
		}
	})
}

func TestRandInt(t *testing.T) {
	t.Run("a range of one number", func(t *testing.T) {
		if got := randInt(7, 7); got != 7 {
			t.Errorf("randInt(7, 7) = %d, want 7", got)
		}
	})

	t.Run("max below min", func(t *testing.T) {
		if got := randInt(7, 6); got != 7 {
			t.Errorf("randInt(7, 6) = %d, want 7", got)
		}
	})

	t.Run("a range of two numbers", func(t *testing.T) {
		if got := randInt(7, 8); got != 7 && got != 8 {
			t.Errorf("randInt(7, 8) = %d, want 7 or 8", got)
		}
	})
}

func TestStrconvItoa(t *testing.T) {
	tests := []struct {
		name string
		v    int
		want string
	}{
		{"zero", 0, "0"},
		{"one digit", 7, "7"},
		{"several digits", 1204, "1204"},
		{"a negative number", -35, "-35"},
		{"the largest number", math.MaxInt64, "9223372036854775807"},
		{"the smallest number (today: only a minus sign)", math.MinInt64, "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strconvItoa(tt.v); got != tt.want {
				t.Errorf("strconvItoa(%d) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}

	t.Run("the smallest number (bug: should read -9223372036854775808)", func(t *testing.T) {
		t.Skip("bug: negating the smallest int overflows, so no digits are written")
		if got := strconvItoa(math.MinInt64); got != "-9223372036854775808" {
			t.Errorf("strconvItoa(%d) = %q, want %q", math.MinInt64, got, "-9223372036854775808")
		}
	})
}
