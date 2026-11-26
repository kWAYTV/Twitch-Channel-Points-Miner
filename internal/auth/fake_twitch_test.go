package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type gqlRequest struct {
	ContentType   string
	Authorization string
	ClientID      string
	DeviceID      string
	UserAgent     string
	Body          string
}

type oauthRequest struct {
	ContentType string
	ClientID    string
	DeviceID    string
	UserAgent   string
	Origin      string
	Referer     string
	Form        string
}

type answer struct {
	Status int
	Body   string
}

// fakeTwitch answers the three endpoints a login calls, and keeps what each
// one receives.
type fakeTwitch struct {
	URL string

	// Users maps an OAuth token to the user ID GQL answers for it.
	Users map[string]string
	// GQL, when set, answers every GQL request instead.
	GQL *answer
	// Device answers the request that starts the device flow.
	Device answer
	// Tokens answers each poll for a token in turn. Once they run out, the
	// authorization is still pending.
	Tokens []answer
	// Dropped closes every connection to a path in it without an answer.
	Dropped map[string]bool

	mu             sync.Mutex
	gqlRequests    []gqlRequest
	deviceRequests []oauthRequest
	tokenRequests  []oauthRequest
}

const (
	deviceGranted = `{"device_code": "device-code-1", "user_code": "ABCD1234", "interval": 5, "expires_in": 1800}`
	pending       = `{"status": 400, "message": "authorization_pending"}`
)

func startFakeTwitch(t *testing.T) *fakeTwitch {
	t.Helper()
	twitch := &fakeTwitch{
		Users:   map[string]string{},
		Device:  answer{Status: http.StatusOK, Body: deviceGranted},
		Dropped: map[string]bool{},
	}
	server := httptest.NewServer(http.HandlerFunc(twitch.serve))
	t.Cleanup(server.Close)
	twitch.URL = server.URL
	return twitch
}

// Grant answers the next poll for a token with an access token.
func (f *fakeTwitch) Grant(token string) {
	f.Tokens = append(f.Tokens, answer{Status: http.StatusOK, Body: fmt.Sprintf(`{"access_token": %q}`, token)})
}

// Pend answers the next poll for a token with "not authorized yet".
func (f *fakeTwitch) Pend() {
	f.Tokens = append(f.Tokens, answer{Status: http.StatusBadRequest, Body: pending})
}

func (f *fakeTwitch) GQLRequests() []gqlRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]gqlRequest(nil), f.gqlRequests...)
}

func (f *fakeTwitch) DeviceRequests() []oauthRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]oauthRequest(nil), f.deviceRequests...)
}

func (f *fakeTwitch) TokenRequests() []oauthRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]oauthRequest(nil), f.tokenRequests...)
}

func (f *fakeTwitch) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	reply := f.record(r, string(body))
	if f.Dropped[r.URL.Path] {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			conn.Close()
		}
		return
	}
	w.WriteHeader(reply.Status)
	fmt.Fprint(w, reply.Body)
}

func (f *fakeTwitch) record(r *http.Request, body string) answer {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch r.URL.Path {
	case "/gql":
		f.gqlRequests = append(f.gqlRequests, gqlRequest{
			ContentType:   r.Header.Get("Content-Type"),
			Authorization: r.Header.Get("Authorization"),
			ClientID:      r.Header.Get("Client-Id"),
			DeviceID:      r.Header.Get("X-Device-Id"),
			UserAgent:     r.Header.Get("User-Agent"),
			Body:          body,
		})
		if f.GQL != nil {
			return *f.GQL
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "OAuth ")
		if id, ok := f.Users[token]; ok {
			return answer{Status: http.StatusOK, Body: fmt.Sprintf(`{"data": {"user": {"id": %q}}}`, id)}
		}
		return answer{Status: http.StatusOK, Body: `{"data": {"user": null}}`}
	case "/oauth2/device":
		f.deviceRequests = append(f.deviceRequests, oauthRequestOf(r, body))
		return f.Device
	case "/oauth2/token":
		f.tokenRequests = append(f.tokenRequests, oauthRequestOf(r, body))
		if len(f.Tokens) == 0 {
			return answer{Status: http.StatusBadRequest, Body: pending}
		}
		next := f.Tokens[0]
		f.Tokens = f.Tokens[1:]
		return next
	}
	return answer{Status: http.StatusNotFound}
}

func oauthRequestOf(r *http.Request, body string) oauthRequest {
	return oauthRequest{
		ContentType: r.Header.Get("Content-Type"),
		ClientID:    r.Header.Get("Client-Id"),
		DeviceID:    r.Header.Get("X-Device-Id"),
		UserAgent:   r.Header.Get("User-Agent"),
		Origin:      r.Header.Get("Origin"),
		Referer:     r.Header.Get("Referer"),
		Form:        body,
	}
}
