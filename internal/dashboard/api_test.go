package dashboard

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGET_api_streamers(t *testing.T) {
	t.Run("200: no streamers", func(t *testing.T) {
		own := startEmptyApp(t)

		req, _ := http.NewRequest(http.MethodGet, own.url+"/api/streamers", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"now":1700000000000,"streamers":[],` +
			`"totals":{"points":0,"session_gain":0,"online":0,"count":0}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("200: two streamers, one of them online", func(t *testing.T) {
		own := startEmptyApp(t)
		own.store.RecordPoint("alice", 1000, 0, "")
		own.clock.Set(t0 + minute)
		own.store.RecordPoint("alice", 1010, 10, "WATCH")
		own.clock.Set(t0 + 2*minute)
		own.store.RecordSnapshot("bob", 500, true)

		req, _ := http.NewRequest(http.MethodGet, own.url+"/api/streamers", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"now":1700000120000,"streamers":[` +
			`{"name":"alice","points":1010,"session_gain":10,"online":false,"updated":1700000060000},` +
			`{"name":"bob","points":500,"session_gain":0,"online":true,"updated":1700000120000}],` +
			`"totals":{"points":1510,"session_gain":10,"online":1,"count":2}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
		wantHeader := map[string]string{
			"Content-Type":     "application/json",
			"Cache-Control":    "no-store",
			"Vary":             "Accept-Encoding",
			"Content-Encoding": "",
		}
		if got := headersOf(resp, wantHeader); !reflect.DeepEqual(got, wantHeader) {
			t.Errorf("headers = %v, want %v", got, wantHeader)
		}
	})
}

func TestGET_api_streamer_name(t *testing.T) {
	t.Run("200: the last day, when no range is given", func(t *testing.T) {
		name := app.newName("day")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 1000, 0, "")
		app.clock.Set(t0 + minute)
		app.store.RecordPoint(name, 1450, 450, "WATCH_STREAK")
		app.clock.Set(t0 + 2*minute)
		app.store.RecordEvent("STREAMER_ONLINE", name+" is online.")
		app.clock.Set(t0 + hour)

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name, nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"name":"` + name + `","range":"24h","from":1699917200000,"to":1700003600000,` +
			`"series":[[1700000000000,1000],[1700000060000,1450]],` +
			`"events":[{"t":1700000060000,"type":"WATCH_STREAK","msg":"Gained 450 points."},` +
			`{"t":1700000120000,"type":"STREAMER_ONLINE","msg":"` + name + ` is online."}],` +
			`"events_truncated":false,` +
			`"stats":{"start":1000,"end":1450,"gain":450,"min":1000,"max":1450}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
		wantHeader := map[string]string{
			"Content-Type":     "application/json",
			"Cache-Control":    "no-store",
			"Vary":             "Accept-Encoding",
			"Content-Encoding": "",
		}
		if got := headersOf(resp, wantHeader); !reflect.DeepEqual(got, wantHeader) {
			t.Errorf("headers = %v, want %v", got, wantHeader)
		}
	})

	t.Run("200: everything recorded", func(t *testing.T) {
		name := app.newName("all")
		app.clock.Set(t0 + 5000)
		app.store.RecordPoint(name, 1000, 0, "")
		app.clock.Set(t0 + 10_000)

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name+"?range=all", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"name":"` + name + `","range":"all","from":1700000005000,"to":1700000010000,` +
			`"series":[[1700000005000,1000]],` +
			`"events":[],"events_truncated":false,` +
			`"stats":{"start":1000,"end":1000,"gain":0,"min":1000,"max":1000}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("200: a week that starts after the first balance", func(t *testing.T) {
		name := app.newName("week")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 900, 0, "")
		app.clock.Set(t0 + 8*day)
		app.store.RecordPoint(name, 1000, 100, "WATCH")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name+"?range=7d", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"name":"` + name + `","range":"7d","from":1700086400000,"to":1700691200000,` +
			`"series":[[1700086400000,900],[1700691200000,1000]],` +
			`"events":[],"events_truncated":false,` +
			`"stats":{"start":900,"end":1000,"gain":100,"min":900,"max":1000}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("200: the name in capitals", func(t *testing.T) {
		name := app.newName("capitals")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 1000, 0, "")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+strings.ToUpper(name)+"?range=all", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		want := `{"name":"` + name + `","range":"all","from":1700000000000,"to":1700000000000,` +
			`"series":[[1700000000000,1000]],` +
			`"events":[],"events_truncated":false,` +
			`"stats":{"start":1000,"end":1000,"gain":0,"min":1000,"max":1000}}`
		if got := string(body); got != want {
			t.Errorf("body = %s, want %s", got, want)
		}
	})

	t.Run("200: sixty points, to a client that takes gzip", func(t *testing.T) {
		name := app.newName("sixty")
		for i := int64(0); i < 60; i++ {
			app.clock.Set(t0 + i*minute)
			app.store.RecordPoint(name, 1000+int(i)*10, 10, "WATCH")
		}

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name+"?range=all", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
			t.Fatalf("Content-Encoding = %q, want %q", got, "gzip")
		}
		plainReq, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name+"?range=all", nil)
		plainResp, plain := send(t, plainReq)
		if got := plainResp.Header.Get("Content-Encoding"); got != "" {
			t.Errorf("Content-Encoding without Accept-Encoding = %q, want none", got)
		}
		if got := string(gunzip(t, body)); got != string(plain) {
			t.Errorf("unpacked body = %s, want the plain body %s", got, plain)
		}
		if len(body) >= len(plain) {
			t.Errorf("packed body is %d bytes, want fewer than the plain body's %d", len(body), len(plain))
		}
	})

	t.Run("400: a range it does not know", func(t *testing.T) {
		name := app.newName("range")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 1000, 0, "")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name+"?range=1y", nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", resp.StatusCode)
		}
		if got, want := string(body), "range must be one of 24h, 7d, 30d, all\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})

	t.Run("404: a streamer nobody recorded", func(t *testing.T) {
		name := app.newName("nobody")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name, nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
		if got, want := string(body), "unknown streamer\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})

	t.Run("500: a points file that can't be read", func(t *testing.T) {
		name := app.newName("points")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 1000, 0, "")
		replaceWithFolder(t, app.dir, name+".points.jsonl")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name, nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
		if got, want := string(body), "failed to read points\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})

	t.Run("500: an events file that can't be read", func(t *testing.T) {
		name := app.newName("events")
		app.clock.Set(t0)
		app.store.RecordPoint(name, 1000, 0, "")
		replaceWithFolder(t, app.dir, name+".events.jsonl")

		req, _ := http.NewRequest(http.MethodGet, app.url+"/api/streamer/"+name, nil)
		resp, body := send(t, req)

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
		if got, want := string(body), "failed to read events\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})
}
