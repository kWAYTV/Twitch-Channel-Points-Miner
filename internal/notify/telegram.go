package notify

import (
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
	return statedFailure(status, body, "description")
}

// formatTelegram renders a message for parse_mode HTML, where a line ends with
// a newline.
func formatTelegram(event events.Event, message Message) string {
	return htmlMessage(event, message, "\n", telegramTextLimit)
}
