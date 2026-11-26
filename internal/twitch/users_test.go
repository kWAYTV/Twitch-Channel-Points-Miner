package twitch

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

const helixUser = "GET https://api.twitch.tv/helix/users?id=123456"

func TestClient_GetUserByID(t *testing.T) {
	t.Run("a known channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(helixUser, ok(`{"data":[{"id":"123456","login":" NewLogin ","display_name":" New Login "}]}`))

		user, err := client.GetUserByID("123456")

		if err != nil {
			t.Fatalf("GetUserByID error = %v, want none", err)
		}
		if want := (&TwitchUserIdentity{ID: "123456", Login: "newlogin", DisplayName: "New Login"}); !equal(user, want) {
			t.Errorf("GetUserByID = %+v, want %+v", user, want)
		}
		wantHeaders := http.Header{"Authorization": {"Bearer token"}, "Client-Id": {"ue6666qo983tsx6so1t0vnawi233wa"}, "User-Agent": {"agent"}}
		if !equal(fake.twitch.Headers, wantHeaders) {
			t.Errorf("request headers = %v, want %v", fake.twitch.Headers, wantHeaders)
		}
	})

	t.Run("an id with padding", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(helixUser, ok(`{"data":[{"id":"123456","login":"newlogin","display_name":"NewLogin"}]}`))

		user, err := client.GetUserByID(" 123456 ")

		if err != nil {
			t.Fatalf("GetUserByID error = %v, want none", err)
		}
		if want := (&TwitchUserIdentity{ID: "123456", Login: "newlogin", DisplayName: "NewLogin"}); !equal(user, want) {
			t.Errorf("GetUserByID = %+v, want %+v", user, want)
		}
	})

	t.Run("an id that needs escaping", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve("GET https://api.twitch.tv/helix/users?id=12%2634", ok(`{"data":[{"id":"12&34","login":"odd","display_name":"Odd"}]}`))

		user, err := client.GetUserByID("12&34")

		if err != nil {
			t.Fatalf("GetUserByID error = %v, want none", err)
		}
		if want := (&TwitchUserIdentity{ID: "12&34", Login: "odd", DisplayName: "Odd"}); !equal(user, want) {
			t.Errorf("GetUserByID = %+v, want %+v", user, want)
		}
	})

	t.Run("a client with no user agent", func(t *testing.T) {
		twitch := newFakeTwitch("token")
		twitch.serve(helixUser, ok(`{"data":[{"id":"123456","login":"newlogin","display_name":"NewLogin"}]}`))
		client := newClient(clientParts{session: &fakeSession{Token: "token"}, http: &http.Client{Transport: twitch}})

		_, err := client.GetUserByID("123456")

		if err != nil {
			t.Fatalf("GetUserByID error = %v, want none", err)
		}
		wantHeaders := http.Header{"Authorization": {"Bearer token"}, "Client-Id": {"ue6666qo983tsx6so1t0vnawi233wa"}}
		if !equal(twitch.Headers, wantHeaders) {
			t.Errorf("request headers = %v, want %v", twitch.Headers, wantHeaders)
		}
	})

	t.Run("an unknown channel", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(helixUser, ok(`{"data":[]}`))

		user, err := client.GetUserByID("123456")

		if !errors.Is(err, ErrChannelNotFound) {
			t.Errorf("GetUserByID error = %v, want it to wrap ErrChannelNotFound", err)
		}
		if want := "channel id 123456 not found: channel not found"; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
		if user != nil {
			t.Errorf("GetUserByID = %+v, want nothing", user)
		}
	})

	t.Run("a blank id", func(t *testing.T) {
		client, _ := setup(t)

		_, err := client.GetUserByID("  ")

		if !errors.Is(err, ErrChannelNotFound) {
			t.Errorf("GetUserByID error = %v, want it to wrap ErrChannelNotFound", err)
		}
		if want := "missing channel id: channel not found"; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})

	t.Run("an account that is not logged in", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.Token = ""
		fake.twitch.serve(helixUser, ok(`{"data":[{"id":"123456","login":"newlogin","display_name":"NewLogin"}]}`))

		_, err := client.GetUserByID("123456")

		if want := `helix user lookup failed: 401 {"error":"Unauthorized"}`; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})

	t.Run("helix down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(helixUser, down("connection refused"))

		_, err := client.GetUserByID("123456")

		if want := `Get "https://api.twitch.tv/helix/users?id=123456": connection refused`; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})

	t.Run("an answer that is not JSON", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.serve(helixUser, ok(`<html>`))

		_, err := client.GetUserByID("123456")

		if want := "invalid character '<' looking for beginning of value"; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})

	t.Run("a client with no session", func(t *testing.T) {
		client := newClient(clientParts{http: &http.Client{Transport: newFakeTwitch("token")}})

		_, err := client.GetUserByID("123456")

		if want := "missing twitch client"; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})

	t.Run("no client", func(t *testing.T) {
		var client *Client

		_, err := client.GetUserByID("123456")

		if want := "missing twitch client"; errorText(err) != want {
			t.Errorf("GetUserByID error = %v, want %s", err, want)
		}
	})
}

func TestClient_GetChannelID(t *testing.T) {
	t.Run("a known login", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("GetIDFromLogin", `{"login":"streamer"}`, ok(`{"data":{"user":{"id":"123456"}}}`))

		id, err := client.GetChannelID("streamer")

		if err != nil {
			t.Fatalf("GetChannelID error = %v, want none", err)
		}
		if id != "123456" {
			t.Errorf("GetChannelID = %q, want %q", id, "123456")
		}
	})

	t.Run("a login in upper case", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("GetIDFromLogin", `{"login":"streamer"}`, ok(`{"data":{"user":{"id":"123456"}}}`))

		id, err := client.GetChannelID("STREAMER")

		if err != nil {
			t.Fatalf("GetChannelID error = %v, want none", err)
		}
		if id != "123456" {
			t.Errorf("GetChannelID = %q, want %q", id, "123456")
		}
	})

	t.Run("an unknown login", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("GetIDFromLogin", `{"login":"nobody"}`, ok(`{"data":{"user":null}}`))

		id, err := client.GetChannelID("Nobody")

		if want := "user Nobody not found"; errorText(err) != want {
			t.Errorf("GetChannelID error = %v, want %s", err, want)
		}
		if id != "" {
			t.Errorf("GetChannelID = %q, want none", id)
		}
	})

	t.Run("a user with an empty id", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("GetIDFromLogin", `{"login":"streamer"}`, ok(`{"data":{"user":{"id":""}}}`))

		_, err := client.GetChannelID("streamer")

		if want := "user streamer not found"; errorText(err) != want {
			t.Errorf("GetChannelID error = %v, want %s", err, want)
		}
	})

	t.Run("twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("GetIDFromLogin", `{"login":"streamer"}`, down("connection refused"))

		_, err := client.GetChannelID("streamer")

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("GetChannelID error = %v, want %s", err, want)
		}
	})
}

func TestClient_GetFollowers(t *testing.T) {
	const firstPage = `{"cursor":"","limit":50,"order":"DESC"}`
	const secondPage = `{"cursor":"c2","limit":50,"order":"DESC"}`

	t.Run("one page of follows", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c1","node":{"login":"Alpha"}},{"cursor":"c2","node":{"login":"beta"}}],"pageInfo":{"hasNextPage":false}}}}}`))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if err != nil {
			t.Fatalf("GetFollowers error = %v, want none", err)
		}
		if want := []string{"alpha", "beta"}; !equal(follows, want) {
			t.Errorf("GetFollowers = %q, want %q", follows, want)
		}
	})

	t.Run("two pages of follows", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c1","node":{"login":"Alpha"}},{"cursor":"c2","node":{"login":"beta"}}],"pageInfo":{"hasNextPage":true}}}}}`))
		fake.twitch.answer("ChannelFollows", secondPage, ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c3","node":{"login":"gamma"}}],"pageInfo":{"hasNextPage":false}}}}}`))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if err != nil {
			t.Fatalf("GetFollowers error = %v, want none", err)
		}
		if want := []string{"alpha", "beta", "gamma"}; !equal(follows, want) {
			t.Errorf("GetFollowers = %q, want %q", follows, want)
		}
	})

	t.Run("an account that follows nobody", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":{"follows":{"edges":[],"pageInfo":{"hasNextPage":false}}}}}`))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if err != nil {
			t.Fatalf("GetFollowers error = %v, want none", err)
		}
		if follows != nil {
			t.Errorf("GetFollowers = %q, want none", follows)
		}
	})

	t.Run("an answer with no follows", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":null}}`))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if err != nil {
			t.Fatalf("GetFollowers error = %v, want none", err)
		}
		if follows != nil {
			t.Errorf("GetFollowers = %q, want none", follows)
		}
	})

	t.Run("twitch down for the second page", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":{"follows":{"edges":[{"cursor":"c2","node":{"login":"Alpha"}}],"pageInfo":{"hasNextPage":true}}}}}`))
		fake.twitch.answer("ChannelFollows", secondPage, down("connection refused"))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if want := `Post "https://gql.twitch.tv/gql": connection refused`; errorText(err) != want {
			t.Errorf("GetFollowers error = %v, want %s", err, want)
		}
		if follows != nil {
			t.Errorf("GetFollowers = %q, want none", follows)
		}
	})

	t.Run("an empty page that claims a next page", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("ChannelFollows", firstPage, ok(`{"data":{"user":{"follows":{"edges":[],"pageInfo":{"hasNextPage":true}}}}}`))

		follows, err := client.GetFollowers(50, domain.FollowersOrderDESC)

		if err != nil {
			t.Fatalf("GetFollowers error = %v, want none", err)
		}
		if follows != nil {
			t.Errorf("GetFollowers = %q, want none", follows)
		}
	})
}
