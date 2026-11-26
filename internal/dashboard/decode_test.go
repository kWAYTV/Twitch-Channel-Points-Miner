package dashboard

import "testing"

func TestParsePoint(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   Point
		wantOK bool
	}{
		{"a balance", `{"t":1700000000000,"p":1000}`, Point{T: 1700000000000, P: 1000}, true},
		{"a gain with its reason", `{"t":1700000000000,"p":1010,"d":10,"r":"WATCH"}`, Point{T: 1700000000000, P: 1010, D: 10, R: "WATCH"}, true},
		{"a loss", `{"t":1700000000000,"p":750,"d":-250,"r":"PREDICTION"}`, Point{T: 1700000000000, P: 750, D: -250, R: "PREDICTION"}, true},
		{"a balance of zero", `{"t":1700000000000,"p":0}`, Point{T: 1700000000000}, true},
		{"a field it does not know", `{"t":1700000000000,"p":1000,"x":1}`, Point{T: 1700000000000, P: 1000}, true},
		{"no time", `{"p":1000}`, Point{}, false},
		{"a time of zero", `{"t":0,"p":1000}`, Point{}, false},
		{"a torn line", `{"t":17000`, Point{}, false},
		{"a time that is a string", `{"t":"1700000000000","p":1000}`, Point{}, false},
		{"not JSON", `not json`, Point{}, false},
		{"an empty line", ``, Point{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parsePoint([]byte(tt.line))
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("parsePoint(%q) = %+v, %v, want %+v, %v", tt.line, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   Event
		wantOK bool
	}{
		{"a bet won", `{"t":1700000000000,"type":"BET_WIN","msg":"Won a bet on alice."}`, Event{T: 1700000000000, Type: "BET_WIN", Msg: "Won a bet on alice."}, true},
		{"a message outside ASCII", `{"t":1700000000000,"type":"DROP_CLAIM","msg":"Añadido 🎁"}`, Event{T: 1700000000000, Type: "DROP_CLAIM", Msg: "Añadido 🎁"}, true},
		{"no message", `{"t":1700000000000,"type":"BET_WIN"}`, Event{T: 1700000000000, Type: "BET_WIN"}, true},
		{"no time", `{"type":"BET_WIN","msg":"Won a bet on alice."}`, Event{Type: "BET_WIN", Msg: "Won a bet on alice."}, true},
		{"no type", `{"t":1700000000000,"msg":"Won a bet on alice."}`, Event{}, false},
		{"an empty type", `{"t":1700000000000,"type":"","msg":"Won a bet on alice."}`, Event{}, false},
		{"a torn line", `{"t":1700000000000,"type":"BET_WIN","msg":"Won a b`, Event{}, false},
		{"not JSON", `not json`, Event{}, false},
		{"an empty line", ``, Event{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseEvent([]byte(tt.line))
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("parseEvent(%q) = %+v, %v, want %+v, %v", tt.line, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestLastPoint(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		want   Point
		wantOK bool
	}{
		{"one line", `{"t":1,"p":10}` + "\n", Point{T: 1, P: 10}, true},
		{"two lines", `{"t":1,"p":10}` + "\n" + `{"t":2,"p":20}` + "\n", Point{T: 2, P: 20}, true},
		{"no newline at the end", `{"t":1,"p":10}` + "\n" + `{"t":2,"p":20}`, Point{T: 2, P: 20}, true},
		{"a torn last line", `{"t":1,"p":10}` + "\n" + `{"t":2,"p":2`, Point{T: 1, P: 10}, true},
		{"blank lines at the end", `{"t":1,"p":10}` + "\n\n\n", Point{T: 1, P: 10}, true},
		{"a first line cut at the start", `"p":10}` + "\n" + `{"t":2,"p":20}` + "\n", Point{T: 2, P: 20}, true},
		{"no line that parses", `not json` + "\n" + `{"p":10}` + "\n", Point{}, false},
		{"nothing", ``, Point{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := lastPoint([]byte(tt.data))
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("lastPoint(%q) = %+v, %v, want %+v, %v", tt.data, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
