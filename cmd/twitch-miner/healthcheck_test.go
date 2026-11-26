package main

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestHealthStatus(t *testing.T) {
	tests := []struct {
		name        string
		age         time.Duration
		err         error
		wantMessage string
		wantHealthy bool
	}{
		{"a beat just now", 0, nil, "healthy: last watch round 0s ago", true},
		{"a beat a minute and a half ago", 90 * time.Second, nil, "healthy: last watch round 1m30s ago", true},
		{"a beat exactly five minutes ago", 5 * time.Minute, nil, "healthy: last watch round 5m0s ago", true},
		{"a beat just over five minutes ago", 5*time.Minute + time.Second, nil, "unhealthy: no watch round for 5m1s", false},
		{"a beat an hour ago", time.Hour, nil, "unhealthy: no watch round for 1h0m0s", false},
		{"a fraction of a second is dropped", 90*time.Second + 400*time.Millisecond, nil, "healthy: last watch round 1m30s ago", true},
		{"no beat yet", 0, os.ErrNotExist, "unhealthy: the miner has not started watching yet", false},
		{"a beat that cannot be read", 0, errors.New("access denied"), "unhealthy: access denied", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, healthy := healthStatus(tt.age, tt.err)
			if message != tt.wantMessage || healthy != tt.wantHealthy {
				t.Errorf("healthStatus(%v, %v) = %q, %t, want %q, %t", tt.age, tt.err, message, healthy, tt.wantMessage, tt.wantHealthy)
			}
		})
	}
}
