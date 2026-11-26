package miner

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestNewSessionID(t *testing.T) {
	uuidV4 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	if got := newSessionID(); !uuidV4.MatchString(got) {
		t.Errorf("newSessionID() = %q, want a version 4 UUID", got)
	}
}

func TestMiner_logSessionEnd(t *testing.T) {
	const ran = time.Hour + 2*time.Minute + 3*time.Second
	ending := []string{"[INFO] Ending session: 'session-1'", "[INFO] Duration 01h 02m 03s"}
	// loadedAlice loads alice with 100 points, so the session knows her starting balance.
	loadedAlice := func(h *harness) *domain.Streamer {
		h.twitch.Channels["alice"] = &fakeChannel{ID: "1", Points: 100}
		alice := h.miner.loadStreamers([]string{"alice"})[0]
		return alice
	}

	t.Run("no streamers", func(t *testing.T) {
		h := setup(t, Options{})

		h.miner.logSessionEnd("session-1", nil, ran)

		want := append(ending, "[INFO] Total Points gained: "+colorGreen+"+0"+colorReset)
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streamer whose balance did not change", func(t *testing.T) {
		h := setup(t, Options{})
		alice := loadedAlice(h)
		h.log.Reset()

		h.miner.logSessionEnd("session-1", []*domain.Streamer{alice}, ran)

		want := append(ending, "[INFO] Total Points gained: "+colorGreen+"+0"+colorReset)
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streamer who gained points", func(t *testing.T) {
		h := setup(t, Options{})
		alice := loadedAlice(h)
		h.miner.handlePubSubGain(alice, 50, "CLAIM", 150)
		h.log.Reset()

		h.miner.logSessionEnd("session-1", []*domain.Streamer{alice}, ran)

		want := append(ending,
			"[INFO] Total Points gained: "+colorGreen+"+50"+colorReset,
			"[INFO] Alice ("+colorCyan+"150"+colorReset+" points), Total Points "+colorGreen+"+50"+colorReset,
			"[INFO]                          CLAIM (1 times, 50 gained)",
		)
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streamer who lost points", func(t *testing.T) {
		h := setup(t, Options{})
		alice := loadedAlice(h)
		h.miner.handlePubSubGain(alice, -30, "PREDICTION", 70)
		h.log.Reset()

		h.miner.logSessionEnd("session-1", []*domain.Streamer{alice}, ran)

		want := append(ending,
			"[INFO] Total Points gained: "+colorRed+"-30"+colorReset,
			"[INFO] Alice ("+colorCyan+"70"+colorReset+" points), Total Points "+colorRed+"-30"+colorReset,
			"[INFO]                          PREDICTION (1 times, -30 gained)",
		)
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streamer whose gains cancelled out", func(t *testing.T) {
		h := setup(t, Options{})
		alice := loadedAlice(h)
		h.miner.handlePubSubGain(alice, -30, "PREDICTION", 70)
		h.miner.handlePubSubGain(alice, 30, "PREDICTION", 100)
		h.log.Reset()

		h.miner.logSessionEnd("session-1", []*domain.Streamer{alice}, ran)

		want := append(ending,
			"[INFO] Total Points gained: "+colorGreen+"+0"+colorReset,
			"[INFO] Alice ("+colorCyan+"100"+colorReset+" points), Total Points "+colorGreen+"+0"+colorReset,
			"[INFO]                          PREDICTION (2 times, 0 gained)",
		)
		if got := h.logLines(t); !reflect.DeepEqual(got, want) {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("a streamer who gained points, with anonymizing on", func(t *testing.T) {
		h := setup(t, anonymizedOptions)
		alice := loadedAlice(h)
		h.miner.handlePubSubGain(alice, 50, "CLAIM", 150)
		h.log.Reset()

		h.miner.logSessionEnd("session-1", []*domain.Streamer{alice}, ran)

		got := h.logLines(t)
		if len(got) != 5 {
			t.Fatalf("log = %q, want 5 lines", got)
		}
		wantStart := append(ending, "[INFO] Total Points gained: [hidden]")
		if !reflect.DeepEqual(got[:3], wantStart) {
			t.Errorf("first lines = %q, want %q", got[:3], wantStart)
		}
		madeUpBalance := regexp.MustCompile(`^\[INFO\] Streamer1 \(` + regexp.QuoteMeta(colorCyan) + `(\d{3,4}|\d(\.\d{1,2})?k)` + regexp.QuoteMeta(colorReset) + ` points\), Total Points \[hidden\]$`)
		if !madeUpBalance.MatchString(got[3]) {
			t.Errorf("streamer line = %q, want Streamer1 with a made-up balance and a hidden total", got[3])
		}
		if want := "[INFO]                          CLAIM (1 times, [hidden])"; got[4] != want {
			t.Errorf("history line = %q, want %q", got[4], want)
		}
	})
}
