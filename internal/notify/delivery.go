package notify

import "sync"

// requestFailed is the reason given when a service could not be reached. The
// error itself is never printed: it holds the address, and with it the secret.
const requestFailed = "request failed (network error or timeout)"

// deliveryWarning reports the first delivery problem of a service only, so a
// bad token or address is visible without flooding the log.
type deliveryWarning struct {
	service string
	logf    func(format string, args ...any)
	once    sync.Once
}

func (w *deliveryWarning) warn(reason string) {
	w.once.Do(func() {
		w.logf("%s: notification not delivered: %s (further failures are not logged)", w.service, reason)
	})
}
