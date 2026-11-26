// Package jsonfile writes files so that a crash never leaves one half
// written, and values as indented JSON.
package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveJSON writes data to path as indented JSON, readable by everyone.
func SaveJSON(path string, data interface{}) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return Write(path, raw, 0o644)
}

// Write puts content at path with the given permissions, creating the folder
// when it is missing. The file is written beside path and moved into place,
// so a crash part-way leaves the old file whole and never a truncated one.
func Write(path string, content []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	_, err = temp.Write(content)
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Chmod(temp.Name(), mode)
	}
	if err == nil {
		err = os.Rename(temp.Name(), path)
	}
	if err != nil {
		os.Remove(temp.Name())
	}
	return err
}
