package main

import (
	"encoding/json"
	"testing"
)

func TestTelegramConfigChatIDAcceptsNumberOrString(t *testing.T) {
	for raw, want := range map[string]string{`-1001234567890`: "-1001234567890", `"@mychannel"`: "@mychannel", `" 42 "`: "42", `null`: ""} {
		var cfg telegramConfig
		if err := json.Unmarshal([]byte(`{"chat_id":`+raw+`,"token":"t"}`), &cfg); err != nil {
			t.Fatalf("chat_id %s: %v", raw, err)
		}
		if got := cfg.settings().ChatID; got != want {
			t.Fatalf("chat_id %s got %q, want %q", raw, got, want)
		}
	}
}
