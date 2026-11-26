package twitch

import (
	"context"
	"errors"
	"testing"
)

func TestBlockedReason(t *testing.T) {
	tests := []struct {
		name      string
		addresses []string
		err       error
		want      string
	}{
		{"a real address", []string{"52.11.230.218"}, nil, ""},
		{"several real addresses", []string{"52.11.230.218", "54.201.61.216"}, nil, ""},
		{"a real IPv6 address", []string{"2600:1f14:abc::1"}, nil, ""},
		{"the unspecified address", []string{"0.0.0.0"}, nil, "resolves to 0.0.0.0, the answer of a DNS blocklist"},
		{"the unspecified IPv6 address", []string{"::"}, nil, "resolves to ::, the answer of a DNS blocklist"},
		{"both unspecified addresses", []string{"::", "0.0.0.0"}, nil, "resolves to ::, the answer of a DNS blocklist"},
		{"this machine itself", []string{"127.0.0.1"}, nil, "resolves to 127.0.0.1, the answer of a DNS blocklist"},
		{"a real address among blocked ones", []string{"0.0.0.0", "52.11.230.218"}, nil, ""},
		{"a lookup that fails", nil, errors.New("no such host"), "does not resolve (no such host)"},
		{"no addresses and no error", nil, nil, "does not resolve (no addresses)"},
		{"an address that is not one", []string{"not-an-address"}, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := blockedReason(tt.addresses, tt.err); got != tt.want {
				t.Errorf("blockedReason(%v, %v) = %q, want %q", tt.addresses, tt.err, got, tt.want)
			}
		})
	}
}

// lookups is a resolver that answers every host the same way and keeps the
// hosts it was asked for.
type lookups struct {
	Addresses []string
	Err       error
	Asked     []string
}

func (l *lookups) host(_ context.Context, host string) ([]string, error) {
	l.Asked = append(l.Asked, host)
	return l.Addresses, l.Err
}

func TestClient_warnIfWatchReportsBlocked(t *testing.T) {
	const blockedLine = `error: Watching will earn no points: spade.example, where watch time is reported, resolves to 0.0.0.0, the answer of a DNS blocklist. Allow it in your DNS filter (see "No points for watching" in docs/troubleshooting.md).`

	t.Run("a blocked host", func(t *testing.T) {
		client, fake := setup(t)
		resolver := &lookups{Addresses: []string{"0.0.0.0"}}
		client.lookupHost = resolver.host

		client.warnIfWatchReportsBlocked("https://spade.example/track")

		if want := []string{blockedLine}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
		if want := []string{"spade.example"}; !equal(resolver.Asked, want) {
			t.Errorf("hosts looked up = %q, want %q", resolver.Asked, want)
		}
	})

	t.Run("a blocked host, found for a second streamer", func(t *testing.T) {
		client, fake := setup(t)
		resolver := &lookups{Addresses: []string{"0.0.0.0"}}
		client.lookupHost = resolver.host
		client.warnIfWatchReportsBlocked("https://spade.example/track")

		client.warnIfWatchReportsBlocked("https://spade.example/track")

		if want := []string{blockedLine}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want it once: %q", fake.logger.Lines, want)
		}
		if want := []string{"spade.example"}; !equal(resolver.Asked, want) {
			t.Errorf("hosts looked up = %q, want one lookup", resolver.Asked)
		}
	})

	t.Run("a host that does not resolve", func(t *testing.T) {
		client, fake := setup(t)
		resolver := &lookups{Err: errors.New("no such host")}
		client.lookupHost = resolver.host

		client.warnIfWatchReportsBlocked("https://spade.example/track")

		want := []string{`error: Watching will earn no points: spade.example, where watch time is reported, does not resolve (no such host). Allow it in your DNS filter (see "No points for watching" in docs/troubleshooting.md).`}
		if !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("a reachable host", func(t *testing.T) {
		client, fake := setup(t)
		resolver := &lookups{Addresses: []string{"52.11.230.218"}}
		client.lookupHost = resolver.host

		client.warnIfWatchReportsBlocked("https://spade.example/track")

		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})

	t.Run("a host with a port", func(t *testing.T) {
		client, _ := setup(t)
		resolver := &lookups{Addresses: []string{"52.11.230.218"}}
		client.lookupHost = resolver.host

		client.warnIfWatchReportsBlocked("https://spade.example:8443/track")

		if want := []string{"spade.example"}; !equal(resolver.Asked, want) {
			t.Errorf("hosts looked up = %q, want %q", resolver.Asked, want)
		}
	})

	t.Run("an address that is not a URL", func(t *testing.T) {
		client, fake := setup(t)
		resolver := &lookups{Addresses: []string{"0.0.0.0"}}
		client.lookupHost = resolver.host

		client.warnIfWatchReportsBlocked("::not a url::")

		if len(fake.logger.Lines) != 0 || len(resolver.Asked) != 0 {
			t.Errorf("logged %q and looked up %q, want neither", fake.logger.Lines, resolver.Asked)
		}
	})

	t.Run("no resolver", func(t *testing.T) {
		client, fake := setup(t)

		client.warnIfWatchReportsBlocked("https://spade.example/track")

		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})
}

func TestClient_GetSpadeURL_checksTheHost(t *testing.T) {
	client, fake := setup(t)
	resolver := &lookups{Addresses: []string{"0.0.0.0"}}
	client.lookupHost = resolver.host
	fake.twitch.serve(channelPage, ok(channelPageBody))
	fake.twitch.serve(settingsScript, ok(settingsScriptBody))

	if err := client.GetSpadeURL(newStreamer()); err != nil {
		t.Fatalf("GetSpadeURL error = %v, want none", err)
	}

	if want := []string{"spade.example"}; !equal(resolver.Asked, want) {
		t.Errorf("hosts looked up = %q, want %q", resolver.Asked, want)
	}
}
