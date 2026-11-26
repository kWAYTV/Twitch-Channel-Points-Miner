package notify

import "testing"

func TestStatedFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status string
		body   string
		field  string
		want   string
	}{
		{"an answer with the field", "403 Forbidden", `{"errcode":"M_FORBIDDEN","error":"User not in room"}`, "error", "User not in room"},
		{"an answer without the field", "403 Forbidden", `{"errcode":"M_FORBIDDEN"}`, "error", "403 Forbidden"},
		{"an answer with the field empty", "403 Forbidden", `{"error":""}`, "error", "403 Forbidden"},
		{"an answer with a number in the field", "403 Forbidden", `{"error":403}`, "error", "403 Forbidden"},
		{"an answer with the reason in another field", "403 Forbidden", `{"description":"User not in room"}`, "error", "403 Forbidden"},
		{"an answer that is a list", "403 Forbidden", `["error"]`, "error", "403 Forbidden"},
		{"an answer that is not JSON", "502 Bad Gateway", "<html>Bad Gateway</html>", "error", "502 Bad Gateway"},
		{"an empty answer", "500 Internal Server Error", "", "error", "500 Internal Server Error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := statedFailure(tc.status, []byte(tc.body), tc.field); got != tc.want {
				t.Errorf("statedFailure(%q, %q, %q) = %q, want %q", tc.status, tc.body, tc.field, got, tc.want)
			}
		})
	}
}
