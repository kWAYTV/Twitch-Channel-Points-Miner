package miner

import (
	"testing"
)

func TestBalanceAfterGain(t *testing.T) {
	tests := []struct {
		name    string
		prev    int
		earned  int
		balance int
		want    int
	}{
		{"points earned", 1_000, 50, 0, 1_050},
		{"points earned with a stale lower balance reported", 1_000, 50, 900, 1_050},
		{"points spent", 1_000, -250, 0, 750},
		{"points spent with a stale higher balance reported", 1_000, -250, 1_000, 750},
		{"spending exactly the whole balance", 1_000, -1_000, 0, 0},
		{"spending more than the balance", 1_000, -1_001, 0, 0},
		{"nothing earned and a higher balance reported", 1_000, 0, 1_200, 1_200},
		{"nothing earned and the same balance reported", 1_000, 0, 1_000, 1_000},
		{"nothing earned and a lower balance reported", 1_000, 0, 999, 1_000},
		{"nothing earned and no balance reported", 1_000, 0, 0, 1_000},
		{"nothing earned and a negative balance reported", 1_000, 0, -5, 1_000},
		{"a first gain from zero", 0, 10, 0, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := balanceAfterGain(tt.prev, tt.earned, tt.balance); got != tt.want {
				t.Errorf("balanceAfterGain(%d, %d, %d) = %d, want %d", tt.prev, tt.earned, tt.balance, got, tt.want)
			}
		})
	}
}
