package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func mustFormat(t *testing.T, settings map[string]interface{}) string {
	t.Helper()
	raw, err := formatSettings(settings)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	return string(raw)
}

// before reports whether each name appears in text in the order given.
func inOrder(text string, names ...string) bool {
	last := -1
	for _, name := range names {
		i := strings.Index(text, `"`+name+`"`)
		if i <= last {
			return false
		}
		last = i
	}
	return true
}

func TestConfigIsWrittenInGroups(t *testing.T) {
	text := mustFormat(t, defaultConfig())

	if !strings.HasPrefix(text, "{\n  \"username\"") {
		t.Fatalf("the account comes first, got %q", text[:40])
	}
	if !inOrder(text, "username", "streamers", "claim_drops", "betting(make_predictions)", "discord", "telegram", "analytics", "emojis", "debug") {
		t.Fatalf("groups are out of order:\n%s", text)
	}
	if !strings.Contains(text, "  },\n  \"telegram\": {") {
		t.Fatalf("discord and telegram should sit next to each other:\n%s", text)
	}
	if !inOrder(text, "strategy", "percentage", "max_points", "delay_mode", "filter_condition", "by", "where", "value") {
		t.Fatalf("bet settings are out of order:\n%s", text)
	}
	if !inOrder(text, "token", "chat_id") || !inOrder(text, "enabled", "host", "port") {
		t.Fatalf("nested settings are out of order:\n%s", text)
	}
	// Seven groups, so six blank lines, and none inside a nested object.
	if got := strings.Count(text, "\n\n"); got != 6 {
		t.Fatalf("got %d blank lines between groups, want 6:\n%s", got, text)
	}
	if !strings.HasSuffix(text, "}\n") || strings.Contains(text, "\r") {
		t.Fatalf("the file should end with one newline and use LF only")
	}
}

func TestFormattedConfigKeepsEveryValue(t *testing.T) {
	settings := defaultConfig()
	settings["username"] = "me"
	settings["discord"] = map[string]interface{}{"webhook_api": "https://discord.com/api/webhooks/1/a?x=1&y=<2>", "events": []interface{}{"BET_WIN"}}
	settings["streamer_overrides"] = map[string]interface{}{
		"zeta":  map[string]interface{}{"bet": map[string]interface{}{"max_points": 5, "strategy": "SMART"}, "make_predictions": true},
		"alpha": map[string]interface{}{"claim_drops": false},
	}
	settings["an_old_key"] = true
	settings["another_old_key"] = "kept"

	text := mustFormat(t, settings)
	var back map[string]interface{}
	if err := json.Unmarshal([]byte(text), &back); err != nil {
		t.Fatalf("formatted config is not valid JSON: %v\n%s", err, text)
	}
	normalized, _ := json.Marshal(settings)
	var want map[string]interface{}
	_ = json.Unmarshal(normalized, &want)
	if !reflect.DeepEqual(back, want) {
		t.Fatalf("formatting changed a value:\n%s", text)
	}

	if !strings.Contains(text, `"https://discord.com/api/webhooks/1/a?x=1&y=<2>"`) {
		t.Fatalf("a URL should be written as typed:\n%s", text)
	}
	if !inOrder(text, "alpha", "zeta") || !inOrder(text, "make_predictions", "bet", "strategy", "max_points") {
		t.Fatalf("overrides should be sorted by streamer and laid out like the rest:\n%s", text)
	}
	if !inOrder(text, "disable_ssl_cert_verification", "an_old_key", "another_old_key") {
		t.Fatalf("keys the layout does not know go last, alphabetically:\n%s", text)
	}
	if again := mustFormat(t, back); again != text {
		t.Fatalf("formatting a formatted config should change nothing")
	}
}

func TestLoadTidiesAConfigInAnotherLayout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	settings := defaultConfig()
	settings["username"] = "me"
	alphabetical, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, alphabetical, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadOrCreateConfig(path)
	if err != nil || cfg.Username != "me" {
		t.Fatalf("load: cfg=%q err=%v", cfg.Username, err)
	}
	tidied, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(tidied), "{\n  \"username\": \"me\"") {
		t.Fatalf("a complete config in the old layout should be rewritten, got %q", string(tidied)[:40])
	}

	info, _ := os.Stat(path)
	if _, err := loadOrCreateConfig(path); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	again, _ := os.Stat(path)
	if string(after) != string(tidied) || !again.ModTime().Equal(info.ModTime()) {
		t.Fatalf("a config already in layout should not be written again")
	}
}
