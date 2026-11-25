package main

import (
	"testing"
)

func TestAutoUpdateDisabledByEnv(t *testing.T) {
	for value, want := range map[string]bool{"": false, "0": false, "1": true, "true": true} {
		t.Setenv("TCPM_DISABLE_AUTO_UPDATE", value)
		if got := autoUpdateDisabledByEnv(); got != want {
			t.Fatalf("TCPM_DISABLE_AUTO_UPDATE=%q got %v, want %v", value, got, want)
		}
	}
}
