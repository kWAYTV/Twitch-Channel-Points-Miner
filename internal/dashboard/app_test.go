package dashboard

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
)

// testApp is the dashboard as it runs: a store on a folder of its own, served
// over HTTP. The store is the side the miner writes to, so a test arranges
// through it, moves the clock, and otherwise only sends requests.
type testApp struct {
	url   string
	dir   string
	store *Store
	clock *fakeClock
	close func()
	names atomic.Int64
}

// newName returns a streamer name no other test, and no earlier run of this
// one, has used.
func (a *testApp) newName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, a.names.Add(1))
}

func startApp(dir string) (*testApp, error) {
	clock := newFakeClock(t0)
	store, err := newStore(dir, clock.Now)
	if err != nil {
		return nil, err
	}
	server := httptest.NewServer(NewServer(store).Handler())
	return &testApp{url: server.URL, dir: dir, store: store, clock: clock, close: server.Close}, nil
}

// app is shared by the route tests: each one records a streamer of its own and
// sets the clock before it sends its request.
var app *testApp

func TestMain(m *testing.M) {
	os.Exit(runWithApp(m))
}

func runWithApp(m *testing.M) int {
	dir, err := os.MkdirTemp("", "dashboard-app")
	if err != nil {
		fmt.Fprintln(os.Stderr, "dashboard tests:", err)
		return 1
	}
	defer os.RemoveAll(dir)
	app, err = startApp(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dashboard tests:", err)
		return 1
	}
	defer app.close()
	return m.Run()
}

// startEmptyApp starts an app with a store of its own, for a route that
// answers about every streamer.
func startEmptyApp(t *testing.T) *testApp {
	t.Helper()
	own, err := startApp(t.TempDir())
	if err != nil {
		t.Fatalf("start app: %v", err)
	}
	t.Cleanup(own.close)
	return own
}

// Compression is asked for by hand, so the client does not undo it.
var client = &http.Client{Transport: &http.Transport{DisableCompression: true}}

func send(t *testing.T, req *http.Request) (*http.Response, []byte) {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: read body: %v", req.Method, req.URL, err)
	}
	return resp, body
}

// headersOf reads from a response the headers named in want.
func headersOf(resp *http.Response, want map[string]string) map[string]string {
	got := map[string]string{}
	for name := range want {
		got[name] = resp.Header.Get(name)
	}
	return got
}
