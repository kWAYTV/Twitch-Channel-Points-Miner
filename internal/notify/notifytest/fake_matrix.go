package notifytest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// MatrixRequest is one message the fake was sent, as it arrived.
type MatrixRequest struct {
	Method string
	// Path is the path as it was written on the wire, with its escapes.
	Path          string
	Authorization string
	ContentType   string
	Body          string
}

// FakeMatrix is a Matrix homeserver's endpoint for sending a message, for one
// account in one room. A request with another access token gets the
// homeserver's 401, one for another room its 403, and neither is kept.
type FakeMatrix struct {
	// URL is the homeserver address to configure.
	URL string
	// Token is the only access token the fake accepts.
	Token string
	// RoomID is the only room the account is in.
	RoomID string

	mu       sync.Mutex
	requests []MatrixRequest
	status   int
	answer   string
	drop     bool
}

// NewFakeMatrix starts the fake and stops it when the test ends.
func NewFakeMatrix(t testing.TB) *FakeMatrix {
	t.Helper()
	fake := &FakeMatrix{Token: "syt_secret", RoomID: "!room:example.org"}
	server := httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(server.Close)
	fake.URL = server.URL
	return fake
}

// Requests returns what the fake has received, oldest first.
func (f *FakeMatrix) Requests() []MatrixRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]MatrixRequest(nil), f.requests...)
}

// Refuse makes every later request fail the way a homeserver reports an error.
func (f *FakeMatrix) Refuse(status int, errcode, reason string) {
	f.RefuseWith(status, matrixErrorAnswer(errcode, reason))
}

// RefuseWith makes every later request fail with this status and raw answer.
func (f *FakeMatrix) RefuseWith(status int, answer string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer = status, answer
}

// Accept makes later requests succeed again.
func (f *FakeMatrix) Accept() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.answer, f.drop = 0, "", false
}

// Drop makes the fake close every later connection without answering.
func (f *FakeMatrix) Drop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drop = true
}

func (f *FakeMatrix) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.drop {
		dropConnection(w)
		return
	}
	roomID, ok := matrixSendRoom(r.URL.EscapedPath())
	if !ok {
		matrixError(w, http.StatusNotFound, "M_UNRECOGNIZED", "Unrecognized request")
		return
	}
	if r.Method != http.MethodPut {
		matrixError(w, http.StatusMethodNotAllowed, "M_UNRECOGNIZED", "Unrecognized request")
		return
	}
	if r.Header.Get("Authorization") != "Bearer "+f.Token {
		matrixError(w, http.StatusUnauthorized, "M_UNKNOWN_TOKEN", "Invalid access token passed.")
		return
	}
	if roomID != f.RoomID {
		matrixError(w, http.StatusForbidden, "M_FORBIDDEN", "User not in room "+roomID)
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, MatrixRequest{
		Method:        r.Method,
		Path:          r.URL.EscapedPath(),
		Authorization: r.Header.Get("Authorization"),
		ContentType:   r.Header.Get("Content-Type"),
		Body:          string(body),
	})
	if f.status != 0 {
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.answer)
		return
	}
	_, _ = io.WriteString(w, `{"event_id":"$sent"}`)
}

// matrixSendRoom reads the room out of the path of a request to send a message:
// /_matrix/client/v3/rooms/{roomId}/send/m.room.message/{txnId}, each part escaped.
func matrixSendRoom(escapedPath string) (roomID string, ok bool) {
	rest, found := strings.CutPrefix(escapedPath, "/_matrix/client/v3/rooms/")
	if !found {
		return "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 4 || parts[1] != "send" || parts[2] != "m.room.message" || parts[3] == "" {
		return "", false
	}
	roomID, err := url.PathUnescape(parts[0])
	if err != nil || roomID == "" {
		return "", false
	}
	return roomID, true
}

func matrixError(w http.ResponseWriter, status int, errcode, reason string) {
	w.WriteHeader(status)
	_, _ = io.WriteString(w, matrixErrorAnswer(errcode, reason))
}

// matrixErrorAnswer is the body a homeserver reports an error with.
func matrixErrorAnswer(errcode, reason string) string {
	return `{"errcode":` + strconv.Quote(errcode) + `,"error":` + strconv.Quote(reason) + `}`
}
