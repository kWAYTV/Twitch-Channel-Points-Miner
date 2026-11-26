package twitch

// fakeSession stands in for a logged-in account.
type fakeSession struct {
	Token string
	ID    string
	// LoginErr is returned by Login when set.
	LoginErr error
	// LoginPaths holds the cookie file of every Login call.
	LoginPaths []string
}

func (s *fakeSession) Login(cookiesPath string) error {
	s.LoginPaths = append(s.LoginPaths, cookiesPath)
	return s.LoginErr
}

func (s *fakeSession) AuthToken() string { return s.Token }
func (s *fakeSession) UserID() string    { return s.ID }
