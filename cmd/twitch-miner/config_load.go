package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// loadOrCreateConfig reads the config, converts the settings of an older
// version, adds any key a newer version introduced, and writes the file back
// when it was missing, incomplete, converted or not laid out the way the
// program writes it.
func loadOrCreateConfig(path string) (config, error) {
	settings, onDisk, err := readSettings(path)
	if err != nil {
		return config{}, err
	}
	// A converted file no longer matches what is on disk, so it is written back
	// below the way a file in another layout is.
	migrateSettings(settings)
	added := fillMissing(settings, defaultConfig())
	if onDisk == nil || added {
		if err := writeSettings(path, settings); err != nil {
			return config{}, err
		}
	} else {
		tidySettings(path, settings, onDisk)
	}
	return decodeConfig(settings)
}

// tidySettings rewrites a complete config that differs from the file on disk
// only in its layout or in settings converted from an older version.
// Nothing depends on it, so a file that cannot be written is left as it is.
func tidySettings(path string, settings map[string]interface{}, onDisk []byte) {
	formatted, err := formatSettings(settings)
	if err != nil || bytes.Equal(formatted, bytes.ReplaceAll(onDisk, []byte("\r\n"), []byte("\n"))) {
		return
	}
	_ = writeSettings(path, settings)
}

// readSettings returns the file as a map, with its bytes as they are on disk.
// A file that cannot be read counts as missing (nil bytes); one that is not
// valid JSON is an error.
func readSettings(path string) (settings map[string]interface{}, onDisk []byte, err error) {
	settings = map[string]interface{}{}
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return settings, nil, nil
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}
	return settings, raw, nil
}
