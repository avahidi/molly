package logging

import (
	"os"
	"path/filepath"

	"bitbucket.org/vahidi/molly/lib/util"
)

var logdir = "logs"

func name(suggestedname string, register bool) string {
	filename := util.SanitizeFilename(suggestedname, nil)
	return filepath.Join(logdir, filename)
}

// SetBase sets the base directory for all logs
func SetBase(path string) {
	logdir = path
}

// Create creates a new log or repoert file
func Create(suggestedname string) (*os.File, error) {
	filename := name(suggestedname, false)

	// make sure the path leading to it exist
	dir, _ := filepath.Split(filename)
	os.MkdirAll(dir, 0700)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)
	return file, err
}
