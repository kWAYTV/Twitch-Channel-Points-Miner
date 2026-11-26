package main

import "testing"

func TestDebugLevel_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want debugLevel
	}{
		{name: "false", json: `false`, want: debugOff},
		{name: "true", json: `true`, want: debugSummary},
		{name: "full", json: `"full"`, want: debugFull},
		{name: "full in capitals with spaces around", json: `" FULL "`, want: debugFull},
		{name: "true with spaces around it", json: ` true `, want: debugSummary},
		{name: "null", json: `null`, want: debugOff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := debugLevel("unset")
			if err := got.UnmarshalJSON([]byte(tt.json)); err != nil {
				t.Fatalf("UnmarshalJSON(%s) error = %v, want none", tt.json, err)
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %q, want %q", tt.json, got, tt.want)
			}
		})
	}

	failures := []struct {
		name string
		json string
	}{
		{name: "another word", json: `"deep"`},
		{name: "true as text", json: `"true"`},
		{name: "an empty string", json: `""`},
		{name: "a number", json: `1`},
		{name: "an object", json: `{}`},
		{name: "a string that never ends", json: `"full`},
		{name: "nothing", json: ``},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			got := debugLevel("unset")
			err := got.UnmarshalJSON([]byte(tt.json))
			if want := `debug must be false, true or "full"`; err == nil || err.Error() != want {
				t.Errorf("UnmarshalJSON(%s) error = %v, want %s", tt.json, err, want)
			}
			if got != "unset" {
				t.Errorf("UnmarshalJSON(%s) left %q, want the value untouched", tt.json, got)
			}
		})
	}
}
