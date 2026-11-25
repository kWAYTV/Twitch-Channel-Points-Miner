// Package project says what this program is and where it lives.
package project

const (
	// Name is how the miner introduces itself in the console and in notifications.
	Name = "Twitch Channel Points Miner"
	// Repository is the GitHub repository releases and assets come from.
	Repository = "kWAYTV/twitch-points-miner"
	// URL is the project's home page.
	URL = "https://github.com/" + Repository
	// AvatarURL is the picture notifications are sent with.
	AvatarURL = "https://raw.githubusercontent.com/" + Repository + "/main/docs/assets/gopher.png"
	// DevVersion is what a build reports when no release version was stamped in.
	DevVersion = "dev"
)

// Version is the running release, compared against the latest one published.
// The release build sets it; a plain "go build" leaves it at DevVersion.
var Version = DevVersion

// VersionLabel is the version as shown to a person.
func VersionLabel() string {
	if Version == DevVersion {
		return "built from source"
	}
	return "v" + Version
}
