package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type persistedCookie struct {
	Value  string `json:"value"`
	Path   string `json:"path,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type cookieStore map[string]persistedCookie

func decodeCookieStore(data []byte) (cookieStore, error) {
	var store cookieStore
	if err := json.Unmarshal(data, &store); err == nil {
		return store, nil
	}

	var legacy []map[string]string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, err
	}

	store = make(cookieStore, len(legacy))
	for _, c := range legacy {
		name := c["name"]
		if name == "" {
			continue
		}
		store[name] = persistedCookie{
			Value:  c["value"],
			Path:   c["path"],
			Domain: c["domain"],
		}
	}

	return store, nil
}

// jarCookie turns a saved cookie into the one sent to Twitch, and the host it
// is kept for. A cookie saved without a domain or path belongs to all of
// twitch.tv.
func jarCookie(name string, c persistedCookie) (string, *http.Cookie) {
	domain := c.Domain
	if domain == "" {
		domain = ".twitch.tv"
	}
	path := c.Path
	if path == "" {
		path = "/"
	}
	cookie := &http.Cookie{
		Name:   name,
		Value:  c.Value,
		Path:   path,
		Domain: domain,
	}
	return strings.TrimPrefix(domain, "."), cookie
}
