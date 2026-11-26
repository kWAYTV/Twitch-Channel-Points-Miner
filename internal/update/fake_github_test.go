package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type githubRequest struct {
	Path      string
	Accept    string
	UserAgent string
}

// fakeGitHub serves a repository's latest release at /releases/latest and its
// assets under /assets/, and keeps the requests it receives.
type fakeGitHub struct {
	URL string

	// ReleaseStatus and ReleaseBody answer /releases/latest.
	ReleaseStatus int
	ReleaseBody   string
	// Assets are served by name; AssetStatus answers every asset when set.
	Assets      map[string]string
	AssetStatus int
	// Unsized serves assets without a Content-Length, and CutShort promises
	// more bytes than it sends.
	Unsized  bool
	CutShort bool
	// Dropped closes every connection without an answer.
	Dropped bool

	mu       sync.Mutex
	requests []githubRequest
}

func startFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()
	github := &fakeGitHub{ReleaseStatus: http.StatusNotFound, Assets: map[string]string{}}
	server := httptest.NewServer(http.HandlerFunc(github.serve))
	t.Cleanup(server.Close)
	github.URL = server.URL
	return github
}

type fakeAsset struct {
	Name    string
	Content string
}

// Publish makes tag the latest release, with the given assets.
func (g *fakeGitHub) Publish(tag string, assets ...fakeAsset) {
	entries := make([]string, 0, len(assets))
	for _, asset := range assets {
		g.Assets[asset.Name] = asset.Content
		entries = append(entries, fmt.Sprintf(`{"name": %q, "browser_download_url": %q}`, asset.Name, g.AssetURL(asset.Name)))
	}
	g.ReleaseStatus = http.StatusOK
	g.ReleaseBody = fmt.Sprintf(`{"tag_name": %q, "assets": [%s]}`, tag, strings.Join(entries, ", "))
}

func (g *fakeGitHub) ReleasesURL() string {
	return g.URL + "/releases/latest"
}

func (g *fakeGitHub) AssetURL(name string) string {
	return g.URL + "/assets/" + name
}

func (g *fakeGitHub) Requests() []githubRequest {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]githubRequest(nil), g.requests...)
}

func (g *fakeGitHub) serve(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	g.requests = append(g.requests, githubRequest{
		Path:      r.URL.Path,
		Accept:    r.Header.Get("Accept"),
		UserAgent: r.Header.Get("User-Agent"),
	})
	g.mu.Unlock()

	if g.Dropped {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			conn.Close()
		}
		return
	}
	if r.URL.Path == "/releases/latest" {
		w.WriteHeader(g.ReleaseStatus)
		fmt.Fprint(w, g.ReleaseBody)
		return
	}
	g.serveAsset(w, strings.TrimPrefix(r.URL.Path, "/assets/"))
}

func (g *fakeGitHub) serveAsset(w http.ResponseWriter, name string) {
	content, ok := g.Assets[name]
	switch {
	case g.AssetStatus != 0:
		w.WriteHeader(g.AssetStatus)
	case !ok:
		w.WriteHeader(http.StatusNotFound)
	case g.CutShort:
		w.Header().Set("Content-Length", fmt.Sprint(len(content)+5))
		fmt.Fprint(w, content)
	case g.Unsized:
		w.(http.Flusher).Flush()
		fmt.Fprint(w, content)
	default:
		w.Header().Set("Content-Length", fmt.Sprint(len(content)))
		fmt.Fprint(w, content)
	}
}
