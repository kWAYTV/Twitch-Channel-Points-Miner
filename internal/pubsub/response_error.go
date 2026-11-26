package pubsub

import (
	"fmt"
	"path/filepath"
	"strings"
)

// logResponseError reports the error of a RESPONSE message, if it carries one.
func (p *Client) logResponseError(connIndex int, envelope map[string]interface{}) {
	respErr, ok := envelope["error"].(string)
	if !ok || strings.TrimSpace(respErr) == "" {
		return
	}
	nonce := strings.TrimSpace(fmt.Sprint(envelope["nonce"]))
	if nonce != "" {
		p.logger.Errorf("PubSub[%d] RESPONSE error nonce=%s: %s", connIndex, nonce, respErr)
	} else {
		p.logger.Errorf("PubSub[%d] RESPONSE error: %s", connIndex, respErr)
	}
	if strings.Contains(respErr, "ERR_BADAUTH") {
		username := ""
		if p != nil && p.account != nil {
			username = strings.TrimSpace(p.account.LoginName())
		}
		p.logger.Errorf("PubSub[%d] ERR_BADAUTH: most likely you have an outdated cookie file %q. Delete this file and try again.", connIndex, cookieFileOf(username))
	}
}

// cookieFileOf is where the login of username keeps its cookies.
func cookieFileOf(username string) string {
	if username == "" {
		return filepath.Join("cookies", "<username>.json")
	}
	return filepath.Join("cookies", fmt.Sprintf("%s.json", username))
}
