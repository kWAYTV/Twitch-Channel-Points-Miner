package notify

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	delivery            deliveryWarning
}

// NewTelegramNotifier returns nil unless both chat_id and token are set.
func NewTelegramNotifier(settings TelegramSettings) *TelegramNotifier {
	return newTelegramNotifier(settings, telegramAPIBase, log.Printf)
}

func newTelegramNotifier(settings TelegramSettings, apiBase string, logf func(format string, args ...any)) *TelegramNotifier {
	chatID := strings.TrimSpace(settings.ChatID)
	token := strings.TrimSpace(settings.Token)
	if chatID == "" || token == "" {
		return nil
	}
	return &TelegramNotifier{
		endpoint:            apiBase + "/bot" + token + "/sendMessage",
		chatID:              chatID,
		events:              selectedEvents(settings.Events),
		disableNotification: settings.DisableNotification,
		client:              &http.Client{Timeout: 5 * time.Second},
		delivery:            deliveryWarning{service: "telegram", logf: logf},
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
		t.delivery.warn(requestFailed)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	t.delivery.warn(telegramFailure(resp.Status, body))
}

// telegramFailure words why Telegram refused a message: the description in its
// answer, or the HTTP status when it gave none.
func telegramFailure(status string, body []byte) string {
	var apiErr struct {
		Description string `json:"description"`
	}
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Description == "" {
		return status
	}
	return apiErr.Description
}

// formatTelegram renders a message for parse_mode HTML: the polarity glyph and the
// headline in bold, then the details on one line. The glyph is the one the
// console uses, and it is sent even when console emojis are turned off.
func formatTelegram(event events.Event, message Message) string {
	text := eventGlyphPrefix(event) + "<b>" + htmlEscape(limitRunes(headlineOf(event, message), headlineLimit)) + "</b>"
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
