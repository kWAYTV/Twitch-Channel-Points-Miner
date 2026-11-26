package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

type Login struct {
	ClientID  string
	DeviceID  string
	Token     string
	Username  string
	Password  string
	UserAgent string

	client *http.Client
	twitch twitchSide
	userID string
	mu     sync.Mutex
}

// twitchSide is everything a login reaches outside the process: where Twitch
// answers, the clock it waits on, and where the activation code is shown.
type twitchSide struct {
	gqlURL    string
	deviceURL string
	tokenURL  string
	now       func() time.Time
	sleep     func(time.Duration)
	prompt    io.Writer
}

func NewLogin(clientID, deviceID, username, userAgent, password string) (*Login, error) {
	return newLogin(clientID, deviceID, username, userAgent, password, twitchSide{
		gqlURL:    twitchapi.GQLOperations.URL,
		deviceURL: "https://id.twitch.tv/oauth2/device",
		tokenURL:  "https://id.twitch.tv/oauth2/token",
		now:       time.Now,
		sleep:     time.Sleep,
		prompt:    os.Stdout,
	}), nil
}

func newLogin(clientID, deviceID, username, userAgent, password string, twitch twitchSide) *Login {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}
	return &Login{
		ClientID:  clientID,
		DeviceID:  deviceID,
		Username:  username,
		Password:  password,
		UserAgent: userAgent,
		client:    client,
		twitch:    twitch,
	}
}

func (t *Login) Client() *http.Client { return t.client }

func (t *Login) Login(cookiesPath string) error {
	if err := t.loadCookies(cookiesPath); err == nil && t.Token != "" {
		if ok := t.checkLogin(); ok {
			return nil
		}
	}
	if err := t.runDeviceFlow(); err != nil {
		return err
	}
	if err := t.saveCookies(cookiesPath); err != nil {
		return err
	}
	return nil
}

func (t *Login) setToken(token string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Token = token
	t.ensureSessionCookiesLocked()
}

func (t *Login) AuthToken() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Token
}

func (t *Login) UserID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.userID != "" {
		return t.userID
	}
	t.mu.Unlock()
	t.checkLogin()
	t.mu.Lock()
	return t.userID
}

func (t *Login) checkLogin() bool {
	payload := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.GetIDFromLogin)
	if payload.Variables == nil {
		payload.Variables = map[string]interface{}{}
	}
	payload.Variables["login"] = t.Username
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, t.twitch.gqlURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", t.Token))
	req.Header.Set("Client-Id", t.ClientID)
	req.Header.Set("X-Device-Id", t.DeviceID)
	req.Header.Set("User-Agent", t.UserAgent)
	resp, err := t.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var res struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false
	}
	if res.Data.User.ID != "" {
		t.mu.Lock()
		t.userID = res.Data.User.ID
		t.ensureSessionCookiesLocked()
		t.mu.Unlock()
		return true
	}
	return false
}
