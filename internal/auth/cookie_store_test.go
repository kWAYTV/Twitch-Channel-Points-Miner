package auth

import (
	"net/http"
	"reflect"
	"testing"
)

func TestDecodeCookieStore(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    cookieStore
		wantErr string
	}{
		{
			"a cookie with every field",
			`{"auth-token": {"value": "tok", "path": "/", "domain": ".twitch.tv"}}`,
			cookieStore{"auth-token": {Value: "tok", Path: "/", Domain: ".twitch.tv"}},
			"",
		},
		{
			"a cookie with only a value",
			`{"auth-token": {"value": "tok"}}`,
			cookieStore{"auth-token": {Value: "tok"}},
			"",
		},
		{
			"two cookies",
			`{"auth-token": {"value": "tok"}, "persistent": {"value": "42"}}`,
			cookieStore{"auth-token": {Value: "tok"}, "persistent": {Value: "42"}},
			"",
		},
		{"no cookies", `{}`, cookieStore{}, ""},
		{"null", `null`, nil, ""},
		{
			"a legacy list with every field",
			`[{"name": "auth-token", "value": "tok", "path": "/", "domain": ".twitch.tv"}]`,
			cookieStore{"auth-token": {Value: "tok", Path: "/", Domain: ".twitch.tv"}},
			"",
		},
		{
			"a legacy list with only names and values",
			`[{"name": "auth-token", "value": "tok"}, {"name": "persistent", "value": "42"}]`,
			cookieStore{"auth-token": {Value: "tok"}, "persistent": {Value: "42"}},
			"",
		},
		{
			"a legacy cookie with no name",
			`[{"value": "lost"}, {"name": "auth-token", "value": "tok"}]`,
			cookieStore{"auth-token": {Value: "tok"}},
			"",
		},
		{
			"a legacy name given twice",
			`[{"name": "auth-token", "value": "old"}, {"name": "auth-token", "value": "new"}]`,
			cookieStore{"auth-token": {Value: "new"}},
			"",
		},
		{"an empty legacy list", `[]`, cookieStore{}, ""},
		{
			"a legacy cookie with a field that is not text",
			`[{"name": "auth-token", "value": "tok", "secure": true}]`,
			nil,
			"json: cannot unmarshal bool into Go value of type string",
		},
		{"text that is not JSON", `<html>`, nil, "invalid character '<' looking for beginning of value"},
		{"a JSON string", `"tok"`, nil, "json: cannot unmarshal string into Go value of type []map[string]string"},
		{"nothing", ``, nil, "unexpected end of JSON input"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCookieStore([]byte(tt.data))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeCookieStore(%s) = %#v, want %#v", tt.data, got, tt.want)
			}
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}

func TestJarCookie(t *testing.T) {
	tests := []struct {
		name       string
		cookieName string
		saved      persistedCookie
		wantHost   string
		want       http.Cookie
	}{
		{
			"only a value",
			"unique_id",
			persistedCookie{Value: "abc"},
			"twitch.tv",
			http.Cookie{Name: "unique_id", Value: "abc", Path: "/", Domain: ".twitch.tv"},
		},
		{
			"a domain with a leading dot",
			"unique_id",
			persistedCookie{Value: "abc", Domain: ".id.twitch.tv"},
			"id.twitch.tv",
			http.Cookie{Name: "unique_id", Value: "abc", Path: "/", Domain: ".id.twitch.tv"},
		},
		{
			"a domain with no leading dot",
			"unique_id",
			persistedCookie{Value: "abc", Domain: "id.twitch.tv"},
			"id.twitch.tv",
			http.Cookie{Name: "unique_id", Value: "abc", Path: "/", Domain: "id.twitch.tv"},
		},
		{
			"a path of its own",
			"unique_id",
			persistedCookie{Value: "abc", Path: "/oauth2"},
			"twitch.tv",
			http.Cookie{Name: "unique_id", Value: "abc", Path: "/oauth2", Domain: ".twitch.tv"},
		},
		{
			"every field",
			"unique_id",
			persistedCookie{Value: "abc", Path: "/oauth2", Domain: ".id.twitch.tv"},
			"id.twitch.tv",
			http.Cookie{Name: "unique_id", Value: "abc", Path: "/oauth2", Domain: ".id.twitch.tv"},
		},
		{
			"nothing",
			"",
			persistedCookie{},
			"twitch.tv",
			http.Cookie{Path: "/", Domain: ".twitch.tv"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, got := jarCookie(tt.cookieName, tt.saved)

			if host != tt.wantHost {
				t.Errorf("host = %q, want %q", host, tt.wantHost)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("cookie = %+v, want %+v", *got, tt.want)
			}
		})
	}
}
