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

// markdownEscaper neutralises Discord's formatting characters, so a streamer
// called d0cc_tv or a chat message with asterisks is shown as written.
var markdownEscaper = strings.NewReplacer(
	`\`, `\\`,
	`*`, `\*`,
	`_`, `\_`,
	`~`, `\~`,
	"`", "\\`",
	`|`, `\|`,
	`>`, `\>`,
)

type DiscordSettings struct {
	WebhookAPI string   `json:"webhook_api"`
	Events     []string `json:"events"`
}

type DiscordWebhook struct {
	webhookAPI string
	events     map[events.Event]struct{}
	client     *http.Client
	logf       func(format string, args ...any)
	warnOnce   sync.Once
}

// NewDiscordWebhook returns nil unless the webhook URL is set.
func NewDiscordWebhook(settings DiscordSettings) *DiscordWebhook {
	return newDiscordWebhook(settings, log.Printf)
}

func newDiscordWebhook(settings DiscordSettings, logf func(format string, args ...any)) *DiscordWebhook {
	webhookAPI := strings.TrimSpace(settings.WebhookAPI)
	if webhookAPI == "" {
		return nil
	}
	return &DiscordWebhook{
		webhookAPI: webhookAPI,
		events:     selectedEvents(settings.Events),
		client:     &http.Client{Timeout: 5 * time.Second},
		logf:       logf,
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
	payload := discordPayload{
		Username:  project.Name,
		AvatarURL: project.AvatarURL,
		Embeds:    []discordEmbed{discordEmbedFor(event, message)},
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
	d.warn(discordFailure(resp.Status, snippet))
}

func (d *DiscordWebhook) warn(reason string) {
	d.warnOnce.Do(func() {
		d.logf("discord: notification not delivered: %s (further failures are not logged)", reason)
	})
}

// discordEmbedFor renders a message as an embed: the polarity glyph and the
// headline as the title, the details on one line as the description.
func discordEmbedFor(event events.Event, message Message) discordEmbed {
	headline := message.Headline
	if headline == "" {
		headline = eventTitle(event)
	}
	return discordEmbed{
		Title:       limitRunes(eventGlyphPrefix(event)+escapeMarkdown(headline), headlineLimit),
		Description: limitRunes(escapeMarkdown(strings.Join(message.Details, detailSeparator)), 4096),
		Color:       eventColor(event),
	}
}

// discordFailure words why Discord refused a message: the start of its answer,
// or the HTTP status when the answer is empty.
func discordFailure(status string, snippet []byte) string {
	if reason := strings.TrimSpace(string(snippet)); reason != "" {
		return reason
	}
	return status
}

func escapeMarkdown(s string) string {
	return markdownEscaper.Replace(s)
}
