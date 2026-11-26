package twitchapi

// The miner signs in with Twitch's Android TV client id, so it presents
// itself as the browser of an Android TV box everywhere except the minute
// watched report.
const (
	AndroidTVUserAgent = "Mozilla/5.0 (Linux; Android 7.1; Smart Box C1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
	// LinuxFirefoxUserAgent is sent with the minute watched report.
	LinuxFirefoxUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:85.0) Gecko/20100101 Firefox/85.0"
)
