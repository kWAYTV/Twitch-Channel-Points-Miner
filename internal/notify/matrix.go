package notify

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/events"
	"github.com/kWAYTV/twitch-points-miner/internal/random"
)

// Matrix client-server API:
// https://spec.matrix.org/latest/client-server-api/#put_matrixclientv3roomsroomidsendeventtypetxnid
const (
	// matrixTextLimit caps the plain text and the HTML of a message, in runes.
	// Both together stay far below the 65536 bytes a Matrix event may take.
	matrixTextLimit = 4096
	// matrixTxnIDBytes is how many random bytes name one message.
	matrixTxnIDBytes = 16
)

type MatrixSettings struct {
	Homeserver  string   `json:"homeserver"`
	AccessToken string   `json:"access_token"`
	RoomID      string   `json:"room_id"`
	Events      []string `json:"events"`
}

// MatrixNotifier sends selected log events to a Matrix room as its account.
type MatrixNotifier struct {
	// sendURL is the address a message is put to, up to the transaction id.
	sendURL     string
	accessToken string
	events      map[events.Event]struct{}
	// newTxnID names one message. The homeserver takes two messages of the same
	// name for one, so every call has to give another.
	newTxnID func() string
	client   *http.Client
	delivery deliveryWarning
}

// NewMatrixNotifier returns nil unless homeserver, access_token and room_id are
// all set and the homeserver is an http or https address.
func NewMatrixNotifier(settings MatrixSettings) *MatrixNotifier {
	return newMatrixNotifier(settings, func() string { return random.Hex(matrixTxnIDBytes) }, log.Printf)
}

func newMatrixNotifier(settings MatrixSettings, newTxnID func() string, logf func(format string, args ...any)) *MatrixNotifier {
	homeserver := strings.TrimSpace(settings.Homeserver)
	accessToken := strings.TrimSpace(settings.AccessToken)
	roomID := strings.TrimSpace(settings.RoomID)
	if homeserver == "" || accessToken == "" || roomID == "" {
		return nil
	}
	sendURL, ok := matrixSendURL(homeserver, roomID)
	if !ok {
		logf("matrix: notifications are off: homeserver has to be an address starting with http:// or https://")
		return nil
	}
	return &MatrixNotifier{
		sendURL:     sendURL,
		accessToken: accessToken,
		events:      selectedEvents(settings.Events),
		newTxnID:    newTxnID,
		client:      &http.Client{Timeout: 5 * time.Second},
		delivery:    deliveryWarning{service: "matrix", logf: logf},
	}
}

// matrixSendURL is where messages for a room are put, ending in the slash the
// transaction id follows. The homeserver may end in a slash or not. It reports
// false for a homeserver that is not an http or https address.
func matrixSendURL(homeserver, roomID string) (string, bool) {
	base := strings.TrimRight(homeserver, "/")
	parsed, err := url.Parse(base)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false
	}
	return base + "/_matrix/client/v3/rooms/" + url.PathEscape(roomID) + "/send/m.room.message/", true
}

// Send puts the message in the room if the event is selected. An empty events
// list sends every event except the noisy ones, matching Telegram. Failures
// never interrupt the miner; the access token is never logged.
func (m *MatrixNotifier) Send(message Message, event events.Event) {
	if m == nil || !wantsEvent(m.events, event) {
		return
	}
	req, err := http.NewRequest(http.MethodPut, m.sendURL+url.PathEscape(m.newTxnID()), bytes.NewReader(matrixBody(event, message)))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		m.delivery.warn(requestFailed)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	m.delivery.warn(statedFailure(resp.Status, body, "error"))
}

// matrixContent is the content of an m.room.message event of type m.text.
type matrixContent struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

// matrixBody is the JSON a message is sent as. The characters HTML reserves are
// left as they are, which JSON allows.
func matrixBody(event events.Event, message Message) []byte {
	plain, html := formatMatrix(event, message)
	var raw bytes.Buffer
	encoder := json.NewEncoder(&raw)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(matrixContent{
		MsgType:       "m.text",
		Body:          plain,
		Format:        "org.matrix.custom.html",
		FormattedBody: html,
	})
	return bytes.TrimSuffix(raw.Bytes(), []byte("\n"))
}

// formatMatrix renders a message twice, as Matrix wants it: plain text for
// clients that show no formatting, and HTML with the headline in bold. Both
// have the polarity glyph, the headline, then the details on a line of their own.
func formatMatrix(event events.Event, message Message) (plain, html string) {
	return plainMessage(event, message, matrixTextLimit), htmlMessage(event, message, "<br>", matrixTextLimit)
}

// plainMessage renders a message as plain text of at most limit runes: the
// polarity glyph and the headline, then the details on the next line.
func plainMessage(event events.Event, message Message, limit int) string {
	text := eventGlyphPrefix(event) + headlineText(event, message)
	details := detailLine(message)
	budget := limit - len([]rune(text)) - 1
	if details == "" || budget < 1 {
		return text
	}
	return text + "\n" + limitRunes(details, budget)
}
