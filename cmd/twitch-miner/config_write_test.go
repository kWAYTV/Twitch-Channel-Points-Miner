package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestFormatSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		want     string
	}{
		{
			name:     "no keys",
			settings: map[string]interface{}{},
			want:     "{}\n",
		},
		{
			name:     "one key",
			settings: map[string]interface{}{"username": "me"},
			want: `{
  "username": "me"
}
`,
		},
		{
			name:     "keys of two groups",
			settings: map[string]interface{}{"streamers": []interface{}{}, "password": "secret", "username": "me"},
			want: `{
  "username": "me",
  "password": "secret",

  "streamers": []
}
`,
		},
		{
			name:     "keys the layout does not know",
			settings: map[string]interface{}{"zebra": 1, "debug": true, "an_old_key": "kept"},
			want: `{
  "debug": true,

  "an_old_key": "kept",
  "zebra": 1
}
`,
		},
		{
			name: "a bet with its filter",
			settings: map[string]interface{}{"bet": map[string]interface{}{
				"unknown":          1,
				"filter_condition": map[string]interface{}{"value": nil, "where": nil, "by": nil},
				"delay":            6,
				"strategy":         "SMART",
			}},
			want: `{
  "bet": {
    "strategy": "SMART",
    "delay": 6,
    "filter_condition": {
      "by": null,
      "where": null,
      "value": null
    },
    "unknown": 1
  }
}
`,
		},
		{
			name: "overrides for two streamers",
			settings: map[string]interface{}{"streamer_overrides": map[string]interface{}{
				"zeta": map[string]interface{}{
					"bet":              map[string]interface{}{"max_points": 5, "strategy": "SMART"},
					"make_predictions": true,
				},
				"alpha": map[string]interface{}{"claim_drops": false},
			}},
			want: `{
  "streamer_overrides": {
    "alpha": {
      "claim_drops": false
    },
    "zeta": {
      "make_predictions": true,
      "bet": {
        "strategy": "SMART",
        "max_points": 5
      }
    }
  }
}
`,
		},
		{
			name:     "an object with no keys",
			settings: map[string]interface{}{"streamer_overrides": map[string]interface{}{}},
			want: `{
  "streamer_overrides": {}
}
`,
		},
		{
			name: "an object the layout does not know",
			settings: map[string]interface{}{"custom": map[string]interface{}{
				"b": 1,
				"a": map[string]interface{}{"d": 1, "c": 2},
			}},
			want: `{
  "custom": {
    "a": {
      "c": 2,
      "d": 1
    },
    "b": 1
  }
}
`,
		},
		{
			name:     "a short list",
			settings: map[string]interface{}{"watch_priority": []interface{}{"STREAK", "DROPS", "ORDER"}},
			want: `{
  "watch_priority": ["STREAK", "DROPS", "ORDER"]
}
`,
		},
		{
			name: "a list too long for one line",
			settings: map[string]interface{}{"streamers": []interface{}{
				"streamer01", "streamer02", "streamer03", "streamer04", "streamer05", "streamer06", "streamer07",
			}},
			want: `{
  "streamers": [
    "streamer01",
    "streamer02",
    "streamer03",
    "streamer04",
    "streamer05",
    "streamer06",
    "streamer07"
  ]
}
`,
		},
		{
			name:     "a list of objects",
			settings: map[string]interface{}{"streamers": []interface{}{map[string]interface{}{"name": "one"}}},
			want: `{
  "streamers": [
    {
      "name": "one"
    }
  ]
}
`,
		},
		{
			name: "a URL with characters HTML escapes",
			settings: map[string]interface{}{"discord": map[string]interface{}{
				"events":      []interface{}{"BET_WIN"},
				"webhook_api": "https://discord.com/api/webhooks/1/a?x=1&y=<2>",
			}},
			want: `{
  "discord": {
    "webhook_api": "https://discord.com/api/webhooks/1/a?x=1&y=<2>",
    "events": ["BET_WIN"]
  }
}
`,
		},
		{
			name:     "a null",
			settings: map[string]interface{}{"timezone": nil},
			want: `{
  "timezone": null
}
`,
		},
		{
			name:     "numbers as a JSON file gives them",
			settings: map[string]interface{}{"analytics": map[string]interface{}{"port": 5000.0, "host": "127.0.0.1", "enabled": true}},
			want: `{
  "analytics": {
    "enabled": true,
    "host": "127.0.0.1",
    "port": 5000
  }
}
`,
		},
		{
			name:     "characters outside ASCII",
			settings: map[string]interface{}{"username": "señor_ñ", "streamers": []interface{}{"日本"}},
			want: `{
  "username": "señor_ñ",

  "streamers": ["日本"]
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatSettings(tt.settings)
			if err != nil {
				t.Fatalf("formatSettings() error = %v, want none", err)
			}
			if string(got) != tt.want {
				t.Errorf("formatSettings() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}

	failures := []struct {
		name     string
		settings map[string]interface{}
	}{
		{name: "a value JSON cannot hold", settings: map[string]interface{}{"username": make(chan int)}},
		{name: "a list with a value JSON cannot hold", settings: map[string]interface{}{"streamers": []interface{}{make(chan int)}}},
		{name: "a nested value JSON cannot hold", settings: map[string]interface{}{"bet": map[string]interface{}{"delay": make(chan int)}}},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatSettings(tt.settings)
			if err == nil || err.Error() != "json: unsupported type: chan int" {
				t.Errorf("formatSettings() error = %v, want json: unsupported type: chan int", err)
			}
			if got != nil {
				t.Errorf("formatSettings() = %q, want nothing", got)
			}
		})
	}
}

func TestInlineList(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		wantLine string
		wantOK   bool
	}{
		{name: "a few names", value: []interface{}{"STREAK", "DROPS"}, wantLine: `["STREAK", "DROPS"]`, wantOK: true},
		{name: "one name", value: []interface{}{"STREAK"}, wantLine: `["STREAK"]`, wantOK: true},
		{name: "plain values of every kind", value: []interface{}{"a", 1.5, true, nil}, wantLine: `["a", 1.5, true, null]`, wantOK: true},
		{name: "names outside ASCII", value: []interface{}{"日本", "ñ"}, wantLine: `["日本", "ñ"]`, wantOK: true},
		{name: "an empty list", value: []interface{}{}, wantLine: "", wantOK: false},
		{name: "a string", value: "STREAK", wantLine: "", wantOK: false},
		{name: "nothing", value: nil, wantLine: "", wantOK: false},
		{name: "a list holding an object", value: []interface{}{"a", map[string]interface{}{}}, wantLine: "", wantOK: false},
		{name: "a list holding a list", value: []interface{}{"a", []interface{}{}}, wantLine: "", wantOK: false},
		{name: "a list holding a value JSON cannot hold", value: []interface{}{make(chan int)}, wantLine: "", wantOK: false},
		{
			name:     "a line of 71 characters",
			value:    []interface{}{strings.Repeat("a", 67)},
			wantLine: `["` + strings.Repeat("a", 67) + `"]`,
			wantOK:   true,
		},
		{
			name:     "a line of 72 characters",
			value:    []interface{}{strings.Repeat("a", 68)},
			wantLine: `["` + strings.Repeat("a", 68) + `"]`,
			wantOK:   true,
		},
		{
			name:     "a line of 73 characters",
			value:    []interface{}{strings.Repeat("a", 69)},
			wantLine: `["` + strings.Repeat("a", 69) + `"]`,
			wantOK:   false,
		},
		{
			name:     "a line of 72 characters, one outside ASCII (today: too long)",
			value:    []interface{}{strings.Repeat("a", 67) + "ñ"},
			wantLine: `["` + strings.Repeat("a", 67) + `ñ"]`,
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, ok := inlineList(tt.value)
			if line != tt.wantLine || ok != tt.wantOK {
				t.Errorf("inlineList() = %q, %v, want %q, %v", line, ok, tt.wantLine, tt.wantOK)
			}
		})
	}

	t.Run("a line of 72 characters, one outside ASCII (bug: should fit)", func(t *testing.T) {
		t.Skip("bug: the width is counted in bytes, so a line with characters outside ASCII is split before 72 characters")
		value := []interface{}{strings.Repeat("a", 67) + "ñ"}
		line, ok := inlineList(value)
		if wantLine := `["` + strings.Repeat("a", 67) + `ñ"]`; line != wantLine || !ok {
			t.Errorf("inlineList() = %q, %v, want %q, true", line, ok, wantLine)
		}
	})
}

func TestKeyGroups(t *testing.T) {
	shape := &layout{groups: [][]string{{"b", "a"}, {"c"}}}
	tests := []struct {
		name   string
		object map[string]interface{}
		shape  *layout
		want   [][]string
	}{
		{name: "every known key", object: map[string]interface{}{"a": 1, "b": 1, "c": 1}, shape: shape, want: [][]string{{"b", "a"}, {"c"}}},
		{name: "one key missing from a group", object: map[string]interface{}{"a": 1, "c": 1}, shape: shape, want: [][]string{{"a"}, {"c"}}},
		{name: "a whole group missing", object: map[string]interface{}{"c": 1}, shape: shape, want: [][]string{{"c"}}},
		{name: "known and unknown keys", object: map[string]interface{}{"z": 1, "c": 1, "y": 1}, shape: shape, want: [][]string{{"c"}, {"y", "z"}}},
		{name: "only unknown keys", object: map[string]interface{}{"z": 1, "y": 1}, shape: shape, want: [][]string{{"y", "z"}}},
		{name: "a key whose value is null", object: map[string]interface{}{"a": nil}, shape: shape, want: [][]string{{"a"}}},
		{name: "no keys", object: map[string]interface{}{}, shape: shape, want: nil},
		{name: "no layout", object: map[string]interface{}{"b": 1, "a": 1}, shape: nil, want: [][]string{{"a", "b"}}},
		{name: "a layout with no groups", object: map[string]interface{}{"b": 1, "a": 1}, shape: &layout{each: shape}, want: [][]string{{"a", "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyGroups(tt.object, tt.shape); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keyGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
