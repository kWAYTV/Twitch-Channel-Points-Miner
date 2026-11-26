// Package jsonfile writes values to disk as indented JSON.
package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveJSON writes data to path as indented JSON. The file is written beside
// path and moved into place, so a crash part-way leaves the old file whole
// and never a truncated one.
func SaveJSON(path string, data interface{}) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	_, err = temp.Write(raw)
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Chmod(temp.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(temp.Name(), path)
	}
	if err != nil {
		os.Remove(temp.Name())
	}
	return err
}
