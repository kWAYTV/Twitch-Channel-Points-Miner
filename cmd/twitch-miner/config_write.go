package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	configIndent = "  "
	// A list of plain values that fits in this many characters stays on one line.
	inlineListWidth = 72
)

// writeSettings saves the config in the order configLayout gives it.
func writeSettings(path string, settings map[string]interface{}) error {
	raw, err := formatSettings(settings)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// formatSettings renders the config as JSON, laid out as configLayout says.
func formatSettings(settings map[string]interface{}) ([]byte, error) {
	var out bytes.Buffer
	if err := writeObject(&out, settings, configLayout, 0); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func writeObject(out *bytes.Buffer, object map[string]interface{}, shape *layout, depth int) error {
	groups := keyGroups(object, shape)
	if len(groups) == 0 {
		out.WriteString("{}")
		return nil
	}
	inner := strings.Repeat(configIndent, depth+1)
	out.WriteString("{\n")
	for g, group := range groups {
		for k, key := range group {
			name, err := encodeJSON(key, 0)
			if err != nil {
				return err
			}
			out.WriteString(inner)
			out.Write(name)
			out.WriteString(": ")
			if err := writeValue(out, object[key], childLayout(shape, key), depth+1); err != nil {
				return err
			}
			if g < len(groups)-1 || k < len(group)-1 {
				out.WriteByte(',')
			}
			out.WriteByte('\n')
		}
		// A blank line between the groups of the top-level object.
		if depth == 0 && g < len(groups)-1 {
			out.WriteByte('\n')
		}
	}
	out.WriteString(strings.Repeat(configIndent, depth))
	out.WriteByte('}')
	return nil
}

func writeValue(out *bytes.Buffer, value interface{}, shape *layout, depth int) error {
	if object, ok := value.(map[string]interface{}); ok {
		return writeObject(out, object, shape, depth)
	}
	if line, ok := inlineList(value); ok {
		out.WriteString(line)
		return nil
	}
	raw, err := encodeJSON(value, depth)
	if err != nil {
		return err
	}
	out.Write(raw)
	return nil
}

// inlineList renders a short list of plain values, such as event names, on one line.
func inlineList(value interface{}) (string, bool) {
	list, ok := value.([]interface{})
	if !ok || len(list) == 0 {
		return "", false
	}
	items := make([]string, 0, len(list))
	for _, item := range list {
		switch item.(type) {
		case map[string]interface{}, []interface{}:
			return "", false
		}
		raw, err := encodeJSON(item, 0)
		if err != nil {
			return "", false
		}
		items = append(items, string(raw))
	}
	line := "[" + strings.Join(items, ", ") + "]"
	return line, len(line) <= inlineListWidth
}

// keyGroups returns the object's keys in layout order: the known groups first,
// then whatever is left, alphabetically. Empty groups are dropped.
func keyGroups(object map[string]interface{}, shape *layout) [][]string {
	placed := make(map[string]bool, len(object))
	var groups [][]string
	if shape != nil {
		for _, known := range shape.groups {
			var present []string
			for _, key := range known {
				if _, ok := object[key]; ok {
					present = append(present, key)
					placed[key] = true
				}
			}
			if len(present) > 0 {
				groups = append(groups, present)
			}
		}
	}
	var rest []string
	for key := range object {
		if !placed[key] {
			rest = append(rest, key)
		}
	}
	if len(rest) > 0 {
		sort.Strings(rest)
		groups = append(groups, rest)
	}
	return groups
}

func childLayout(shape *layout, key string) *layout {
	if shape == nil {
		return nil
	}
	if child, ok := shape.children[key]; ok {
		return child
	}
	return shape.each
}

// encodeJSON renders a value as indented JSON without the escaping
// encoding/json applies for web pages, so a webhook URL reads as it was typed.
func encodeJSON(value interface{}, depth int) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(strings.Repeat(configIndent, depth), configIndent)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
