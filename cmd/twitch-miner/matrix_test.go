package main

import (
	"reflect"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/notify"
)

func TestDefaultMatrixConfig(t *testing.T) {
	t.Run("decoded, it is Matrix turned off", func(t *testing.T) {
		cfg, err := decodeConfig(map[string]interface{}{"matrix": defaultMatrixConfig()})
		if err != nil {
			t.Fatalf("decodeConfig() error = %v, want none", err)
		}
		if want := (matrixConfig{Events: []string{}}); !reflect.DeepEqual(cfg.Matrix, want) {
			t.Errorf("decodeConfig() Matrix = %+v, want %+v", cfg.Matrix, want)
		}
		if got := notify.NewMatrixNotifier(cfg.Matrix.settings()); got != nil {
			t.Errorf("NewMatrixNotifier() = %+v, want nil", got)
		}
	})
}

func TestMatrixConfig_settings(t *testing.T) {
	t.Run("every field set", func(t *testing.T) {
		cfg := matrixConfig{
			Homeserver:  "https://matrix.example.org",
			AccessToken: "syt_secret",
			RoomID:      "!room:example.org",
			Events:      []string{"BET_WIN", "STREAMER_ONLINE"},
		}
		want := notify.MatrixSettings{
			Homeserver:  "https://matrix.example.org",
			AccessToken: "syt_secret",
			RoomID:      "!room:example.org",
			Events:      []string{"BET_WIN", "STREAMER_ONLINE"},
		}
		if got := cfg.settings(); !reflect.DeepEqual(got, want) {
			t.Errorf("settings() = %+v, want %+v", got, want)
		}
	})

	t.Run("nothing set", func(t *testing.T) {
		cfg := matrixConfig{}
		want := notify.MatrixSettings{}
		if got := cfg.settings(); !reflect.DeepEqual(got, want) {
			t.Errorf("settings() = %+v, want %+v", got, want)
		}
	})
}
