package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

// debugLevel is how much the debug setting logs. In config.json it is false,
// true or "full".
type debugLevel string

const (
	// debugOff logs no debug lines.
	debugOff debugLevel = ""
	// debugSummary logs a summary of each request.
	debugSummary debugLevel = "summary"
	// debugFull also logs full payloads.
	debugFull debugLevel = "full"
)

var errDebugLevel = errors.New(`debug must be false, true or "full"`)

func (d *debugLevel) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	switch string(data) {
	case "false", "null":
		*d = debugOff
		return nil
	case "true":
		*d = debugSummary
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return errDebugLevel
	}
	if strings.ToLower(strings.TrimSpace(text)) != string(debugFull) {
		return errDebugLevel
	}
	*d = debugFull
	return nil
}
