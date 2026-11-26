package prediction

import (
	"fmt"
	"strings"
)

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
