package miner

import (
	"strings"
	"testing"
)

func TestNewSessionIDShape(t *testing.T) {
	id := newSessionID()
	if id == "" {
		t.Fatalf("id should not be empty")
	}
	if strings.Count(id, "-") != 4 {
		t.Fatalf("expected 4 hyphens got %d (%s)", strings.Count(id, "-"), id)
	}
}
