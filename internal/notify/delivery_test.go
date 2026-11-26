package notify

import (
	"reflect"
	"testing"
)

func TestDeliveryWarning_warn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		service string
		reasons []string
		want    []string
	}{
		{
			"discord unreachable",
			"discord",
			[]string{requestFailed},
			[]string{"discord: notification not delivered: request failed (network error or timeout) (further failures are not logged)"},
		},
		{
			"telegram refusing the chat",
			"telegram",
			[]string{"Bad Request: chat not found"},
			[]string{"telegram: notification not delivered: Bad Request: chat not found (further failures are not logged)"},
		},
		{
			"a second failure for another reason",
			"telegram",
			[]string{"Unauthorized", requestFailed},
			[]string{"telegram: notification not delivered: Unauthorized (further failures are not logged)"},
		},
		{"no failure", "discord", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := &fakeLog{}
			warning := &deliveryWarning{service: tc.service, logf: log.Printf}

			for _, reason := range tc.reasons {
				warning.warn(reason)
			}

			if !reflect.DeepEqual(log.Lines, tc.want) {
				t.Errorf("logged %q, want %q", log.Lines, tc.want)
			}
		})
	}
}
