package miner

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
)

func displayName(name string) string {
	if name == "" {
		return ""
	}
	first, size := utf8.DecodeRuneInString(name)
	return string(unicode.ToUpper(first)) + name[size:]
}

func (m *Miner) anonymizeLogs() bool {
	return m != nil && m.anonymizer != nil && m.anonymizer.Enabled()
}

func (m *Miner) rawStreamerName(raw string) string {
	if m.anonymizeLogs() {
		return m.anonymizer.Name(raw)
	}
	return raw
}

func (m *Miner) styledStreamerName(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	name := displayName(streamer.Username)
	if m.anonymizeLogs() {
		name = m.anonymizer.StreamerName(streamer)
	}
	if streamer.HasActiveMultipliers() {
		return fmt.Sprintf("%s%s%s", colorDropsAccent, name, colorReset)
	}
	return name
}

func (m *Miner) formattedStreamerPoints(streamer *domain.Streamer) string {
	if streamer == nil {
		return ""
	}
	points := streamer.ChannelPoints
	if m.anonymizeLogs() {
		points = m.anonymizer.PseudoChannelPoints(streamer)
	}
	return formatChannelPoints(points)
}
