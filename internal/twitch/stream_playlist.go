package twitch

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/twitchapi"
)

// parseStreamPlaylist reads an HLS playlist fetched from address and returns the
// URLs it lists, each resolved against address. Only http and https entries with
// a host and no credentials are accepted.
func parseStreamPlaylist(address *url.URL, body io.Reader) ([]*url.URL, error) {
	scanner := bufio.NewScanner(body)
	var entries []*url.URL
	header := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !header {
			if line != "#EXTM3U" {
				return nil, errors.New("invalid stream playlist")
			}
			header = true
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		entry, err := url.Parse(line)
		if err != nil {
			return nil, errors.New("invalid stream playlist entry")
		}
		entry = address.ResolveReference(entry)
		if (entry.Scheme != "http" && entry.Scheme != "https") || entry.Host == "" || entry.User != nil {
			return nil, errors.New("invalid stream playlist entry")
		}
		entries = append(entries, entry)
	}
	if scanner.Err() != nil || len(entries) == 0 {
		return nil, errors.New("unreadable or empty stream playlist")
	}
	return entries, nil
}

// check listed media segments
func (t *Client) sendWatchPlaylist(streamer *domain.Streamer) error {
	masterURL := ""
	if streamer.Stream.PlaylistURL == "" {
		op := twitchapi.ClonePersistedOperation(twitchapi.GQLOperations.PlaybackAccessToken)
		op.Variables = map[string]interface{}{
			"isLive": true, "isVod": false, "login": streamer.Username,
			"platform": "web", "playerType": "site", "vodID": "",
		}
		var response struct {
			Data struct {
				Token *struct {
					Value     string `json:"value"`
					Signature string `json:"signature"`
				} `json:"streamPlaybackAccessToken"`
			} `json:"data"`
		}
		if err := t.PostGQLDecode(op, &response); err != nil {
			return err
		}
		token := response.Data.Token
		if token == nil || token.Value == "" || token.Signature == "" {
			return errors.New("stream playback token unavailable")
		}
		masterURL = "https://usher.ttvnw.net/api/channel/hls/" + url.PathEscape(streamer.Username) + ".m3u8?" +
			url.Values{"sig": {token.Signature}, "token": {token.Value}}.Encode()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if masterURL != "" {
		variants, err := t.readStreamPlaylist(ctx, masterURL)
		if err != nil {
			return err
		}
		selected := variants[len(variants)-1]
		if !strings.HasSuffix(selected.Path, ".m3u8") {
			return errors.New("invalid stream master playlist")
		}
		streamer.Stream.PlaylistURL = selected.String()
	}
	segments, err := t.readStreamPlaylist(ctx, streamer.Stream.PlaylistURL)
	if err != nil {
		streamer.Stream.PlaylistURL = ""
		return err
	}
	// The newest segment is the one a player would fetch next. Asking for
	// every listed segment in turn could outlast the time allowed on a slow
	// connection.
	newest := segments[len(segments)-1]
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, newest.String(), nil)
	if err != nil {
		return errors.New("invalid stream segment URL")
	}
	req.Header.Set("User-Agent", t.userAgent)
	resp, err := t.client.Do(req)
	if err != nil {
		streamer.Stream.PlaylistURL = ""
		return fmt.Errorf("stream segment request failed: %w", causeWithoutURL(err))
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		streamer.Stream.PlaylistURL = ""
		return fmt.Errorf("stream segment returned status %d", resp.StatusCode)
	}
	return nil
}

// causeWithoutURL is why a request failed, without the address it was sent
// to: a stream's address names the channel and carries its access token.
func causeWithoutURL(err error) error {
	var failed *url.Error
	if errors.As(err, &failed) {
		return failed.Err
	}
	return err
}

func (t *Client) readStreamPlaylist(ctx context.Context, address string) ([]*url.URL, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, errors.New("invalid stream playlist URL")
	}
	req.Header.Set("User-Agent", t.userAgent)
	req.Close = true
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stream playlist request failed: %w", causeWithoutURL(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stream playlist returned status %d", resp.StatusCode)
	}
	return parseStreamPlaylist(req.URL, resp.Body)
}
