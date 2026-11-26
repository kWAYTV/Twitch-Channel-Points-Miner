package twitch

import "testing"

func TestClient_PostGQL_loginExpired(t *testing.T) {
	const expiredLine = "error LOGIN_EXPIRED: Twitch login expired: restart the miner and enter the new code at twitch.tv/activate"

	t.Run("Twitch refusing the login", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.Token = ""
		fake.twitch.answer("JoinRaid", "", ok(`{}`))

		client.PostGQL(raid)

		if want := []string{expiredLine}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("Twitch refusing the login a second time", func(t *testing.T) {
		client, fake := setup(t)
		fake.session.Token = ""
		fake.twitch.answer("JoinRaid", "", ok(`{}`))
		client.PostGQL(raid)

		client.PostGQL(raid)

		if want := []string{expiredLine}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want it once: %q", fake.logger.Lines, want)
		}
	})

	t.Run("the login refused again after a request that worked", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`{}`))
		fake.session.Token = ""
		client.PostGQL(raid)
		fake.session.Token = "token"
		client.PostGQL(raid)
		fake.session.Token = ""

		client.PostGQL(raid)

		if want := []string{expiredLine, expiredLine}; !equal(fake.logger.Lines, want) {
			t.Errorf("logged %q, want %q", fake.logger.Lines, want)
		}
	})

	t.Run("Twitch down", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", code(503, `{"error":"Service Unavailable"}`))

		client.PostGQL(raid)

		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing: an outage is not an expired login", fake.logger.Lines)
		}
	})

	t.Run("a request that works", func(t *testing.T) {
		client, fake := setup(t)
		fake.twitch.answer("JoinRaid", "", ok(`{}`))

		client.PostGQL(raid)

		if len(fake.logger.Lines) != 0 {
			t.Errorf("logged %q, want nothing", fake.logger.Lines)
		}
	})
}
