package jsonvalue

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFloat(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want float64
	}{
		{"a float64", float64(2.5), 2.5},
		{"an int", int(7), 7},
		{"a json.Number", json.Number("12.5"), 12.5},
		{"a json.Number with an exponent", json.Number("1e3"), 1000},
		{"a json.Number that is not a number", json.Number("twelve"), 0},
		{"zero", float64(0), 0},
		{"a negative float64", float64(-2.5), -2.5},
		{"a negative int", int(-7), -7},
		{"an int64", int64(7), 0},
		{"a string of digits", "3", 0},
		{"a bool", true, 0},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Float(tt.v); got != tt.want {
				t.Errorf("Float(%#v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
	}{
		{"a string", "hello", "hello"},
		{"an empty string", "", ""},
		{"a string outside ASCII", "añadido 🎁", "añadido 🎁"},
		{"a number", float64(5), ""},
		{"a bool", true, ""},
		{"a list of strings", []interface{}{"a"}, ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.v); got != tt.want {
				t.Errorf("String(%#v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"true", true, true},
		{"false", false, false},
		{"the string true", "true", false},
		{"the number one", float64(1), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bool(tt.v); got != tt.want {
				t.Errorf("Bool(%#v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestAt(t *testing.T) {
	user := map[string]interface{}{"id": "42", "roles": nil, "name": "", "banned": false}
	data := map[string]interface{}{"data": map[string]interface{}{"user": user}, "count": float64(0)}

	tests := []struct {
		name string
		data interface{}
		path string
		want interface{}
	}{
		{"a value three levels down", data, "data.user.id", "42"},
		{"an object on the way", data, "data.user", user},
		{"a key at the top", data, "count", float64(0)},
		{"an empty string", data, "data.user.name", ""},
		{"false", data, "data.user.banned", false},
		{"a null value", data, "data.user.roles", nil},
		{"a step that is missing", data, "data.missing.id", nil},
		{"a last key that is missing", data, "data.user.email", nil},
		{"a path through a string", data, "data.user.id.deeper", nil},
		{"a top key that is missing", data, "nope", nil},
		{"an empty path", data, "", nil},
		{"a path that ends in a dot", data, "data.", nil},
		{"data that is a list", []interface{}{data}, "data", nil},
		{"the first of a list", []interface{}{data}, "0.data.user.id", "42"},
		{"the last of a list", map[string]interface{}{"edges": []interface{}{"a", "b"}}, "edges.1", "b"},
		{"one past the end of a list", map[string]interface{}{"edges": []interface{}{"a", "b"}}, "edges.2", nil},
		{"a negative position in a list", map[string]interface{}{"edges": []interface{}{"a", "b"}}, "edges.-1", nil},
		{"an empty list", map[string]interface{}{"edges": []interface{}{}}, "edges.0", nil},
		{"nil data", nil, "data", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := At(tt.data, tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("At(data, %q) = %#v, want %#v", tt.path, got, tt.want)
			}
		})
	}
}

func TestFirstString(t *testing.T) {
	data := map[string]interface{}{"empty": "", "login": "alice", "name": "Alice", "id": float64(42)}

	tests := []struct {
		name string
		data map[string]interface{}
		keys []string
		want string
	}{
		{"one key that holds a string", data, []string{"login"}, "alice"},
		{"two keys that hold strings", data, []string{"name", "login"}, "Alice"},
		{"a missing key first", data, []string{"missing", "login"}, "alice"},
		{"an empty string first", data, []string{"empty", "login"}, "alice"},
		{"a number first", data, []string{"id", "login"}, "alice"},
		{"only a number", data, []string{"id"}, ""},
		{"only a missing key", data, []string{"missing"}, ""},
		{"no keys", data, nil, ""},
		{"an empty object", map[string]interface{}{}, []string{"login"}, ""},
		{"a nil object", nil, []string{"login"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstString(tt.data, tt.keys...); got != tt.want {
				t.Errorf("FirstString(data, %q) = %q, want %q", tt.keys, got, tt.want)
			}
		})
	}
}

func TestFirstInt(t *testing.T) {
	data := map[string]interface{}{
		"zero":     float64(0),
		"points":   float64(9),
		"balance":  float64(1200),
		"loss":     float64(-3),
		"fraction": float64(2.9),
		"half":     float64(0.5),
		"number":   json.Number("15"),
		"int":      int(4),
		"text":     "7",
	}

	tests := []struct {
		name string
		data map[string]interface{}
		keys []string
		want int
	}{
		{"one key that holds a number", data, []string{"points"}, 9},
		{"two keys that hold numbers", data, []string{"balance", "points"}, 1200},
		{"a missing key first", data, []string{"missing", "points"}, 9},
		{"a zero first", data, []string{"zero", "points"}, 9},
		{"a string of digits first", data, []string{"text", "points"}, 9},
		{"a number below one first", data, []string{"half", "points"}, 9},
		{"a negative number", data, []string{"loss"}, -3},
		{"a fraction", data, []string{"fraction"}, 2},
		{"a json.Number", data, []string{"number"}, 15},
		{"an int", data, []string{"int"}, 4},
		{"only a zero", data, []string{"zero"}, 0},
		{"only a missing key", data, []string{"missing"}, 0},
		{"no keys", data, nil, 0},
		{"an empty object", map[string]interface{}{}, []string{"points"}, 0},
		{"a nil object", nil, []string{"points"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstInt(tt.data, tt.keys...); got != tt.want {
				t.Errorf("FirstInt(data, %q) = %d, want %d", tt.keys, got, tt.want)
			}
		})
	}
}
