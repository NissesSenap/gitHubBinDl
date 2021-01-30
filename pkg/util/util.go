package util

import "os"

const DateFormat = "2006-01-02"

// MakeDirectoryIfNotExists create a folder if it's missing
func MakeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}
