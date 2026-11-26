// Package jsonvalue reads loosely typed values out of decoded JSON.
package jsonvalue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Float reads a number decoded from JSON, whatever numeric type it arrived as.
func Float(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

// Int reads a whole number decoded from JSON, whatever numeric type it arrived
// as. A fraction is cut off, and anything that is not a number is 0.
func Int(v interface{}) int {
	if n, ok := v.(int64); ok {
		return int(n)
	}
	return int(Float(v))
}

// Text writes any JSON value as text: a string as it is, a number in full
// digits, and nothing for a missing value. Surrounding spaces are dropped.
func Text(v interface{}) string {
	if v == nil {
		return ""
	}
	// A number decoded from JSON is a float64, which fmt would write as
	// 1.23456789e+08 from seven digits up.
	if number, ok := v.(float64); ok {
		return strconv.FormatFloat(number, 'f', -1, 64)
	}
	text := strings.TrimSpace(fmt.Sprint(v))
	if text == "<nil>" {
		return ""
	}
	return text
}

// String returns v when it is a string, and "" otherwise.
func String(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Bool returns v when it is a bool, and false otherwise.
func Bool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// At follows a dot-separated path through nested objects and lists, where a
// number is a position in a list, and returns nil when any step is missing.
func At(data interface{}, path string) interface{} {
	if data == nil {
		return nil
	}
	current := data
	parts := strings.Split(path, ".")
	for _, p := range parts {
		switch node := current.(type) {
		case map[string]interface{}:
			current = node[p]
		case []interface{}:
			index, err := strconv.Atoi(p)
			if err != nil || index < 0 || index >= len(node) {
				return nil
			}
			current = node[index]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}
	return current
}

// FirstString returns the first non-empty string found under one of keys.
func FirstString(data map[string]interface{}, keys ...string) string {
	if data == nil {
		return ""
	}
	for _, key := range keys {
		if val, ok := data[key].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

// FirstInt returns the first non-zero number found under one of keys.
func FirstInt(data map[string]interface{}, keys ...string) int {
	if data == nil {
		return 0
	}
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if intVal := Int(val); intVal != 0 {
				return intVal
			}
		}
	}
	return 0
}
