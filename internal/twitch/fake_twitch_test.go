package twitch

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// reply is one canned answer of the fake: a status and a body, or a failed connection.
type reply struct {
	status int
	body   string
	err    error
}

func ok(body string) reply               { return reply{status: http.StatusOK, body: body} }
func code(status int, body string) reply { return reply{status: status, body: body} }
func down(message string) reply          { return reply{err: errors.New(message)} }

// gqlRequest is one GQL operation the fake received, its variables as JSON with sorted keys.
type gqlRequest struct {
	Operation string
	Variables string
}

// spadePost is one batch of events posted to a spade URL, decoded to its JSON.
type spadePost struct {
	URL    string
	Events string
}

// fakeTwitch plays every Twitch host the client talks to. It sits where the
// client's requests leave the process, as its HTTP transport, because the
// client's URLs are fixed.
//
// A GQL operation is answered by its name and variables, as Twitch answers by
// what was asked. GQL and Helix accept only Token and the miner's client id.
type fakeTwitch struct {
	Token string

	// Requests holds every GQL operation received, in order.
	Requests []gqlRequest
	// Headers holds the headers of the last request received.
	Headers http.Header
	// Tracked holds every batch of events posted to a spade URL.
	Tracked []spadePost
	// Watched holds the URL of every stream segment asked for.
	Watched []string

	answers map[string][]reply
}

func newFakeTwitch(token string) *fakeTwitch {
	return &fakeTwitch{Token: token, answers: map[string][]reply{}}
}

// answer sets the replies to a GQL operation asked with exactly these variables,
// written as a JSON object. Empty variables answer the operation whatever it is
// asked with. Replies are given in order and the last one repeats.
func (f *fakeTwitch) answer(operation, variables string, replies ...reply) {
	key := operation + " *"
	if variables != "" {
		key = operation + " " + sortedJSON([]byte(variables))
	}
	f.answers[key] = replies
}

// serve sets the replies to a request outside GQL, named as "GET https://host/path".
func (f *fakeTwitch) serve(request string, replies ...reply) {
	f.answers[request] = replies
}

// sent returns the variables of every request received for one GQL operation.
func (f *fakeTwitch) sent(operation string) []string {
	var variables []string
	for _, request := range f.Requests {
		if request.Operation == operation {
			variables = append(variables, request.Variables)
		}
	}
	return variables
}

func (f *fakeTwitch) RoundTrip(req *http.Request) (*http.Response, error) {
	var raw []byte
	if req.Body != nil {
		raw, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	f.Headers = req.Header.Clone()
	address := req.URL.String()
	switch {
	case address == twitchapi.GQLOperations.URL:
		return f.gql(req, raw)
	case req.URL.Host == "api.twitch.tv":
		if req.Header.Get("Authorization") != "Bearer "+f.Token || req.Header.Get("Client-Id") != twitchapi.ClientID {
			return respond(req, code(http.StatusUnauthorized, `{"error":"Unauthorized"}`))
		}
	case req.Method == http.MethodPost:
		form, _ := url.ParseQuery(string(raw))
		events, err := base64.StdEncoding.DecodeString(form.Get("data"))
		if err != nil {
			return nil, fmt.Errorf("fake twitch: spade data is not base64: %v", err)
		}
		f.Tracked = append(f.Tracked, spadePost{URL: address, Events: sortedJSON(events)})
	case req.Method == http.MethodHead:
		f.Watched = append(f.Watched, address)
	}
	return f.next(req, req.Method+" "+address)
}

func (f *fakeTwitch) gql(req *http.Request, raw []byte) (*http.Response, error) {
	var body struct {
		OperationName string          `json:"operationName"`
		Variables     json.RawMessage `json:"variables"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("fake twitch: GQL body %q: %v", raw, err)
	}
	variables := "{}"
	if len(body.Variables) > 0 {
		variables = sortedJSON(body.Variables)
	}
	if req.Header.Get("Authorization") != "OAuth "+f.Token || req.Header.Get("Client-Id") != twitchapi.ClientID {
		return respond(req, code(http.StatusUnauthorized, `{"error":"Unauthorized"}`))
	}
	f.Requests = append(f.Requests, gqlRequest{Operation: body.OperationName, Variables: variables})
	key := body.OperationName + " " + variables
	if _, found := f.answers[key]; !found {
		key = body.OperationName + " *"
	}
	return f.next(req, key)
}

func (f *fakeTwitch) next(req *http.Request, key string) (*http.Response, error) {
	replies := f.answers[key]
	if len(replies) == 0 {
		return nil, fmt.Errorf("fake twitch: no answer for %s", strings.TrimSuffix(key, " *"))
	}
	next := replies[0]
	if len(replies) > 1 {
		f.answers[key] = replies[1:]
	}
	return respond(req, next)
}

func respond(req *http.Request, r reply) (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &http.Response{
		StatusCode: r.status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(r.body)),
		Request:    req,
	}, nil
}

// sortedJSON rewrites a JSON value with the keys of every object in order.
func sortedJSON(raw []byte) string {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	out, _ := json.Marshal(value)
	return string(out)
}
