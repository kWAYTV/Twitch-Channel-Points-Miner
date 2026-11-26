package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func downloadAsset(url, dir string, logf func(format string, args ...interface{})) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "TwitchChannelPointsMiner-Updater")

	client := newHTTPClient(5 * time.Minute)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		return "", fmt.Errorf("download asset: unexpected status %s", resp.Status)
	}

	temp, err := os.CreateTemp(dir, "miner-update-*")
	if err != nil {
		return "", err
	}
	defer temp.Close()

	var reader io.Reader = resp.Body
	var progress *progressWriter
	if resp.ContentLength > 0 {
		logf("auto-update: downloading update (%.1f MB)...", bytesToMB(resp.ContentLength))
		progress = newProgressWriter(resp.ContentLength, logf)
		reader = io.TeeReader(resp.Body, progress)
	}

	if _, err := io.Copy(temp, reader); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return "", err
	}
	if progress != nil {
		progress.Done()
	}

	return temp.Name(), nil
}
