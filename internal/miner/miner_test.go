package miner

import (
	"reflect"
	"testing"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func TestNew(t *testing.T) {
	t.Run("every public option set", func(t *testing.T) {
		overrides := map[string]domain.StreamerSettings{"alice": {ClaimDrops: true}}

		m := New(Options{
			Username:                   "account",
			Password:                   "secret",
			ClaimDropsStartup:          true,
			DisableSSLCertVerification: true,
			StreamerSettings:           domain.StreamerSettings{WatchStreak: true, IRCMode: domain.IRCModeNever},
			StreamerOverrides:          overrides,
			WatchStreakWarmStartCache:  true,
		})

		type public struct {
			Username, Password                            string
			ClaimDropsStartup, DisableSSLCertVerification bool
			WatchStreak                                   bool
			IRCMode                                       domain.IRCMode
			StreamerOverrides                             map[string]domain.StreamerSettings
			WatchStreakWarmStartCache                     bool
		}
		got := public{m.Username, m.Password, m.ClaimDropsStartup, m.DisableSSLCertVerification, m.StreamerSettings.WatchStreak, m.StreamerSettings.IRCMode, m.StreamerOverrides, m.WatchStreakWarmStartCache}
		want := public{"account", "secret", true, true, true, domain.IRCModeNever, overrides, true}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("New() = %+v, want %+v", got, want)
		}
	})

	t.Run("@ in the nickname turned off", func(t *testing.T) {
		m := New(Options{DisableAtInNickname: true})

		if !m.disableAtInNickname {
			t.Errorf("disableAtInNickname = false, want true")
		}
	})

	t.Run("no options set", func(t *testing.T) {
		m := New(Options{})

		if m.StreamerSettings.IRCMode != domain.IRCModeOnline {
			t.Errorf("StreamerSettings.IRCMode = %q, want %q", m.StreamerSettings.IRCMode, domain.IRCModeOnline)
		}
	})
}

func TestMiner_now(t *testing.T) {
	t.Run("a clock given", func(t *testing.T) {
		h := setup(t, Options{})

		if got := h.miner.now(); !got.Equal(testNow) {
			t.Errorf("now() = %s, want %s", got, testNow)
		}
	})

	t.Run("no clock given", func(t *testing.T) {
		m := New(Options{})
		before := time.Now()

		got := m.now()

		if after := time.Now(); got.Before(before) || got.After(after) {
			t.Errorf("now() = %s, want the wall clock (%s to %s)", got, before, after)
		}
	})
}

func TestMiner_chatToken(t *testing.T) {
	t.Run("before login", func(t *testing.T) {
		m := New(Options{})

		if got := m.chatToken(); got != "" {
			t.Errorf("chatToken() = %q, want none", got)
		}
	})

	t.Run("after login", func(t *testing.T) {
		h := setup(t, Options{})
		h.twitch.Token = "oauth-token"

		if got := h.miner.chatToken(); got != "oauth-token" {
			t.Errorf("chatToken() = %q, want %q", got, "oauth-token")
		}
	})
}

func TestMiner_filterExcludedTargets(t *testing.T) {
	excluding := Options{StreamersExclude: []string{" Foo ", "BAR"}}
	tests := []struct {
		name         string
		options      Options
		targets      []string
		want         []string
		wantExcluded int
	}{
		{"no targets", excluding, nil, nil, 0},
		{"nobody excluded by the list", excluding, []string{"baz", "qux"}, []string{"baz", "qux"}, 0},
		{"two excluded, in another case and with spaces", excluding, []string{"FOO", "baz", "  bar  ", "qux"}, []string{"baz", "qux"}, 2},
		{"everybody excluded", excluding, []string{"foo", "bar"}, []string{}, 2},
		{"a blank target, with an exclusion list", excluding, []string{"baz", "  "}, []string{"baz"}, 0},
		{"a blank target, with no exclusion list", Options{}, []string{"baz", "  "}, []string{"baz", "  "}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t, tt.options)

			got, excluded := h.miner.filterExcludedTargets(tt.targets)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterExcludedTargets(%q) kept %q, want %q", tt.targets, got, tt.want)
			}
			if excluded != tt.wantExcluded {
				t.Errorf("filterExcludedTargets(%q) excluded %d, want %d", tt.targets, excluded, tt.wantExcluded)
			}
		})
	}
}
