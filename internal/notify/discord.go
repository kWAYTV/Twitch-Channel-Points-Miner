package notify

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/project"
)

type DiscordSettings struct {
	WebhookAPI string   `json:"webhook_api"`
	Events     []string `json:"events"`
}

type DiscordWebhook struct {
	webhookAPI string
	events     map[events.Event]struct{}
	client     *http.Client
	warnOnce   sync.Once
}

func NewDiscordWebhook(settings DiscordSettings) *DiscordWebhook {
	webhookAPI := strings.TrimSpace(settings.WebhookAPI)
	if webhookAPI == "" {
		return nil
	}
	selected := make(map[events.Event]struct{})
	for _, raw := range settings.Events {
		event := events.NormalizeEventName(raw)
		if event == "" {
			continue
		}
		selected[event] = struct{}{}
	}
	return &DiscordWebhook{
		webhookAPI: webhookAPI,
		events:     selected,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color"`
}

type discordPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Embeds    []discordEmbed `json:"embeds"`
}

// Send posts the message as an embed if the event is selected: the headline is
// the title and the details are the description. An empty events list sends
// every event except the noisy ones, matching Telegram. Failures never interrupt
// the miner; the webhook URL (it contains a secret) is never logged.
func (d *DiscordWebhook) Send(message Message, event events.Event) {
	if d == nil || !wantsEvent(d.events, event) {
		return
	}
	headline := message.Headline
	if headline == "" {
		headline = eventTitle(event)
	}
	title := limitRunes(eventGlyphPrefix(event)+escapeMarkdown(headline), headlineLimit)
	body := limitRunes(escapeMarkdown(strings.Join(message.Details, detailSeparator)), 4096)
	payload := discordPayload{
		Username:  project.Name,
		AvatarURL: project.AvatarURL,
		Embeds: []discordEmbed{{
			Title:       title,
			Description: body,
			Color:       eventColor(event),
		}},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, d.webhookAPI, bytes.NewReader(raw))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		d.warn("request failed (network error or timeout)")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 180))
	reason := strings.TrimSpace(string(snippet))
	if reason == "" {
		reason = resp.Status
	}
	d.warn(reason)
}

func (d *DiscordWebhook) warn(reason string) {
	d.warnOnce.Do(func() {
		log.Printf("discord: notification not delivered: %s (further failures are not logged)", reason)
	})
}
