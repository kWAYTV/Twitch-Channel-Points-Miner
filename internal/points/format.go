// Package points writes channel-point amounts for people to read.
package points

import (
	"fmt"
	"strings"
)

// Format shortens an amount to at most two decimals of thousands or millions:
// 950, 1.5k, 2M. A negative amount keeps its sign.
func Format(amount int) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	switch {
	// From 999,995 the thousands would round up to "1000k".
	case amount >= 999_995:
		return sign + short(float64(amount)/1_000_000) + "M"
	case amount >= 1_000:
		return sign + short(float64(amount)/1_000) + "k"
	default:
		return fmt.Sprintf("%s%d", sign, amount)
	}
}

func short(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}
