package twitch

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"
)

// Watch time is reported to a host that ad and tracker blocklists often
// block. Everything else then keeps working, bonuses, bets and raids, while
// watching earns nothing and the only sign is a connection error on every
// report. So the host is looked up once, when it is first known, and a
// blocked one is named in plain words.

// blockedReason says why a host cannot be reached, from what looking it up
// returned, or "" when it can. A blocklist answers with no address at all,
// with the unspecified address, or with this machine's own.
func blockedReason(addresses []string, err error) string {
	if err != nil {
		return fmt.Sprintf("does not resolve (%v)", err)
	}
	if len(addresses) == 0 {
		return "does not resolve (no addresses)"
	}
	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip == nil || (!ip.IsUnspecified() && !ip.IsLoopback()) {
			return ""
		}
	}
	return fmt.Sprintf("resolves to %s, the answer of a DNS blocklist", addresses[0])
}

// warnIfWatchReportsBlocked looks up the host of the address watch time is
// reported to, the first time it is called, and logs it when the host is
// blocked. Without a resolver it does nothing.
func (t *Client) warnIfWatchReportsBlocked(spadeURL string) {
	if t.lookupHost == nil {
		return
	}
	address, err := url.Parse(spadeURL)
	if err != nil || address.Hostname() == "" {
		return
	}
	t.watchReportCheck.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		reason := blockedReason(t.lookupHost(ctx, address.Hostname()))
		if reason == "" || t.logger == nil {
			return
		}
		t.logger.Errorf(
			`Watching will earn no points: %s, where watch time is reported, %s. Allow it in your DNS filter (see "No points for watching" in docs/troubleshooting.md).`,
			address.Hostname(), reason,
		)
	})
}
