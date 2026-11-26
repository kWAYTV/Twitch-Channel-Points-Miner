package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func (t *Login) saveCookies(cookiesPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(cookiesPath), 0o755); err != nil {
		return err
	}
	t.ensureSessionCookiesLocked()

	hostnames := []string{"twitch.tv", "id.twitch.tv"}
	store := make(cookieStore)
	for _, host := range hostnames {
		u := &url.URL{Scheme: "https", Host: host}
		for _, c := range t.client.Jar.Cookies(u) {
			if c == nil || c.Name == "" {
				continue
			}
			store[c.Name] = persistedCookie{
				Value:  c.Value,
				Path:   c.Path,
				Domain: c.Domain,
			}
		}
	}

	if _, ok := store["auth-token"]; !ok && t.Token != "" {
		store["auth-token"] = persistedCookie{Value: t.Token}
	}
	if _, ok := store["persistent"]; !ok && t.userID != "" {
		store["persistent"] = persistedCookie{Value: t.userID}
	}

	raw, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cookiesPath, raw, 0o600)
}

func (t *Login) loadCookies(cookiesPath string) error {
	data, err := os.ReadFile(cookiesPath)
	if err != nil {
		return err
	}
	store, err := decodeCookieStore(data)
	if err != nil {
		return err
	}

	sessionCookies := make(map[string][]*http.Cookie)
	for name, c := range store {
		if name == "auth-token" {
			if c.Value != "" {
				t.setToken(c.Value)
			}
			continue
		}
		if name == "persistent" && c.Value != "" && t.userID == "" {
			t.userID = c.Value
		}
		if c.Value == "" {
			continue
		}

		host, cookie := jarCookie(name, c)
		sessionCookies[host] = append(sessionCookies[host], cookie)
	}

	for host, cookies := range sessionCookies {
		u := &url.URL{Scheme: "https", Host: host}
		t.client.Jar.SetCookies(u, append(t.client.Jar.Cookies(u), cookies...))
	}

	// Re-apply auth token and persistent cookies into the jar for future HTTP requests.
	t.mu.Lock()
	t.ensureSessionCookiesLocked()
	t.mu.Unlock()
	return nil
}

// Caller must hold t.mu.
func (t *Login) ensureSessionCookiesLocked() {
	if t.client == nil || t.client.Jar == nil {
		return
	}
	if t.Token != "" {
		t.upsertSessionCookiesLocked("auth-token", t.Token)
	}
	if t.userID != "" {
		t.upsertSessionCookiesLocked("persistent", t.userID)
	}
}

func (t *Login) upsertSessionCookiesLocked(name, value string) {
	if value == "" {
		return
	}
	for _, host := range []string{"twitch.tv", "id.twitch.tv"} {
		u := &url.URL{Scheme: "https", Host: host}
		existing := t.client.Jar.Cookies(u)
		c := &http.Cookie{
			Name:   name,
			Value:  value,
			Path:   "/",
			Domain: "." + host,
		}
		t.client.Jar.SetCookies(u, append(existing, c))
	}
}
