package dashboard

import (
	"net/http"
	"testing"
)

func TestGET_unknown_path(t *testing.T) {
	t.Run("404: a path that is not a route", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/nothing", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
		if got, want := string(body), "404 page not found\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})
}

func TestPOST_root(t *testing.T) {
	t.Run("405: the page is read only", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, app.url+"/", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", resp.StatusCode)
		}
		if got, want := resp.Header.Get("Allow"), "GET, HEAD"; got != want {
			t.Errorf("Allow = %q, want %q", got, want)
		}
		if got, want := string(body), "Method Not Allowed\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})
}
