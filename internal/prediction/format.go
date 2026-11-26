package prediction

import (
	"fmt"
	"strings"
)

func FormatNumber(value int) string {
	sign := ""
	v := value
	if v < 0 {
		sign = "-"
		v = -v
	}
	switch {
	// From 999,995 the thousands would round up to "1000k".
	case v >= 999_995:
		return sign + trimZeros(fmt.Sprintf("%.2f", float64(v)/1_000_000)) + "M"
	case v >= 1_000:
		return sign + trimZeros(fmt.Sprintf("%.2f", float64(v)/1_000)) + "k"
	default:
		return fmt.Sprintf("%s%d", sign, v)
	}
}

func FormatFloat(val float64) string {
	return trimZeros(fmt.Sprintf("%.2f", val))
}

func trimZeros(val string) string {
	val = strings.TrimRight(strings.TrimRight(val, "0"), ".")
	if val == "" || val == "-" || val == "-0" {
		return "0"
	}
	return val
}
