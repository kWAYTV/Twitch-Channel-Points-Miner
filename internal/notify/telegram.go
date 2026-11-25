package notify

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
)

// Telegram Bot API: https://core.telegram.org/bots/api#sendmessage
const (
	telegramAPIBase    = "https://api.telegram.org"
	telegramTextLimit  = 4096 // sendMessage accepts 1-4096 characters
	telegramNoPreviews = `{"is_disabled":true}`
)

type TelegramSettings struct {
	ChatID              string   `json:"chat_id"`
	Token               string   `json:"token"`
	Events              []string `json:"events"`
	DisableNotification bool     `json:"disable_notification"`
}

// TelegramNotifier sends selected log events to a Telegram chat through a bot.
type TelegramNotifier struct {
	endpoint            string
	chatID              string
	events              map[events.Event]struct{}
	disableNotification bool
	client              *http.Client
	warnOnce            sync.Once
}

// NewTelegramNotifier returns nil unless both chat_id and token are set.
func NewTelegramNotifier(settings TelegramSettings) *TelegramNotifier {
	return newTelegramNotifier(settings, telegramAPIBase)
}

func newTelegramNotifier(settings TelegramSettings, apiBase string) *TelegramNotifier {
	chatID := strings.TrimSpace(settings.ChatID)
	token := strings.TrimSpace(settings.Token)
	if chatID == "" || token == "" {
		return nil
	}
	selected := make(map[events.Event]struct{})
	for _, raw := range settings.Events {
		if event := events.NormalizeEventName(raw); event != "" {
			selected[event] = struct{}{}
		}
	}
	return &TelegramNotifier{
		endpoint:            apiBase + "/bot" + token + "/sendMessage",
		chatID:              chatID,
		events:              selected,
		disableNotification: settings.DisableNotification,
		client:              &http.Client{Timeout: 5 * time.Second},
	}
}

// Send posts the message if the event is selected. An empty events
// list sends every event except the noisy ones (watch ticks, drop progress,
// bet-filter skips); "*" or "ALL" sends those too. Failures never interrupt the miner.
func (t *TelegramNotifier) Send(message Message, event events.Event) {
	if t == nil || !wantsEvent(t.events, event) {
		return
	}
	text := formatTelegram(event, message)
	if text == "" {
		return
	}
	payload := url.Values{
		"chat_id":              []string{t.chatID},
		"text":                 []string{text},
		"parse_mode":           []string{"HTML"},
		"disable_notification": []string{strconv.FormatBool(t.disableNotification)},
		"link_preview_options": []string{telegramNoPreviews},
	}
	resp, err := t.client.PostForm(t.endpoint, payload)
	if err != nil {
		// The error text contains the request URL and therefore the bot token: do not print it.
		t.warn("request failed (network error or timeout)")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return
	}
	var apiErr struct {
		Description string `json:"description"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Description == "" {
		apiErr.Description = resp.Status
	}
	t.warn(apiErr.Description)
}

// warn reports the first delivery problem only, so a bad token or chat id is
// visible without flooding the log.
func (t *TelegramNotifier) warn(reason string) {
	t.warnOnce.Do(func() {
		log.Printf("telegram: notification not delivered: %s (further failures are not logged)", reason)
	})
}

// formatTelegram renders a message for parse_mode HTML: the polarity glyph and the
// headline in bold, then the details on one line. The glyph is the one the
// console uses, and it is sent even when console emojis are turned off.
func formatTelegram(event events.Event, message Message) string {
	headline := message.Headline
	if headline == "" {
		headline = eventTitle(event)
	}
	text := eventGlyphPrefix(event) + "<b>" + htmlEscape(limitRunes(headline, headlineLimit)) + "</b>"
	details := htmlEscape(strings.Join(message.Details, detailSeparator))
	budget := telegramTextLimit - len([]rune(text)) - 1
	if details == "" || budget < 1 {
		return text
	}
	return text + "\n" + limitEscaped(details, budget)
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// limitEscaped truncates an HTML-escaped string without splitting an entity.
func limitEscaped(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 1 {
		return ""
	}
	cut := string(r[:n-1])
	if i := strings.LastIndex(cut, "&"); i >= 0 && !strings.Contains(cut[i:], ";") {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "…"
}
