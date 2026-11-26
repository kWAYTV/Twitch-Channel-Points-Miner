package main

import "testing"

func TestAutoUpdateDisabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "unset", value: "", want: false},
		{name: "1", value: "1", want: true},
		{name: "true", value: "true", want: true},
		{name: "yes", value: "yes", want: true},
		{name: "on", value: "on", want: true},
		{name: "upper case", value: "TRUE", want: true},
		{name: "padded", value: " on ", want: true},
		{name: "0", value: "0", want: false},
		{name: "false", value: "false", want: false},
		{name: "another word", value: "enabled", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := autoUpdateDisabled(tt.value); got != tt.want {
				t.Errorf("autoUpdateDisabled(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
